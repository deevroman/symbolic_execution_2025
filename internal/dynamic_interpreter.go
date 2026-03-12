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
	CallStack         []CallStackFrame
	Analyser          *Analyser
	PathCondition     symbolic.SymbolicExpression
	Heap              memory.Memory
	Cache             map[ssa.Value]symbolic.SymbolicExpression
	CallResults       map[ssa.Value][]symbolic.SymbolicExpression
	Parameters        map[ssa.Value]*symbolic.Ref
	Returns           []symbolic.SymbolicExpression
	CurrentBlock      *ssa.BasicBlock
	PredBlock         *ssa.BasicBlock
	instrIndex        int
	visitCount        map[string]int
	structsTypesCache map[string]*symbolic.ExpressionType

	debugId     int
	predDebugId int
}

var interpreterCounter = 0

func nextDebugId() int {
	interpreterCounter++
	return interpreterCounter - 1
}

func NewInterpreter(analyser *Analyser, pathCond symbolic.SymbolicExpression) Interpreter {
	return Interpreter{
		CallStack:         make([]CallStackFrame, 0),
		Analyser:          analyser,
		PathCondition:     pathCond,
		Heap:              memory.NewSymbolicMemory(),
		Cache:             make(map[ssa.Value]symbolic.SymbolicExpression),
		CallResults:       make(map[ssa.Value][]symbolic.SymbolicExpression),
		CurrentBlock:      nil,
		PredBlock:         nil,
		instrIndex:        0,
		visitCount:        make(map[string]int),
		Parameters:        make(map[ssa.Value]*symbolic.Ref),
		structsTypesCache: make(map[string]*symbolic.ExpressionType),

		debugId:     nextDebugId(),
		predDebugId: -1,
	}
}

type CallStackFrame struct {
	Function         *ssa.Function
	LocalMemory      map[string]symbolic.SymbolicExpression
	ReturnValue      symbolic.SymbolicExpression
	CallerBlock      *ssa.BasicBlock
	CallerInstrIndex int
	CallerPredBlock  *ssa.BasicBlock
	CallInstr        ssa.Value
}

