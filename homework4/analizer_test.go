package main

import (
	"fmt"
	"os"
	. "symbolic-execution-course/internal"
	. "symbolic-execution-course/internal/ssa"
	"symbolic-execution-course/internal/translator"
	"testing"

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

func readSourceFile(t *testing.T, filepath string) string {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	return string(bytes)
}

func run(t *testing.T, functionName string) {
	source := readSourceFile(t, "examples/test_functions.go")
	ssaPkg, err2 := NewBuilder().ParseAndBuildSSAPkg(source)
	if err2 != nil {
		panic(err2)
	}

	analyser := Analyser{
		Package:      ssaPkg,
		StatesQueue:  make(PriorityQueue, 0),
		PathSelector: &RandomPathSelector{},
		Results:      make([]Interpreter, 0),
		Z3Translator: translator.NewZ3Translator(),
	}
	result, err := analyser.Analyse(functionName)
	if err != nil {
		t.Fatalf("Analysis failed: %v", err)
	}
	for _, interpreter := range result {
		fmt.Print("PathCondition:")
		fmt.Println(interpreter.PathCondition)
		analyser.Z3Translator.Assert(interpreter.PathCondition)
		cs := interpreter.Heap.GetAliasingConstraints()
		fmt.Println("Ограничения:")
		for _, c := range cs {
			fmt.Println(c)
			analyser.Z3Translator.Assert(c)
		}
		if m, sat := checkSat(t, analyser.Z3Translator); sat {
			fmt.Printf("Model:\n%s\n", m.String())
		}
		fmt.Println()
	}
}

func TestAnalyzeTest1(t *testing.T) {
	run(t, "test1")
}

func TestAnalyzeTest2(t *testing.T) {
	run(t, "test2")
}

func TestAnalyzeTestArithmetic(t *testing.T) {
	run(t, "testArithmetic")
}

func TestAnalyzeTestUnary(t *testing.T) {
	run(t, "testUnary")
}

func TestAnalyzeTestComparisons(t *testing.T) {
	run(t, "testComparisons")
}

func TestAnalyzeTestLogicalOps(t *testing.T) {
	run(t, "testLogicalOps")
}

func TestAnalyzeTestWhileLoop(t *testing.T) {
	run(t, "testWhileLoop")
}

func TestAnalyzeTestForLoop(t *testing.T) {
	run(t, "testForLoop")
}

func TestAnalyzeTestInfiniteLoopBreak(t *testing.T) {
	run(t, "testInfiniteLoopBreak")
}

func TestAnalyzeTestLoopWithConcreteBoundAndSymbolicBranching(t *testing.T) {
	run(t, "testLoopWithConcreteBoundAndSymbolicBranching")
}

func TestAnalyzeTestLoopWithSymbolicBoundAndSymbolicBranching(t *testing.T) {
	run(t, "testLoopWithSymbolicBoundAndSymbolicBranching")
}

func TestAnalyzeTestComplexConditions(t *testing.T) {
	run(t, "testComplexConditions")
}

func TestAnalyzeTestNestedIf(t *testing.T) {
	run(t, "testNestedIf")
}

func TestAnalyzeTestBitwise(t *testing.T) {
	run(t, "testBitwise")
}

func TestAnalyzeTestCombined(t *testing.T) {
	run(t, "testCombined")
}

func TestAnalyzeTestEdgeCases(t *testing.T) {
	run(t, "testEdgeCases")
}

func TestAnalyzeTestMultipleReturns(t *testing.T) {
	run(t, "testMultipleReturns")
}

func TestAnalyzeTestSimpleSum(t *testing.T) {
	run(t, "testSimpleSum")
}
