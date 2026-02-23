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
	return ExpressionType{ExprType: IntType}
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
	return ExpressionType{ExprType: BoolType}
}

// String возвращает строковое представление константы
func (bc *BoolConstant) String() string {
	return fmt.Sprintf("%t", bc.Value)
}

// Accept реализует Visitor pattern
func (bc *BoolConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitBoolConstant(bc)
}

// StringConstant представляет строковую константу
type StringConstant struct {
	Value string
}

// NewStringConstant создаёт новую строковую константу
func NewStringConstant(value string) *StringConstant {
	return &StringConstant{Value: value}
}

// Type возвращает тип константы
func (sc *StringConstant) Type() ExpressionType {
	return ExpressionType{ExprType: StringType}
}

// String возвращает строковое представление константы
func (sc *StringConstant) String() string {
	return fmt.Sprintf("%q", sc.Value)
}

// Accept реализует Visitor pattern
func (sc *StringConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitStringConstant(sc)
}

// FloatConstant представляет константу с плавающей точкой
type FloatConstant struct {
	Value float64
}

// NewFloatConstant создаёт новую константу с плавающей точкой
func NewFloatConstant(value float64) *FloatConstant {
	return &FloatConstant{Value: value}
}

// Type возвращает тип константы
func (fc *FloatConstant) Type() ExpressionType {
	return ExpressionType{ExprType: FloatType}
}

// String возвращает строковое представление константы
func (fc *FloatConstant) String() string {
	return fmt.Sprintf("%g", fc.Value)
}

// Accept реализует Visitor pattern
func (fc *FloatConstant) Accept(visitor Visitor) interface{} {
	return visitor.VisitFloatConstant(fc)
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
	if left.Type().ExprType != right.Type().ExprType {
		panic("left and right types do not match")
	}
	return &BinaryOperation{Left: left, Right: right, Operator: op}
}

// Type возвращает результирующий тип операции
func (bo *BinaryOperation) Type() ExpressionType {
	// Определить результирующий тип на основе операции и типов операндов
	// Например: int + int = int, int < int = bool
	switch bo.Operator {
	case ADD, SUB, MUL, DIV, MOD:
		if bo.Left.Type().ExprType == FloatType || bo.Right.Type().ExprType == FloatType {
			return FloatExpr()
		} else {
			return IntExpr()
		}
	case EQ, NE, LT, LE, GT, GE:
		return BoolExpr()
	case AND, OR, XOR, SHL, SHR, AND_NOT, IMPLIES:
		if bo.Left.Type().ExprType == IntType || bo.Right.Type().ExprType == IntType {
			return IntExpr()
		} else {
			if bo.Operator == AND || bo.Operator == OR || bo.Operator == IMPLIES {
				return BoolExpr()
			}
			panic("unsupported operator for bool type")
		}
	default:
		panic("не реализовано")
	}
}

// String возвращает строковое представление операции
func (bo *BinaryOperation) String() string {
	// Формат: "(left operator right)"
	return fmt.Sprintf("(%s %s %s)", bo.Left.String(), bo.Operator.String(), bo.Right.String())
}

// Accept реализует Visitor pattern
func (bo *BinaryOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitBinaryOperation(bo)
}

// LogicalOperation представляет логическую операцию
type LogicalOperation struct {
	Operands []SymbolicExpression
	Operator BinaryOperator
}

// NewLogicalOperation создаёт новую логическую операцию
func NewLogicalOperation(operands []SymbolicExpression, op BinaryOperator) *LogicalOperation {
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
	}
	for _, operand := range operands {
		if operand.Type().ExprType != BoolType {
			panic("not bool operand")
		}
	}
	return &LogicalOperation{Operands: operands, Operator: op}
}

// Type возвращает тип логической операции (всегда bool)
func (lo *LogicalOperation) Type() ExpressionType {
	return BoolExpr()
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

	AND
	OR
	XOR
	SHL
	SHR
	AND_NOT
	IMPLIES
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
	case AND:
		return "&&"
	case OR:
		return "||"
	case IMPLIES:
		return "=>"
	default:
		return "unknown"
	}
}

type UnaryOperator int

const (
	UNARY_MINUS UnaryOperator = iota
	INVERT
	NOT
)

type UnaryOperation struct {
	Operand  SymbolicExpression
	Operator UnaryOperator
}

func (uo *UnaryOperation) String() string {
	return fmt.Sprintf("%s%s", uo.Operator.String(), uo.Operand.String())
}

func (uo *UnaryOperation) Accept(visitor Visitor) interface{} {
	return visitor.VisitUnaryOperation(uo)
}

// - UnaryOperation (унарные операции: -x, !x)
func (uo UnaryOperator) String() string {
	switch uo {
	case UNARY_MINUS:
		return "-"
	case INVERT:
		return "^"
	case NOT:
		return "!"
	default:
		return "unknown"
	}
}

func (uo *UnaryOperation) Type() ExpressionType {
	switch uo.Operator {
	case UNARY_MINUS, INVERT:
		return IntExpr()
	case NOT:
		return BoolExpr()
	default:
		panic("не реализовано")
	}
}

func NewUnaryOperation(operand SymbolicExpression, op UnaryOperator) *UnaryOperation {
	if (op == UNARY_MINUS || op == INVERT) && operand.Type().ExprType != IntType {
		panic("operand type is not Int")
	}
	if op == NOT && operand.Type().ExprType != BoolType {
		panic("operand type is not Bool")
	}
	return &UnaryOperation{Operand: operand, Operator: op}
}

