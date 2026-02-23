package internal

import (
	"fmt"
	"go/constant"
	"go/token"
	"go/types"
	"symbolic-execution-course/internal/memory"
	"symbolic-execution-course/internal/symbolic"

	"golang.org/x/tools/go/ssa"
)

type Interpreter struct {
	CallStack     []CallStackFrame
	Analyser      *Analyser
	PathCondition symbolic.SymbolicExpression
	Heap          memory.Memory
	Cache         map[ssa.Value]symbolic.SymbolicExpression
	CurrentBlock  *ssa.BasicBlock
	PredBlock     *ssa.BasicBlock
	instrIndex    int
	visitCount    map[string]int
}

func NewInterpreter(analyser *Analyser, pathCond symbolic.SymbolicExpression) Interpreter {
	return Interpreter{
		CallStack:     make([]CallStackFrame, 0),
		Analyser:      analyser,
		PathCondition: pathCond,
		Heap:          memory.NewSymbolicMemory(),
		Cache:         make(map[ssa.Value]symbolic.SymbolicExpression),
		CurrentBlock:  nil,
		PredBlock:     nil,
		instrIndex:    0,
		visitCount:    make(map[string]int),
	}
}

type CallStackFrame struct {
	Function    *ssa.Function
	LocalMemory map[string]symbolic.SymbolicExpression
	ReturnValue symbolic.SymbolicExpression
}

func (interpreter *Interpreter) interpretDynamically(element ssa.Instruction) []*Interpreter {
	fmt.Printf("interpret: %s\n", element)

	switch i := element.(type) {
	case *ssa.Alloc, *ssa.BinOp, *ssa.Convert,
		*ssa.FieldAddr, *ssa.IndexAddr,
		*ssa.MakeSlice, *ssa.Slice, *ssa.UnOp:
		if interpreter.instrIndex+1 < len(interpreter.CurrentBlock.Instrs) {
			it := interpreter.clone()
			it.instrIndex = interpreter.instrIndex + 1
			return []*Interpreter{&it}
		}
		return make([]*Interpreter, 0, len(interpreter.CurrentBlock.Succs))
	case *ssa.If:
		cond := interpreter.resolveExpression(i.Cond)
		res := make([]*Interpreter, 0, 2)
		if len(i.Block().Succs) >= 1 {
			thenState := interpreter.clone()
			thenState.PathCondition = symbolic.NewBinaryOperation(
				thenState.PathCondition,
				cond,
				symbolic.AND,
			)
			thenState.instrIndex = 0
			thenState.CurrentBlock = interpreter.CurrentBlock.Succs[0]
			thenState.PredBlock = interpreter.CurrentBlock
			res = append(res, &thenState)
		}
		if len(i.Block().Succs) >= 2 {
			elseState := interpreter.clone()
			notCond := symbolic.NewUnaryOperation(cond, symbolic.NOT)
			elseState.PathCondition = symbolic.NewBinaryOperation(
				elseState.PathCondition,
				notCond,
				symbolic.AND,
			)
			elseState.instrIndex = 0
			elseState.CurrentBlock = interpreter.CurrentBlock.Succs[1]
			elseState.PredBlock = interpreter.CurrentBlock
			res = append(res, &elseState)
		}
		return res
	case *ssa.Jump:
		if len(interpreter.CurrentBlock.Succs) == 1 {
			newIt := interpreter.clone()
			newIt.instrIndex = 0
			newIt.CurrentBlock = interpreter.CurrentBlock.Succs[0]
			newIt.PredBlock = interpreter.CurrentBlock
			return []*Interpreter{&newIt}
		}
		return make([]*Interpreter, 0, len(interpreter.CurrentBlock.Succs))
	case *ssa.MakeInterface:
		interpreter.instrIndex = 0
		interpreter.CurrentBlock = nil
		return []*Interpreter{interpreter}
	case *ssa.Panic:
		interpreter.instrIndex = 0
		interpreter.CurrentBlock = nil
		return []*Interpreter{interpreter}
	case *ssa.Phi:
		if interpreter.PredBlock == nil {
			panic("pred is nil")
		}
		var edge *ssa.Value
		for idx, p := range interpreter.CurrentBlock.Preds {
			if p == interpreter.PredBlock {
				if idx < 0 || idx >= len(i.Edges) {
					panic("pred not found")
				}
				edge = &i.Edges[idx]
			}
		}
		if edge == nil {
			panic("pred not found)")
		}
		phiVal := interpreter.resolveExpression(*edge)
		interpreter.Cache[i] = phiVal
		if interpreter.instrIndex+1 < len(interpreter.CurrentBlock.Instrs) {
			newIt := interpreter.clone()
			newIt.instrIndex++
			return []*Interpreter{&newIt}
		}
		return make([]*Interpreter, 0, len(interpreter.CurrentBlock.Succs))
	case *ssa.Return:
		for _, r := range i.Results {
			interpreter.resolveExpression(r)
		}
		interpreter.instrIndex = -1
		interpreter.CurrentBlock = nil
		return []*Interpreter{interpreter}
	default:
		panic(fmt.Sprintf("not implemented: %v", element))
	}
}

