package main

import (
	"fmt"
	"log"
	. "symbolic-execution-course/internal"
	. "symbolic-execution-course/internal/ssa"
	"symbolic-execution-course/internal/translator"

	"github.com/ebukreev/go-z3/z3"
)

func main() {
	source := `
package main

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func testRecursive(x int) int {
	if x < 5 {
		return factorial(x)
	}
	return -1
}

func ShlWithBigLongShift(shift int64) int {
	if shift < 40 {
		return 1
	}
	if (0x77777777 << shift) == 0x77777770 {
		return 2
	}
	return 3
}

func intAliasing(x int, y int) int {
	if y == x {
		return 2
	}
	return 5
}

type ObjectWithPrimitivesClass struct {
	ValueByDefault int
	x, y           int
	ShortValue     int16
	Weight         float64
}

func Memory(objectExample *ObjectWithPrimitivesClass, value int) *ObjectWithPrimitivesClass {
	if value > 0 {
		objectExample.x = 1
		objectExample.y = 2
		objectExample.Weight = 1.2
	} else {
		objectExample.x = -1
		objectExample.y = -2
		objectExample.Weight = -1.2
	}
	return objectExample
}
`
	ssaPkg, err2 := NewBuilder().ParseAndBuildSSAPkg([]string{source})
	if err2 != nil {
		panic(err2)
	}

	functionsToAnalyze := []string{"testRecursive", "ShlWithBigLongShift", "intAliasing", "Memory"}

	for _, funcName := range functionsToAnalyze {
		fmt.Printf("========== Анализ функции: %s ==========\n", funcName)

		analyser := Analyser{
			Package:      ssaPkg,
			StatesQueue:  make(PriorityQueue, 0),
			PathSelector: &RandomPathSelector{},
			Results:      make([]Interpreter, 0),
		}
		result, err := analyser.Analyse(funcName)
		if err != nil {
			log.Printf("Ошибка при анализе функции %s %v", funcName, err)
			continue
		}
		for i, interpreter := range result {
			analyser.Z3Translator = translator.NewZ3Translator()
			fmt.Printf("№%d\n", i)
			fmt.Print("PathCondition: ")
			fmt.Println(interpreter.PathCondition)
			analyser.Z3Translator.Assert(interpreter.PathCondition)
			cs := interpreter.Heap.GetAliasingConstraints()
			fmt.Println("Ограничения:")
			for _, c := range cs {
				fmt.Println(c)
				analyser.Z3Translator.Assert(c)
			}
			if m, sat := check(analyser.Z3Translator); sat {
				fmt.Println()
				fmt.Printf("Model:\n%s\n", m.String())
			} else {
				fmt.Printf("unsasat")
			}
			fmt.Println()
		}
		fmt.Printf("========== Завершён анализ функции %s ==========\n\n", funcName)
	}
}

func check(tr *translator.Z3Translator) (*z3.Model, bool) {
	sat, err := tr.IsSat()
	if err != nil {
		log.Fatalf("SAT check error: %v", err)
	}
	if !sat {
		return nil, false
	}
	return tr.GetSolver().Model(), true
}
