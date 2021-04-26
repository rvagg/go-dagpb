package dagpb

import (
	"fmt"
	"reflect"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/schema"
)

// WrapNoSchema implements an ipld.Node given a pointer to a Go value.
//
// Same rules as PrototypeNoSchema apply.
func WrapNoSchema(ptr interface{}) ipld.Node {
	panic("TODO")
	// ptrVal := reflect.ValueOf(ptr)
	// if ptrVal.Kind() != reflect.Ptr {
	// 	panic("must be a pointer")
	// }
	// return &_node{val: ptrVal.Elem()}
}

// Unwrap takes an ipld.Node implemented by one of the Wrap* APIs, and returns
// a pointer to the inner Go value.
//
// Unwrap returns nil if the node isn't implemented by this package.
func Unwrap(node ipld.Node) (ptr interface{}) {
	if w, ok := node.(*_node); ok {
		if w.val.Kind() == reflect.Ptr {
			panic("didn't expect w.val to be a pointer")
		}
		return w.val.Addr().Interface()
	}
	return nil
}

// PrototypeNoSchema implements an ipld.NodePrototype given a Go pointer type.
//
// In this form, no IPLD schema is used; it is entirely inferred from the Go
// type.
//
// Go types map to schema types in simple ways: Go string to schema String, Go
// []byte to schema Bytes, Go struct to schema Map, and so on.
//
// A Go struct field is optional when its type is a pointer. Nullable fields are
// not supported in this mode.
func PrototypeNoSchema(ptrType interface{}) ipld.NodePrototype {
	panic("TODO")
	// typ := reflect.TypeOf(ptrType)
	// if typ.Kind() != reflect.Ptr {
	// 	panic("must be a pointer")
	// }
	// return &_prototype{goType: typ.Elem()}
}

// PrototypeOnlySchema implements an ipld.NodePrototype given an IPLD schema type.
//
// In this form, Go values are constructed with types inferred from the IPLD
// schema, like a reverse of PrototypeNoSchema.
func PrototypeOnlySchema(schemaType schema.Type) ipld.NodePrototype {
	panic("TODO")
}

// Prototype implements an ipld.NodePrototype given a Go pointer type and an
// IPLD schema type.
//
// In this form, it is assumed that the Go type and IPLD schema type are
// compatible. TODO: check upfront and panic otherwise
func Prototype(ptrType interface{}, schemaType schema.Type) ipld.NodePrototype {
	typ := reflect.TypeOf(ptrType)
	if typ.Kind() != reflect.Ptr {
		panic("ptrType must be a pointer")
	}
	if schemaType == nil {
		panic("schemaType must not be nil")
	}
	return &_prototype{schemaType: schemaType, goType: typ.Elem()}
}

type _prototype struct {
	schemaType schema.Type
	goType     reflect.Type // non-pointer
}

func (w *_prototype) NewBuilder() ipld.NodeBuilder {
	return &_builder{_assembler{
		schemaType: w.schemaType,
		val:        reflect.New(w.goType).Elem(),
	}}
}

var (
	goTypeString = reflect.TypeOf("")
	goTypeBytes  = reflect.TypeOf([]byte{})
	goTypeLink   = reflect.TypeOf((*ipld.Link)(nil)).Elem()

	schemaTypeString schema.Type
)

type _node struct {
	schemaType schema.Type

	optional, nullable bool

	val reflect.Value // non-pointer
}

func (w *_node) Kind() ipld.Kind {
	// TODO: avoid this check usually?
	if w.IsAbsent() {
		return ipld.Kind_Null
	}
	return w.schemaType.TypeKind().ActsLike()
}

func (w *_node) LookupByString(key string) (ipld.Node, error) {
	switch typ := w.schemaType.(type) {
	case *schema.TypeStruct:
		field := typ.Field(key)
		if field == nil {
			return nil, &ipld.ErrInvalidKey{
				TypeName: typ.Name().String(),
				Key:      basicnode.NewString(key),
			}
		}
		fval := w.val.FieldByName(key)
		if !fval.IsValid() {
			panic("TODO: go-schema mismatch")
		}
		node := &_node{
			schemaType: field.Type(),
			optional:   field.IsOptional(),
			nullable:   field.IsNullable(),
			val:        fval,
		}
		if node.optional && node.nullable {
			panic("TODO: optional and nullable")
		}
		return node, nil
	case *schema.TypeMap:
	}
	return nil, &ipld.ErrWrongKind{
		TypeName:   w.schemaType.Name().String(),
		MethodName: "LookupByString",
		// TODO
	}
}

