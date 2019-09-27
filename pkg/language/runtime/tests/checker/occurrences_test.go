package checker

import (
	"github.com/dapperlabs/flow-go/pkg/language/runtime/common"
	"github.com/dapperlabs/flow-go/pkg/language/runtime/sema"
	. "github.com/dapperlabs/flow-go/pkg/language/runtime/tests/utils"
	. "github.com/onsi/gomega"
	"testing"
)

// TODO: implement occurrences for type references
//  (e.g. conformances, conditional casting expression)

func TestCheckOccurrencesVariableDeclarations(t *testing.T) {
	RegisterTestingT(t)

	checker, err := ParseAndCheck(`
        let x = 1
        var y = x
    `)

	Expect(err).
		To(Not(HaveOccurred()))

	occurrences := checker.Occurrences.All()

	matchers := []*OccurrenceMatcher{
		{
			StartPos:        sema.Position{Line: 2, Column: 12},
			EndPos:          sema.Position{Line: 2, Column: 12},
			OriginStartPos:  &sema.Position{Line: 2, Column: 12},
			OriginEndPos:    &sema.Position{Line: 2, Column: 12},
			DeclarationKind: common.DeclarationKindConstant,
		},
		{
			StartPos:        sema.Position{Line: 3, Column: 12},
			EndPos:          sema.Position{Line: 3, Column: 12},
			OriginStartPos:  &sema.Position{Line: 3, Column: 12},
			OriginEndPos:    &sema.Position{Line: 3, Column: 12},
			DeclarationKind: common.DeclarationKindVariable,
		},
		{
			StartPos:        sema.Position{Line: 3, Column: 16},
			EndPos:          sema.Position{Line: 3, Column: 16},
			OriginStartPos:  &sema.Position{Line: 2, Column: 12},
			OriginEndPos:    &sema.Position{Line: 2, Column: 12},
			DeclarationKind: common.DeclarationKindConstant,
		},
	}

	ms := make([]interface{}, len(matchers))
	for i := range matchers {
		ms[i] = matchers[i]
	}

	Expect(occurrences).
		To(ConsistOf(ms...))

	for _, matcher := range matchers {
		Expect(checker.Occurrences.Find(matcher.StartPos)).To(Not(BeNil()))
		Expect(checker.Occurrences.Find(matcher.EndPos)).To(Not(BeNil()))
	}
}