func (interpreter *Interpreter) interpretDynamically(element ssa.Instruction) []*Interpreter {
	//fmt.Printf("interpreter #%d\n", interpreter.debugId)
	fmt.Printf("interpret: %s\n", element)

	switch i := element.(type) {
	case *ssa.Alloc, *ssa.BinOp, *ssa.Convert,
		*ssa.Extract, *ssa.FieldAddr, *ssa.IndexAddr,
		*ssa.MakeSlice, *ssa.Slice, *ssa.UnOp:
		if interpreter.instrIndex+1 < len(interpreter.CurrentBlock.Instrs) {
			it := *interpreter
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
		rets := make([]symbolic.SymbolicExpression, len(i.Results))
		for i, r := range i.Results {
			rets[i] = interpreter.resolveExpression(r)
		}
		if len(interpreter.CallStack) > 0 {
			top := interpreter.GetCurrentFrame()
			interpreter.CallStack = interpreter.CallStack[:len(interpreter.CallStack)-1]

			if top.CallInstr != nil {
				retsCopy := make([]symbolic.SymbolicExpression, len(rets))
				copy(retsCopy, rets)
				interpreter.CallResults[top.CallInstr] = retsCopy
				if len(rets) == 0 {
					interpreter.Cache[top.CallInstr] = symbolic.NewNilConstant()
				} else {
					interpreter.Cache[top.CallInstr] = rets[0]
				}
			}

			newIt := interpreter.clone()
			newIt.instrIndex = -1
			if top.CallerBlock == nil {
				newIt.CurrentBlock = nil
			} else {
				newIt.CurrentBlock = top.CallerBlock
				newIt.PredBlock = top.CallerPredBlock
				newIt.instrIndex = top.CallerInstrIndex
			}
			return []*Interpreter{&newIt}
		}
		interpreter.Returns = rets
		interpreter.instrIndex = -1
		interpreter.CurrentBlock = nil
		return []*Interpreter{interpreter}
	case *ssa.Call:
		switch v := i.Call.Value.(type) {
		case *ssa.Builtin:
			interpreter.Cache[i] = interpreter.resolveExpression(i)
			return nextBlock(*interpreter)
		case *ssa.Function:
			if v == nil || len(v.Blocks) == 0 {
				interpreter.Cache[i] = symbolic.NewSymbolicVariable("call_"+i.String(), interpreter.toSymbolicType(i.Type()))
				return nextBlock(*interpreter)
			}

			frame := CallStackFrame{
				Function:         v,
				LocalMemory:      make(map[string]symbolic.SymbolicExpression),
				CallerBlock:      interpreter.CurrentBlock,
				CallerInstrIndex: interpreter.instrIndex + 1,
				CallerPredBlock:  interpreter.PredBlock,
				CallInstr:        i,
			}
			for pi, p := range v.Params {
				frame.LocalMemory[p.Name()] = interpreter.resolveExpression(i.Call.Args[pi])
			}
			interpreter.CallStack = append(interpreter.CallStack, frame)

			newIt := interpreter.clone()
			newIt.CurrentBlock = v.Blocks[0]
			newIt.PredBlock = nil
			newIt.instrIndex = 0
			return []*Interpreter{&newIt}
		default:
			return nextBlock(*interpreter)
		}
	case *ssa.Store:
		val := interpreter.resolveExpression(i.Val)
		switch addr := i.Addr.(type) {
		case *ssa.FieldAddr:
			interpreter.Heap.AssignValue(
				interpreter.Heap.FieldRef(interpreter.resolveExpression(addr.X).(*symbolic.Ref), addr.Field),
				val,
			)
		case *ssa.IndexAddr:
			r, ok := interpreter.resolveExpression(addr.X).(*symbolic.Ref)
			if !ok {
				panic("IndexAddr base not ref")
			}
			index := interpreter.resolveExpression(addr.Index)
			if index.Type().ExprType != symbolic.IntType {
				panic("Index must be int")
			}
			if r.RefType.ExprType != symbolic.ArrayType {
				inferred, ok := interpreter.inferArrayLikeRefType(addr.X.Type())
				if !ok {
					panic(fmt.Sprintf("elem %s is not array-like (type=%s), elem=%s",
						addr.X.String(), addr.X.Type().String(), r.RefType.String(),
					))
				}
				r.RefType = inferred
			}
			interpreter.Heap.AssignValue(interpreter.Heap.ArrayElemRef(r, index), val)
		default:
			interpreter.Heap.AssignValue(interpreter.resolveExpression(i.Addr).(*symbolic.Ref), val)
		}
		return nextBlock(*interpreter)
	default:
		panic(fmt.Sprintf("not implemented: %v", element))
	}
}

func nextBlock(interpreter Interpreter) []*Interpreter {
	block := interpreter.CurrentBlock
	if interpreter.instrIndex+1 < len(interpreter.CurrentBlock.Instrs) {
		interpreter.instrIndex++
		return []*Interpreter{&interpreter}
	}

	if len(block.Succs) == 1 {
		newIt := interpreter
		newIt.instrIndex = 0
		newIt.CurrentBlock = block.Succs[0]
		newIt.PredBlock = block
		return []*Interpreter{&newIt}
	}

	out := make([]*Interpreter, 0, len(block.Succs))
	for _, s := range block.Succs {
		newIt := interpreter.clone()
		newIt.instrIndex = 0
		newIt.CurrentBlock = s
		newIt.PredBlock = block
		out = append(out, &newIt)
	}
	return out
}

func isPtr(t types.Type) bool {
	_, ok := t.Underlying().(*types.Pointer)
	return ok
}

func isSlice(t types.Type) bool {
	_, ok := t.Underlying().(*types.Slice)
	return ok
}

func isArray(t types.Type) bool {
	_, ok := t.Underlying().(*types.Array)
	return ok
}

func sliceInfo(t types.Type) (int, types.Type) {
	t = unwindPtr(t)
	dim := 0
	for {
		s, ok := t.Underlying().(*types.Slice)
		if !ok {
			return dim, t
		}
		dim++
		t = s.Elem()
	}
}

func toStruct(t types.Type, interpreter *Interpreter) (*symbolic.ExpressionType, bool) {
	t = unwindPtr(t)

	typeName := t.String()
	if n, isNamed := t.(*types.Named); isNamed {
		typeName = n.String()
		t = n.Underlying()
	}

	s, isStruct := t.(*types.Struct)
	if !isStruct {
		return nil, false
	}

	fieldTypes := make([]symbolic.ExpressionType, 0, s.NumFields())
	for i := 0; i < s.NumFields(); i++ {
		fieldTypes = append(fieldTypes, interpreter.toSymbolicType(s.Field(i).Type()))
	}
	res := symbolic.StructExpr(typeName, fieldTypes)
	return &res, true
}

func unwindPtr(t types.Type) types.Type {
	res := t
	for {
		if p, ok := res.Underlying().(*types.Pointer); ok {
			res = p.Elem()
			continue
		}
		return res
	}
}

func (interpreter *Interpreter) inferArrayLikeRefType(t types.Type) (symbolic.ExpressionType, bool) {
	res := interpreter.toSymbolicType(unwindPtr(t))
	if res.ExprType != symbolic.ArrayType {
		return symbolic.ExpressionType{}, false
	}
	return res, true
}

func (interpreter *Interpreter) resolveExpression(value ssa.Value) symbolic.SymbolicExpression {
	if value == nil {
		panic(value)
	}
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
		if v.IsNil() {
			return symbolic.NewNilConstant()
		}
		if v.Value == nil {
			panic("???")
		}
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
		default:
			panic("unhandled const case")
		}
		panic(fmt.Sprintf("unsupported const: %v", v.Value.Kind()))
	case *ssa.Call:
		switch b := v.Call.Value.(type) {
		case *ssa.Builtin:
			switch b.Name() {
			case "len":
				arg := interpreter.resolveExpression(v.Call.Args[0])
				if ref, ok := arg.(*symbolic.Ref); ok {
					length := interpreter.Heap.GetArrayLength(ref)
					if length == nil {
						return symbolic.NewSymbolicVariable("len_result_"+arg.String(), symbolic.IntExpr())
					}
					return symbolic.NewIntConstant(int64(*length))
				}
				return symbolic.NewSymbolicVariable("len_result_"+arg.String(), symbolic.IntExpr())
			case "append":
				return interpreter.resolveExpression(v.Call.Args[0])
			case "println":
				return interpreter.resolveExpression(v.Call.Args[0])
			default:
				panic(fmt.Sprintf("unsupported function call: %v", b.Name()))
			}
		default:
			return symbolic.NewSymbolicVariable("call_"+v.String(), interpreter.toSymbolicType(v.Type()))
		}
	case *ssa.Parameter:
		if len(interpreter.CallStack) > 0 {
			if v, ok := interpreter.GetCurrentFrame().LocalMemory[v.Name()]; ok {
				return v
			}
		}
		if ref, ok := interpreter.Parameters[v]; ok {
			if isPtr(v.Type()) || isSlice(v.Type()) || isArray(v.Type()) {
				return ref
			}
			if _, isStruct := toStruct(v.Type(), interpreter); isStruct {
				return ref
			}
			return interpreter.Heap.GetValue(ref)
		}
		if s, ok := toStruct(v.Type(), interpreter); ok {
			r := interpreter.Heap.Allocate(*s, true)
			interpreter.Parameters[v] = r
			return r
		} else if isSlice(v.Type()) || isArray(v.Type()) {
			r := interpreter.Heap.Allocate(interpreter.toSymbolicType(v.Type()), true)
			interpreter.Parameters[v] = r
			return r
		} else if p, ok := v.Type().Underlying().(*types.Pointer); ok {
			r := interpreter.Heap.Allocate(interpreter.toSymbolicType(p.Elem()), true)
			interpreter.Parameters[v] = r
			return r
		} else {
			// TODO
			r := interpreter.Heap.Allocate(interpreter.toSymbolicType(v.Type()), true)
			interpreter.Parameters[v] = r
			return interpreter.Heap.GetValue(r)
		}
	case *ssa.Extract:
		if values, ok := interpreter.CallResults[v.Tuple]; ok {
			if v.Index < 0 || v.Index >= len(values) {
				panic(fmt.Sprintf("extract index out of range: idx=%d, values=%d", v.Index, len(values)))
			}
			interpreter.Cache[v] = values[v.Index]
			return values[v.Index]
		}
		if v.Index == 0 {
			first := interpreter.resolveExpression(v.Tuple)
			interpreter.Cache[v] = first
			return first
		}
		res := symbolic.NewSymbolicVariable("extract_"+v.String(), interpreter.toSymbolicType(v.Type()))
		interpreter.Cache[v] = res
		return res
	case *ssa.Phi:
		if cached, ok := interpreter.Cache[v]; ok {
			return cached
		}
		if len(v.Edges) == 0 {
			panic("phi has no edges")
		}
		edgeIndex := 0
		if interpreter.PredBlock != nil {
			found := false
			for idx, p := range v.Block().Preds {
				if p == interpreter.PredBlock {
					edgeIndex = idx
					found = true
					break
				}
			}
			if !found {
				panic("phi pred not found in resolveExpression")
			}
		}
		if edgeIndex < 0 || edgeIndex >= len(v.Edges) {
			panic("phi edge index out of range")
		}
		res := interpreter.resolveExpression(v.Edges[edgeIndex])
		interpreter.Cache[v] = res
		return res
	case *ssa.UnOp:
		if v.Op != token.MUL {
			return symbolic.NewUnaryOperation(interpreter.resolveExpression(v.X), tokenToUnaryOp(v.Op))
		}
		switch addr := v.X.(type) {
		case *ssa.FieldAddr:
			raw := interpreter.Heap.GetValue(interpreter.Heap.FieldRef(
				interpreter.resolveExpression(addr.X).(*symbolic.Ref),
				addr.Field,
			))
			t := addr.Type()
			if p, ok := t.Underlying().(*types.Pointer); ok {
				t = p.Elem()
			}
			if isSlice(t) {
				arrayType := interpreter.toSymbolicType(t)
				if raw.Type().ExprType == symbolic.NilType {
					raw = symbolic.NewIntConstant(0)
				}
				return symbolic.NewRefFromExpr(raw, arrayType)
			}
			if isPtr(t) {
				p := t.Underlying().(*types.Pointer)
				if raw.Type().ExprType == symbolic.NilType {
					return raw
				}
				if s, ok := toStruct(p.Elem(), interpreter); ok {
					return symbolic.NewRefFromExpr(raw, *s)
				}
				return symbolic.NewRefFromExpr(raw, interpreter.toSymbolicType(p.Elem()))
			}
			return raw
		case *ssa.IndexAddr:
			v := interpreter.resolveExpression(addr.X)
			br, ok := v.(*symbolic.Ref)
			if !ok {
				panic(fmt.Sprintf("IndexAddr in UnOp base not ref, %v", v))
			}

			index := interpreter.resolveExpression(addr.Index)
			if index.Type().ExprType != symbolic.IntType {
				panic("Index must be int")
			}
			if br.RefType.ExprType != symbolic.ArrayType {
				inferred, ok := interpreter.inferArrayLikeRefType(addr.X.Type())
				if !ok {
					panic(fmt.Sprintf("base %s is not array-like (type=%s), baseRefType=%s",
						addr.X.String(), addr.X.Type().String(), br.RefType.String(),
					))
				}
				br.RefType = inferred
			}
			r := interpreter.Heap.ArrayElemRef(br, index)
			if (r.RefType.ExprType == symbolic.StructType && r.RefType.FieldIndex == nil) || r.RefType.ExprType == symbolic.ArrayType || r.RefType.ExprType == symbolic.RefType {
				return r
			}
			return interpreter.Heap.GetValue(r)
		default:
			base := interpreter.resolveExpression(v.X)
			if ref, ok := base.(*symbolic.Ref); ok {
				if ref.RefType.ExprType == symbolic.StructType && ref.RefType.FieldIndex == nil {
					return ref
				}
				return interpreter.Heap.GetValue(ref)
			}
			if p, ok := v.Type().Underlying().(*types.Pointer); ok {
				if s, ok := toStruct(p.Elem(), interpreter); ok {
					return symbolic.NewRefFromExpr(base, *s)
				}
				return symbolic.NewRefFromExpr(base, interpreter.toSymbolicType(p.Elem()))
			}
			panic(fmt.Sprintf("UnOp * base is not ref, got %T (%v)", base, base))
		}
	case *ssa.Alloc:
		var r *symbolic.Ref
		if s, ok := toStruct(v.Type(), interpreter); ok {
			r = interpreter.Heap.Allocate(*s, false)
		} else if isSlice(v.Type()) {
			r = interpreter.Heap.Allocate(interpreter.toSymbolicType(v.Type()), false)
		} else if p, ok := v.Type().Underlying().(*types.Pointer); ok {
			if s, ok := toStruct(p.Elem(), interpreter); ok {
				r = interpreter.Heap.Allocate(*s, false)
			} else {
				r = interpreter.Heap.Allocate(interpreter.toSymbolicType(p.Elem()), false)
			}
		} else {
			r = interpreter.Heap.Allocate(interpreter.toSymbolicType(v.Type()), false)
		}
		interpreter.Cache[v] = r
		return r
	case *ssa.Slice:
		res := interpreter.resolveExpression(v.X)
		interpreter.Cache[v] = res
		return res
	case *ssa.MakeSlice:
		r := interpreter.Heap.Allocate(interpreter.toSymbolicType(v.Type()), false)
		interpreter.Cache[v] = r
		return r
	case *ssa.Convert:
		fromType := interpreter.resolveExpression(v.X)
		toType := interpreter.toSymbolicType(v.Type())
		if fromType.Type().ExprType != toType.ExprType {
			panic(fmt.Sprintf("Unsupported convert: %s -> %s (%s)", fromType.Type().String(), toType.String(), v.String()))
		}
		interpreter.Cache[v] = fromType
		return fromType
	case *ssa.FieldAddr:
		base := interpreter.resolveExpression(v.X)
		if ref, ok := base.(*symbolic.Ref); ok {
			return interpreter.Heap.FieldRef(ref, v.Field)
		}
		return interpreter.Heap.Allocate(interpreter.toSymbolicType(v.Type()), false)
	default:
		panic(fmt.Sprintf("not implemented %T", v))
	}
}