func (w *_node) LookupByNode(key ipld.Node) (ipld.Node, error) {
	panic("TODO: LookupByNode")
}

func (w *_node) LookupByIndex(idx int64) (ipld.Node, error) {
	switch typ := w.schemaType.(type) {
	case *schema.TypeList:
		// TODO: bounds check
		val := w.val.Index(int(idx))
		return &_node{schemaType: typ.ValueType(), val: val}, nil
	}
	return nil, &ipld.ErrWrongKind{
		TypeName:   w.schemaType.Name().String(),
		MethodName: "LookupByIndex",
		// TODO
	}
}

func (w *_node) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	panic("TODO: LookupBySegment")
}

func (w *_node) MapIterator() ipld.MapIterator {
	switch typ := w.schemaType.(type) {
	case *schema.TypeStruct:
		return &_structIterator{
			fields: typ.Fields(),
			val:    w.val,
		}
	case *schema.TypeMap:
		panic("TODO: ")
	}
	return nil
}

func (w *_node) ListIterator() ipld.ListIterator {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if !val.IsNil() {
			val = val.Elem()
		}
	}
	switch typ := w.schemaType.(type) {
	case *schema.TypeList:
		return &_listIterator{schemaType: typ, val: val}
	}
	return nil
}

func (w *_node) Length() int64 {
	switch w.Kind() {
	case ipld.Kind_List, ipld.Kind_Map:
		return int64(w.val.Len())
	}
	return -1
}

// TODO: better story around pointers and absent/null

func (w *_node) IsAbsent() bool {
	return w.optional && w.val.IsNil()
}

func (w *_node) IsNull() bool {
	return w.nullable && w.val.IsNil()
}

func (w *_node) AsBool() (bool, error) {
	panic("TODO: ")
}

func (w *_node) AsInt() (int64, error) {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if val.IsNil() {
			panic("Kind == Int but nil ptr?")
		}
		val = val.Elem()
	}
	if typ.Kind() != reflect.Int {
		return 0, &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AsInt",
			// TODO
		}
	}
	return val.Int(), nil
}

func (w *_node) AsFloat() (float64, error) {
	panic("TODO: ")
}

func (w *_node) AsString() (string, error) {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if !val.IsNil() {
			val = val.Elem()
		}
	}
	if typ.Kind() != reflect.String {
		return "", &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AsString",
			// TODO
		}
	}
	return val.String(), nil
}

func (w *_node) AsBytes() ([]byte, error) {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if !val.IsNil() {
			val = val.Elem()
		}
	}
	if typ.Kind() != reflect.Slice || typ.Elem().Kind() != reflect.Uint8 {
		return nil, &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AsBytes",
			// TODO
		}
	}
	return val.Bytes(), nil
}

func (w *_node) AsLink() (ipld.Link, error) {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if !val.IsNil() {
			val = val.Elem()
		}
	}
	if typ != goTypeLink {
		return nil, &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AsLink",
			// TODO
		}
	}
	link, _ := w.val.Interface().(ipld.Link)
	return link, nil
}

func (w *_node) Prototype() ipld.NodePrototype {
	panic("TODO: ")
}

type _builder struct {
	_assembler
}

func (w *_builder) Build() ipld.Node {
	return &_node{schemaType: w.schemaType, val: w.val}
}

func (w *_builder) Reset() {
	panic("TODO")
}

type _assembler struct {
	schemaType schema.Type
	val        reflect.Value // non-pointer
}

func (w *_assembler) BeginMap(sizeHint int64) (ipld.MapAssembler, error) {
	switch typ := w.schemaType.(type) {
	case *schema.TypeStruct:
		doneFields := make([]bool, w.val.NumField())
		return &_structAssembler{
			schemaType: typ,
			val:        w.val,
			doneFields: doneFields,
		}, nil
	case *schema.TypeMap:
	}
	return nil, &ipld.ErrWrongKind{
		TypeName:   w.schemaType.Name().String(),
		MethodName: "BeginMap",
		// TODO
	}
}

