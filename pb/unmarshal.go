package pb

import (
	"bytes"
	"fmt"
	"io"

	cid "github.com/ipfs/go-cid"
	"github.com/polydawn/refmt/shared"
)

type PBLinkBuilder interface {
	SetHash(*cid.Cid) error
	SetName(*string) error
	SetTsize(uint64) error
	Done() error
}

type PBNodeBuilder interface {
	SetData([]byte) error
	AddLink() (PBLinkBuilder, error)
	Done() error
}

func UnmarshalPBNode(byts []byte) (*PBNode, error) {
	node := NewPBNode()
	tokenReceiver := func(tok Token) error {
		switch tok.Type {
		case TypeData:
			node.Data = tok.Bytes
		case TypeLink:
			node.Links = append(node.Links, &PBLink{})
		case TypeHash:
			node.Links[len(node.Links)-1].Hash = tok.Cid
		case TypeName:
			s := string(tok.Bytes)
			node.Links[len(node.Links)-1].Name = &s
		case TypeTSize:
			node.Links[len(node.Links)-1].Tsize = &tok.Int
		}
		return nil
	}
	if err := Unmarshal(bytes.NewReader(byts), tokenReceiver); err != nil {
		return nil, err
	}

	return node, nil
}

func Unmarshal(in io.Reader, tokenReceiver func(tok Token) error) error {
	haveData := false
	reader := shared.NewReader(in)
	for {
		_, err := reader.Readn1()
		if err == io.EOF {
			break
		}
		reader.Unreadn1()

		fieldNum, wireType, err := decodeKey(reader)
		if err != nil {
			return err
		}
		if wireType != 2 {
			return fmt.Errorf("protobuf: (PBNode) invalid wireType, expected 2, got %d", wireType)
		}

		if fieldNum == 1 {
			if haveData {
				return fmt.Errorf("protobuf: (PBNode) duplicate Data section")
			}
			var chunk []byte
			if chunk, err = decodeBytes(reader); err != nil {
				return err
			}
			if err := tokenReceiver(Token{Type: TypeData, Bytes: chunk}); err != nil {
				return err
			}
			haveData = true
		} else if fieldNum == 2 {
			if haveData {
				return fmt.Errorf("protobuf: (PBNode) invalid order, found Data before Links content")
			}

			bytesLen, err := decodeVarint(reader)
			if err != nil {
				return err
			}
			if err = tokenReceiver(Token{Type: TypeLink}); err != nil {
				return err
			}
			if err = unmarshalLink(reader, int(bytesLen), tokenReceiver); err != nil {
				return err
			}
			if err = tokenReceiver(Token{Type: TypeLinkEnd}); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("protobuf: (PBNode) invalid fieldNumber, expected 1 or 2, got %d", fieldNum)
		}
	}

	return nil
}

func unmarshalLink(reader shared.SlickReader, length int, tokenReceiver func(tok Token) error) error {
	haveHash := false
	haveName := false
	haveTsize := false
	startOffset := reader.NumRead()
	for {
		readBytes := reader.NumRead() - startOffset
		if readBytes == length {
			break
		} else if readBytes > length {
			return fmt.Errorf("protobuf: (PBLink) bad length for link")
		}
		fieldNum, wireType, err := decodeKey(reader)
		if err != nil {
			return err
		}

		if fieldNum == 1 {
			if haveHash {
				return fmt.Errorf("protobuf: (PBLink) duplicate Hash section")
			}
			if haveName {
				return fmt.Errorf("protobuf: (PBLink) invalid order, found Name before Hash")
			}
			if haveTsize {
				return fmt.Errorf("protobuf: (PBLink) invalid order, found Tsize before Hash")
			}
			if wireType != 2 {
				return fmt.Errorf("protobuf: (PBLink) wrong wireType (%d) for Hash", wireType)
			}

			var chunk []byte
			if chunk, err = decodeBytes(reader); err != nil {
				return err
			}
			var c cid.Cid
			if _, c, err = cid.CidFromBytes(chunk); err != nil {
				return fmt.Errorf("invalid Hash field found in link, expected CID (%v)", err)
			}
			if err := tokenReceiver(Token{Type: TypeHash, Cid: &c}); err != nil {
				return err
			}
			haveHash = true
		} else if fieldNum == 2 {
			if haveName {
				return fmt.Errorf("protobuf: (PBLink) duplicate Name section")
			}
			if haveTsize {
				return fmt.Errorf("protobuf: (PBLink) invalid order, found Tsize before Name")
			}
			if wireType != 2 {
				return fmt.Errorf("protobuf: (PBLink) wrong wireType (%d) for Name", wireType)
			}

			var chunk []byte
			if chunk, err = decodeBytes(reader); err != nil {
				return err
			}
			if err := tokenReceiver(Token{Type: TypeName, Bytes: chunk}); err != nil {
				return err
			}
			haveName = true
		} else if fieldNum == 3 {
			if haveTsize {
				return fmt.Errorf("protobuf: (PBLink) duplicate Tsize section")
			}
			if wireType != 0 {
				return fmt.Errorf("protobuf: (PBLink) wrong wireType (%d) for Tsize", wireType)
			}

			var v uint64
			if v, err = decodeVarint(reader); err != nil {
				return err
			}
			if err := tokenReceiver(Token{Type: TypeTSize, Int: v}); err != nil {
				return err
			}
			haveTsize = true
		} else {
			return fmt.Errorf("protobuf: (PBLink) invalid fieldNumber, expected 1, 2 or 3, got %d", fieldNum)
		}
	}

	if !haveHash {
		return fmt.Errorf("invalid Hash field found in link, expected CID")
	}

	return nil
}

func decodeKey(reader shared.SlickReader) (int, int, error) {
	var wire uint64
	var err error
	if wire, err = decodeVarint(reader); err != nil {
		return 0, 0, err
	}
	fieldNum := int(wire >> 3)
	wireType := int(wire & 0x7)
	return fieldNum, wireType, nil
}

func decodeBytes(reader shared.SlickReader) ([]byte, error) {
	bytesLen, err := decodeVarint(reader)
	if err != nil {
		return nil, err
	}
	byts, err := reader.Readn(int(bytesLen))
	if err != nil {
		return nil, fmt.Errorf("protobuf: unexpected read error: %w", err)
	}
	return byts, nil
}

func decodeVarint(reader shared.SlickReader) (uint64, error) {
	var v uint64
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			return 0, ErrIntOverflow
		}
		b, err := reader.Readn1()
		if err != nil {
			return 0, fmt.Errorf("protobuf: unexpected read error: %w", err)
		}
		v |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	return v, nil
}

// ErrIntOverflow TODO
var ErrIntOverflow = fmt.Errorf("protobuf: varint overflow")
