// Package translator содержит реализацию транслятора в Z3
package translator

import (
	"fmt"
	"math/big"
	"symbolic-execution-course/internal/memory"
	"symbolic-execution-course/internal/symbolic"

	"github.com/ebukreev/go-z3/z3"
)

// Z3Translator транслирует символьные выражения в Z3 формулы
type Z3Translator struct {
	ctx       *z3.Context
	solver    *z3.Solver
	config    *z3.Config
	vars      map[string]z3.Value // Кэш переменных
	functions map[string]z3.FuncDecl
	mem       *memory.SymbolicMemory
}

func (zt *Z3Translator) VisitRef(expr *symbolic.Ref) interface{} {
	v, err := zt.TranslateExpression(expr.Id.(symbolic.SymbolicExpression))
	if err != nil {
		panic(err)
	}
	bv, ok := v.(z3.BV)
	if !ok {
		panic(fmt.Sprintf("VisitRef: expected BV (got %T)", v))
	}
	return bv
}

// NewZ3Translator создаёт новый экземпляр Z3 транслятора
func NewZ3Translator() *Z3Translator {
	config := &z3.Config{}
	ctx := z3.NewContext(config)

	return &Z3Translator{
		ctx:       ctx,
		solver:    z3.NewSolver(ctx),
		config:    config,
		vars:      make(map[string]z3.Value),
		functions: make(map[string]z3.FuncDecl),
		mem:       memory.NewSymbolicMemory(),
	}
}

// GetContext возвращает Z3 контекст
func (zt *Z3Translator) GetContext() interface{} {
	return zt.ctx
}

// GetSolver возвращает Z3 солвер
func (zt *Z3Translator) GetSolver() *z3.Solver {
	return zt.solver
}

func (zt *Z3Translator) IsSat() (bool, error) {
	return zt.solver.Check()
}

