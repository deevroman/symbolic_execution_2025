// Package translator содержит реализацию транслятора в Z3
package translator

import (
	"fmt"
	"math/big"
	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

// Z3Translator транслирует символьные выражения в Z3 формулы
type Z3Translator struct {
	ctx       *z3.Context
	config    *z3.Config
	vars      map[string]z3.Value // Кэш переменных
	functions map[string]z3.FuncDecl
}

// NewZ3Translator создаёт новый экземпляр Z3 транслятора
func NewZ3Translator() *Z3Translator {
	config := &z3.Config{}
	ctx := z3.NewContext(config)

	return &Z3Translator{
		ctx:       ctx,
		config:    config,
		vars:      make(map[string]z3.Value),
		functions: make(map[string]z3.FuncDecl),
	}
}

// GetContext возвращает Z3 контекст
func (zt *Z3Translator) GetContext() interface{} {
	return zt.ctx
}

// Reset сбрасывает состояние транслятора
func (zt *Z3Translator) Reset() {
	zt.vars = make(map[string]z3.Value)
}

// Close освобождает ресурсы
func (zt *Z3Translator) Close() {
	// Z3 контекст закрывается автоматически
}

// TranslateExpression транслирует символьное выражение в Z3
func (zt *Z3Translator) TranslateExpression(expr symbolic.SymbolicExpression) (interface{}, error) {
	return expr.Accept(zt), nil
}

// VisitVariable транслирует символьную переменную в Z3
func (zt *Z3Translator) VisitVariable(expr *symbolic.SymbolicVariable) interface{} {
	// Проверить, есть ли переменная в кэше
	// Если нет - создать новую Z3 переменную соответствующего типа
	// Добавить в кэш и вернуть

	// Подсказки:
	// - Используйте zt.ctx.IntConst(name) для int переменных
	// - Используйте zt.ctx.BoolConst(name) для bool переменных
	// - Храните переменные в zt.vars для повторного использования
	v, ok := zt.vars[expr.Name]
	if ok {
		return v
	}
	v = zt.createZ3Variable(expr.Name, expr.ExprType)
	zt.vars[expr.Name] = v
	return v
}

// VisitIntConstant транслирует целочисленную константу в Z3
func (zt *Z3Translator) VisitIntConstant(expr *symbolic.IntConstant) interface{} {
	// Создать Z3 константу с помощью zt.ctx.FromBigInt или аналогичного метода
	v := zt.ctx.FromBigInt(big.NewInt(expr.Value), zt.ctx.BVSort(32))
	zt.vars[expr.String()] = v
	return v
}

// VisitBoolConstant транслирует булеву константу в Z3
func (zt *Z3Translator) VisitBoolConstant(expr *symbolic.BoolConstant) interface{} {
	// Использовать zt.ctx.FromBool для создания Z3 булевой константы
	v := zt.ctx.FromBool(expr.Value)
	zt.vars[expr.String()] = v
	return v
}

// VisitBinaryOperation транслирует бинарную операцию в Z3
func (zt *Z3Translator) VisitBinaryOperation(expr *symbolic.BinaryOperation) interface{} {
	// 1. Транслировать левый и правый операнды
	// 2. В зависимости от оператора создать соответствующую Z3 операцию

	// Подсказки по операциям в Z3:
	// - Арифметические: left.Add(right), left.Sub(right), left.Mul(right), left.Div(right)
	// - Сравнения: left.Eq(right), left.LT(right), left.LE(right), etc.
	// - Приводите типы: left.(z3.BV), right.(z3.BV) для int операций
	switch expr.Operator {
	case symbolic.ADD:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.Add(r)
	case symbolic.SUB:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.Sub(r)
	case symbolic.MUL:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.Mul(r)
	case symbolic.DIV:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.SDiv(r)
	case symbolic.MOD:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.SMod(r)
	case symbolic.EQ:
		l := expr.Left.Accept(zt)
		r := expr.Right.Accept(zt)
		switch l := l.(type) {
		case z3.BV:
			return l.Eq(r.(z3.BV))
		case z3.Bool:
			return l.Eq(r.(z3.Bool))
		default:
			panic("не реализовано")
		}
	case symbolic.NE:
		l := expr.Left.Accept(zt)
		r := expr.Right.Accept(zt)
		switch l := l.(type) {
		case z3.BV:
			return l.NE(r.(z3.BV))
		case z3.Bool:
			return l.NE(r.(z3.Bool))
		default:
			panic("не реализовано")
		}
	case symbolic.LT:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.SLT(r)
	case symbolic.LE:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.SLE(r)
	case symbolic.GT:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.SGT(r)
	case symbolic.GE:
		l := expr.Left.Accept(zt).(z3.BV)
		r := expr.Right.Accept(zt).(z3.BV)
		return l.SGE(r)
	default:
		panic("не реализовано")
	}
}

// VisitLogicalOperation транслирует логическую операцию в Z3
func (zt *Z3Translator) VisitLogicalOperation(expr *symbolic.LogicalOperation) interface{} {
	// 1. Транслировать все операнды
	// 2. Применить соответствующую логическую операцию

	// Подсказки:
	// - AND: zt.ctx.And(operands...)
	// - OR: zt.ctx.Or(operands...)
	// - NOT: operand.Not() (для единственного операнда)
	// - IMPLIES: antecedent.Implies(consequent)
	switch expr.Operator {
	case symbolic.AND:
		res := expr.Operands[0].Accept(zt).(z3.Bool)
		for _, operand := range expr.Operands[1:] {
			res = res.And(operand.Accept(zt).(z3.Bool))
		}
		return res
	case symbolic.OR:
		res := expr.Operands[0].Accept(zt).(z3.Bool)
		for _, operand := range expr.Operands[1:] {
			res = res.Or(operand.Accept(zt).(z3.Bool))
		}
		return res
	case symbolic.NOT:
		return expr.Operands[0].Accept(zt).(z3.Bool).Not()
	case symbolic.IMPLIES:
		return expr.Operands[0].Accept(zt).(z3.Bool).Implies(expr.Operands[1].Accept(zt).(z3.Bool))
	default:
		panic("не реализовано")
	}
}

func (zt *Z3Translator) VisitConditionalExpression(expr *symbolic.ConditionalExpression) interface{} {
	return expr.Condition.Accept(zt).(z3.Bool).IfThenElse(
		expr.ThenBranch.Accept(zt).(z3.Value),
		expr.ElseBranch.Accept(zt).(z3.Value),
	)
}

func (zt *Z3Translator) makeSort(expr symbolic.ExpressionType) z3.Sort {
	switch expr.ExprType {
	case symbolic.IntType:
		return zt.ctx.BVSort(32)
	case symbolic.BoolType:
		return zt.ctx.BoolSort()
	case symbolic.ArrayType:
		switch expr.Param.ExprType {
		case symbolic.BoolType:
			return zt.ctx.ArraySort(zt.ctx.BoolSort(), zt.ctx.BoolSort())
		case symbolic.IntType:
			return zt.ctx.ArraySort(zt.ctx.BVSort(32), zt.ctx.BVSort(32))
		default:
			panic("не реализовано")
		}
	default:
		panic("не реализовано")
	}
}

func (zt *Z3Translator) VisitFunction(expr *symbolic.Function) interface{} {
	if v, hasCache := zt.functions[expr.Name]; hasCache {
		return v
	}

	args := make([]z3.Sort, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = zt.makeSort(arg)
	}

	zt.functions[expr.Name] = zt.ctx.FuncDecl(expr.Name, args, zt.makeSort(expr.ReturnType))

	return zt.functions[expr.Name]
}

func (zt *Z3Translator) VisitFunctionCall(expr *symbolic.FunctionCall) interface{} {
	fun := expr.Func.Accept(zt).(z3.FuncDecl)

	args := make([]z3.Value, len(expr.Args))
	for i, arg := range expr.Args {
		args[i] = arg.Accept(zt).(z3.Value)
	}
	return fun.Apply(args...)
}

func (zt *Z3Translator) VisitArray(expr *symbolic.SymbolicArray) interface{} {
	switch expr.ElemType.ExprType {
	case symbolic.BoolType:
		zt.vars[expr.Name] = zt.ctx.FreshConst(expr.Name, zt.ctx.ArraySort(zt.ctx.BoolSort(), zt.ctx.BoolSort()))
	case symbolic.IntType:
		zt.vars[expr.Name] = zt.ctx.FreshConst(expr.Name, zt.ctx.ArraySort(zt.ctx.BVSort(32), zt.ctx.BVSort(32)))
	default:
		panic("не реализовано")
	}
	return zt.vars[expr.Name]
}

func (zt *Z3Translator) VisitUnaryOperation(expr *symbolic.UnaryOperation) interface{} {
	switch expr.Operator {
	case symbolic.UNARY_MINUS:
		o := expr.Operand.Accept(zt).(z3.BV)
		minusOne := symbolic.NewIntConstant(-1)
		return o.Mul(minusOne.Accept(zt).(z3.BV))
	default:
		panic("не реализовано")
	}
}

func (zt *Z3Translator) VisitArraySelect(expr *symbolic.ArraySelect) interface{} {
	return expr.Array.Accept(zt).(z3.Array).Select(
		expr.Index.Accept(zt).(z3.BV),
	)
}

func (zt *Z3Translator) VisitArrayStore(expr *symbolic.ArrayStore) interface{} {
	return expr.Array.Accept(zt).(z3.Array).Store(
		expr.Index.Accept(zt).(z3.BV),
		expr.Value.Accept(zt).(z3.Value),
	)
}

// Вспомогательные методы

// createZ3Variable создаёт Z3 переменную соответствующего типа
func (zt *Z3Translator) createZ3Variable(name string, exprType symbolic.ExpressionType) z3.Value {
	// Создать Z3 переменную на основе типа
	switch exprType.ExprType {
	case symbolic.IntType:
		return zt.ctx.BVConst(name, 32)
	case symbolic.BoolType:
		return zt.ctx.BoolConst(name)
	case symbolic.ArrayType:
		switch exprType.Param.ExprType {
		case symbolic.BoolType:
			return zt.ctx.ConstArray(zt.ctx.BoolSort(), zt.ctx.BVConst(name, 32))
		case symbolic.IntType:
			return zt.ctx.ConstArray(zt.ctx.BVSort(32), zt.ctx.BVConst(name, 32))
		default:
			panic("не реализовано")
		}
	default:
		panic("не реализовано")
	}
}

// castToZ3Type приводит значение к нужному Z3 типу
func (zt *Z3Translator) castToZ3Type(value interface{}, targetType symbolic.ExpressionType) (z3.Value, error) {
	// Безопасно привести interface{} к конкретному Z3 типу
	switch targetType.ExprType {
	case symbolic.IntType:
		v, ok := value.(z3.BV)
		if !ok {
			return nil, fmt.Errorf("bad cast")
		}
		return v, nil
	case symbolic.BoolType:
		v, ok := value.(z3.Bool)
		if !ok {
			return nil, fmt.Errorf("bad cast")
		}
		return v, nil
	case symbolic.ArrayType:
		v, ok := value.(z3.Array)
		if !ok {
			return nil, fmt.Errorf("bad cast")
		}
		return v, nil
	default:
		panic("не реализовано")
	}
}