func TestCheckOccurrencesFunction(t *testing.T) {
	RegisterTestingT(t)

	checker, err := ParseAndCheck(`
		fun f1(paramX: Int, paramY: Bool) {
		   let x = 1
		   var y: Int? = x
		   fun f2() {
		       if let y = y {
		       }
		   }
           f1(paramX: 1, paramY: true)
		}

        fun f3() {
            f1(paramX: 2, paramY: false)
        }
	`)

	Expect(err).
		To(Not(HaveOccurred()))

	occurrences := checker.Occurrences.All()

	matchers := []*OccurrenceMatcher{
		{
			StartPos:        sema.Position{Line: 2, Column: 6},
			EndPos:          sema.Position{Line: 2, Column: 7},
			OriginStartPos:  &sema.Position{Line: 2, Column: 6},
			OriginEndPos:    &sema.Position{Line: 2, Column: 6},
			DeclarationKind: common.DeclarationKindFunction,
		},
		{
			StartPos:        sema.Position{Line: 2, Column: 9},
			EndPos:          sema.Position{Line: 2, Column: 14},
			OriginStartPos:  &sema.Position{Line: 2, Column: 9},
			OriginEndPos:    &sema.Position{Line: 2, Column: 9},
			DeclarationKind: common.DeclarationKindParameter,
		},
		{
			StartPos:        sema.Position{Line: 2, Column: 22},
			EndPos:          sema.Position{Line: 2, Column: 27},
			OriginStartPos:  &sema.Position{Line: 2, Column: 22},
			OriginEndPos:    &sema.Position{Line: 2, Column: 22},
			DeclarationKind: common.DeclarationKindParameter,
		},
		{
			StartPos:        sema.Position{Line: 3, Column: 9},
			EndPos:          sema.Position{Line: 3, Column: 9},
			OriginStartPos:  &sema.Position{Line: 3, Column: 9},
			OriginEndPos:    &sema.Position{Line: 3, Column: 9},
			DeclarationKind: common.DeclarationKindConstant,
		},
		{
			StartPos:        sema.Position{Line: 4, Column: 9},
			EndPos:          sema.Position{Line: 4, Column: 9},
			OriginStartPos:  &sema.Position{Line: 4, Column: 9},
			OriginEndPos:    &sema.Position{Line: 4, Column: 9},
			DeclarationKind: common.DeclarationKindVariable,
		},
		{
			StartPos:        sema.Position{Line: 4, Column: 19},
			EndPos:          sema.Position{Line: 4, Column: 19},
			OriginStartPos:  &sema.Position{Line: 3, Column: 9},
			OriginEndPos:    &sema.Position{Line: 3, Column: 9},
			DeclarationKind: common.DeclarationKindConstant,
		},
		{
			StartPos:        sema.Position{Line: 5, Column: 9},
			EndPos:          sema.Position{Line: 5, Column: 10},
			OriginStartPos:  &sema.Position{Line: 5, Column: 9},
			OriginEndPos:    &sema.Position{Line: 5, Column: 9},
			DeclarationKind: common.DeclarationKindFunction,
		},
		{
			StartPos:        sema.Position{Line: 6, Column: 16},
			EndPos:          sema.Position{Line: 6, Column: 16},
			OriginStartPos:  &sema.Position{Line: 6, Column: 16},
			OriginEndPos:    &sema.Position{Line: 6, Column: 16},
			DeclarationKind: common.DeclarationKindConstant,
		},
		{
			StartPos:        sema.Position{Line: 6, Column: 20},
			EndPos:          sema.Position{Line: 6, Column: 20},
			OriginStartPos:  &sema.Position{Line: 4, Column: 9},
			OriginEndPos:    &sema.Position{Line: 4, Column: 9},
			DeclarationKind: common.DeclarationKindVariable,
		},
		{
			StartPos:        sema.Position{Line: 9, Column: 11},
			EndPos:          sema.Position{Line: 9, Column: 12},
			OriginStartPos:  &sema.Position{Line: 2, Column: 6},
			OriginEndPos:    &sema.Position{Line: 2, Column: 6},
			DeclarationKind: common.DeclarationKindFunction,
		},
		{
			StartPos:        sema.Position{Line: 12, Column: 12},
			EndPos:          sema.Position{Line: 12, Column: 13},
			OriginStartPos:  &sema.Position{Line: 12, Column: 12},
			OriginEndPos:    &sema.Position{Line: 12, Column: 12},
			DeclarationKind: common.DeclarationKindFunction,
		},
		{
			StartPos:        sema.Position{Line: 13, Column: 12},
			EndPos:          sema.Position{Line: 13, Column: 13},
			OriginStartPos:  &sema.Position{Line: 2, Column: 6},
			OriginEndPos:    &sema.Position{Line: 2, Column: 6},
			DeclarationKind: common.DeclarationKindFunction,
		},
	}

	ms := make([]interface{}, len(matchers))
	for i := range matchers {
		ms[i] = matchers[i]
	}

	Expect(occurrences).
		To(ConsistOf(ms...))

	for _, matcher := range matchers {
		Expect(checker.Occurrences.Find(matcher.StartPos)).To(Not(BeNil()))
		Expect(checker.Occurrences.Find(matcher.EndPos)).To(Not(BeNil()))
	}
}