func (interpreter *Interpreter) GetCurrentFrame() *CallStackFrame {
	if len(interpreter.CallStack) == 0 {
		return nil
	}
	return &interpreter.CallStack[len(interpreter.CallStack)-1]
}

func (interpreter *Interpreter) toSymbolicType(t types.Type) symbolic.ExpressionType {
	switch tu := t.Underlying().(type) {
	case *types.Basic:
		if tu.Info()&types.IsBoolean != 0 {
			return symbolic.BoolExpr()
		} else if tu.Info()&types.IsInteger != 0 {
			return symbolic.IntExpr()
		} else if tu.Info()&types.IsString != 0 {
			return symbolic.StringExpr()
		} else if tu.Info()&types.IsFloat != 0 {
			return symbolic.FloatExpr()
		} else {
			panic(fmt.Sprintf("unsupported type: %v", tu))
		}
	case *types.Pointer:
		name := tu.Elem().String()
		if s, ok := interpreter.structsTypesCache[name]; ok {
			return symbolic.ExpressionType{
				ExprType: symbolic.RefType,
				Param:    s,
				Name:     &name,
			}
		}
		elemType := interpreter.toSymbolicType(tu.Elem())
		return symbolic.ExpressionType{
			ExprType: symbolic.RefType,
			Param:    &elemType,
			Name:     &name,
		}
	case *types.Struct:
		fields := make([]symbolic.ExpressionType, 0, tu.NumFields())
		newType := symbolic.StructExpr(t.String(), fields)
		interpreter.structsTypesCache[t.String()] = &newType
		for i := 0; i < tu.NumFields(); i++ {
			fields = append(fields, interpreter.toSymbolicType(tu.Field(i).Type()))
		}
		return newType
	case *types.Interface, *types.Named:
		return symbolic.StructExpr(tu.String(), make([]symbolic.ExpressionType, 0))
	case *types.Array:
		return symbolic.ArrayExpr(interpreter.toSymbolicType(tu.Elem()), 1)
	case *types.Slice:
		return symbolic.ArrayExpr(interpreter.toSymbolicType(tu.Elem()), 1)
	default:
		panic(fmt.Sprintf("not implemented: %T %s", tu, t.String()))
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
	//fmt.Println("Cloning state...")
	newIt := *interpreter
	newIt.predDebugId = interpreter.debugId
	newIt.debugId = nextDebugId()

	newIt.Cache = make(map[ssa.Value]symbolic.SymbolicExpression, len(interpreter.Cache))
	for k, v := range interpreter.Cache {
		newIt.Cache[k] = v
	}

	newIt.Parameters = make(map[ssa.Value]*symbolic.Ref, len(interpreter.Parameters))
	for k, v := range interpreter.Parameters {
		newIt.Parameters[k] = v
	}

	newIt.CallResults = make(map[ssa.Value][]symbolic.SymbolicExpression, len(interpreter.CallResults))
	for k, values := range interpreter.CallResults {
		cp := make([]symbolic.SymbolicExpression, len(values))
		copy(cp, values)
		newIt.CallResults[k] = cp
	}

	newIt.Returns = make([]symbolic.SymbolicExpression, len(interpreter.Returns))
	for k, v := range interpreter.Returns {
		newIt.Returns[k] = v
	}

	newIt.Heap = interpreter.Heap.Clone()
	//fmt.Printf("\tpred: %d, new: %d\n", newIt.predDebugId, newIt.debugId)
	return newIt
}
