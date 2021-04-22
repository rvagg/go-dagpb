package dagpb

import (
	"fmt"
	"reflect"

	ipld "github.com/ipld/go-ipld-prime"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

var (
	typeString = reflect.TypeOf("")
	typeBytes  = reflect.TypeOf([]byte{})
	typeLink   = reflect.TypeOf((*ipld.Link)(nil)).Elem()
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
	return &_node{val}
}

type _node struct {
	val reflect.Value
}

func Unwrap(node ipld.Node) (ptr interface{}) {
	if w, ok := node.(*_node); ok {
		if w.val.Kind() == reflect.Ptr {
			panic("didn't expect w.val to be a pointer")
		}
		return w.val.Addr().Interface()
	}
	return nil
}

func (w *_node) under() reflect.Value {
	val := w.val
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

func (w *_node) Kind() ipld.Kind {
	typ := w.val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if w.val.IsNil() {
			return ipld.Kind_Null
		}
	}

	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return ipld.Kind_Int

	case reflect.Float32, reflect.Float64:
		return ipld.Kind_Float

	case reflect.String:
		return ipld.Kind_String

	case reflect.Struct, reflect.Map:
		return ipld.Kind_Map

	case reflect.Slice, reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			return ipld.Kind_Bytes
		}
		return ipld.Kind_List

	case reflect.Interface:
		if typ == typeLink {
			return ipld.Kind_Link
		}
	}
	return ipld.Kind_Invalid
}

func (w *_node) LookupByString(key string) (ipld.Node, error) {
	switch w.val.Kind() {
	case reflect.Struct:
		fval := w.val.FieldByName(key)
		if !fval.IsValid() {
			return nil, &ipld.ErrInvalidKey{
				TypeName: w.val.Type().Name(),
				Key:      basicnode.NewString(key),
			}
		}
		return &_node{fval}, nil
	case reflect.Map:
	}
	return nil, &ipld.ErrWrongKind{
		MethodName: "LookupByString",
		// TODO
	}
}

func (w *_node) LookupByNode(key ipld.Node) (ipld.Node, error) {
	panic("TODO: LookupByNode")
}

func (w *_node) LookupByIndex(idx int64) (ipld.Node, error) {
	switch w.val.Kind() {
	case reflect.Slice, reflect.Array:
		// TODO: bounds check
		val := w.val.Index(int(idx))
		return &_node{val}, nil
	}
	return nil, &ipld.ErrWrongKind{
		MethodName: "LookupByIndex",
		// TODO
	}
}

func (w *_node) LookupBySegment(seg ipld.PathSegment) (ipld.Node, error) {
	panic("TODO: LookupBySegment")
}

func (w *_node) MapIterator() ipld.MapIterator {
	switch w.val.Kind() {
	case reflect.Struct:
		return &_structIterator{val: w.val}
	case reflect.Map:
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
	switch val.Kind() {
	case reflect.Slice:
		return &_listIterator{val: val}
	}
	return nil
}

func (w *_node) Length() int64 {
	switch w.val.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return int64(w.val.Len())
	}
	return -1
}

// TODO: better story around pointers and absent/null

func (w *_node) IsAbsent() bool {
	switch w.val.Kind() {
	case reflect.Ptr:
		return w.val.IsNil()
	}
	return false
}

func (w *_node) IsNull() bool {
	return false // TODO
}

func (w *_node) AsBool() (bool, error) {
	panic("TODO: ")
}

