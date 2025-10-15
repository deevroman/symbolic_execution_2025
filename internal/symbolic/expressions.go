// Package symbolic содержит конкретные реализации символьных выражений
package symbolic

import (
	"fmt"
	"strings"
)

// SymbolicExpression - базовый интерфейс для всех символьных выражений
type SymbolicExpression interface {
	// Type возвращает тип выражения
	Type() ExpressionType

	// String возвращает строковое представление выражения
	String() string

	// Accept принимает visitor для обхода дерева выражений
	Accept(visitor Visitor) interface{}
}

// SymbolicVariable представляет символьную переменную
type SymbolicVariable struct {
	Name     string
	ExprType ExpressionType
}

// NewSymbolicVariable создаёт новую символьную переменную
func NewSymbolicVariable(name string, exprType ExpressionType) *SymbolicVariable {
	return &SymbolicVariable{
		Name:     name,
		ExprType: exprType,
	}
}

// Type возвращает тип переменной
func (sv *SymbolicVariable) Type() ExpressionType {
	return sv.ExprType
}

// String возвращает строковое представление переменной
func (sv *SymbolicVariable) String() string {
	return sv.Name
}

// Accept реализует Visitor pattern
func (sv *SymbolicVariable) Accept(visitor Visitor) interface{} {
	return visitor.VisitVariable(sv)
}

// IntConstant представляет целочисленную константу
type IntConstant struct {
	Value int64
}

// NewIntConstant создаёт новую целочисленную константу
func NewIntConstant(value int64) *IntConstant {
	return &IntConstant{Value: value}
}

// Type возвращает тип константы
func (ic *IntConstant) Type() ExpressionType {
	return IntType
}

// String возвращает строковое представление константы
func (ic *IntConstant) String() string {
	return fmt.Sprintf("%d", ic.Value)
}

// Accept реализует Visitor pattern
func (ic *IntConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitIntConstant(ic)
}

// BoolConstant представляет булеву константу
type BoolConstant struct {
	Value bool
}

// NewBoolConstant создаёт новую булеву константу
func NewBoolConstant(value bool) *BoolConstant {
	return &BoolConstant{Value: value}
}

// Type возвращает тип константы
func (bc *BoolConstant) Type() ExpressionType {
	return BoolType
}

// String возвращает строковое представление константы
func (bc *BoolConstant) String() string {
	return fmt.Sprintf("%t", bc.Value)
}

// Accept реализует Visitor pattern
func (bc *BoolConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitBoolConstant(bc)
}

// BinaryOperation представляет бинарную операцию
type BinaryOperation struct {
	Left     SymbolicExpression
	Right    SymbolicExpression
	Operator BinaryOperator
}

// NewBinaryOperation создаёт новую бинарную операцию
func NewBinaryOperation(left, right SymbolicExpression, op BinaryOperator) *BinaryOperation {
	// Создать новую бинарную операцию и проверить совместимость типов
	if left.Type() != IntType && right.Type() != IntType {
		panic("left and right types don't match")
	}
	return &BinaryOperation{Left: left, Right: right, Operator: op}
}

// Type возвращает результирующий тип операции
func (bo *BinaryOperation) Type() ExpressionType {
	// Определить результирующий тип на основе операции и типов операндов
	// Например: int + int = int, int < int = bool
	switch bo.Operator {
	case ADD:
		return IntType
	case SUB:
		return IntType
	case MUL:
		return IntType
	case DIV:
		return IntType
	case MOD:
		return IntType
	case EQ:
		return BoolType
	case NE:
		return BoolType
	case LT:
		return BoolType
	case LE:
		return BoolType
	case GT:
		return BoolType
	case GE:
		return BoolType
	default:
		panic("не реализовано")
	}
}

// String возвращает строковое представление операции
func (bo *BinaryOperation) String() string {
	// Формат: "(left operator right)"
	return fmt.Sprintf("(%s %s %s)", bo.Left.String(), bo.Right.String(), bo.Operator.String())
}

