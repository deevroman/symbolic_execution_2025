package main

import (
	"fmt"
	"log"
	"os"
	"symbolic-execution-course/internal/ssa"
	"testing"
)

func TestReadExamples(t *testing.T) {
	bytes, err := os.ReadFile("examples/test_functions.go")
	source := string(bytes)
	builder := ssa.NewBuilder()
	_, err = builder.ParseAndBuildSSAPkg(source)
	if err != nil {
		log.Fatalf("Ошибка построения SSA: %v", err)
	}
}

func TestEmptyFun(t *testing.T) {
	bytes, err := os.ReadFile("examples/test_functions.go")
	source := string(bytes)
	builder := ssa.NewBuilder()
	_, err = builder.ParseAndBuildSSA(source, "main")
	if err != nil {
		log.Fatalf("Ошибка построения SSA: %v", err)
	}
}

func TestAllExamples(t *testing.T) {
	bytes, err := os.ReadFile("examples/test_functions.go")
	source := string(bytes)
	builder := ssa.NewBuilder()
	objects, err := builder.ParseAndBuildSSAPkg(source)
	if err != nil {
		log.Fatalf("Ошибка построения SSA: %v", err)
		return
	}
	for funcName := range objects.Members {
		fn := objects.Func(funcName)
		if fn == nil {
			fmt.Errorf("func %q not found", funcName)
		}
	}
	if err != nil {
		log.Fatalf("Ошибка построения SSA: %v", err)
	}
}