func (interpreter *Interpreter) resolveExpression(value ssa.Value) symbolic.SymbolicExpression {
	if e, ok := interpreter.Cache[value]; ok {
		return e
	}
	switch v := value.(type) {
	case *ssa.BinOp:
		return symbolic.NewBinaryOperation(
			interpreter.resolveExpression(v.X),
			interpreter.resolveExpression(v.Y),
			tokenToBinaryOp(v.Op),
		)
	case *ssa.Const:
		switch v.Value.Kind() {
		case constant.Bool:
			return symbolic.NewBoolConstant(constant.BoolVal(v.Value))
		case constant.Int:
			if i, ok := constant.Int64Val(v.Value); ok {
				return symbolic.NewIntConstant(i)
			}
		case constant.String:
			return symbolic.NewStringConstant(constant.StringVal(v.Value))
		case constant.Float:
			if f, ok := constant.Float64Val(v.Value); ok {
				return symbolic.NewFloatConstant(f)
			}
		}
		panic(fmt.Sprintf("unsupported const: %v", v.Value.Kind()))
	case *ssa.Parameter:
		// TODO
		var ty symbolic.ExpressionType
		switch tt := v.Type().Underlying().(type) {
		case *types.Basic:
			if tt.Info()&types.IsBoolean != 0 {
				ty = symbolic.BoolExpr()
			} else if tt.Info()&types.IsInteger != 0 {
				ty = symbolic.IntExpr()
			} else if tt.Info()&types.IsString != 0 {
				ty = symbolic.StringExpr()
			} else if tt.Info()&types.IsFloat != 0 {
				ty = symbolic.FloatExpr()
			}
		default:
			panic(fmt.Sprintf("not implemented: %T %s", tt, v.Type().String()))
		}
		return interpreter.Heap.GetValue(interpreter.Heap.Allocate(ty, true))
	case *ssa.UnOp:
		return symbolic.NewUnaryOperation(interpreter.resolveExpression(v.X), tokenToUnaryOp(v.Op))
	default:
		panic(fmt.Sprintf("not implemented %T", v))
	}
}

func tokenToBinaryOp(op token.Token) symbolic.BinaryOperator {
	switch op {
	case token.ADD:
		return symbolic.ADD
	case token.SUB:
		return symbolic.SUB
	case token.MUL:
		return symbolic.MUL
	case token.QUO:
		return symbolic.DIV
	case token.REM:
		return symbolic.MOD
	case token.AND:
		return symbolic.AND
	case token.OR:
		return symbolic.OR
	case token.XOR:
		return symbolic.XOR
	case token.SHL:
		return symbolic.SHL
	case token.SHR:
		return symbolic.SHR
	case token.AND_NOT:
		return symbolic.AND_NOT
	case token.EQL:
		return symbolic.EQ
	case token.NEQ:
		return symbolic.NE
	case token.LSS:
		return symbolic.LT
	case token.LEQ:
		return symbolic.LE
	case token.GTR:
		return symbolic.GT
	case token.GEQ:
		return symbolic.GE
	default:
		panic("unsupported token: %v" + op.String())
	}
}

func tokenToUnaryOp(op token.Token) symbolic.UnaryOperator {
	switch op {
	case token.NOT:
		return symbolic.NOT
	case token.SUB:
		return symbolic.UNARY_MINUS
	case token.XOR:
		return symbolic.INVERT
	default:
		panic("not supported unary operator: " + op.String())
	}
}

func (interpreter *Interpreter) clone() Interpreter {
	newIt := *interpreter
	newIt.Cache = make(map[ssa.Value]symbolic.SymbolicExpression, len(interpreter.Cache))
	for k, v := range interpreter.Cache {
		newIt.Cache[k] = v
	}
	newIt.Heap = interpreter.Heap.Clone()
	return newIt
}
