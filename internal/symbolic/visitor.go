package symbolic

// Visitor интерфейс для обхода символьных выражений (Visitor Pattern)
type Visitor interface {
	VisitVariable(expr *SymbolicVariable) interface{}
	VisitIntConstant(expr *IntConstant) interface{}
	VisitBoolConstant(expr *BoolConstant) interface{}
	VisitFloatConstant(expr *FloatConstant) interface{}
	VisitStringConstant(expr *StringConstant) interface{}
	VisitNilConstant(expr *NilConstant) interface{}
	VisitUnaryOperation(expr *UnaryOperation) interface{}
	VisitBinaryOperation(expr *BinaryOperation) interface{}
	VisitLogicalOperation(expr *LogicalOperation) interface{}
	VisitConditionalExpression(expr *ConditionalExpression) interface{}
	VisitFunction(expr *Function) interface{}
	VisitFunctionCall(expr *FunctionCall) interface{}
	VisitArray(expr *SymbolicArray) interface{}
	VisitRef(expr *Ref) interface{}
}