// SymbolicArray Symbolic array type
type SymbolicArray struct {
	Name       string
	ElemType   ExpressionType
	Operations []ArrayOperation
}

func NewSymbolicArray(name string, elemType ExpressionType) *SymbolicArray {
	return &SymbolicArray{name, elemType, make([]ArrayOperation, 0)}
}

func (a *SymbolicArray) Type() ExpressionType {
	return ExpressionType{ExprType: ArrayType, Param: &a.ElemType}
}

func (a *SymbolicArray) String() string {
	res := fmt.Sprintf("%s[%s]", a.Name, a.ElemType)
	for _, op := range a.Operations {
		if op.IsStore {
			res += "[" + op.Index.String() + "<=" + op.Value.String() + ")"
		} else {
			res += "[" + op.Index.String() + "]~>" + op.Value.String() + ")"
		}
	}
	return res
}

func (a *SymbolicArray) Accept(visitor Visitor) interface{} {
	return visitor.VisitArray(a)
}

type ArrayOperation struct {
	IsStore bool
	Index   SymbolicExpression
	Value   SymbolicExpression
}

func (a *SymbolicArray) Select(index SymbolicExpression) SymbolicExpression {
	op := ArrayOperation{
		IsStore: false,
		Index:   index,
		Value: NewSymbolicVariable(
			fmt.Sprintf("%s_select%d", a.Name, len(a.Operations)),
			a.ElemType,
		),
	}
	a.Operations = append(a.Operations, op)
	return op.Value
}

func (a *SymbolicArray) Store(idx, val SymbolicExpression) {
	a.Operations = append(a.Operations, ArrayOperation{IsStore: true, Index: idx, Value: val})
}

// - FunctionCall (вызовы функций: f(x, y))

type Function struct {
	Name       string
	Args       []ExpressionType
	ReturnType ExpressionType
}

func NewFunction(name string, args []ExpressionType, returnType ExpressionType) *Function {
	return &Function{
		Name:       name,
		Args:       args,
		ReturnType: returnType,
	}
}

func (f *Function) Type() ExpressionType {
	return f.ReturnType
}

func (f *Function) String() string {
	return fmt.Sprintf("%s %s", f.Type(), f.Name)
}

func (f *Function) Accept(visitor Visitor) interface{} {
	return visitor.VisitFunction(f)
}

type FunctionCall struct {
	Func Function
	Args []SymbolicExpression
}

func NewFunctionCall(fun Function, args []SymbolicExpression) *FunctionCall {
	return &FunctionCall{Func: fun, Args: args}
}

func (fc *FunctionCall) Type() ExpressionType {
	return fc.Func.Type()
}

// String возвращает строковое представление операции
func (fc *FunctionCall) String() string {
	args := make([]string, len(fc.Args))
	for i := range fc.Args {
		args[i] = fc.Args[i].String()
	}
	return fmt.Sprintf("%s(%s)", fc.Func.Name, strings.Join(args, ", "))
}

func (fc *FunctionCall) Accept(visitor Visitor) interface{} {
	return visitor.VisitFunctionCall(fc)
}

// - ConditionalExpression тернарный оператор
type ConditionalExpression struct {
	Condition  SymbolicExpression
	ThenBranch SymbolicExpression
	ElseBranch SymbolicExpression
}

func (ce *ConditionalExpression) Accept(visitor Visitor) interface{} {
	return visitor.VisitConditionalExpression(ce)
}

func NewConditionalExpression(
	condition SymbolicExpression,
	thenBranch SymbolicExpression,
	elseBranch SymbolicExpression,
) *ConditionalExpression {
	if condition.Type().ExprType != BoolType {
		return nil
	}
	if thenBranch.Type() != elseBranch.Type() {
		return nil
	}

	return &ConditionalExpression{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func (ce *ConditionalExpression) Type() ExpressionType {
	return ce.ThenBranch.Type()
}

func (ce *ConditionalExpression) String() string {
	return fmt.Sprintf("(%s ? %s : %s)", ce.Condition.String(), ce.ThenBranch.String(), ce.ElseBranch.String())
}

type Ref struct {
	Id      SymbolicExpression
	RefType ExpressionType
}

func NewRef(refId Id, refType ExpressionType) *Ref {
	return &Ref{
		Id:      NewSymbolicVariable(fmt.Sprintf("_ref_%d", refId), IntExpr()),
		RefType: refType,
	}
}

func NewRefFromExpr(ref SymbolicExpression, refType ExpressionType) *Ref {
	return &Ref{
		Id:      ref,
		RefType: refType,
	}
}

func NewStructRef(refId Id, structTypeName string) *Ref {
	return &Ref{
		Id:      NewSymbolicVariable(fmt.Sprintf("_ref_%d", refId), IntExpr()),
		RefType: ExpressionType{ExprType: StructType, Name: &structTypeName},
	}
}

func (ref *Ref) Type() ExpressionType {
	return ExpressionType{ExprType: RefType}
}

func (ref *Ref) String() string {
	return fmt.Sprintf("&(%s)", ref.Id.String())
}

func (ref *Ref) Accept(visitor Visitor) interface{} {
	return visitor.VisitRef(ref)
}