func (w *_assembler) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Slice &&
		typ.Elem().Elem().Kind() == reflect.Uint8 {
		val.Set(reflect.New(typ.Elem()))
		val = val.Elem()
	}
	switch typ := w.schemaType.(type) {
	case *schema.TypeList:
		return &_listAssembler{schemaType: typ, val: val}, nil
	}
	return nil, &ipld.ErrWrongKind{
		TypeName:   w.schemaType.Name().String(),
		MethodName: "BeginList",
		// TODO
	}
}

func (w *_assembler) AssignNull() error {
	return nil // TODO?
}

func (w *_assembler) AssignBool(bool) error {
	panic("TODO")
}

func (w *_assembler) AssignInt(i int64) error {
	typ := w.val.Type()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Int {
		w.val.Set(reflect.New(typ.Elem()))
		w.val.Elem().SetInt(i)
		return nil
	}
	if typ.Kind() != reflect.Int {
		return &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AssignInt",
			// TODO
		}
	}
	w.val.SetInt(i)
	return nil
}

func (w *_assembler) AssignFloat(float64) error {
	panic("TODO")
}

func (w *_assembler) AssignString(s string) error {
	typ := w.val.Type()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.String {
		w.val.Set(reflect.New(typ.Elem()))
		w.val.Elem().SetString(s)
		return nil
	}
	if typ.Kind() != reflect.String {
		return &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AssignString",
			// TODO
		}
	}
	w.val.SetString(s)
	return nil
}

func (w *_assembler) AssignBytes(p []byte) error {
	typ := w.val.Type()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Slice &&
		typ.Elem().Elem().Kind() == reflect.Uint8 {
		w.val.Set(reflect.New(typ.Elem()))
		w.val.Elem().SetBytes(p)
		return nil
	}
	if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Uint8 {
		return &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AssignBytes",
			// TODO
		}
	}
	w.val.SetBytes(p)
	return nil
}

func (w *_assembler) AssignLink(link ipld.Link) error {
	newVal := reflect.ValueOf(link)
	if !newVal.Type().AssignableTo(w.val.Type()) {
		return &ipld.ErrWrongKind{
			TypeName:   w.schemaType.Name().String(),
			MethodName: "AssignLink",
			// TODO
		}
	}
	w.val.Set(newVal)
	return nil
}

func (w *_assembler) AssignNode(node ipld.Node) error {
	// TODO: does this ever trigger?
	// newVal := reflect.ValueOf(node)
	// if newVal.Type().AssignableTo(w.val.Type()) {
	// 	w.val.Set(newVal)
	// 	return nil
	// }
	switch node.Kind() {
	case ipld.Kind_Map:
		itr := node.MapIterator()
		am, err := w.BeginMap(-1) // TODO: length?
		if err != nil {
			return err
		}
		for !itr.Done() {
			k, v, err := itr.Next()
			if err != nil {
				return err
			}
			if err := am.AssembleKey().AssignNode(k); err != nil {
				return err
			}
			if err := am.AssembleValue().AssignNode(v); err != nil {
				return err
			}
		}
		return am.Finish()
	case ipld.Kind_List:
		itr := node.ListIterator()
		am, err := w.BeginList(-1) // TODO: length?
		if err != nil {
			return err
		}
		for !itr.Done() {
			_, v, err := itr.Next()
			if err != nil {
				return err
			}
			if err := am.AssembleValue().AssignNode(v); err != nil {
				return err
			}
		}
		return am.Finish()

	case ipld.Kind_Int:
		i, err := node.AsInt()
		if err != nil {
			return err
		}
		return w.AssignInt(i)
	case ipld.Kind_String:
		s, err := node.AsString()
		if err != nil {
			return err
		}
		return w.AssignString(s)
	case ipld.Kind_Bytes:
		p, err := node.AsBytes()
		if err != nil {
			return err
		}
		return w.AssignBytes(p)
	case ipld.Kind_Link:
		l, err := node.AsLink()
		if err != nil {
			return err
		}
		return w.AssignLink(l)
	case ipld.Kind_Null:
		return w.AssignNull()
	}
	// fmt.Println(w.val.Type(), reflect.TypeOf(node))
	panic(fmt.Sprintf("TODO: %v %v", w.val.Type(), node.Kind()))
}

