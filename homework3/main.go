package main

import (
	"fmt"
	"symbolic-execution-course/internal/memory"
	. "symbolic-execution-course/internal/symbolic"
)

func main() {
	mem := memory.NewSymbolicMemory()

	// Простая аллокация и работа с примитивами
	fmt.Println("1. Работа с примитивными типами:")
	intRef := mem.Allocate(IntExpr(), false)
	mem.AssignValue(intRef, NewIntConstant(42))
	intVal := mem.GetValue(intRef)
	fmt.Printf("   int ref: %s = %s\n", intRef.Id, intVal)

	boolRef := mem.Allocate(BoolExpr(), false)
	mem.AssignValue(boolRef, NewBoolConstant(true))
	boolVal := mem.GetValue(boolRef)
	fmt.Printf("   bool ref: %s = %s\n\n", boolRef.Id, boolVal)

	// Работа с массивом
	fmt.Println("2. Работа с одномерным массивом:")
	arr := mem.Allocate(ArrayExpr(IntExpr(), 1), false)

	for i := 0; i < 3; i++ {
		elemAddr := mem.ArrayElemRef(arr, NewIntConstant(int64(i)))
		mem.AssignValue(elemAddr, NewIntConstant(int64(i*10)))
	}

	for i := 0; i < 3; i++ {
		elemAddr := mem.ArrayElemRef(arr, NewIntConstant(int64(i)))
		val := mem.GetValue(elemAddr)
		fmt.Printf("   arr[%d] = %s\n", i, val)
	}
	fmt.Println()

	fmt.Println("3. Работа со структурами:")
	structFieldTypes := []ExpressionType{
		{ExprType: IntType},
		{ExprType: BoolType},
		{ExprType: StringType},
	}
	structRef := mem.Allocate(ExpressionType{
		ExprType:   StructType,
		Param:      nil,
		Name:       &[]string{"Person"}[0],
		Fields:     &structFieldTypes,
		FieldIndex: nil,
	}, false)

	// Записываем поля структуры
	field0 := mem.FieldRef(structRef, 0)
	mem.AssignValue(field0, NewIntConstant(25))

	field1 := mem.FieldRef(structRef, 1)
	mem.AssignValue(field1, NewBoolConstant(true))

	field2 := mem.FieldRef(structRef, 2)
	mem.AssignValue(field2, NewStringConstant("John"))

	// Читаем поля структуры
	fmt.Printf("   Person.field[0] (age) = %s\n", mem.GetValue(field0))
	fmt.Printf("   Person.field[1] (active) = %s\n", mem.GetValue(field1))
	fmt.Printf("   Person.field[2] (name) = %s\n\n", mem.GetValue(field2))

	fmt.Println("4. Ограничения алиасинга:")
	_ = mem.Allocate(IntExpr(), true)  // параметр (может быть алиасом)
	_ = mem.Allocate(IntExpr(), false) // локальная аллокация

	constraints := mem.GetAliasingConstraints()
	fmt.Printf(" Количество ограничений: %d\n", len(constraints))
	for i, c := range constraints {
		fmt.Printf(" Ограничение %d: %s\n", i, c)
	}
	fmt.Println()

	fmt.Println("5. Сегменты памяти:")
	segments := constraints
	displayed := 0
	for i, seg := range segments {
		if _, ok := seg.(*SymbolicArray); ok {
			fmt.Printf("   Сегмент %d: %s\n", i, seg)
			displayed++
		}
	}
}
