package ast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgram_ResolveImports(t *testing.T) {

	makeImportingProgram := func(imported string) *Program {
		return &Program{
			Declarations: []Declaration{
				&ImportDeclaration{
					Location: StringLocation(imported),
				},
			},
		}
	}

	a := makeImportingProgram("b")
	b := makeImportingProgram("c")
	c := &Program{}

	err := a.ResolveImports(func(location Location) (*Program, error) {
		switch location {
		case StringLocation("b"):
			return b, nil
		case StringLocation("c"):
			return c, nil
		default:
			return nil, fmt.Errorf("tried to resolve unknown import location: %s", location)
		}
	})

	assert.Nil(t, err)

	importsA := a.ImportedPrograms()

	actual := importsA[StringLocation("b").ID()]
	if actual != b {
		assert.Fail(t, "not b", actual)
	}

	importsB := b.ImportedPrograms()

	actual = importsB[StringLocation("c").ID()]
	if actual != c {
		assert.Fail(t, "not c", actual)
	}
}

func TestProgram_ResolveImportsCycle(t *testing.T) {

	makeImportingProgram := func(imported string) *Program {
		return &Program{
			Declarations: []Declaration{
				&ImportDeclaration{
					Location: StringLocation(imported),
				},
			},
		}
	}

	a := makeImportingProgram("b")
	b := makeImportingProgram("c")
	c := makeImportingProgram("a")

	err := a.ResolveImports(func(location Location) (*Program, error) {
		switch location {
		case StringLocation("a"):
			return a, nil
		case StringLocation("b"):
			return b, nil
		case StringLocation("c"):
			return c, nil
		default:
			return nil, fmt.Errorf("tried to resolve unknown import location: %s", location)
		}
	})

	assert.Equal(t,
		err,
		CyclicImportsError{
			Location: StringLocation("b"),
		},
	)
}