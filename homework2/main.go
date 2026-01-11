// Демонстрационная программа для тестирования символьных выражений
package main

import (
	"fmt"
	"log"
	. "symbolic-execution-course/internal/symbolic"
	. "symbolic-execution-course/internal/translator"

	"github.com/ebukreev/go-z3/z3"
)

func main() {
	fmt.Println("=== Symbolic Expressions Demo ===")

	// Создаём простые символьные выражения
	x := NewSymbolicVariable("x", ExpressionType{ExprType: IntType})
	y := NewSymbolicVariable("y", ExpressionType{ExprType: IntType})
	five := NewIntConstant(5)

	// Создаём выражение: x + y > 5
	sum := NewBinaryOperation(x, y, ADD)
	condition := NewBinaryOperation(sum, five, GT)

	fmt.Printf("Выражение: %s\n", condition.String())
	fmt.Printf("Тип выражения: %s\n", condition.Type().String())

	// Создаём Z3 транслятор
	translator := NewZ3Translator()
	defer translator.Close()

	// Транслируем в Z3
	z3Expr, err := translator.TranslateExpression(condition)
	if err != nil {
		log.Fatalf("Ошибка трансляции: %v", err)
	}

	fmt.Printf("Z3 выражение создано: %T\n", z3Expr)
	solveXY(translator, z3Expr, x, y)

	// Создаём более сложное выражение: (x > 0) && (y < 10) && (y > 0) && ((x + y) == 5))
	zero := NewIntConstant(0)
	ten := NewIntConstant(10)

	cond1 := NewBinaryOperation(x, zero, GT)
	cond2 := NewBinaryOperation(y, ten, LT)
	cond3 := NewBinaryOperation(y, zero, GT)
	cond4 := NewBinaryOperation(NewBinaryOperation(x, y, ADD), NewIntConstant(5), EQ)

	andExpr := NewLogicalOperation([]SymbolicExpression{cond1, cond2, cond3, cond4}, AND)

	fmt.Printf("Сложное выражение: %s\n", andExpr.String())

	// Транслируем сложное выражение
	z3AndExpr, err := translator.TranslateExpression(andExpr)
	if err != nil {
		log.Fatalf("Ошибка трансляции сложного выражения: %v", err)
	}

	fmt.Printf("Сложное Z3 выражение создано: %T\n", z3AndExpr)
	solveXY(translator, z3AndExpr, x, y)
}

func solveXY(translator *Z3Translator, z3Expr interface{}, x *SymbolicVariable, y *SymbolicVariable) (int64, int64, bool) {
	solver := z3.NewSolver(translator.GetContext().(*z3.Context))

	solver.Assert(z3Expr.(z3.Bool))
	sat, err := solver.Check()
	if err != nil {
		log.Fatal(err)
	}

	z3x, _ := translator.TranslateExpression(x)
	z3y, _ := translator.TranslateExpression(y)

	if sat {
		model := solver.Model()

		xVal, _, _ := model.Eval(z3x.(z3.BV), false).(z3.BV).AsInt64()
		yVal, _, _ := model.Eval(z3y.(z3.BV), false).(z3.BV).AsInt64()

		fmt.Printf("Решение найдено: x = %d, y = %d\n", xVal, yVal)
		return xVal, yVal, true
	} else {
		fmt.Println("Решение не найдено")
		return 0, 0, false
	}
}
