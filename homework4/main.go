package main

import (
	"fmt"
	"log"

	. "symbolic-execution-course/internal"
	. "symbolic-execution-course/internal/ssa"
	"symbolic-execution-course/internal/translator"
)

func main() {
	source := `
package main

func testFunction(x int) int {
	if x > 0 {
		return x * 2
	} else {
		return x * -1
	}
}
`
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
	result, err := analyser.Analyse("testFunction")
	if err != nil {
		log.Fatal(err)
		return
	}
	for i, interpreter := range result {
		fmt.Printf("State №%d\n", i)
		fmt.Print("PathCondition:")
		fmt.Println(interpreter.PathCondition)
		analyser.Z3Translator.Assert(interpreter.PathCondition)
		cs := interpreter.Heap.GetAliasingConstraints()
		fmt.Println("Ограничения:")
		for _, c := range cs {
			fmt.Println(c)
			analyser.Z3Translator.Assert(c)
		}
		fmt.Println()
	}
}
