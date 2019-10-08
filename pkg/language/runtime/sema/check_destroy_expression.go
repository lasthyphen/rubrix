package sema

import (
	"github.com/dapperlabs/flow-go/pkg/language/runtime/ast"
	"github.com/dapperlabs/flow-go/pkg/language/runtime/common"
)

func (checker *Checker) VisitDestroyExpression(expression *ast.DestroyExpression) (resultType ast.Repr) {
	resultType = &VoidType{}

	valueType := expression.Expression.Accept(checker).(Type)

	// NOTE: not using `isResourceType`,
	// as only direct resource types can be destructed
	if compositeType, ok := valueType.(*CompositeType); !ok ||
		compositeType.Kind != common.CompositeKindResource {

		checker.report(
			&InvalidDestructionError{
				StartPos: expression.Expression.StartPosition(),
				EndPos:   expression.Expression.EndPosition(),
			},
		)

		return
	}

	checker.recordResourceInvalidation(
		expression.Expression,
		valueType,
		ResourceInvalidationKindDestroy,
	)

	return
}