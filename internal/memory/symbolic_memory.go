package memory

import (
	"fmt"
	"reflect"
	"strconv"
	. "symbolic-execution-course/internal/symbolic"
)

type Memory interface {
	Allocate(tpe ExpressionType, aliasable bool) *Ref

	AssignValue(ref *Ref, value SymbolicExpression)
	GetValue(ref *Ref) SymbolicExpression

	FieldRef(ref *Ref, fieldIndex int) *Ref
	ArrayElemRef(ref *Ref, elemIndex SymbolicExpression) *Ref

	Clone() Memory
	GetAliasingConstraints() []SymbolicExpression
}

type SymbolicMemory struct {
	IdCounter Id

	Segments       map[PrimitiveType]*SymbolicArray
	StructSegments map[string]map[int]*SymbolicArray

	Aliasable    []*Ref
	NonAliasable []*Ref
}

func NewSymbolicMemory() *SymbolicMemory {
	return &SymbolicMemory{
		IdCounter: 1,

		Segments: map[PrimitiveType]*SymbolicArray{
			IntType:    NewSymbolicArray("IntSegment", ExpressionType{ExprType: IntType}),
			BoolType:   NewSymbolicArray("BoolSegment", ExpressionType{ExprType: BoolType}),
			StringType: NewSymbolicArray("StringSegment", ExpressionType{ExprType: StringType}),
			FloatType:  NewSymbolicArray("FloatSegment", ExpressionType{ExprType: FloatType}),
			ArrayType:  NewSymbolicArray("ArraySegment", ExpressionType{ExprType: IntType}),
		},

		StructSegments: make(map[string]map[int]*SymbolicArray),

		Aliasable:    make([]*Ref, 0),
		NonAliasable: make([]*Ref, 0),
	}
}

func (mem *SymbolicMemory) nextId() Id {
	mem.IdCounter++
	return mem.IdCounter
}

func (mem *SymbolicMemory) Allocate(tpe ExpressionType, aliasable bool) (res *Ref) {
	switch tpe.ExprType {
	case IntType, BoolType, StringType, FloatType, ArrayType:
		res = NewRef(mem.nextId(), tpe)
	case StructType:
		if mem.StructSegments[*tpe.Name] == nil {
			mem.StructSegments[*tpe.Name] = make(map[int]*SymbolicArray)
			for i, f := range *tpe.Fields {
				arr := NewSymbolicArray(fmt.Sprintf("%s_%s", *tpe.Name, strconv.Itoa(i)), f)
				mem.StructSegments[*tpe.Name][i] = arr
			}
		}
		res = NewStructRef(mem.nextId(), *tpe.Name)
	default:
		panic("not implemented")
	}
	if aliasable {
		mem.Aliasable = append(mem.Aliasable, res)
	} else {
		mem.NonAliasable = append(mem.NonAliasable, res)
	}
	return
}

func (mem *SymbolicMemory) FieldRef(ref *Ref, fieldIndex int) *Ref {
	if ref == nil {
		panic("nil ref")
	}
	if ref.RefType.ExprType != StructType {
		panic("ref is not struct")
	}
	if ref.RefType.Name == nil {
		panic("empty struct type name")
	}
	if mem.StructSegments[*ref.RefType.Name] == nil {
		panic("unknown struct name")
	}
	if mem.StructSegments[*ref.RefType.Name][fieldIndex] == nil {
		panic("unknown struct field")
	}

	fr := *ref
	fr.RefType.FieldIndex = &fieldIndex
	return &fr
}

func (mem *SymbolicMemory) ArrayElemRef(ref *Ref, elemIndex SymbolicExpression) *Ref {
	if ref == nil {
		panic("ArrayElemRef: nil ref")
	}
	if ref.RefType.ExprType != ArrayType {
		panic("ArrayElemRef: ref is not array")
	}
	if elemIndex == nil {
		panic("ArrayElemRef: nil elemIndex")
	}

	switch ref.RefType.Param.ExprType {
	case IntType, BoolType, StringType, FloatType:
		return NewRefFromExpr(NewBinaryOperation(ref.Id, elemIndex, ADD), *ref.RefType.Param)
	case StructType:
		return NewRefFromExpr(
			mem.Segments[ArrayType].Select(NewBinaryOperation(ref.Id, elemIndex, ADD)),
			ExpressionType{ExprType: StructType, Name: ref.RefType.Name},
		)
	default:
		panic("not implemented")
	}
}

