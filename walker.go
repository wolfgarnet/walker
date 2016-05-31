package walker

import (
	"fmt"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/file"
	"runtime/debug"
)

type Hook struct {
	OnNode      func(node ast.Node, metadata []Metadata) error
	OnNodeLeave func(node ast.Node, metadata []Metadata) error

	OnFinished func(node ast.Node, metadata Metadata) error
}

// Walker can walk a given AST with a visitor
type Walker struct {
	Visitor         Visitor
	Current, Parent ast.Node
	CatchPanic      bool
	program         *ast.Program
}

func NewWalker(visitor Visitor) *Walker {
	return &Walker{
		Visitor:    visitor,
		CatchPanic: false,
		program:    nil,
	}
}

// Visitor interface for the walker.
type Visitor interface {
	VisitArray(walker *Walker, node *ast.ArrayLiteral, metadata []Metadata) Metadata
	VisitAssign(walker *Walker, node *ast.AssignExpression, metadata []Metadata) Metadata
	VisitBad(walker *Walker, node *ast.BadExpression, metadata []Metadata) Metadata
	VisitBadStatement(walker *Walker, node *ast.BadStatement, metadata []Metadata) Metadata
	VisitBinary(walker *Walker, node *ast.BinaryExpression, metadata []Metadata) Metadata
	VisitBlock(walker *Walker, node *ast.BlockStatement, metadata []Metadata) Metadata
	VisitBoolean(walker *Walker, node *ast.BooleanLiteral, metadata []Metadata) Metadata
	VisitBracket(walker *Walker, node *ast.BracketExpression, metadata []Metadata) Metadata
	VisitBranch(walker *Walker, node *ast.BranchStatement, metadata []Metadata) Metadata
	VisitCall(walker *Walker, node *ast.CallExpression, metadata []Metadata) Metadata
	VisitCase(walker *Walker, node *ast.CaseStatement, metadata []Metadata) Metadata
	VisitCase2(walker *Walker, node *ast.CaseStatement2, metadata []Metadata) Metadata
	VisitCatch(walker *Walker, node *ast.CatchStatement, metadata []Metadata) Metadata
	VisitConditional(walker *Walker, node *ast.ConditionalExpression, metadata []Metadata) Metadata
	VisitDebugger(walker *Walker, node *ast.DebuggerStatement, metadata []Metadata) Metadata
	VisitDot(walker *Walker, node *ast.DotExpression, metadata []Metadata) Metadata
	VisitDoWhile(walker *Walker, node *ast.DoWhileStatement, metadata []Metadata) Metadata
	VisitEmpty(walker *Walker, node *ast.EmptyExpression, metadata []Metadata) Metadata
	VisitEmptyStatement(walker *Walker, node *ast.EmptyStatement, metadata []Metadata) Metadata
	VisitExpression(walker *Walker, node *ast.ExpressionStatement, metadata []Metadata) Metadata
	VisitForIn(walker *Walker, node *ast.ForInStatement, metadata []Metadata) Metadata
	VisitFor(walker *Walker, node *ast.ForStatement, metadata []Metadata) Metadata
	VisitFunction(walker *Walker, node *ast.FunctionLiteral, metadata []Metadata) Metadata
	VisitFunctionStatement(walker *Walker, node *ast.FunctionStatement, metadata []Metadata) Metadata
	VisitIdentifier(walker *Walker, node *ast.Identifier, metadata []Metadata) Metadata
	VisitIf(walker *Walker, node *ast.IfStatement, metadata []Metadata) Metadata
	VisitLabelled(walker *Walker, node *ast.LabelledStatement, metadata []Metadata) Metadata
	VisitNew(walker *Walker, node *ast.NewExpression, metadata []Metadata) Metadata
	VisitNull(walker *Walker, node *ast.NullLiteral, metadata []Metadata) Metadata
	VisitNumber(walker *Walker, node *ast.NumberLiteral, metadata []Metadata) Metadata
	VisitObject(walker *Walker, node *ast.ObjectLiteral, metadata []Metadata) Metadata
	VisitProgram(walker *Walker, node *ast.Program, metadata []Metadata) Metadata
	VisitProgram2(walker *Walker, node *ast.Program2, metadata []Metadata) Metadata
	VisitReturn(walker *Walker, node *ast.ReturnStatement, metadata []Metadata) Metadata
	VisitRegex(walker *Walker, node *ast.RegExpLiteral, metadata []Metadata) Metadata
	VisitSequence(walker *Walker, node *ast.SequenceExpression, metadata []Metadata) Metadata
	VisitString(walker *Walker, node *ast.StringLiteral, metadata []Metadata) Metadata
	VisitSwitch(walker *Walker, node *ast.SwitchStatement, metadata []Metadata) Metadata
	VisitThis(walker *Walker, node *ast.ThisExpression, metadata []Metadata) Metadata
	VisitThrow(walker *Walker, node *ast.ThrowStatement, metadata []Metadata) Metadata
	VisitTry(walker *Walker, node *ast.TryStatement, metadata []Metadata) Metadata
	VisitUnary(walker *Walker, node *ast.UnaryExpression, metadata []Metadata) Metadata
	VisitVariable(walker *Walker, node *ast.VariableExpression, metadata []Metadata) Metadata
	VisitVariableStatement(walker *Walker, node *ast.VariableStatement, metadata []Metadata) Metadata
	VisitWhile(walker *Walker, node *ast.WhileStatement, metadata []Metadata) Metadata
	VisitWith(walker *Walker, node *ast.WithStatement, metadata []Metadata) Metadata

	getHooks() []*Hook
	AddHook(hook *Hook)
	ResetHooks()
}

