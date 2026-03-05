package main

import (
	"testing"

	"symbolic-execution-course/internal/memory"
	. "symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"

	"github.com/ebukreev/go-z3/z3"
)

func checkSat(t *testing.T, tr *translator.Z3Translator) (*z3.Model, bool) {
	t.Helper()
	sat, err := tr.IsSat()
	if err != nil {
		t.Fatalf("SAT check error: %v", err)
	}
	if !sat {
		return nil, false
	}
	return tr.GetSolver().Model(), true
}

func TestIntPreventAliasingArray(t *testing.T) {
	tr := translator.NewZ3Translator()
	mem := memory.NewSymbolicMemory()

	x := mem.Allocate(ArrayExpr(IntExpr(), 1), false)
	y := mem.Allocate(ArrayExpr(IntExpr(), 1), false)

	yElem := mem.ArrayElemRef(y, NewIntConstant(1))
	mem.AssignValue(yElem, NewIntConstant(5))

	xElem := mem.ArrayElemRef(x, NewIntConstant(1))
	mem.AssignValue(xElem, NewIntConstant(2))

	cond := NewBinaryOperation(
		mem.GetValue(yElem),
		NewIntConstant(2),
		EQ,
	)

	tr.Assert(cond)
	for _, v := range mem.GetAliasingConstraints() {
		tr.Assert(v)
	}

	if _, sat := checkSat(t, tr); sat {
		t.Fatal("expected UNSAT")
	}
}

func TestIntAliasingArray(t *testing.T) {
	tr := translator.NewZ3Translator()
	mem := memory.NewSymbolicMemory()

	x := mem.Allocate(ArrayExpr(IntExpr(), 1), true)
	y := mem.Allocate(ArrayExpr(IntExpr(), 1), true)

	yElem := mem.ArrayElemRef(y, NewIntConstant(1))
	mem.AssignValue(yElem, NewIntConstant(5))

	xElem := mem.ArrayElemRef(x, NewIntConstant(1))
	mem.AssignValue(xElem, NewIntConstant(2))

	cond := NewBinaryOperation(
		mem.GetValue(yElem),
		NewIntConstant(2),
		EQ,
	)

	tr.Assert(cond)
	for _, v := range mem.GetAliasingConstraints() {
		tr.Assert(v)
	}

	if _, sat := checkSat(t, tr); !sat {
		t.Fatal("expected SAT")
	}
}

func TestIntPreventAliasingScalar(t *testing.T) {
	tr := translator.NewZ3Translator()
	mem := memory.NewSymbolicMemory()

	x := mem.Allocate(IntExpr(), false)
	y := mem.Allocate(IntExpr(), false)

	mem.AssignValue(y, NewIntConstant(5))
	mem.AssignValue(x, NewIntConstant(2))

	valY := mem.GetValue(y)
	cond := NewBinaryOperation(
		valY,
		NewIntConstant(2),
		EQ,
	)

	tr.Assert(cond)
	for _, c := range mem.GetAliasingConstraints() {
		tr.Assert(c)
	}

	if _, sat := checkSat(t, tr); sat {
		t.Fatal("expected UNSAT")
	}
}

func TestIntAliasingScalar(t *testing.T) {
	tr := translator.NewZ3Translator()
	mem := memory.NewSymbolicMemory()

	x := mem.Allocate(IntExpr(), true)
	y := mem.Allocate(IntExpr(), true)

	mem.AssignValue(y, NewIntConstant(5))
	mem.AssignValue(x, NewIntConstant(2))

	valY := mem.GetValue(y)
	cond := NewBinaryOperation(
		valY,
		NewIntConstant(2),
		EQ,
	)

	tr.Assert(cond)
	for _, c := range mem.GetAliasingConstraints() {
		tr.Assert(c)
	}

	if m, sat := checkSat(t, tr); !sat {
		t.Fatal("expected SAT")
	} else if m != nil {
		t.Logf("model:\n%s", m.String())
	}
}

func TestParamStructAliasing(t *testing.T) {
	tr := translator.NewZ3Translator()
	mem := memory.NewSymbolicMemory()

	s1 := mem.Allocate(ExpressionType{
		ExprType:   StructType,
		Param:      nil,
		Name:       &[]string{"Foo"}[0],
		Fields:     &[]ExpressionType{BoolExpr()},
		FieldIndex: nil,
	}, true)
	s2 := mem.Allocate(ExpressionType{
		ExprType:   StructType,
		Param:      nil,
		Name:       &[]string{"Foo"}[0],
		Fields:     &[]ExpressionType{BoolExpr()},
		FieldIndex: nil,
	}, true)

	s1f0 := mem.FieldRef(s1, 0)
	mem.AssignValue(s1f0, NewBoolConstant(true))

	s2f0 := mem.FieldRef(s2, 0)
	mem.AssignValue(s2f0, NewBoolConstant(false))

	val := mem.GetValue(s1f0)
	cond := NewBinaryOperation(val, NewBoolConstant(false), EQ)
	tr.Assert(cond)

	for _, c := range mem.GetAliasingConstraints() {
		tr.Assert(c)
	}

	if _, sat := checkSat(t, tr); !sat {
		t.Fatal("should be UNSAT")
	}
}