func (mem *SymbolicMemory) AssignValue(ref *Ref, value SymbolicExpression) {
	switch ref.RefType.ExprType {
	case IntType, BoolType, StringType, FloatType:
		mem.Segments[ref.RefType.ExprType].Store(ref.Id, value)

	case StructType:
		if ref.RefType.FieldIndex == nil {
			src, ok := value.(*Ref)
			if !ok || src.RefType.ExprType != StructType {
				panic("AssignValue: assigning non-struct into whole struct")
			}
			if *src.RefType.Name != *ref.RefType.Name {
				panic("AssignValue: struct type mismatch in whole-struct assignment")
			}

			fields := mem.StructSegments[*ref.RefType.Name]
			if fields == nil {
				panic("AssignValue: unknown struct type: " + *ref.RefType.Name)
			}

			for fieldIndex := range fields {
				dstField := mem.FieldRef(ref, fieldIndex)
				srcField := mem.FieldRef(src, fieldIndex)
				mem.AssignValue(dstField, mem.GetValue(srcField))
			}
			return
		}

		fieldSegment := mem.StructSegments[*ref.RefType.Name][*ref.RefType.FieldIndex]
		if fieldSegment == nil {
			panic("AssignValue: struct field segment is nil")
		}
		fieldSegment.Store(ref.Id, value)
	default:
		panic("unknown ref type")
	}
}

func (mem *SymbolicMemory) GetValue(ref *Ref) SymbolicExpression {
	switch ref.RefType.ExprType {
	case IntType, BoolType, StringType, FloatType:
		return mem.Segments[ref.RefType.ExprType].Select(ref.Id)
	case StructType:
		if ref.RefType.FieldIndex == nil {
			panic("GetValue: can't read whole struct; take field addr and read field")
		}
		return mem.StructSegments[*ref.RefType.Name][*ref.RefType.FieldIndex].Select(ref.Id)
	default:
		panic("unknown base type")
	}
}

func (mem *SymbolicMemory) GetAliasingConstraints() []SymbolicExpression {
	ops := make([]SymbolicExpression, 0)

	for _, a := range mem.Aliasable {
		for _, na := range mem.NonAliasable {
			ops = append(ops, NewBinaryOperation(a.Id, na.Id, NE))
		}
	}

	for i := 0; i < len(mem.NonAliasable); i++ {
		for j := i + 1; j < len(mem.NonAliasable); j++ {
			ops = append(ops, NewBinaryOperation(mem.NonAliasable[i].Id, mem.NonAliasable[j].Id, NE))
		}
	}

	res := []SymbolicExpression{}
	for _, seg := range mem.Segments {
		res = append(res, seg)
	}
	for _, fields := range mem.StructSegments {
		for _, seg := range fields {
			res = append(res, seg)
		}
	}
	return append(ops, res...)
}

func (mem *SymbolicMemory) Clone() Memory {
	return deepCopy(reflect.ValueOf(mem), make(map[uintptr]reflect.Value)).Interface().(Memory)
}

func deepCopy(v reflect.Value, pointers map[uintptr]reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Pointer:
		return copyPointer(v, pointers)
	case reflect.Interface:
		return copyInterface(v, pointers)
	case reflect.Struct:
		return copyStruct(v, pointers)
	case reflect.Map:
		return copyMap(v, pointers)
	case reflect.Slice:
		return copySlice(v, pointers)
	default:
		return v
	}
}

func copyMap(v reflect.Value, pointers map[uintptr]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}
	nv := reflect.MakeMapWithSize(v.Type(), v.Len())
	iter := v.MapRange()
	for iter.Next() {
		k := deepCopy(iter.Key(), pointers)
		val := deepCopy(iter.Value(), pointers)
		nv.SetMapIndex(k, val)
	}
	return nv
}

func copySlice(v reflect.Value, pointers map[uintptr]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}
	nv := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
	for i := 0; i < v.Len(); i++ {
		nv.Index(i).Set(deepCopy(v.Index(i), pointers))
	}
	return nv
}

func copyStruct(v reflect.Value, pointers map[uintptr]reflect.Value) reflect.Value {
	nv := reflect.New(v.Type()).Elem()
	for i := 0; i < v.NumField(); i++ {
		sf := v.Field(i)
		cf := deepCopy(sf, pointers)
		if nv.Field(i).CanSet() {
			nv.Field(i).Set(cf)
		} else {
			panic("not implemented")
		}
	}
	return nv
}

func copyInterface(v reflect.Value, pointers map[uintptr]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}
	ev := deepCopy(v.Elem(), pointers)
	nv := reflect.New(v.Type()).Elem()
	nv.Set(ev)
	return nv
}

func copyPointer(v reflect.Value, pointers map[uintptr]reflect.Value) reflect.Value {
	if v.IsNil() {
		return reflect.Zero(v.Type())
	}
	if got, ok := pointers[v.Pointer()]; ok {
		return got
	}
	nv := reflect.New(v.Elem().Type())
	pointers[v.Pointer()] = nv
	nv.Elem().Set(deepCopy(v.Elem(), pointers))
	return nv
}
