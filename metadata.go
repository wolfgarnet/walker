package walker

import (
	"fmt"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/file"
)

const (
	Vars      string = "vars"
	NodeField        = "node"
)

// Metadata contains information about a node.
// It is a map of values, by default the parent of the current node is inserted.
type Metadata map[string]interface{}

// NewMetadata returns a new instance
func NewMetadata(node ast.Node) Metadata {
	md := Metadata{NodeField: node}
	return md
}

// Parent retrieves the parent of the node
func (md Metadata) Node() ast.Node {
	parent, ok := md[NodeField].(ast.Node)
	if !ok {
		return nil
	}

	return parent
}

// AddParent inserts the given node as the parent
func (md Metadata) AddParent(parent ast.Node) {
	md[NodeField] = parent
}

// CurrentMetadata returns the last added element as the current metadata
func CurrentMetadata(metadata []Metadata) Metadata {
	l := len(metadata)
	if l == 0 {
		return nil
	}

	return metadata[l-1]
}

// ParentMetadata returns the second last added element as the parent metadata
func ParentMetadata(metadata []Metadata) Metadata {
	l := len(metadata)
	if l < 2 {
		return nil
	}

	return metadata[l-2]
}

// FindIthParentStatement can return a program, which is not a statement
func FindIthParentStatement(metadata []Metadata, i int) ast.Node {
	for j := len(metadata) - 1; j >= 0; j-- {
		parent := metadata[j][NodeField]
		statement, ok := parent.(ast.Statement)
		if ok {
			if i == 0 {
				return statement
			}

			i--
		}

		// Alternatively, check if it's a program
		program, isProgram := parent.(*ast.Program)
		if isProgram {
			if i == 0 {
				return program
			}

			i--
		}
	}

	return nil
}

func FindIthParentStatementMetadata(metadata []Metadata, i int) []Metadata {
	for j := len(metadata) - 1; j >= 0; j-- {
		parent := metadata[j][NodeField]
		_, ok := parent.(ast.Statement)
		if ok {
			if i == 0 {
				return metadata[:j+1]
			}

			i--
		}

		// Alternatively, check if it's a program
		_, isProgram := parent.(*ast.Program)
		if isProgram {
			if i == 0 {
				return metadata[:j+1]
			}

			i--
		}
	}

	return nil
}

func FindParentStatement(metadata []Metadata) ast.Statement {
	for i := len(metadata) - 1; i >= 0; i-- {
		parent := metadata[i][NodeField]
		statement, ok := parent.(ast.Statement)
		if ok {
			return statement
		}
	}

	return nil
}

func FindParentFunction(metadata []Metadata) *ast.FunctionLiteral {
	for i := len(metadata) - 1; i >= 0; i-- {
		parent := metadata[i][NodeField]
		statement, ok := parent.(*ast.FunctionLiteral)
		if ok {
			return statement
		}
	}

	return nil
}

// String displays information about the metadata
func (md Metadata) String() string {
	return fmt.Sprintf("{node:%T@%p}", md[NodeField], md[NodeField])
}

type Variables map[string]file.Idx

func NewVariables() Variables {
	return Variables{}
}