func (zt *Z3Translator) Assert(e symbolic.SymbolicExpression) {
	v, err := zt.TranslateExpression(e)
	if err != nil {
		panic("Ошибка трансляции: " + err.Error())
	}
	switch v := v.(type) {
	case z3.Bool:
		zt.solver.Assert(v)
	case z3.Value:
		if b, ok := v.(z3.Bool); ok {
			zt.solver.Assert(b)
			return
		}
		panic(fmt.Sprintf("non-bool z3.Value (%T) passed", v))
	default:
		panic(fmt.Sprintf("unsupported type %T", v))
	}
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

// VisitStringConstant транслирует строковую константу в Z3
func (zt *Z3Translator) VisitStringConstant(expr *symbolic.StringConstant) interface{} {
	return zt.ctx.Const("string_"+expr.Value, zt.ctx.UninterpretedSort("String"))
}

// VisitFloatConstant транслирует константу с плавающей точкой в Z3
func (zt *Z3Translator) VisitFloatConstant(expr *symbolic.FloatConstant) interface{} {
	v := zt.ctx.FromFloat64(expr.Value, zt.ctx.FloatSort(11, 53))
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
	case symbolic.StringType:
		return zt.ctx.UninterpretedSort("String")
	case symbolic.FloatType:
		return zt.ctx.RealSort()
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

func (zt *Z3Translator) makeSortInArray(expr *symbolic.SymbolicArray) z3.Sort {
	switch expr.ElemType.ExprType {
	case symbolic.IntType:
		return zt.ctx.BVSort(32)
	case symbolic.BoolType:
		return zt.ctx.BoolSort()
	case symbolic.StringType:
		return zt.ctx.UninterpretedSort("String")
	case symbolic.FloatType:
		return zt.ctx.RealSort()
	case symbolic.StructType:
		return zt.ctx.BVSort(32)
	default:
		panic(fmt.Sprintf("unsupported array elem type: %v", expr.ElemType.ExprType))
	}
}

func (zt *Z3Translator) VisitArray(expr *symbolic.SymbolicArray) interface{} {
	arrSort := zt.ctx.ArraySort(zt.ctx.BVSort(32), zt.makeSortInArray(expr))
	arr := zt.ctx.Const(expr.Name, arrSort).(z3.Array)
	res := make([]z3.Bool, 0)

	for _, op := range expr.Operations {
		if op.IsStore {
			arr = zt.visitSelect(op, arr)
		} else {
			res = zt.visitStore(op, arr, res)
		}
	}

	acc := zt.ctx.FromBool(true)
	for _, e := range res {
		acc = acc.And(e)
	}
	return acc
}

func (zt *Z3Translator) visitStore(op symbolic.ArrayOperation, arr z3.Array, res []z3.Bool) []z3.Bool {
	idxV, err := zt.TranslateExpression(op.Index)
	if err != nil {
		panic(fmt.Sprintf("translate op : %v", err))
	}
	idx, ok := idxV.(z3.Value)
	if !ok {
		panic(fmt.Sprintf("op is not z3.Value (got %T)", idxV))
	}

	sel := arr.Select(idx)

	resV, err := zt.TranslateExpression(op.Value)
	if err != nil {
		panic(fmt.Sprintf("translate select result: %v", err))
	}

	switch resT := resV.(type) {

	case z3.BV:
		selT, ok := sel.(z3.BV)
		if !ok {
			panic(fmt.Sprintf("select type mismatch: res is BV, sel is %T", sel))
		}
		res = append(res, resT.Eq(selT))

	case z3.Bool:
		selT, ok := sel.(z3.Bool)
		if !ok {
			panic(fmt.Sprintf("select type mismatch: res is Bool, sel is %T", sel))
		}
		res = append(res, resT.Eq(selT))

	case z3.Uninterpreted:
		selT, ok := sel.(z3.Uninterpreted)
		if !ok {
			panic(fmt.Sprintf("select type mismatch: res is Uninterpreted, sel is %T", sel))
		}
		res = append(res, resT.Eq(selT))

	case z3.Real:
		selT, ok := sel.(z3.Real)
		if !ok {
			panic(fmt.Sprintf("select type mismatch: res is Real, sel is %T", sel))
		}
		res = append(res, resT.Eq(selT))

	case z3.Array:
		selT, ok := sel.(z3.Array)
		if !ok {
			panic(fmt.Sprintf("select type mismatch: res is Array, sel is %T", sel))
		}
		res = append(res, resT.Eq(selT))

	case z3.Value:
		panic(fmt.Sprintf("select unsupported Value subtype: %T (sel=%T)", resV, sel))

	default:
		panic(fmt.Sprintf("select result is not a z3.Value subtype (got %T)", resV))
	}
	return res
}

func (zt *Z3Translator) visitSelect(op symbolic.ArrayOperation, arr z3.Array) z3.Array {
	idxV, err := zt.TranslateExpression(op.Index)
	if err != nil {
		panic(fmt.Sprintf("translate op: %v", err))
	}
	idx, ok := idxV.(z3.Value)
	if !ok {
		panic(fmt.Sprintf("op is not z3.Value (got %T)", idxV))
	}

	valV, err := zt.TranslateExpression(op.Value)
	if err != nil {
		panic(fmt.Sprintf("translate store value: %v", err))
	}
	val, ok := valV.(z3.Value)
	if !ok {
		panic(fmt.Sprintf("store value is not z3.Value (got %T)", valV))
	}

	arr = arr.Store(idx, val)
	return arr
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

// Вспомогательные методы

// createZ3Variable создаёт Z3 переменную соответствующего типа
func (zt *Z3Translator) createZ3Variable(name string, exprType symbolic.ExpressionType) z3.Value {
	// Создать Z3 переменную на основе типа
	switch exprType.ExprType {
	case symbolic.IntType:
		return zt.ctx.BVConst(name, 32)
	case symbolic.BoolType:
		return zt.ctx.BoolConst(name)
	case symbolic.StringType:
		return zt.ctx.Const(name, zt.ctx.UninterpretedSort("String"))
	case symbolic.FloatType:
		return zt.ctx.Const(name, zt.ctx.FloatSort(11, 53))
	case symbolic.ArrayType:
		return zt.ctx.Const(name, zt.ctx.ArraySort(zt.ctx.BVSort(32), zt.ctx.BVSort(32)))
	case symbolic.StructType:
		return zt.ctx.Const(name, zt.ctx.BVSort(32))
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
