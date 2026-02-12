// Package symbolic определяет базовые типы символьных выражений
package symbolic

import "fmt"

type Id uint64

// PrimitiveType представляет тип символьного выражения
type PrimitiveType int

const (
	IntType PrimitiveType = iota
	BoolType
	StringType
	FloatType
	ArrayType
	RefType
	StructType
)

type ExpressionType struct {
	ExprType   PrimitiveType
	Param      *ExpressionType
	Name       *string
	Fields     *[]ExpressionType
	FieldIndex *int
}

func IntExpr() ExpressionType {
	return ExpressionType{ExprType: IntType}
}

func BoolExpr() ExpressionType {
	return ExpressionType{ExprType: BoolType}
}

func ArrayExpr(param ExpressionType) ExpressionType {
	return ExpressionType{ExprType: ArrayType, Param: &param}
}

func (g ExpressionType) String() string {
	if g.Param == nil {
		return g.ExprType.String()
	}

	return fmt.Sprintf("%s[%s]", g.ExprType.String(), g.Param.String())
}

// String возвращает строковое представление типа
func (et PrimitiveType) String() string {
	switch et {
	case IntType:
		return "int"
	case BoolType:
		return "bool"
	case StringType:
		return "string"
	case FloatType:
		return "float"
	case ArrayType:
		return "array"
	case RefType:
		return "ref"
	case StructType:
		return "struct"
	default:
		return "unknown"
	}
}