func (w *_assembler) Prototype() ipld.NodePrototype {
	panic("TODO")
}

type _structAssembler struct {
	// TODO: embed _assembler?

	schemaType *schema.TypeStruct
	val        reflect.Value // non-pointer

	// TODO: more state checks
	doneFields []bool

	// TODO: optimize for structs

	curKey _assembler
}

func (w *_structAssembler) AssembleKey() ipld.NodeAssembler {
	w.curKey = _assembler{
		schemaType: schemaTypeString,
		val:        reflect.New(goTypeString).Elem(),
	}
	return &w.curKey
}

func (w *_structAssembler) AssembleValue() ipld.NodeAssembler {
	// TODO: optimize this to do one lookup by name
	name := w.curKey.val.String()
	field := w.schemaType.Field(name)
	if field == nil {
		panic("TODO: missing fields")
	}
	ftyp, ok := w.val.Type().FieldByName(name)
	if !ok {
		panic("TODO: go-schema mismatch")
	}
	if len(ftyp.Index) > 1 {
		panic("TODO: embedded fields")
	}
	w.doneFields[ftyp.Index[0]] = true
	fval := w.val.FieldByIndex(ftyp.Index)
	// TODO: reuse same assembler for perf?
	return &_assembler{schemaType: field.Type(), val: fval}
}

func (w *_structAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	panic("TODO")
}

func (w *_structAssembler) Finish() error {
	fields := w.schemaType.Fields()
	for i, field := range fields {
		if !field.IsOptional() && !w.doneFields[i] {
			// TODO: return a complete list
			return &ipld.ErrMissingRequiredField{Missing: []string{field.Name()}}
		}
	}
	return nil
}

func (w *_structAssembler) KeyPrototype() ipld.NodePrototype {
	return &_prototype{schemaType: schemaTypeString, goType: goTypeString}
}

func (w *_structAssembler) ValuePrototype(k string) ipld.NodePrototype {
	panic("TODO")
}

type _listAssembler struct {
	schemaType *schema.TypeList

	val reflect.Value // non-pointer
}

func (w *_listAssembler) AssembleValue() ipld.NodeAssembler {
	goType := w.val.Type().Elem()
	w.val.Set(reflect.Append(w.val, reflect.New(goType).Elem()))
	return &_assembler{schemaType: w.schemaType.ValueType(), val: w.val.Index(w.val.Len() - 1)}
}

func (w *_listAssembler) Finish() error {
	return nil
}

func (w *_listAssembler) ValuePrototype(idx int64) ipld.NodePrototype {
	panic("TODO")
}

type _structIterator struct {
	// TODO: support embedded fields?
	fields    []schema.StructField
	val       reflect.Value // non-pointer
	nextIndex int
}

func (w *_structIterator) Next() (key, value ipld.Node, _ error) {
	field := w.fields[w.nextIndex]
	val := w.val.Field(w.nextIndex)
	w.nextIndex++
	node := &_node{
		schemaType: field.Type(),
		optional:   field.IsOptional(),
		nullable:   field.IsNullable(),
		val:        val,
	}
	if node.optional && node.nullable {
		panic("TODO: optional and nullable")
	}
	return basicnode.NewString(field.Name()), node, nil
}

func (w *_structIterator) Done() bool {
	return w.nextIndex >= len(w.fields)
}

type _listIterator struct {
	schemaType *schema.TypeList
	val        reflect.Value // non-pointer
	nextIndex  int
}

func (w *_listIterator) Next() (index int64, value ipld.Node, _ error) {
	idx := int64(w.nextIndex)
	val := w.val.Index(w.nextIndex)
	w.nextIndex++
	return idx, &_node{schemaType: w.schemaType.ValueType(), val: val}, nil
}

func (w *_listIterator) Done() bool {
	return w.nextIndex >= w.val.Len()
}

// TODO: consider making our own Node interface, like:
//
// type WrappedNode interface {
//     ipld.Node
//     Unwrap() (ptr interface)
// }
//
// Pros: API is easier to understand, harder to mix up with other ipld.Nodes.
// Cons: One usually only has an ipld.Node, and type assertions can be weird.
