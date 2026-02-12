package main

import (
	"fmt"
	"log"
	. "symbolic-execution-course/internal/symbolic"
	. "symbolic-execution-course/internal/translator"
	"testing"
)

/*
	func TestArrayExample(t *testing.T) {
		x := NewSymbolicVariable("x", IntExpr())
		y := NewSymbolicVariable("y", IntExpr())
		one := NewIntConstant(1)
		five := NewIntConstant(5)
		sorokDva := NewIntConstant(42)
		arr := NewSymbolicArray("arr", IntExpr())

		// Создаём выражения:
		// 1. -x + y == (arr[5] = 1)[5]
		// 2. x * y == 42
		// Ожидаемый ответ x = -7, y = -6

		//sum := NewBinaryOperation(x, y, ADD)
		sum := NewBinaryOperation(NewUnaryOperation(x, UNARY_MINUS), y, ADD)
		first := NewBinaryOperation(sum, NewArraySelect(NewArrayStore(arr, five, one), five), EQ)

		mul := NewBinaryOperation(x, y, MUL)
		second := NewBinaryOperation(mul, sorokDva, EQ)
		condition := NewLogicalOperation([]SymbolicExpression{first, second}, AND)

		fmt.Printf("Выражение: %s\n", condition.String())
		fmt.Printf("Тип выражения: %s\n", condition.Type().ExprType.String())

		// Создаём Z3 транслятор
		translator := NewZ3Translator()
		defer translator.Close()

		// Транслируем в Z3
		z3Expr, err := translator.TranslateExpression(condition)
		if err != nil {
			log.Fatalf("Ошибка трансляции: %v", err)
		}

		fmt.Printf("Z3 выражение создано: %T\n", z3Expr)
		resX, resY, found := solveXY(translator, z3Expr, x, y)
		if !found || resX != -7 || resY != -6 {
			log.Fatalf("test failed")
		}
	}
*/
func TestCondition(t *testing.T) {
	x := NewSymbolicVariable("x", IntExpr())
	y := NewSymbolicVariable("y", IntExpr())
	one := NewIntConstant(1)
	five := NewIntConstant(5)
	dva := NewIntConstant(2)

	// Создаём выражение
	// ((x > 5) ? y : 1) == 2)

	gt := NewBinaryOperation(x, five, GT)
	ternar := NewConditionalExpression(gt, y, one)

	condition := NewBinaryOperation(ternar, dva, EQ)

	fmt.Printf("Выражение: %s\n", condition.String())
	fmt.Printf("Тип выражения: %s\n", condition.Type().ExprType.String())

	// Создаём Z3 транслятор
	translator := NewZ3Translator()
	defer translator.Close()

	// Транслируем в Z3
	z3Expr, err := translator.TranslateExpression(condition)
	if err != nil {
		log.Fatalf("Ошибка трансляции: %v", err)
	}

	fmt.Printf("Z3 выражение создано: %T\n", z3Expr)
	resX, resY, found := solveXY(translator, z3Expr, x, y)
	if !found || resX <= 5 || resY != 2 {
		log.Fatalf("test failed")
	}
}

func TestFN(t *testing.T) {
	x := NewSymbolicVariable("x", IntExpr())
	y := NewSymbolicVariable("y", IntExpr())
	SorokDva := NewIntConstant(42)

	// Создаём выражение
	// x == y && fn(x) + fn(y) == 42 && y == fn(x)

	eq := NewBinaryOperation(x, y, EQ)

	fn := NewFunction("fn", []ExpressionType{IntExpr()}, IntExpr())
	fnRes := NewFunctionCall(*fn, []SymbolicExpression{x})
	fn2Res := NewFunctionCall(*fn, []SymbolicExpression{y})
	eq2 := NewBinaryOperation(NewBinaryOperation(fnRes, fn2Res, ADD), SorokDva, EQ)

	eq3 := NewBinaryOperation(NewBinaryOperation(y, fnRes, ADD), SorokDva, EQ)

	condition := NewLogicalOperation([]SymbolicExpression{eq, eq2, eq3}, AND)

	fmt.Printf("Выражение: %s\n", condition.String())
	fmt.Printf("Тип выражения: %s\n", condition.Type().ExprType.String())

	// Создаём Z3 транслятор
	translator := NewZ3Translator()
	defer translator.Close()

	// Транслируем в Z3
	z3Expr, err := translator.TranslateExpression(condition)
	if err != nil {
		log.Fatalf("Ошибка трансляции: %v", err)
	}

	fmt.Printf("Z3 выражение создано: %T\n", z3Expr)
	resX, resY, found := solveXY(translator, z3Expr, x, y)
	if !found || resX != 21 || resY != 21 {
		log.Fatalf("test failed")
	}
}