func (w *Walker) GetPosition(idx file.Idx) *file.Position {
	if w.program == nil {
		return nil
	}

	return w.program.File.Position(idx)
}

// Begin the walk of the given AST node
func (w *Walker) Begin(node ast.Node) {
	if w.CatchPanic {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from %v\n", r)
				fmt.Printf("Panicked at node: %v\n", w.Current)
				fmt.Printf("Parent node is %v\n", w.Parent)

				program, isProgram := node.(*ast.Program)
				if isProgram {
					var pos *file.Position
					if w.Current != nil {
						pos = program.File.Position(w.Current.Idx0())
					} else if w.Parent != nil {
						pos = program.File.Position(w.Parent.Idx0())
					}
					if pos != nil {
						fmt.Printf("Position is %v\n", pos)
					} else {
						fmt.Printf("Unknown position!\n")
					}
				}
				fmt.Printf("%s\n", debug.Stack())
			}
		}()
	}
	md := []Metadata{NewMetadata(nil)}
	metadata := w.Walk(node, md)

	for _, hook := range w.Visitor.getHooks() {
		if hook.OnFinished != nil {
			hook.OnFinished(node, metadata)
		}
	}
}

// CollectScope collects information about the given scope
func CollectScope(metadata Metadata, declarations []ast.Declaration) {
	// Initialize the scope variables field in the metadata
	vars, ok := metadata[Vars].(Variables)
	if !ok {
		vars = NewVariables()
		metadata[Vars] = vars
	}

	for _, vd := range declarations {
		switch d := vd.(type) {
		case *ast.VariableDeclaration:
			for _, v := range d.List {
				vars[v.Name] = v.Idx
			}
		}
	}
}

func FindVariable(metadata []Metadata, name string) file.Idx {
	md := metadata[len(metadata)-1]

	vars, ok := md[Vars].(Variables)
	if ok {
		for v, i := range vars {
			if v == name {
				return i
			}
		}
	}

	if len(metadata) > 1 {
		metadata = metadata[:len(metadata)-1]
		return FindVariable(metadata, name)
	}

	return -1
}

