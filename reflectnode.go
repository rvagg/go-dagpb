package dagpb

import (
	"fmt"
	"reflect"

	ipld "github.com/ipld/go-ipld-prime"
)

func Wrap(ptr interface{}) ipld.Node {
	ptrVal := reflect.ValueOf(ptr)
	if ptrVal.Kind() != reflect.Ptr {
		panic("must be a pointer")
	}
	return wrap(ptrVal.Elem())
}

func wrap(val reflect.Value) ipld.Node {
	if !val.CanAddr() {
		panic("must be addressable")
	}
	return _node{val}
}

type _node struct {
	val reflect.Value
}

func Unwrap(node ipld.Node) interface{} {
	if w, ok := node.(_node); ok {
		return w.val.Interface()
	}
	return nil
}

func (w _node) under() reflect.Value {
	val := w.val
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

func (w _node) Kind() ipld.Kind {
	switch w.val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return ipld.Kind_Int

	case reflect.Float32, reflect.Float64:
		return ipld.Kind_Float

	case reflect.Struct, reflect.Map:
		return ipld.Kind_Map

	case reflect.Slice, reflect.Array:
		return ipld.Kind_List
	}
	return ipld.Kind_Invalid
}
func (w _node) LookupByString(key string) (ipld.Node, error) {
	switch w.val.Kind() {
	case reflect.Struct:
	case reflect.Map:
	}
	return nil, &ipld.ErrWrongKind{
		// TODO
	}

}
func (w _node) LookupByNode(key ipld.Node) (ipld.Node, error) {
	panic("TODO: LookupByNode")
}
func (w _node) LookupByIndex(idx int64) (ipld.Node, error) {
	switch w.val.Kind() {
	case reflect.Slice, reflect.Array:
		// TODO: bounds check
		val := w.val.Index(int(idx))
		return _node{val}, nil
	}
	return nil, &ipld.ErrWrongKind{
		// TODO
	}
}
func (w _node) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	panic("TODO: LookupBySegment")
}
func (w _node) MapIterator() ipld.MapIterator {
	panic("TODO: ")
}
func (w _node) ListIterator() ipld.ListIterator {
	panic("TODO: ")
}
func (w _node) Length() int64 {
	panic("TODO: ")
}
func (w _node) IsAbsent() bool {
	panic("TODO: ")
}
func (w _node) IsNull() bool {
	panic("TODO: ")
}
func (w _node) AsBool() (bool, error) {
	panic("TODO: ")
}
func (w _node) AsInt() (int64, error) {
	panic("TODO: ")
}
func (w _node) AsFloat() (float64, error) {
	panic("TODO: ")
}
func (w _node) AsString() (string, error) {
	panic("TODO: ")
}
func (w _node) AsBytes() ([]byte, error) {
	panic("TODO: ")
}
func (w _node) AsLink() (ipld.Link, error) {
	panic("TODO: ")
}
func (w _node) Prototype() ipld.NodePrototype {
	panic("TODO: ")
}

type _prototype struct {
	typ reflect.Type
}

func (w _prototype) NewBuilder() ipld.NodeBuilder {
	return &_builder{_assembler{reflect.New(w.typ).Elem()}}
}

type _builder struct {
	_assembler
}

func (w *_builder) Build() ipld.Node {
	return _node{w.val}
}
func (w *_builder) Reset() {
	panic("TODO")
}

type _assembler struct {
	val reflect.Value
}

func (w *_assembler) BeginMap(sizeHint int64) (ipld.MapAssembler, error) {
	switch w.val.Kind() {
	case reflect.Struct:
		return &_mapAssembler{val: w.val}, nil
	case reflect.Map:
	}
	return nil, &ipld.ErrWrongKind{
		// TODO
	}
}

func (w *_assembler) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	panic("TODO")
}
func (w *_assembler) AssignNull() error {
	panic("TODO")
}
func (w *_assembler) AssignBool(bool) error {
	panic("TODO")
}
func (w *_assembler) AssignInt(int64) error {
	panic("TODO")
}
func (w *_assembler) AssignFloat(float64) error {
	panic("TODO")
}
func (w *_assembler) AssignString(s string) error {
	switch w.val.Kind() {
	case reflect.String:
		w.val.SetString(s)
	}
	return &ipld.ErrWrongKind{
		// TODO
	}
}
func (w *_assembler) AssignBytes([]byte) error {
	panic("TODO")
}
func (w *_assembler) AssignLink(ipld.Link) error {
	panic("TODO")
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
		am, err := w.BeginMap(-1)
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
	case ipld.Kind_String:
		s, err := node.AsString()
		if err != nil {
			return err
		}
		return w.AssignString(s)
	}
	// fmt.Println(w.val.Type(), newVal.Type())
	panic(fmt.Sprintf("TODO: %v %v", w.val.Type(), node.Kind()))
}
func (w *_assembler) Prototype() ipld.NodePrototype {
	panic("TODO")
}

type _mapAssembler struct {
	val reflect.Value

	// TODO: state checks
	// TODO: optimize for structs

	curKey _builder
}

func (w *_mapAssembler) AssembleKey() ipld.NodeAssembler {
	switch typ := w.val.Type(); typ.Kind() {
	case reflect.Struct:
		w.curKey = _builder{_assembler{reflect.New(reflect.TypeOf("")).Elem()}}
	case reflect.Map:
		w.curKey = _builder{_assembler{reflect.New(typ.Key()).Elem()}}
	default:
		panic("unreachable")
	}
	return &w.curKey
}
func (w *_mapAssembler) AssembleValue() ipld.NodeAssembler {
	panic("TODO")
}
func (w *_mapAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	panic("TODO")
}
func (w *_mapAssembler) Finish() error {
	return nil
}
func (w *_mapAssembler) KeyPrototype() ipld.NodePrototype {
	switch w.val.Kind() {
	case reflect.Struct:
		return _prototype{reflect.TypeOf("")}
	case reflect.Map:
		return _prototype{w.val.Type().Key()}
	}
	panic("unreachable")
}
func (w *_mapAssembler) ValuePrototype(k string) ipld.NodePrototype {
	panic("TODO")
}