// Accept реализует Visitor pattern
func (bo *BinaryOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitBinaryOperation(bo)
}

// LogicalOperation представляет логическую операцию
type LogicalOperation struct {
	Operands []SymbolicExpression
	Operator LogicalOperator
}

// NewLogicalOperation создаёт новую логическую операцию
func NewLogicalOperation(operands []SymbolicExpression, op LogicalOperator) *LogicalOperation {
	// Создать логическую операцию и проверить типы операндов
	switch op {
	case AND:
		if len(operands) < 2 {
			panic("incorrect number of arguments for AND")
		}
	case OR:
		if len(operands) < 2 {
			panic("incorrect number of arguments for OR")
		}
	case IMPLIES:
		if len(operands) != 2 {
			panic("incorrect number of arguments for IMPLIES")
		}
	case NOT:
		if len(operands) != 1 {
			panic("incorrect number of arguments for NOT")
		}
	}
	for _, operand := range operands {
		if operand.Type() != BoolType {
			panic("not bool operand")
		}
	}
	return &LogicalOperation{Operands: operands, Operator: op}
}

// Type возвращает тип логической операции (всегда bool)
func (lo *LogicalOperation) Type() ExpressionType {
	return BoolType
}

// String возвращает строковое представление логической операции
func (lo *LogicalOperation) String() string {
	// Для NOT: "!operand"
	// Для AND/OR: "(operand1 && operand2 && ...)"
	// Для IMPLIES: "(operand1 => operand2)"
	switch lo.Operator {
	case AND:
		ops := make([]string, len(lo.Operands))
		for i, o := range lo.Operands {
			ops[i] = o.String()
		}
		return "(" + strings.Join(ops, " "+lo.Operator.String()+" ") + ")"
	case OR:
		ops := make([]string, len(lo.Operands))
		for i, o := range lo.Operands {
			ops[i] = o.String()
		}
		return "(" + strings.Join(ops, " "+lo.Operator.String()+" ") + ")"
	case NOT:
		return fmt.Sprintf("%s%s", lo.Operator.String(), lo.Operands[0].String())
	case IMPLIES:
		return fmt.Sprintf("%s %s %s", lo.Operands[0].String(), lo.Operator.String(), lo.Operands[1].String())
	default:
		panic("не реализовано")
	}
}

// Accept реализует Visitor pattern
func (lo *LogicalOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitLogicalOperation(lo)
}

// Операторы для бинарных выражений
type BinaryOperator int

const (
	// Арифметические операторы
	ADD BinaryOperator = iota
	SUB
	MUL
	DIV
	MOD

	// Операторы сравнения
	EQ // равно
	NE // не равно
	LT // меньше
	LE // меньше или равно
	GT // больше
	GE // больше или равно
)

// String возвращает строковое представление оператора
func (op BinaryOperator) String() string {
	switch op {
	case ADD:
		return "+"
	case SUB:
		return "-"
	case MUL:
		return "*"
	case DIV:
		return "/"
	case MOD:
		return "%"
	case EQ:
		return "=="
	case NE:
		return "!="
	case LT:
		return "<"
	case LE:
		return "<="
	case GT:
		return ">"
	case GE:
		return ">="
	default:
		return "unknown"
	}
}

// Логические операторы
type LogicalOperator int

const (
	AND LogicalOperator = iota
	OR
	NOT
	IMPLIES
)

// String возвращает строковое представление логического оператора
func (op LogicalOperator) String() string {
	switch op {
	case AND:
		return "&&"
	case OR:
		return "||"
	case NOT:
		return "!"
	case IMPLIES:
		return "=>"
	default:
		return "unknown"
	}
}

// TODO: Добавьте дополнительные типы выражений по необходимости:
// - UnaryOperation (унарные операции: -x, !x)
// - ArrayAccess (доступ к элементам массива: arr[index])
// - FunctionCall (вызовы функций: f(x, y))
// - ConditionalExpression (тернарный оператор: condition ? true_expr : false_expr)
