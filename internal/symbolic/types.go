// Package symbolic определяет базовые типы символьных выражений
package symbolic

import "fmt"

// PrimitiveType представляет тип символьного выражения
type PrimitiveType int

const (
	IntType PrimitiveType = iota
	BoolType
	ArrayType
	// Добавьте другие типы по необходимости
)

type ExpressionType struct {
	ExprType PrimitiveType
	Param    *ExpressionType
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
	case ArrayType:
		return "array"
	default:
		return "unknown"
	}
}