// Walk the AST, including metadata
func (w *Walker) Walk(node ast.Node, metadata []Metadata) (result Metadata) {
	w.Current = node
	w.Parent = ParentMetadata(metadata).Node()

	// Create metadata for current node
	md := NewMetadata(node)

	// Scope things
	switch n := node.(type) {
	case *ast.Program:
		CollectScope(md, n.DeclarationList)
	case *ast.FunctionLiteral:
		CollectScope(md, n.DeclarationList)
	}

	// Append the node
	metadata = append(metadata, md)

	for _, hook := range w.Visitor.getHooks() {
		if hook.OnNode != nil {
			hook.OnNode(node, metadata)
		}
	}

	switch n := node.(type) {
	case *ast.ArrayLiteral:
		result = w.Visitor.VisitArray(w, n, metadata)
	case *ast.AssignExpression:
		result = w.Visitor.VisitAssign(w, n, metadata)
	case *ast.BadExpression:
		result = w.Visitor.VisitBad(w, n, metadata)
	case *ast.BadStatement:
		result = w.Visitor.VisitBadStatement(w, n, metadata)
	case *ast.BinaryExpression:
		result = w.Visitor.VisitBinary(w, n, metadata)
	case *ast.BlockStatement:
		result = w.Visitor.VisitBlock(w, n, metadata)
	case *ast.BooleanLiteral:
		result = w.Visitor.VisitBoolean(w, n, metadata)
	case *ast.BracketExpression:
		result = w.Visitor.VisitBracket(w, n, metadata)
	case *ast.BranchStatement:
		result = w.Visitor.VisitBranch(w, n, metadata)
	case *ast.CallExpression:
		result = w.Visitor.VisitCall(w, n, metadata)
	case *ast.CaseStatement:
		result = w.Visitor.VisitCase(w, n, metadata)
	case *ast.CatchStatement:
		result = w.Visitor.VisitCatch(w, n, metadata)
	case *ast.ConditionalExpression:
		result = w.Visitor.VisitConditional(w, n, metadata)
	case *ast.DebuggerStatement:
		result = w.Visitor.VisitDebugger(w, n, metadata)
	case *ast.DotExpression:
		result = w.Visitor.VisitDot(w, n, metadata)
	case *ast.DoWhileStatement:
		result = w.Visitor.VisitDoWhile(w, n, metadata)
	case *ast.EmptyExpression:
		result = w.Visitor.VisitEmpty(w, n, metadata)
	case *ast.EmptyStatement:
		result = w.Visitor.VisitEmptyStatement(w, n, metadata)
	case *ast.ExpressionStatement:
		result = w.Visitor.VisitExpression(w, n, metadata)
	case *ast.ForInStatement:
		result = w.Visitor.VisitForIn(w, n, metadata)
	case *ast.ForStatement:
		result = w.Visitor.VisitFor(w, n, metadata)
	case *ast.FunctionLiteral:
		result = w.Visitor.VisitFunction(w, n, metadata)
	case *ast.FunctionStatement:
		result = w.Visitor.VisitFunctionStatement(w, n, metadata)
	case *ast.Identifier:
		result = w.Visitor.VisitIdentifier(w, n, metadata)
	case *ast.IfStatement:
		result = w.Visitor.VisitIf(w, n, metadata)
	case *ast.LabelledStatement:
		result = w.Visitor.VisitLabelled(w, n, metadata)
	case *ast.NewExpression:
		result = w.Visitor.VisitNew(w, n, metadata)
	case *ast.NullLiteral:
		result = w.Visitor.VisitNull(w, n, metadata)
	case *ast.NumberLiteral:
		result = w.Visitor.VisitNumber(w, n, metadata)
	case *ast.ObjectLiteral:
		result = w.Visitor.VisitObject(w, n, metadata)
	case *ast.Program:
		w.program = n
		result = w.Visitor.VisitProgram(w, n, metadata)
	case *ast.ReturnStatement:
		result = w.Visitor.VisitReturn(w, n, metadata)
	case *ast.RegExpLiteral:
		result = w.Visitor.VisitRegex(w, n, metadata)
	case *ast.SequenceExpression:
		result = w.Visitor.VisitSequence(w, n, metadata)
	case *ast.StringLiteral:
		result = w.Visitor.VisitString(w, n, metadata)
	case *ast.SwitchStatement:
		result = w.Visitor.VisitSwitch(w, n, metadata)
	case *ast.ThisExpression:
		result = w.Visitor.VisitThis(w, n, metadata)
	case *ast.ThrowStatement:
		result = w.Visitor.VisitThrow(w, n, metadata)
	case *ast.TryStatement:
		result = w.Visitor.VisitTry(w, n, metadata)
	case *ast.UnaryExpression:
		result = w.Visitor.VisitUnary(w, n, metadata)
	case *ast.VariableExpression:
		result = w.Visitor.VisitVariable(w, n, metadata)
	case *ast.VariableStatement:
		result = w.Visitor.VisitVariableStatement(w, n, metadata)
	case *ast.WhileStatement:
		result = w.Visitor.VisitWhile(w, n, metadata)
	case *ast.WithStatement:
		result = w.Visitor.VisitWith(w, n, metadata)
	default:
		result = nil
	}

	for _, hook := range w.Visitor.getHooks() {
		if hook.OnNodeLeave != nil {
			hook.OnNodeLeave(node, metadata)
		}
	}

	return
}

