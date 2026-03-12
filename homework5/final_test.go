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

func readSourceFiles(t *testing.T, filepaths []string) []string {
	sources := make([]string, 0, len(filepaths))
	for _, filepath := range filepaths {
		bytes, err := os.ReadFile(filepath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		sources = append(sources, string(bytes))
	}
	return sources
}

func run(t *testing.T, functionName string, filepaths []string) {
	ssaPkg, err2 := NewBuilder().ParseAndBuildSSAPkg(readSourceFiles(t, filepaths))
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
	for i, interpreter := range result {
		analyser.Z3Translator = translator.NewZ3Translator()
		fmt.Printf("State №%d\n", i)
		fmt.Print("PathCondition: ")
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

func TestAnalyzeFactorialExample(t *testing.T) {
	run(t, "factorial", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestRecursive(t *testing.T) {
	run(t, "testRecursive", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestMutualRecursion(t *testing.T) {
	run(t, "testMutualRecursion", []string{"examples/test_functions.go"})
}

func TestAnalyzeIsEven(t *testing.T) {
	run(t, "isEven", []string{"examples/test_functions.go"})
}

func TestAnalyzeIsOdd(t *testing.T) {
	run(t, "isOdd", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestArrayOperations(t *testing.T) {
	run(t, "testArrayOperations", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestSliceDynamic(t *testing.T) {
	run(t, "testSliceDynamic", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestStructOperations(t *testing.T) {
	run(t, "testStructOperations", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestPointers(t *testing.T) {
	run(t, "testPointers", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestPointerParameters(t *testing.T) {
	run(t, "testPointerParameters", []string{"examples/test_functions.go"})
}

func TestAnalyzeModifyViaPointer(t *testing.T) {
	run(t, "modifyViaPointer", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestComplexLoop(t *testing.T) {
	run(t, "testComplexLoop", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestMultipleReturnValues(t *testing.T) {
	run(t, "testMultipleReturnValues", []string{"examples/test_functions.go"})
}

func TestAnalyzeMultipleReturns(t *testing.T) {
	run(t, "multipleReturns", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestInterface(t *testing.T) {
	run(t, "testInterface", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestErrorHandling(t *testing.T) {
	run(t, "testErrorHandling", []string{"examples/test_functions.go"})
}

func TestAnalyzeSafeDivide(t *testing.T) {
	run(t, "safeDivide", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestMixedConditions(t *testing.T) {
	run(t, "testMixedConditions", []string{"examples/test_functions.go"})
}

func TestAnalyzeTestTreeStructure(t *testing.T) {
	run(t, "testTreeStructure", []string{"examples/test_functions.go"})
}

func TestAnalyzeComprehensiveTest(t *testing.T) {
	run(t, "comprehensiveTest", []string{"examples/test_functions.go"})
}

func TestAnalyzeFactorial(t *testing.T) {
	run(t, "Factorial", []string{
		"../final_tests/recursion.go",
	})
}

// Aliasing tests
func TestAnalyzeAliasing(t *testing.T) {
	run(t, "Aliasing", []string{
		"../final_tests/aliasing.go",
	})
}

func TestAnalyzeArrayAliasing(t *testing.T) {
	run(t, "ArrayAliasing", []string{
		"../final_tests/aliasing.go",
	})
}

// Arrays tests
func TestAnalyzeDefaultBooleanValues(t *testing.T) {
	run(t, "DefaultBooleanValues", []string{
		"../final_tests/structs.go", "../final_tests/arrays.go",
	})
}

func TestAnalyzeByteArray(t *testing.T) {
	run(t, "ByteArray", []string{
		"../final_tests/structs.go", "../final_tests/arrays.go",
	})
}

func TestAnalyzeCharSizeAndIndex(t *testing.T) {
	run(t, "CharSizeAndIndex", []string{
		"../final_tests/structs.go", "../final_tests/arrays.go",
	})
}

func TestAnalyzeBooleanArray(t *testing.T) {
	run(t, "BooleanArray", []string{
		"../final_tests/structs.go", "../final_tests/arrays.go",
	})
}

func TestAnalyzeCreateArray(t *testing.T) {
	run(t, "CreateArray", []string{
		"../final_tests/arrays.go",
		"../final_tests/structs.go",
	})
}

func TestAnalyzeIsIdentityMatrix(t *testing.T) {
	run(t, "IsIdentityMatrix", []string{
		"../final_tests/structs.go", "../final_tests/arrays.go",
	})
}

func TestAnalyzeReallyMultiDimensionalArray(t *testing.T) {
	run(t, "ReallyMultiDimensionalArray", []string{
		"../final_tests/structs.go", "../final_tests/arrays.go",
	})
}

func TestAnalyzeFillMultiArrayWithArray(t *testing.T) {
	run(t, "FillMultiArrayWithArray", []string{
		"../final_tests/structs.go", "../final_tests/arrays.go",
	})
}

// Bit tests
func TestAnalyzeComplement(t *testing.T) {
	run(t, "Complement", []string{
		"../final_tests/bit.go",
	})
}

func TestAnalyzeXor(t *testing.T) {
	run(t, "Xor", []string{
		"../final_tests/bit.go",
	})
}

func TestAnalyzeOr(t *testing.T) {
	run(t, "Or", []string{
		"../final_tests/bit.go",
	})
}

func TestAnalyzeAnd(t *testing.T) {
	run(t, "And", []string{
		"../final_tests/bit.go",
	})
}

func TestAnalyzeBooleanNot(t *testing.T) {
	run(t, "BooleanNot", []string{
		"../final_tests/bit.go",
	})
}

func TestAnalyzeBooleanXorCompare(t *testing.T) {
	run(t, "BooleanXorCompare", []string{
		"../final_tests/bit.go",
	})
}

func TestAnalyzeShlWithBigLongShift(t *testing.T) {
	run(t, "ShlWithBigLongShift", []string{
		"../final_tests/bit.go",
	})
}

// Calls tests
func TestAnalyzeSimpleFormula(t *testing.T) {
	run(t, "SimpleFormula", []string{
		"../final_tests/calls.go",
	})
}

func TestAnalyzeCreateObjectFromValue(t *testing.T) {
	run(t, "CreateObjectFromValue", []string{
		"../final_tests/calls.go",
	})
}

func TestAnalyzeChangeObjectValueByMethod(t *testing.T) {
	run(t, "ChangeObjectValueByMethod", []string{
		"../final_tests/calls.go",
	})
}

func TestAnalyzeParticularValue(t *testing.T) {
	run(t, "ParticularValue", []string{
		"../final_tests/calls.go",
	})
}

func TestAnalyzeGetNullOrValue(t *testing.T) {
	run(t, "GetNullOrValue", []string{
		"../final_tests/calls.go",
	})
}

// Doubles tests
func TestAnalyzeCompareWithDiv(t *testing.T) {
	run(t, "CompareWithDiv", []string{
		"../final_tests/doubles.go",
	})
}

func TestAnalyzeMul(t *testing.T) {
	run(t, "Mul", []string{
		"../final_tests/doubles.go",
	})
}

// Loops tests
func TestAnalyzeLoopWithConcreteBound(t *testing.T) {
	run(t, "LoopWithConcreteBound", []string{
		"../final_tests/loops.go",
	})
}

func TestAnalyzeLoopWithSymbolicBound(t *testing.T) {
	run(t, "LoopWithSymbolicBound", []string{
		"../final_tests/loops.go",
	})
}

func TestAnalyzeLoopWithSymbolicBoundAndSymbolicBranching(t *testing.T) {
	run(t, "LoopWithSymbolicBoundAndSymbolicBranching", []string{
		"../final_tests/loops.go",
	})
}

func TestAnalyzeLoopWithSymbolicBoundAndComplexControlFlow(t *testing.T) {
	run(t, "LoopWithSymbolicBoundAndComplexControlFlow", []string{
		"../final_tests/loops.go",
	})
}

func TestAnalyzeWhileCycle(t *testing.T) {
	run(t, "WhileCycle", []string{
		"../final_tests/loops.go",
	})
}

func TestAnalyzeLoopInsideLoop(t *testing.T) {
	run(t, "LoopInsideLoop", []string{
		"../final_tests/loops.go",
	})
}

func TestAnalyzeMax(t *testing.T) {
	run(t, "Max", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeExample(t *testing.T) {
	run(t, "Example", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeCreateObject(t *testing.T) {
	run(t, "CreateObject", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeMemory(t *testing.T) {
	run(t, "Memory", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeCompareTwoNullObjects(t *testing.T) {
	run(t, "CompareTwoNullObjects", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeWriteToRefTypeField(t *testing.T) {
	run(t, "WriteToRefTypeField", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeWriteToArrayField(t *testing.T) {
	run(t, "WriteToArrayField", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeReadFromArrayField(t *testing.T) {
	run(t, "ReadFromArrayField", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeCompareTwoDifferentObjectsFromArguments(t *testing.T) {
	run(t, "CompareTwoDifferentObjectsFromArguments", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeNextValue(t *testing.T) {
	run(t, "NextValue", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeWriteObjectField(t *testing.T) {
	run(t, "WriteObjectField", []string{
		"../final_tests/structs.go",
	})
}

func TestAnalyzeTestPathConstraintMutability(t *testing.T) {
	run(t, "TestPathConstraintMutability", []string{
		"../final_tests/structs.go",
	})
}
