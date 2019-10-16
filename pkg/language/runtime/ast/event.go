package ast

import (
	"github.com/dapperlabs/flow-go/pkg/language/runtime/common"
)

// EventDeclaration

type EventDeclaration struct {
	Identifier Identifier
	Parameters Parameters
	StartPos   Position
	EndPos     Position
}

func (e *EventDeclaration) StartPosition() Position {
	return e.StartPos
}

func (e *EventDeclaration) EndPosition() Position {
	return e.EndPos
}

func (e *EventDeclaration) Accept(visitor Visitor) Repr {
	return visitor.VisitEventDeclaration(e)
}

func (*EventDeclaration) isDeclaration() {}
func (*EventDeclaration) isStatement()   {}

func (e *EventDeclaration) DeclarationName() string {
	return e.Identifier.Identifier
}

func (e *EventDeclaration) DeclarationKind() common.DeclarationKind {
	return common.DeclarationKindEvent
}