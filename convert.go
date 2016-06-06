package walker

import "github.com/robertkrimen/otto/ast"

type ASTReWriter struct {
	VisitorImpl
	allBlocks bool
}

func NewASTReWriter(allBlocks bool) *ASTReWriter {
	return &ASTReWriter{
		allBlocks:allBlocks,
	}
}

// TODO
func (v *ASTReWriter) VisitCase(w *Walker, node *ast.CaseStatement, metadata []Metadata) Metadata {

	if node.Test != nil {
		w.Walk(node.Test, metadata)
	}
	for _, e := range node.Consequent {
		if e != nil {
			w.Walk(e, metadata)
		}
	}

	node2 := &ast.CaseStatement2{
		Test:node.Test,
	}

	switch {
	case len(node.Consequent) == 0:
		node2.Consequent = nil
	default:
		node2.Consequent = &ast.BlockStatement{
			List: node.Consequent,
		}
	}

	md := v.VisitorImpl.VisitCatch(w, node, metadata)
	md["case"] = node2
	return md
}

func (v *ASTReWriter) VisitSwitch(w *Walker, node *ast.SwitchStatement, metadata []Metadata) Metadata {
	w.Walk(node.Discriminant, metadata)
	for i, e := range node.Body {
		if e != nil {
			md := w.Walk(e, metadata)
			case2, ok := md.Get("case")
			if !ok {
				panic("Case statement not walked correctly")
			}

		}
	}

	return CurrentMetadata(metadata)
}

func (v *ASTReWriter) VisitCatch(w *Walker, node *ast.CatchStatement, metadata []Metadata) Metadata {

	_, isBlock := node.Body.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Body,
		}

		node.Body = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitCatch(w, node, metadata)
}

func (v *ASTReWriter) VisitDoWhile(w *Walker, node *ast.DoWhileStatement, metadata []Metadata) Metadata {

	_, isBlock := node.Body.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Body,
		}

		node.Body = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitDoWhile(w, node, metadata)
}

func (v *ASTReWriter) VisitFunction(w *Walker, node *ast.FunctionLiteral, metadata []Metadata) Metadata {

	_, isBlock := node.Body.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Body,
		}

		node.Body = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitFunction(w, node, metadata)
}

func (v *ASTReWriter) VisitForIn(w *Walker, node *ast.ForInStatement, metadata []Metadata) Metadata {

	_, isBlock := node.Body.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Body,
		}

		node.Body = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitForIn(w, node, metadata)
}

func (v *ASTReWriter) VisitFor(w *Walker, node *ast.ForStatement, metadata []Metadata) Metadata {

	_, isBlock := node.Body.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Body,
		}

		node.Body = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitFor(w, node, metadata)
}

// TODO
func (v *ASTReWriter) VisitIf(w *Walker, node *ast.IfStatement, metadata []Metadata) Metadata {

	if v.allBlocks {
		block1 := &ast.BlockStatement{
			List: node.Consequent,
		}
		block2 := &ast.BlockStatement{
			List: node.Alternate,
		}

		node.Consequent = []ast.Statement{block1}
		node.Alternate = []ast.Statement{block2}
	}

	return v.VisitorImpl.VisitIf(w, node, metadata)
}

func (v *ASTReWriter) VisitLabelled(w *Walker, node *ast.LabelledStatement, metadata []Metadata) Metadata {

	_, isBlock := node.Statement.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Statement,
		}

		node.Statement = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitLabelled(w, node, metadata)
}


// TODO
func (v *ASTReWriter) VisitProgram(w *Walker, node *ast.Program, metadata []Metadata) Metadata {

	block := &ast.BlockStatement{
		List: node.Body,
	}

	node.Body = []ast.Statement{block}

	return v.VisitorImpl.VisitProgram(w, node, metadata)
}

// TODO
func (v *ASTReWriter) VisitTry(w *Walker, node *ast.TryStatement, metadata []Metadata) Metadata {

	if v.allBlocks {
		block1 := &ast.BlockStatement{
			List: node.Body,
		}
		block2 := &ast.BlockStatement{
			List: node.Finally,
		}

		node.Body = []ast.Statement{block1}
		node.Finally = []ast.Statement{block2}
	}

	return v.VisitorImpl.VisitTry(w, node, metadata)
}

func (v *ASTReWriter) VisitWhile(w *Walker, node *ast.WhileStatement, metadata []Metadata) Metadata {

	_, isBlock := node.Body.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Body,
		}

		node.Body = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitWhile(w, node, metadata)
}

func (v *ASTReWriter) VisitWith(w *Walker, node *ast.WithStatement, metadata []Metadata) Metadata {

	_, isBlock := node.Body.(*ast.BlockStatement)
	if v.allBlocks && !isBlock {
		block := &ast.BlockStatement{
			List: node.Body,
		}

		node.Body = []ast.Statement{block}
	}

	return v.VisitorImpl.VisitProgram(w, node, metadata)
}