func TestCheckOccurrencesStructAndInterface(t *testing.T) {
	RegisterTestingT(t)

	checker, err := ParseAndCheck(`
		struct interface I1 {}

	    struct S1: I1 {
	       let x: Int
	       init() {
	          self.x = 1
	          self.test()
	       }
	       fun test() {}
	    }

	    fun f(): S1 {
	       return S1()
	    }
	`)

	Expect(err).
		To(Not(HaveOccurred()))

	occurrences := checker.Occurrences.All()

	matchers := []*OccurrenceMatcher{
		{
			StartPos:        sema.Position{Line: 2, Column: 19},
			EndPos:          sema.Position{Line: 2, Column: 20},
			OriginStartPos:  &sema.Position{Line: 2, Column: 19},
			OriginEndPos:    &sema.Position{Line: 2, Column: 19},
			DeclarationKind: common.DeclarationKindStructureInterface,
		},
		{
			StartPos:        sema.Position{Line: 4, Column: 12},
			EndPos:          sema.Position{Line: 4, Column: 13},
			OriginStartPos:  &sema.Position{Line: 4, Column: 12},
			OriginEndPos:    &sema.Position{Line: 4, Column: 12},
			DeclarationKind: common.DeclarationKindStructure,
		},
		{
			StartPos:        sema.Position{Line: 5, Column: 8},
			EndPos:          sema.Position{Line: 5, Column: 17},
			OriginStartPos:  &sema.Position{Line: 5, Column: 12},
			OriginEndPos:    &sema.Position{Line: 5, Column: 12},
			DeclarationKind: common.DeclarationKindField,
		},
		// self
		{
			StartPos:        sema.Position{Line: 7, Column: 11},
			EndPos:          sema.Position{Line: 7, Column: 14},
			OriginStartPos:  nil,
			OriginEndPos:    nil,
			DeclarationKind: common.DeclarationKindConstant,
		},
		{
			StartPos:        sema.Position{Line: 7, Column: 16},
			EndPos:          sema.Position{Line: 7, Column: 16},
			OriginStartPos:  &sema.Position{Line: 5, Column: 12},
			OriginEndPos:    &sema.Position{Line: 5, Column: 12},
			DeclarationKind: common.DeclarationKindField,
		},
		// self
		{
			StartPos:        sema.Position{Line: 8, Column: 11},
			EndPos:          sema.Position{Line: 8, Column: 14},
			OriginStartPos:  nil,
			OriginEndPos:    nil,
			DeclarationKind: common.DeclarationKindConstant,
		},
		{
			StartPos:        sema.Position{Line: 8, Column: 16},
			EndPos:          sema.Position{Line: 8, Column: 19},
			OriginStartPos:  &sema.Position{Line: 10, Column: 12},
			OriginEndPos:    &sema.Position{Line: 10, Column: 15},
			DeclarationKind: common.DeclarationKindFunction,
		},
		{
			StartPos:        sema.Position{Line: 10, Column: 12},
			EndPos:          sema.Position{Line: 10, Column: 15},
			OriginStartPos:  &sema.Position{Line: 10, Column: 12},
			OriginEndPos:    &sema.Position{Line: 10, Column: 12},
			DeclarationKind: common.DeclarationKindFunction,
		},
		// TODO: why the duplicate?
		{
			StartPos:        sema.Position{Line: 10, Column: 12},
			EndPos:          sema.Position{Line: 10, Column: 15},
			OriginStartPos:  &sema.Position{Line: 10, Column: 12},
			OriginEndPos:    &sema.Position{Line: 10, Column: 15},
			DeclarationKind: common.DeclarationKindFunction,
		},
		{
			StartPos:        sema.Position{Line: 13, Column: 9},
			EndPos:          sema.Position{Line: 13, Column: 9},
			OriginStartPos:  &sema.Position{Line: 13, Column: 9},
			OriginEndPos:    &sema.Position{Line: 13, Column: 9},
			DeclarationKind: common.DeclarationKindFunction,
		},
		{
			StartPos:        sema.Position{Line: 14, Column: 15},
			EndPos:          sema.Position{Line: 14, Column: 16},
			OriginStartPos:  &sema.Position{Line: 4, Column: 12},
			OriginEndPos:    &sema.Position{Line: 4, Column: 12},
			DeclarationKind: common.DeclarationKindFunction,
		},
	}

	ms := make([]interface{}, len(matchers))
	for i := range matchers {
		ms[i] = matchers[i]
	}

	Expect(occurrences).
		To(ConsistOf(ms...))

	for _, matcher := range matchers {
		Expect(checker.Occurrences.Find(matcher.StartPos)).To(Not(BeNil()))
		Expect(checker.Occurrences.Find(matcher.EndPos)).To(Not(BeNil()))
	}
}