// VisitorImpl is a default implementation of the Visitor interface
type VisitorImpl struct {
	Hooks []*Hook
}

// getHooks returns the hooks for this visitor
func (v *VisitorImpl) getHooks() []*Hook {
	return v.Hooks
}

func (v *VisitorImpl) AddHook(hook *Hook) {
	v.Hooks = append(v.Hooks, hook)
}

func (v *VisitorImpl) ResetHooks() {
	v.Hooks = nil
}

func (v *VisitorImpl) VisitProgram(w *Walker, node *ast.Program, metadata []Metadata) Metadata {
	for _, e := range node.Body {
		w.Walk(e, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitProgram2(w *Walker, node *ast.Program2, metadata []Metadata) Metadata {
	if node.Body != nil {
		w.Walk(node.Body, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitArray(w *Walker, node *ast.ArrayLiteral, metadata []Metadata) Metadata {
	for _, e := range node.Value {
		w.Walk(e, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitAssign(w *Walker, node *ast.AssignExpression, metadata []Metadata) Metadata {
	w.Walk(node.Left, metadata)
	w.Walk(node.Right, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitBad(w *Walker, node *ast.BadExpression, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitBadStatement(w *Walker, node *ast.BadStatement, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitBinary(w *Walker, node *ast.BinaryExpression, metadata []Metadata) Metadata {
	w.Walk(node.Left, metadata)
	w.Walk(node.Right, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitBlock(w *Walker, node *ast.BlockStatement, metadata []Metadata) Metadata {
	for _, value := range node.List {
		w.Walk(value, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitBoolean(w *Walker, node *ast.BooleanLiteral, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitBracket(w *Walker, node *ast.BracketExpression, metadata []Metadata) Metadata {
	w.Walk(node.Left, metadata)
	w.Walk(node.Member, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitBranch(w *Walker, node *ast.BranchStatement, metadata []Metadata) Metadata {
	if node.Label != nil {
		if node.Label != nil {
			w.Walk(node.Label, metadata)
		}
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitCall(w *Walker, node *ast.CallExpression, metadata []Metadata) Metadata {
	w.Walk(node.Callee, metadata)
	for _, value := range node.ArgumentList {
		w.Walk(value, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitCase(w *Walker, node *ast.CaseStatement, metadata []Metadata) Metadata {
	if node.Test != nil {
		w.Walk(node.Test, metadata)
	}
	for _, e := range node.Consequent {
		if e != nil {
			w.Walk(e, metadata)
		}
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitCase2(w *Walker, node *ast.CaseStatement2, metadata []Metadata) Metadata {
	if node.Test != nil {
		w.Walk(node.Test, metadata)
	}
	if node.Consequent != nil {
		w.Walk(node.Consequent, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitCatch(w *Walker, node *ast.CatchStatement, metadata []Metadata) Metadata {
	w.Walk(node.Parameter, metadata)
	w.Walk(node.Body, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitConditional(w *Walker, node *ast.ConditionalExpression, metadata []Metadata) Metadata {
	w.Walk(node.Test, metadata)
	w.Walk(node.Consequent, metadata)
	w.Walk(node.Alternate, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitDebugger(w *Walker, node *ast.DebuggerStatement, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitDot(w *Walker, node *ast.DotExpression, metadata []Metadata) Metadata {
	w.Walk(node.Left, metadata)
	w.Walk(node.Identifier, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitDoWhile(w *Walker, node *ast.DoWhileStatement, metadata []Metadata) Metadata {
	w.Walk(node.Test, metadata)
	w.Walk(node.Body, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitEmpty(w *Walker, node *ast.EmptyExpression, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitEmptyStatement(w *Walker, node *ast.EmptyStatement, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitExpression(w *Walker, node *ast.ExpressionStatement, metadata []Metadata) Metadata {
	w.Walk(node.Expression, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitForIn(w *Walker, node *ast.ForInStatement, metadata []Metadata) Metadata {
	w.Walk(node.Into, metadata)
	w.Walk(node.Source, metadata)
	w.Walk(node.Body, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitFor(w *Walker, node *ast.ForStatement, metadata []Metadata) Metadata {
	if node.Initializer != nil {
		w.Walk(node.Initializer, metadata)
	}
	if node.Test != nil {
		w.Walk(node.Test, metadata)
	}
	if node.Update != nil {
		w.Walk(node.Update, metadata)
	}
	w.Walk(node.Body, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitFunction(w *Walker, node *ast.FunctionLiteral, metadata []Metadata) Metadata {
	if node.Name != nil {
		w.Walk(node.Name, metadata)
	}
	for _, value := range node.ParameterList.List {
		w.Walk(value, metadata)
	}
	w.Walk(node.Body, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitFunctionStatement(w *Walker, node *ast.FunctionStatement, metadata []Metadata) Metadata {
	if node.Function != nil {
		w.Walk(node.Function, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitIdentifier(w *Walker, node *ast.Identifier, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitIf(w *Walker, node *ast.IfStatement, metadata []Metadata) Metadata {
	w.Walk(node.Test, metadata)
	w.Walk(node.Consequent, metadata)
	if node.Alternate != nil {
		w.Walk(node.Alternate, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitLabelled(w *Walker, node *ast.LabelledStatement, metadata []Metadata) Metadata {
	w.Walk(node.Label, metadata)
	w.Walk(node.Statement, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitNew(w *Walker, node *ast.NewExpression, metadata []Metadata) Metadata {
	w.Walk(node.Callee, metadata)
	for _, e := range node.ArgumentList {
		w.Walk(e, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitNull(w *Walker, node *ast.NullLiteral, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitNumber(w *Walker, node *ast.NumberLiteral, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitObject(w *Walker, node *ast.ObjectLiteral, metadata []Metadata) Metadata {
	for _, v := range node.Value {
		w.Walk(v.Value, metadata)
	}
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitReturn(w *Walker, node *ast.ReturnStatement, metadata []Metadata) Metadata {
	if node.Argument != nil {
		w.Walk(node.Argument, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitRegex(w *Walker, node *ast.RegExpLiteral, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitSequence(w *Walker, node *ast.SequenceExpression, metadata []Metadata) Metadata {
	if node.Sequence != nil {
		for _, e := range node.Sequence {
			w.Walk(e, metadata)
		}
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitString(w *Walker, node *ast.StringLiteral, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitSwitch(w *Walker, node *ast.SwitchStatement, metadata []Metadata) Metadata {
	w.Walk(node.Discriminant, metadata)
	for _, e := range node.Body {
		if e != nil {
			w.Walk(e, metadata)
		}
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitThis(w *Walker, node *ast.ThisExpression, metadata []Metadata) Metadata {
	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitThrow(w *Walker, node *ast.ThrowStatement, metadata []Metadata) Metadata {
	w.Walk(node.Argument, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitTry(w *Walker, node *ast.TryStatement, metadata []Metadata) Metadata {
	w.Walk(node.Body, metadata)
	if node.Catch != nil {
		w.Walk(node.Catch, metadata)
	}
	if node.Finally != nil {
		w.Walk(node.Finally, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitUnary(w *Walker, node *ast.UnaryExpression, metadata []Metadata) Metadata {
	w.Walk(node.Operand, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitVariable(w *Walker, node *ast.VariableExpression, metadata []Metadata) Metadata {
	if node.Initializer != nil {
		w.Walk(node.Initializer, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitVariableStatement(w *Walker, node *ast.VariableStatement, metadata []Metadata) Metadata {
	for _, e := range node.List {
		w.Walk(e, metadata)
	}

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitWhile(w *Walker, node *ast.WhileStatement, metadata []Metadata) Metadata {
	w.Walk(node.Test, metadata)
	w.Walk(node.Body, metadata)

	return CurrentMetadata(metadata)
}

func (v *VisitorImpl) VisitWith(w *Walker, node *ast.WithStatement, metadata []Metadata) Metadata {
	w.Walk(node.Object, metadata)
	w.Walk(node.Body, metadata)

	return CurrentMetadata(metadata)
}
