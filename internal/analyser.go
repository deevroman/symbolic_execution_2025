package internal

import (
	"container/heap"
	"fmt"
	"maps"
	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"

	"golang.org/x/tools/go/ssa"
)

type Analyser struct {
	Package      *ssa.Package
	StatesQueue  PriorityQueue
	PathSelector PathSelector
	Results      []Interpreter
	Z3Translator *translator.Z3Translator
}

func (analyser *Analyser) Analyse(functionName string) ([]Interpreter, error) {
	fn := analyser.Package.Func(functionName)
	if fn == nil {
		return nil, fmt.Errorf("function %s not found in package", functionName)
	}

	q := make(PriorityQueue, 0)
	heap.Init(&q)
	initState := NewInterpreter(analyser, symbolic.NewBoolConstant(true))
	initState.CurrentBlock = fn.Blocks[0]
	heap.Push(&q, &Item{
		value:    initState,
		priority: analyser.PathSelector.CalculatePriority(initState),
	})
	results := make([]Interpreter, 0)
	for q.Len() > 0 {
		it := heap.Pop(&q).(*Item)
		if it.value.CurrentBlock == nil {
			results = append(results, it.value)
			continue
		}
		location := fmt.Sprintf("%d_%d", it.value.CurrentBlock.Index, it.value.instrIndex)
		if it.value.visitCount[location] >= 5 {
			continue
		}
		prefVisits := maps.Clone(it.value.visitCount)
		prefVisits[location] = it.value.visitCount[location] + 1

		nextInter := (&it.value).interpretDynamically(it.value.CurrentBlock.Instrs[it.value.instrIndex])
		for _, ni := range nextInter {
			ni.visitCount = maps.Clone(prefVisits)
			heap.Push(&q, &Item{
				value:    *ni,
				priority: analyser.PathSelector.CalculatePriority(*ni),
				index:    0,
			})
		}
	}
	return results, nil
}