func (w *_node) AsInt() (int64, error) {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if !val.IsNil() {
			val = val.Elem()
		}
	}
	if typ.Kind() != reflect.Int {
		return 0, &ipld.ErrWrongKind{
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
	if typ != typeLink {
		return nil, &ipld.ErrWrongKind{
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
	return &_node{w.val}
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
		doneFields := make([]bool, w.val.NumField())
		return &_structAssembler{
			val:        w.val,
			doneFields: doneFields,
		}, nil
	case reflect.Map:
	}
	return nil, &ipld.ErrWrongKind{
		MethodName: "BeginMap",
		// TODO
	}
}

func (w *_assembler) BeginList(sizeHint int64) (ipld.ListAssembler, error) {
	val := w.val
	typ := val.Type()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Slice && typ.Elem().Elem().Kind() == reflect.Uint8 {
		val.Set(reflect.New(typ.Elem()))
		val = val.Elem()
	}
	switch val.Kind() {
	case reflect.Slice:
		return &_listAssembler{val: val}, nil
	}
	return nil, &ipld.ErrWrongKind{
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
			MethodName: "AssignString",
			// TODO
		}
	}
	w.val.SetString(s)
	return nil
}

func (w *_assembler) AssignBytes(p []byte) error {
	typ := w.val.Type()
	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Slice && typ.Elem().Elem().Kind() == reflect.Uint8 {
		w.val.Set(reflect.New(typ.Elem()))
		w.val.Elem().SetBytes(p)
		return nil
	}
	if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Uint8 {
		return &ipld.ErrWrongKind{
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
	fmt.Println(w.val.Type(), reflect.TypeOf(node))
	panic(fmt.Sprintf("TODO: %v %v", w.val.Type(), node.Kind()))
}

func (w *_assembler) Prototype() ipld.NodePrototype {
	panic("TODO")
}

type _structAssembler struct {
	val reflect.Value

	// TODO: more state checks
	doneFields []bool

	// TODO: optimize for structs

	curKey _assembler
}

func (w *_structAssembler) AssembleKey() ipld.NodeAssembler {
	w.curKey = _assembler{reflect.New(typeString).Elem()}
	return &w.curKey
	// case reflect.Map: w.curKey = _assembler{reflect.New(typ.Key()).Elem()}
}

func (w *_structAssembler) AssembleValue() ipld.NodeAssembler {
	name := w.curKey.val.String()
	ftyp, ok := w.val.Type().FieldByName(name)
	if !ok {
		panic("TODO: missing fields")
	}
	if len(ftyp.Index) > 1 {
		panic("TODO: embedded fields")
	}
	w.doneFields[ftyp.Index[0]] = true
	fval := w.val.FieldByIndex(ftyp.Index)
	// TODO: reuse same assembler for perf?
	return &_assembler{fval}
}

func (w *_structAssembler) AssembleEntry(k string) (ipld.NodeAssembler, error) {
	panic("TODO")
}

func (w *_structAssembler) Finish() error {
	typ := w.val.Type()
	for i := 0; i < typ.NumField(); i++ {
		ftyp := typ.Field(i)
		if ftyp.Type.Kind() != reflect.Ptr && !w.doneFields[i] {
			return &ipld.ErrMissingRequiredField{Missing: []string{ftyp.Name}}
		}
	}
	return nil
}

func (w *_structAssembler) KeyPrototype() ipld.NodePrototype {
	return _prototype{typeString}
	// reflect.Map: return _prototype{w.val.Type().Key()}
}

func (w *_structAssembler) ValuePrototype(k string) ipld.NodePrototype {
	panic("TODO")
}

type _listAssembler struct {
	val reflect.Value
}

func (w *_listAssembler) AssembleValue() ipld.NodeAssembler {
	switch typ := w.val.Type(); typ.Kind() {
	case reflect.Slice:
		w.val.Set(reflect.Append(w.val, reflect.New(typ.Elem()).Elem()))
		return &_assembler{w.val.Index(w.val.Len() - 1)}
	default:
		panic("unreachable")
	}
}

func (w *_listAssembler) Finish() error {
	return nil
}

func (w *_listAssembler) ValuePrototype(idx int64) ipld.NodePrototype {
	panic("TODO")
}

type _structIterator struct {
	// TODO: support embedded fields?
	val       reflect.Value
	nextIndex int
}

func (w *_structIterator) Next() (key, value ipld.Node, _ error) {
	name := w.val.Type().Field(w.nextIndex).Name
	val := w.val.Field(w.nextIndex)
	w.nextIndex++
	return basicnode.NewString(name), &_node{val}, nil
}

func (w *_structIterator) Done() bool {
	return w.nextIndex >= w.val.NumField()
}

type _listIterator struct {
	val       reflect.Value
	nextIndex int
}

func (w *_listIterator) Next() (index int64, value ipld.Node, _ error) {
	idx := int64(w.nextIndex)
	val := w.val.Index(w.nextIndex)
	w.nextIndex++
	return idx, &_node{val}, nil
}

func (w *_listIterator) Done() bool {
	return w.nextIndex >= w.val.Len()
}
