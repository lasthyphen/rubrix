package sema

import (
	"github.com/dapperlabs/flow-go/pkg/language/runtime/ast"
	"github.com/dapperlabs/flow-go/pkg/language/runtime/common"
	"github.com/dapperlabs/flow-go/pkg/language/runtime/errors"
	"github.com/dapperlabs/flow-go/pkg/language/runtime/sema/exit_detector"
	"github.com/dapperlabs/flow-go/pkg/language/runtime/sema/self_field_analyzer"
)

const ArgumentLabelNotRequired = "_"
const InitializerIdentifier = "init"
const SelfIdentifier = "self"
const BeforeIdentifier = "before"
const ResultIdentifier = "result"

type functionContext struct {
	returnType Type
	loops      int
}

// TODO: move annotations

var beforeType = &FunctionType{
	ParameterTypeAnnotations: NewTypeAnnotations(
		&AnyType{},
	),
	ReturnTypeAnnotation: NewTypeAnnotation(
		&AnyType{},
	),
	GetReturnType: func(argumentTypes []Type) Type {
		return argumentTypes[0]
	},
}

// Checker

type Checker struct {
	Program           *ast.Program
	PredeclaredValues map[string]ValueDeclaration
	PredeclaredTypes  map[string]TypeDeclaration
	ImportCheckers    map[ast.ImportLocation]*Checker
	errors            []error
	valueActivations  *ValueActivations
	typeActivations   *TypeActivations
	functionContexts  []*functionContext
	GlobalValues      map[string]*Variable
	GlobalTypes       map[string]Type
	inCondition       bool
	Occurrences       *Occurrences
	variableOrigins   map[*Variable]*Origin
	memberOrigins     map[Type]map[string]*Origin
	seenImports       map[ast.ImportLocation]bool
	isChecked         bool
	inCreate          bool
	Elaboration       *Elaboration
}

func NewChecker(
	program *ast.Program,
	predeclaredValues map[string]ValueDeclaration,
	predeclaredTypes map[string]TypeDeclaration,
) (*Checker, error) {

	checker := &Checker{
		Program:           program,
		PredeclaredValues: predeclaredValues,
		PredeclaredTypes:  predeclaredTypes,
		ImportCheckers:    map[ast.ImportLocation]*Checker{},
		valueActivations:  NewValueActivations(),
		typeActivations:   NewTypeActivations(baseTypes),
		GlobalValues:      map[string]*Variable{},
		GlobalTypes:       map[string]Type{},
		Occurrences:       NewOccurrences(),
		variableOrigins:   map[*Variable]*Origin{},
		memberOrigins:     map[Type]map[string]*Origin{},
		seenImports:       map[ast.ImportLocation]bool{},
		Elaboration:       NewElaboration(),
	}

	for name, declaration := range predeclaredValues {
		checker.declareValue(name, declaration)
		checker.declareGlobalValue(name)
	}

	for name, declaration := range predeclaredTypes {
		checker.declareTypeDeclaration(name, declaration)
	}

	err := checker.checkerError()
	if err != nil {
		return nil, err
	}

	return checker, nil
}

func (checker *Checker) declareValue(name string, declaration ValueDeclaration) {
	variable, err := checker.valueActivations.Declare(
		name,
		declaration.ValueDeclarationType(),
		declaration.ValueDeclarationKind(),
		declaration.ValueDeclarationPosition(),
		declaration.ValueDeclarationIsConstant(),
		declaration.ValueDeclarationArgumentLabels(),
	)
	checker.report(err)
	checker.recordVariableDeclarationOccurrence(name, variable)
}

func (checker *Checker) declareTypeDeclaration(name string, declaration TypeDeclaration) {
	identifier := ast.Identifier{
		Identifier: name,
		Pos:        declaration.TypeDeclarationPosition(),
	}

	ty := declaration.TypeDeclarationType()
	err := checker.typeActivations.Declare(identifier, ty)
	checker.report(err)
	checker.recordVariableDeclarationOccurrence(
		identifier.Identifier,
		&Variable{
			Kind:       declaration.TypeDeclarationKind(),
			IsConstant: true,
			Type:       ty,
			Pos:        &identifier.Pos,
		},
	)
}

func (checker *Checker) FindType(name string) Type {
	return checker.typeActivations.Find(name)
}

func (checker *Checker) IsChecked() bool {
	return checker.isChecked
}

func (checker *Checker) Check() error {
	if !checker.IsChecked() {
		checker.errors = nil
		checker.Program.Accept(checker)
		checker.isChecked = true
	}
	err := checker.checkerError()
	if err != nil {
		return err
	}
	return nil
}

func (checker *Checker) checkerError() *CheckerError {
	if len(checker.errors) > 0 {
		return &CheckerError{
			Errors: checker.errors,
		}
	}
	return nil
}

func (checker *Checker) report(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		checker.errors = append(checker.errors, errs...)
	}
}

func (checker *Checker) VisitProgram(program *ast.Program) ast.Repr {

	// pre-declare interfaces, composites, and functions (check afterwards)

	for _, declaration := range program.InterfaceDeclarations() {
		checker.declareInterfaceDeclaration(declaration)
	}

	for _, declaration := range program.CompositeDeclarations() {
		checker.declareCompositeDeclaration(declaration)
	}

	for _, declaration := range program.FunctionDeclarations() {
		checker.declareFunctionDeclaration(declaration)
	}

	// check all declarations

	for _, declaration := range program.Declarations {
		declaration.Accept(checker)
		checker.declareGlobalDeclaration(declaration)
	}

	return nil
}

func (checker *Checker) VisitFunctionDeclaration(declaration *ast.FunctionDeclaration) ast.Repr {
	return checker.visitFunctionDeclaration(declaration, true)
}

func (checker *Checker) visitFunctionDeclaration(declaration *ast.FunctionDeclaration, mustExit bool) ast.Repr {
	checker.checkFunctionAccessModifier(declaration)

	// global functions were previously declared, see `declareFunctionDeclaration`

	functionType := checker.Elaboration.FunctionDeclarationFunctionTypes[declaration]
	if functionType == nil {
		functionType = checker.declareFunctionDeclaration(declaration)
	}

	checker.checkFunction(
		declaration.Parameters,
		declaration.ReturnTypeAnnotation.StartPos,
		functionType,
		declaration.FunctionBlock,
		mustExit,
	)

	return nil
}

func (checker *Checker) declareFunctionDeclaration(declaration *ast.FunctionDeclaration) *FunctionType {

	functionType := checker.functionType(declaration.Parameters, declaration.ReturnTypeAnnotation)
	argumentLabels := declaration.Parameters.ArgumentLabels()

	checker.Elaboration.FunctionDeclarationFunctionTypes[declaration] = functionType

	variable, err := checker.valueActivations.DeclareFunction(
		declaration.Identifier,
		functionType,
		argumentLabels,
	)
	checker.report(err)

	checker.recordVariableDeclarationOccurrence(declaration.Identifier.Identifier, variable)

	return functionType
}

func (checker *Checker) checkFunctionAccessModifier(declaration *ast.FunctionDeclaration) {
	switch declaration.Access {
	case ast.AccessNotSpecified, ast.AccessPublic:
		return
	default:
		checker.report(
			&InvalidAccessModifierError{
				DeclarationKind: common.DeclarationKindFunction,
				Access:          declaration.Access,
				Pos:             declaration.StartPosition(),
			},
		)
	}
}

func (checker *Checker) checkFunction(
	parameters ast.Parameters,
	returnTypePosition ast.Position,
	functionType *FunctionType,
	functionBlock *ast.FunctionBlock,
	mustExit bool,
) {
	checker.valueActivations.Enter()
	defer checker.valueActivations.Leave()

	// check argument labels
	checker.checkArgumentLabels(parameters)

	checker.declareParameters(parameters, functionType.ParameterTypeAnnotations)

	checker.checkParameters(parameters, functionType.ParameterTypeAnnotations)
	if functionType.ReturnTypeAnnotation != nil {
		checker.checkTypeAnnotation(functionType.ReturnTypeAnnotation, returnTypePosition)
	}

	if functionBlock != nil {
		func() {
			// check the function's block
			checker.enterFunction(functionType)
			defer checker.leaveFunction()

			checker.visitFunctionBlock(functionBlock, functionType.ReturnTypeAnnotation)
		}()

		if mustExit && !functionType.ReturnTypeAnnotation.Type.Equal(&VoidType{}) {
			if !exit_detector.FunctionBlockExits(functionBlock) {
				checker.report(
					&MissingReturnStatementError{
						StartPos: functionBlock.StartPosition(),
						EndPos:   functionBlock.EndPosition(),
					},
				)
			}
		}
	}
}

func (checker *Checker) checkParameters(parameters ast.Parameters, parameterTypeAnnotations []*TypeAnnotation) {
	for i, parameter := range parameters {
		parameterTypeAnnotation := parameterTypeAnnotations[i]
		checker.checkTypeAnnotation(parameterTypeAnnotation, parameter.StartPos)
	}
}

func (checker *Checker) checkTypeAnnotation(typeAnnotation *TypeAnnotation, pos ast.Position) {
	checker.checkMoveAnnotation(
		typeAnnotation.Type,
		typeAnnotation.Move,
		pos,
	)
}

func (checker *Checker) checkMoveAnnotation(ty Type, move bool, pos ast.Position) {
	if ty.IsResourceType() {
		if !move {
			checker.report(
				&MissingMoveAnnotationError{
					Pos: pos,
				},
			)
		}
	} else {
		if move {
			checker.report(
				&InvalidMoveAnnotationError{
					Pos: pos,
				},
			)
		}
	}
}

// checkArgumentLabels checks that all argument labels (if any) are unique
//
func (checker *Checker) checkArgumentLabels(parameters ast.Parameters) {

	argumentLabelPositions := map[string]ast.Position{}

	for _, parameter := range parameters {
		label := parameter.Label
		if label == "" || label == ArgumentLabelNotRequired {
			continue
		}

		labelPos := parameter.StartPos

		if previousPos, ok := argumentLabelPositions[label]; ok {
			checker.report(
				&RedeclarationError{
					Kind:        common.DeclarationKindArgumentLabel,
					Name:        label,
					Pos:         labelPos,
					PreviousPos: &previousPos,
				},
			)
		}

		argumentLabelPositions[label] = labelPos
	}
}

// declareParameters declares a constant for each parameter,
// ensuring names are unique and constants don't already exist
//
func (checker *Checker) declareParameters(parameters ast.Parameters, parameterTypeAnnotations []*TypeAnnotation) {

	depth := checker.valueActivations.Depth()

	for i, parameter := range parameters {
		identifier := parameter.Identifier

		// check if variable with this identifier is already declared in the current scope
		existingVariable := checker.valueActivations.Find(identifier.Identifier)
		if existingVariable != nil && existingVariable.Depth == depth {
			checker.report(
				&RedeclarationError{
					Kind:        common.DeclarationKindParameter,
					Name:        identifier.Identifier,
					Pos:         identifier.Pos,
					PreviousPos: existingVariable.Pos,
				},
			)

			continue
		}

		parameterTypeAnnotation := parameterTypeAnnotations[i]
		parameterType := parameterTypeAnnotation.Type

		variable := &Variable{
			Kind:       common.DeclarationKindParameter,
			IsConstant: true,
			Type:       parameterType,
			Depth:      depth,
			Pos:        &identifier.Pos,
		}
		checker.valueActivations.Set(identifier.Identifier, variable)
		checker.recordVariableDeclarationOccurrence(identifier.Identifier, variable)
	}
}

func (checker *Checker) VisitVariableDeclaration(declaration *ast.VariableDeclaration) ast.Repr {
	checker.visitVariableDeclaration(declaration, false)
	return nil
}

func (checker *Checker) visitVariableDeclaration(declaration *ast.VariableDeclaration, isOptionalBinding bool) {
	valueType := declaration.Value.Accept(checker).(Type)

	checker.Elaboration.VariableDeclarationValueTypes[declaration] = valueType

	valueIsInvalid := IsInvalidType(valueType)

	// if the variable declaration is a optional binding, the value must be optional

	var valueIsOptional bool
	var optionalValueType *OptionalType

	if isOptionalBinding && !valueIsInvalid {
		optionalValueType, valueIsOptional = valueType.(*OptionalType)
		if !valueIsOptional {
			checker.report(
				&TypeMismatchError{
					ExpectedType: &OptionalType{},
					ActualType:   valueType,
					StartPos:     declaration.Value.StartPosition(),
					EndPos:       declaration.Value.EndPosition(),
				},
			)
		}
	}

	declarationType := valueType

	// does the declaration have an explicit type annotation?
	if declaration.TypeAnnotation != nil {
		typeAnnotation := checker.ConvertTypeAnnotation(declaration.TypeAnnotation)
		declarationType = typeAnnotation.Type

		checker.checkTypeAnnotation(typeAnnotation, declaration.TypeAnnotation.StartPos)

		// check the value type is a subtype of the declaration type
		if declarationType != nil && valueType != nil && !valueIsInvalid && !IsInvalidType(declarationType) {

			if isOptionalBinding {
				if optionalValueType != nil &&
					(optionalValueType.Equal(declarationType) ||
						!IsSubType(optionalValueType.Type, declarationType)) {

					checker.report(
						&TypeMismatchError{
							ExpectedType: declarationType,
							ActualType:   optionalValueType.Type,
							StartPos:     declaration.Value.StartPosition(),
							EndPos:       declaration.Value.EndPosition(),
						},
					)
				}

			} else {
				if !checker.IsTypeCompatible(declaration.Value, valueType, declarationType) {
					checker.report(
						&TypeMismatchError{
							ExpectedType: declarationType,
							ActualType:   valueType,
							StartPos:     declaration.Value.StartPosition(),
							EndPos:       declaration.Value.EndPosition(),
						},
					)
				}
			}
		}
	} else if isOptionalBinding && optionalValueType != nil {
		declarationType = optionalValueType.Type
	}

	if declarationType != nil {
		checker.checkTransfer(declaration.Transfer, declarationType)
	}

	checker.Elaboration.VariableDeclarationTargetTypes[declaration] = declarationType

	variable, err := checker.valueActivations.Declare(
		declaration.Identifier.Identifier,
		declarationType,
		declaration.DeclarationKind(),
		declaration.Identifier.Pos,
		declaration.IsConstant,
		nil,
	)
	checker.report(err)
	checker.recordVariableDeclarationOccurrence(declaration.Identifier.Identifier, variable)
}

func (checker *Checker) checkTransfer(transfer *ast.Transfer, valueType Type) {
	if valueType.IsResourceType() {
		if transfer.Operation != ast.TransferOperationMove {
			checker.report(
				&IncorrectTransferOperationError{
					ActualOperation:   transfer.Operation,
					ExpectedOperation: ast.TransferOperationMove,
					Pos:               transfer.Pos,
				},
			)
		}
	} else {
		if transfer.Operation == ast.TransferOperationMove {
			checker.report(
				&IncorrectTransferOperationError{
					ActualOperation:   transfer.Operation,
					ExpectedOperation: ast.TransferOperationCopy,
					Pos:               transfer.Pos,
				},
			)
		}
	}
}

func (checker *Checker) IsTypeCompatible(expression ast.Expression, valueType Type, targetType Type) bool {
	switch typedExpression := expression.(type) {
	case *ast.IntExpression:
		unwrappedTargetType := checker.unwrapOptionalType(targetType)

		// check if literal value fits range can't be checked when target is Never
		//
		if IsSubType(unwrappedTargetType, &IntegerType{}) &&
			!IsSubType(unwrappedTargetType, &NeverType{}) {

			checker.checkIntegerLiteral(typedExpression, unwrappedTargetType)

			return true
		}
	}

	return IsSubType(valueType, targetType)
}

// checkIntegerLiteral checks that the value of the integer literal
// fits into range of the target integer type
//
func (checker *Checker) checkIntegerLiteral(expression *ast.IntExpression, integerType Type) {
	intRange := integerType.(Ranged)
	literalValue := expression.Value
	rangeMin := intRange.Min()
	rangeMax := intRange.Max()
	if (rangeMin != nil && literalValue.Cmp(rangeMin) == -1) ||
		(rangeMax != nil && literalValue.Cmp(rangeMax) == 1) {

		checker.report(
			&InvalidIntegerLiteralRangeError{
				ExpectedType:     integerType,
				ExpectedRangeMin: rangeMin,
				ExpectedRangeMax: rangeMax,
				StartPos:         expression.StartPosition(),
				EndPos:           expression.EndPosition(),
			},
		)
	}
}

func (checker *Checker) declareGlobalDeclaration(declaration ast.Declaration) {
	name := declaration.DeclarationName()
	if name == "" {
		return
	}
	checker.declareGlobalValue(name)
	checker.declareGlobalType(name)
}

func (checker *Checker) declareGlobalValue(name string) {
	variable := checker.valueActivations.Find(name)
	if variable == nil {
		return
	}
	checker.GlobalValues[name] = variable
}

func (checker *Checker) declareGlobalType(name string) {
	ty := checker.typeActivations.Find(name)
	if ty == nil {
		return
	}
	checker.GlobalTypes[name] = ty
}

func (checker *Checker) VisitBlock(block *ast.Block) ast.Repr {
	checker.valueActivations.Scoped(func() {
		checker.visitStatements(block.Statements)
	})
	return nil
}

func (checker *Checker) visitStatements(statements []ast.Statement) {

	// check all statements
	for _, statement := range statements {

		// check statement is not a local composite or interface declaration

		if compositeDeclaration, ok := statement.(*ast.CompositeDeclaration); ok {
			checker.report(
				&InvalidDeclarationError{
					Kind:     compositeDeclaration.DeclarationKind(),
					StartPos: statement.StartPosition(),
					EndPos:   statement.EndPosition(),
				},
			)

			continue
		}

		if interfaceDeclaration, ok := statement.(*ast.InterfaceDeclaration); ok {
			checker.report(
				&InvalidDeclarationError{
					Kind:     interfaceDeclaration.DeclarationKind(),
					StartPos: statement.StartPosition(),
					EndPos:   statement.EndPosition(),
				},
			)

			continue
		}

		// check statement

		statement.Accept(checker)
	}
}

func (checker *Checker) VisitFunctionBlock(functionBlock *ast.FunctionBlock) ast.Repr {
	// NOTE: see visitFunctionBlock
	panic(&errors.UnreachableError{})
}

func (checker *Checker) visitFunctionBlock(functionBlock *ast.FunctionBlock, returnTypeAnnotation *TypeAnnotation) {

	checker.valueActivations.Enter()
	defer checker.valueActivations.Leave()

	checker.visitConditions(functionBlock.PreConditions)

	// NOTE: not checking block as it enters a new scope
	// and post-conditions need to be able to refer to block's declarations

	checker.visitStatements(functionBlock.Block.Statements)

	// if there is a post-condition, declare the function `before`

	// TODO: improve: only declare when a condition actually refers to `before`?

	if len(functionBlock.PostConditions) > 0 {
		checker.declareBefore()
	}

	// if there is a return type, declare the constant `result`
	// which has the return type

	if _, ok := returnTypeAnnotation.Type.(*VoidType); !ok {
		checker.declareResult(returnTypeAnnotation.Type)
	}

	checker.visitConditions(functionBlock.PostConditions)
}

func (checker *Checker) declareResult(ty Type) {
	_, err := checker.valueActivations.DeclareImplicitConstant(
		ResultIdentifier,
		ty,
		common.DeclarationKindConstant,
	)
	checker.report(err)
	// TODO: record occurrence - but what position?
}

func (checker *Checker) declareBefore() {
	_, err := checker.valueActivations.DeclareImplicitConstant(
		BeforeIdentifier,
		beforeType,
		common.DeclarationKindFunction,
	)
	checker.report(err)
	// TODO: record occurrence – but what position?
}

func (checker *Checker) VisitReturnStatement(statement *ast.ReturnStatement) ast.Repr {

	// check value type matches enclosing function's return type

	if statement.Expression == nil {
		return nil
	}

	valueType := statement.Expression.Accept(checker).(Type)
	valueIsInvalid := IsInvalidType(valueType)

	returnType := checker.currentFunction().returnType

	checker.Elaboration.ReturnStatementValueTypes[statement] = valueType
	checker.Elaboration.ReturnStatementReturnTypes[statement] = returnType

	if valueType == nil {
		return nil
	} else if valueIsInvalid {
		// return statement has expression, but function has Void return type?
		if _, ok := returnType.(*VoidType); ok {
			checker.report(
				&InvalidReturnValueError{
					StartPos: statement.Expression.StartPosition(),
					EndPos:   statement.Expression.EndPosition(),
				},
			)
		}
	} else {

		if !IsInvalidType(returnType) &&
			!checker.IsTypeCompatible(statement.Expression, valueType, returnType) {

			checker.report(
				&TypeMismatchError{
					ExpectedType: returnType,
					ActualType:   valueType,
					StartPos:     statement.Expression.StartPosition(),
					EndPos:       statement.Expression.EndPosition(),
				},
			)
		}

		checker.checkResourceMoveOperation(statement.Expression, valueType)
	}

	return nil
}

func (checker *Checker) checkResourceMoveOperation(valueExpression ast.Expression, valueType Type) {
	if !valueType.IsResourceType() {
		return
	}

	if unaryExpression, ok := valueExpression.(*ast.UnaryExpression); !ok ||
		unaryExpression.Operation != ast.OperationMove {

		checker.report(
			&MissingMoveOperationError{
				Pos: valueExpression.StartPosition(),
			},
		)
	}
}

func (checker *Checker) VisitBreakStatement(statement *ast.BreakStatement) ast.Repr {

	// check statement is inside loop

	if checker.currentFunction().loops == 0 {
		checker.report(
			&ControlStatementError{
				ControlStatement: common.ControlStatementBreak,
				StartPos:         statement.StartPos,
				EndPos:           statement.EndPos,
			},
		)
	}

	return nil
}

func (checker *Checker) VisitContinueStatement(statement *ast.ContinueStatement) ast.Repr {

	// check statement is inside loop

	if checker.currentFunction().loops == 0 {
		checker.report(
			&ControlStatementError{
				ControlStatement: common.ControlStatementContinue,
				StartPos:         statement.StartPos,
				EndPos:           statement.EndPos,
			},
		)
	}

	return nil
}

func (checker *Checker) VisitIfStatement(statement *ast.IfStatement) ast.Repr {

	thenElement := statement.Then

	var elseElement ast.Element = ast.NotAnElement{}
	if statement.Else != nil {
		elseElement = statement.Else
	}

	switch test := statement.Test.(type) {
	case ast.Expression:
		checker.visitConditional(test, thenElement, elseElement)

	case *ast.VariableDeclaration:
		checker.valueActivations.Scoped(func() {

			checker.visitVariableDeclaration(test, true)

			thenElement.Accept(checker)
		})

		elseElement.Accept(checker)

	default:
		panic(&errors.UnreachableError{})
	}

	return nil
}

func (checker *Checker) VisitWhileStatement(statement *ast.WhileStatement) ast.Repr {

	testExpression := statement.Test
	testType := testExpression.Accept(checker).(Type)

	if !IsSubType(testType, &BoolType{}) {
		checker.report(
			&TypeMismatchError{
				ExpectedType: &BoolType{},
				ActualType:   testType,
				StartPos:     testExpression.StartPosition(),
				EndPos:       testExpression.EndPosition(),
			},
		)
	}

	checker.currentFunction().loops += 1
	defer func() {
		checker.currentFunction().loops -= 1
	}()

	statement.Block.Accept(checker)

	return nil
}

func (checker *Checker) VisitAssignment(assignment *ast.AssignmentStatement) ast.Repr {
	valueType := assignment.Value.Accept(checker).(Type)
	checker.Elaboration.AssignmentStatementValueTypes[assignment] = valueType

	targetType := checker.visitAssignmentValueType(assignment, valueType)
	checker.Elaboration.AssignmentStatementTargetTypes[assignment] = targetType

	checker.checkTransfer(assignment.Transfer, valueType)

	return nil
}

func (checker *Checker) visitAssignmentValueType(assignment *ast.AssignmentStatement, valueType Type) (targetType Type) {
	switch target := assignment.Target.(type) {
	case *ast.IdentifierExpression:
		return checker.visitIdentifierExpressionAssignment(assignment, target, valueType)

	case *ast.IndexExpression:
		return checker.visitIndexExpressionAssignment(assignment, target, valueType)

	case *ast.MemberExpression:
		return checker.visitMemberExpressionAssignment(assignment, target, valueType)

	default:
		panic(&unsupportedAssignmentTargetExpression{
			target: target,
		})
	}

	panic(&errors.UnreachableError{})
}

func (checker *Checker) visitIdentifierExpressionAssignment(
	assignment *ast.AssignmentStatement,
	target *ast.IdentifierExpression,
	valueType Type,
) (targetType Type) {
	identifier := target.Identifier.Identifier

	// check identifier was declared before
	variable := checker.valueActivations.Find(identifier)
	if variable == nil {
		checker.report(
			&NotDeclaredError{
				ExpectedKind: common.DeclarationKindVariable,
				Name:         identifier,
				Pos:          target.StartPosition(),
			},
		)

		return &InvalidType{}
	} else {
		// check identifier is not a constant
		if variable.IsConstant {
			checker.report(
				&AssignmentToConstantError{
					Name:     identifier,
					StartPos: target.StartPosition(),
					EndPos:   target.EndPosition(),
				},
			)
		}

		// check value type is subtype of variable type
		if !IsInvalidType(valueType) &&
			!checker.IsTypeCompatible(assignment.Value, valueType, variable.Type) {

			checker.report(
				&TypeMismatchError{
					ExpectedType: variable.Type,
					ActualType:   valueType,
					StartPos:     assignment.Value.StartPosition(),
					EndPos:       assignment.Value.EndPosition(),
				},
			)
		}

		return variable.Type
	}
}

func (checker *Checker) visitIndexExpressionAssignment(
	assignment *ast.AssignmentStatement,
	target *ast.IndexExpression,
	valueType Type,
) (elementType Type) {

	elementType = checker.visitIndexingExpression(target.Expression, target.Index, true)

	if elementType == nil {
		return &InvalidType{}
	}

	if !IsInvalidType(elementType) &&
		!checker.IsTypeCompatible(assignment.Value, valueType, elementType) {

		checker.report(
			&TypeMismatchError{
				ExpectedType: elementType,
				ActualType:   valueType,
				StartPos:     assignment.Value.StartPosition(),
				EndPos:       assignment.Value.EndPosition(),
			},
		)
	}

	return elementType
}

func (checker *Checker) visitMemberExpressionAssignment(
	assignment *ast.AssignmentStatement,
	target *ast.MemberExpression,
	valueType Type,
) (memberType Type) {

	member := checker.visitMember(target)

	if member == nil {
		return
	}

	// check member is not constant

	if member.VariableKind == ast.VariableKindConstant {
		if member.IsInitialized {
			checker.report(
				&AssignmentToConstantMemberError{
					Name:     target.Identifier.Identifier,
					StartPos: assignment.Value.StartPosition(),
					EndPos:   assignment.Value.EndPosition(),
				},
			)
		}
	}

	member.IsInitialized = true

	// if value type is valid, check value can be assigned to member
	if !IsInvalidType(valueType) &&
		!checker.IsTypeCompatible(assignment.Value, valueType, member.Type) {

		checker.report(
			&TypeMismatchError{
				ExpectedType: member.Type,
				ActualType:   valueType,
				StartPos:     assignment.Value.StartPosition(),
				EndPos:       assignment.Value.EndPosition(),
			},
		)
	}

	return member.Type
}

// visitIndexingExpression checks if the indexed expression is indexable,
// checks if the indexing expression can be used to index into the indexed expression,
// and returns the expected element type
//
func (checker *Checker) visitIndexingExpression(
	indexedExpression ast.Expression,
	indexingExpression ast.Expression,
	isAssignment bool,
) Type {

	indexedType := indexedExpression.Accept(checker).(Type)
	indexingType := indexingExpression.Accept(checker).(Type)

	// NOTE: check indexed type first for UX reasons

	// check indexed expression's type is indexable
	// by getting the expected element

	if IsInvalidType(indexedType) {
		return &InvalidType{}
	}

	elementType := IndexableElementType(indexedType, isAssignment)
	if elementType == nil {
		elementType = &InvalidType{}

		checker.report(
			&NotIndexableTypeError{
				Type:     indexedType,
				StartPos: indexedExpression.StartPosition(),
				EndPos:   indexedExpression.EndPosition(),
			},
		)
	} else {

		// check indexing expression's type can be used to index
		// into indexed expression's type

		if !IsInvalidType(indexingType) &&
			!IsIndexingType(indexingType, indexedType) {

			checker.report(
				&NotIndexingTypeError{
					Type:     indexingType,
					StartPos: indexingExpression.StartPosition(),
					EndPos:   indexingExpression.EndPosition(),
				},
			)
		}
	}

	return elementType
}

func (checker *Checker) VisitIdentifierExpression(expression *ast.IdentifierExpression) ast.Repr {
	variable := checker.findAndCheckVariable(expression.Identifier, true)
	if variable == nil {
		return &InvalidType{}
	}

	return variable.Type
}

func (checker *Checker) findAndCheckVariable(identifier ast.Identifier, recordOccurrence bool) *Variable {
	variable := checker.valueActivations.Find(identifier.Identifier)
	if variable == nil {
		checker.report(
			&NotDeclaredError{
				ExpectedKind: common.DeclarationKindValue,
				Name:         identifier.Identifier,
				Pos:          identifier.StartPosition(),
			},
		)
		return nil
	}

	if recordOccurrence {
		checker.recordVariableReferenceOccurrence(
			identifier.StartPosition(),
			identifier.EndPosition(),
			variable,
		)
	}

	return variable
}

func (checker *Checker) visitBinaryOperation(expr *ast.BinaryExpression) (left, right Type) {
	left = expr.Left.Accept(checker).(Type)
	right = expr.Right.Accept(checker).(Type)
	return
}

func (checker *Checker) VisitBinaryExpression(expression *ast.BinaryExpression) ast.Repr {

	leftType, rightType := checker.visitBinaryOperation(expression)

	leftIsInvalid := IsInvalidType(leftType)
	rightIsInvalid := IsInvalidType(rightType)
	anyInvalid := leftIsInvalid || rightIsInvalid

	operation := expression.Operation
	operationKind := binaryOperationKind(operation)

	switch operationKind {
	case BinaryOperationKindIntegerArithmetic,
		BinaryOperationKindIntegerComparison:

		return checker.checkBinaryExpressionIntegerArithmeticOrComparison(
			expression, operation, operationKind,
			leftType, rightType,
			leftIsInvalid, rightIsInvalid, anyInvalid,
		)

	case BinaryOperationKindEquality:

		return checker.checkBinaryExpressionEquality(
			expression, operation, operationKind,
			leftType, rightType,
			leftIsInvalid, rightIsInvalid, anyInvalid,
		)

	case BinaryOperationKindBooleanLogic:

		return checker.checkBinaryExpressionBooleanLogic(
			expression, operation, operationKind,
			leftType, rightType,
			leftIsInvalid, rightIsInvalid, anyInvalid,
		)

	case BinaryOperationKindNilCoalescing:
		resultType := checker.checkBinaryExpressionNilCoalescing(
			expression, operation, operationKind,
			leftType, rightType,
			leftIsInvalid, rightIsInvalid, anyInvalid,
		)

		checker.Elaboration.BinaryExpressionResultTypes[expression] = resultType
		checker.Elaboration.BinaryExpressionRightTypes[expression] = rightType

		return resultType

	case BinaryOperationKindConcatenation:
		return checker.checkBinaryExpressionConcatenation(
			expression, operation, operationKind,
			leftType, rightType,
			leftIsInvalid, rightIsInvalid, anyInvalid,
		)
	}

	panic(&unsupportedOperation{
		kind:      common.OperationKindBinary,
		operation: operation,
		startPos:  expression.StartPosition(),
		endPos:    expression.EndPosition(),
	})
}

func (checker *Checker) checkBinaryExpressionIntegerArithmeticOrComparison(
	expression *ast.BinaryExpression,
	operation ast.Operation,
	operationKind BinaryOperationKind,
	leftType, rightType Type,
	leftIsInvalid, rightIsInvalid, anyInvalid bool,
) Type {
	// check both types are integer subtypes

	leftIsInteger := IsSubType(leftType, &IntegerType{})
	rightIsInteger := IsSubType(rightType, &IntegerType{})

	if !leftIsInteger && !rightIsInteger {
		if !anyInvalid {
			checker.report(
				&InvalidBinaryOperandsError{
					Operation: operation,
					LeftType:  leftType,
					RightType: rightType,
					StartPos:  expression.StartPosition(),
					EndPos:    expression.EndPosition(),
				},
			)
		}
	} else if !leftIsInteger {
		if !leftIsInvalid {
			checker.report(
				&InvalidBinaryOperandError{
					Operation:    operation,
					Side:         common.OperandSideLeft,
					ExpectedType: &IntegerType{},
					ActualType:   leftType,
					StartPos:     expression.Left.StartPosition(),
					EndPos:       expression.Left.EndPosition(),
				},
			)
		}
	} else if !rightIsInteger {
		if !rightIsInvalid {
			checker.report(
				&InvalidBinaryOperandError{
					Operation:    operation,
					Side:         common.OperandSideRight,
					ExpectedType: &IntegerType{},
					ActualType:   rightType,
					StartPos:     expression.Right.StartPosition(),
					EndPos:       expression.Right.EndPosition(),
				},
			)
		}
	}

	// check both types are equal
	if !anyInvalid && !leftType.Equal(rightType) {
		checker.report(
			&InvalidBinaryOperandsError{
				Operation: operation,
				LeftType:  leftType,
				RightType: rightType,
				StartPos:  expression.StartPosition(),
				EndPos:    expression.EndPosition(),
			},
		)
	}

	switch operationKind {
	case BinaryOperationKindIntegerArithmetic:
		return leftType
	case BinaryOperationKindIntegerComparison:
		return &BoolType{}
	}

	panic(&errors.UnreachableError{})
}

func (checker *Checker) checkBinaryExpressionEquality(
	expression *ast.BinaryExpression,
	operation ast.Operation,
	operationKind BinaryOperationKind,
	leftType, rightType Type,
	leftIsInvalid, rightIsInvalid, anyInvalid bool,
) (resultType Type) {
	// check both types are equal, and boolean subtypes or integer subtypes

	resultType = &BoolType{}

	if !anyInvalid &&
		leftType != nil &&
		!(checker.isValidEqualityType(leftType) &&
			checker.compatibleEqualityTypes(leftType, rightType)) {

		checker.report(
			&InvalidBinaryOperandsError{
				Operation: operation,
				LeftType:  leftType,
				RightType: rightType,
				StartPos:  expression.StartPosition(),
				EndPos:    expression.EndPosition(),
			},
		)
	}

	return
}

func (checker *Checker) isValidEqualityType(ty Type) bool {
	if IsSubType(ty, &BoolType{}) {
		return true
	}

	if IsSubType(ty, &IntegerType{}) {
		return true
	}

	if IsSubType(ty, &StringType{}) {
		return true
	}

	if IsSubType(ty, &CharacterType{}) {
		return true
	}

	if _, ok := ty.(*OptionalType); ok {
		return true
	}

	return false
}

func (checker *Checker) compatibleEqualityTypes(leftType, rightType Type) bool {
	unwrappedLeft := checker.unwrapOptionalType(leftType)
	unwrappedRight := checker.unwrapOptionalType(rightType)

	if unwrappedLeft.Equal(unwrappedRight) {
		return true
	}

	if _, ok := unwrappedLeft.(*NeverType); ok {
		return true
	}

	if _, ok := unwrappedRight.(*NeverType); ok {
		return true
	}

	return false
}

// unwrapOptionalType returns the type if it is not an optional type,
// or the inner-most type if it is (optional types are repeatedly unwrapped)
//
func (checker *Checker) unwrapOptionalType(ty Type) Type {
	for {
		optionalType, ok := ty.(*OptionalType)
		if !ok {
			return ty
		}
		ty = optionalType.Type
	}
}

func (checker *Checker) checkBinaryExpressionBooleanLogic(
	expression *ast.BinaryExpression,
	operation ast.Operation,
	operationKind BinaryOperationKind,
	leftType, rightType Type,
	leftIsInvalid, rightIsInvalid, anyInvalid bool,
) Type {
	// check both types are integer subtypes

	leftIsBool := IsSubType(leftType, &BoolType{})
	rightIsBool := IsSubType(rightType, &BoolType{})

	if !leftIsBool && !rightIsBool {
		if !anyInvalid {
			checker.report(
				&InvalidBinaryOperandsError{
					Operation: operation,
					LeftType:  leftType,
					RightType: rightType,
					StartPos:  expression.StartPosition(),
					EndPos:    expression.EndPosition(),
				},
			)
		}
	} else if !leftIsBool {
		if !leftIsInvalid {
			checker.report(
				&InvalidBinaryOperandError{
					Operation:    operation,
					Side:         common.OperandSideLeft,
					ExpectedType: &BoolType{},
					ActualType:   leftType,
					StartPos:     expression.Left.StartPosition(),
					EndPos:       expression.Left.EndPosition(),
				},
			)
		}
	} else if !rightIsBool {
		if !rightIsInvalid {
			checker.report(
				&InvalidBinaryOperandError{
					Operation:    operation,
					Side:         common.OperandSideRight,
					ExpectedType: &BoolType{},
					ActualType:   rightType,
					StartPos:     expression.Right.StartPosition(),
					EndPos:       expression.Right.EndPosition(),
				},
			)
		}
	}

	return &BoolType{}
}

func (checker *Checker) checkBinaryExpressionNilCoalescing(
	expression *ast.BinaryExpression,
	operation ast.Operation,
	operationKind BinaryOperationKind,
	leftType, rightType Type,
	leftIsInvalid, rightIsInvalid, anyInvalid bool,
) Type {
	leftOptional, leftIsOptional := leftType.(*OptionalType)

	if !leftIsInvalid {
		if !leftIsOptional {
			checker.report(
				&InvalidBinaryOperandError{
					Operation:    operation,
					Side:         common.OperandSideLeft,
					ExpectedType: &OptionalType{},
					ActualType:   leftType,
					StartPos:     expression.Left.StartPosition(),
					EndPos:       expression.Left.EndPosition(),
				},
			)
		}
	}

	if leftIsInvalid || !leftIsOptional {
		return &InvalidType{}
	}

	leftInner := leftOptional.Type

	if _, ok := leftInner.(*NeverType); ok {
		return rightType
	} else {
		canNarrow := false

		if !rightIsInvalid {
			if !IsSubType(rightType, leftOptional) {
				checker.report(
					&InvalidBinaryOperandError{
						Operation:    operation,
						Side:         common.OperandSideRight,
						ExpectedType: leftOptional,
						ActualType:   rightType,
						StartPos:     expression.Right.StartPosition(),
						EndPos:       expression.Right.EndPosition(),
					},
				)
			} else {
				canNarrow = IsSubType(rightType, leftInner)
			}
		}

		if !canNarrow {
			return leftOptional
		}
		return leftInner
	}
}

func (checker *Checker) checkBinaryExpressionConcatenation(
	expression *ast.BinaryExpression,
	operation ast.Operation,
	operationKind BinaryOperationKind,
	leftType, rightType Type,
	leftIsInvalid, rightIsInvalid, anyInvalid bool,
) Type {

	// check both types are concatenatable
	leftIsConcat := IsConcatenatableType(leftType)
	rightIsConcat := IsConcatenatableType(rightType)

	if !leftIsConcat && !rightIsConcat {
		if !anyInvalid {
			checker.report(
				&InvalidBinaryOperandsError{
					Operation: operation,
					LeftType:  leftType,
					RightType: rightType,
					StartPos:  expression.StartPosition(),
					EndPos:    expression.EndPosition(),
				},
			)
		}
	} else if !leftIsConcat {
		if !leftIsInvalid {
			checker.report(
				&InvalidBinaryOperandError{
					Operation:    operation,
					Side:         common.OperandSideLeft,
					ExpectedType: rightType,
					ActualType:   leftType,
					StartPos:     expression.Left.StartPosition(),
					EndPos:       expression.Left.EndPosition(),
				},
			)
		}
	} else if !rightIsConcat {
		if !rightIsInvalid {
			checker.report(
				&InvalidBinaryOperandError{
					Operation:    operation,
					Side:         common.OperandSideRight,
					ExpectedType: leftType,
					ActualType:   rightType,
					StartPos:     expression.Right.StartPosition(),
					EndPos:       expression.Right.EndPosition(),
				},
			)
		}
	}

	// check both types are equal
	if !leftType.Equal(rightType) {
		checker.report(
			&InvalidBinaryOperandsError{
				Operation: operation,
				LeftType:  leftType,
				RightType: rightType,
				StartPos:  expression.StartPosition(),
				EndPos:    expression.EndPosition(),
			},
		)
	}

	return leftType
}

func (checker *Checker) VisitUnaryExpression(expression *ast.UnaryExpression) ast.Repr {

	valueType := expression.Expression.Accept(checker).(Type)

	switch expression.Operation {
	case ast.OperationNegate:
		if !IsSubType(valueType, &BoolType{}) {
			checker.report(
				&InvalidUnaryOperandError{
					Operation:    expression.Operation,
					ExpectedType: &BoolType{},
					ActualType:   valueType,
					StartPos:     expression.Expression.StartPosition(),
					EndPos:       expression.Expression.EndPosition(),
				},
			)
		}
		return valueType

	case ast.OperationMinus:
		if !IsSubType(valueType, &IntegerType{}) {
			checker.report(
				&InvalidUnaryOperandError{
					Operation:    expression.Operation,
					ExpectedType: &IntegerType{},
					ActualType:   valueType,
					StartPos:     expression.Expression.StartPosition(),
					EndPos:       expression.Expression.EndPosition(),
				},
			)
		}
		return valueType

	case ast.OperationMove:
		if !valueType.IsResourceType() {
			checker.report(
				&InvalidMoveOperationError{
					StartPos: expression.StartPos,
					EndPos:   expression.Expression.StartPosition(),
				},
			)
		}

		return valueType
	}

	panic(&unsupportedOperation{
		kind:      common.OperationKindUnary,
		operation: expression.Operation,
		startPos:  expression.StartPos,
		endPos:    expression.EndPos,
	})
}

func (checker *Checker) VisitExpressionStatement(statement *ast.ExpressionStatement) ast.Repr {
	result := statement.Expression.Accept(checker)

	if ty, ok := result.(Type); ok &&
		ty.IsResourceType() {

		checker.report(
			&ResourceLossError{
				StartPos: statement.Expression.StartPosition(),
				EndPos:   statement.Expression.EndPosition(),
			},
		)
	}

	return nil
}

func (checker *Checker) VisitBoolExpression(expression *ast.BoolExpression) ast.Repr {
	return &BoolType{}
}

func (checker *Checker) VisitNilExpression(expression *ast.NilExpression) ast.Repr {
	// TODO: verify
	return &OptionalType{
		Type: &NeverType{},
	}
}

func (checker *Checker) VisitIntExpression(expression *ast.IntExpression) ast.Repr {
	return &IntType{}
}

func (checker *Checker) VisitStringExpression(expression *ast.StringExpression) ast.Repr {
	return &StringType{}
}

func (checker *Checker) VisitArrayExpression(expression *ast.ArrayExpression) ast.Repr {

	// visit all elements, ensure they are all the same type

	var elementType Type

	for _, value := range expression.Values {
		valueType := value.Accept(checker).(Type)

		// infer element type from first element
		// TODO: find common super type?
		if elementType == nil {
			elementType = valueType
		} else if !IsSubType(valueType, elementType) {
			checker.report(
				&TypeMismatchError{
					ExpectedType: elementType,
					ActualType:   valueType,
					StartPos:     value.StartPosition(),
					EndPos:       value.EndPosition(),
				},
			)
		}
	}

	if elementType == nil {
		elementType = &NeverType{}
	}

	return &VariableSizedType{
		Type: elementType,
	}
}

func (checker *Checker) VisitDictionaryExpression(expression *ast.DictionaryExpression) ast.Repr {

	// visit all entries, ensure key are all the same type,
	// and values are all the same type

	var keyType, valueType Type

	for _, entry := range expression.Entries {
		entryKeyType := entry.Key.Accept(checker).(Type)
		entryValueType := entry.Value.Accept(checker).(Type)

		// infer key type from first entry's key
		// TODO: find common super type?
		if keyType == nil {
			keyType = entryKeyType
		} else if !IsSubType(entryKeyType, keyType) {
			checker.report(
				&TypeMismatchError{
					ExpectedType: keyType,
					ActualType:   entryKeyType,
					StartPos:     entry.Key.StartPosition(),
					EndPos:       entry.Key.EndPosition(),
				},
			)
		}

		// infer value type from first entry's value
		// TODO: find common super type?
		if valueType == nil {
			valueType = entryValueType
		} else if !IsSubType(entryValueType, valueType) {
			checker.report(
				&TypeMismatchError{
					ExpectedType: valueType,
					ActualType:   entryValueType,
					StartPos:     entry.Value.StartPosition(),
					EndPos:       entry.Value.EndPosition(),
				},
			)
		}
	}

	if keyType == nil {
		keyType = &NeverType{}
	}
	if valueType == nil {
		valueType = &NeverType{}
	}

	return &DictionaryType{
		KeyType:   keyType,
		ValueType: valueType,
	}
}

func (checker *Checker) VisitMemberExpression(expression *ast.MemberExpression) ast.Repr {
	member := checker.visitMember(expression)

	var memberType Type = &InvalidType{}
	if member != nil {
		memberType = member.Type
	}

	return memberType
}

func (checker *Checker) visitMember(expression *ast.MemberExpression) *Member {
	member, ok := checker.Elaboration.MemberExpressionMembers[expression]
	if ok {
		return member
	}

	expressionType := expression.Expression.Accept(checker).(Type)

	if expressionType.IsResourceType() {
		if _, isIdentifier := expression.Expression.(*ast.IdentifierExpression); !isIdentifier {
			checker.report(&ResourceLossError{
				StartPos: expression.Expression.StartPosition(),
				EndPos:   expression.Expression.EndPosition(),
			})
		}
	}

	origins := checker.memberOrigins[expressionType]

	identifier := expression.Identifier.Identifier
	identifierStartPosition := expression.Identifier.StartPosition()
	identifierEndPosition := expression.Identifier.EndPosition()

	if ty, ok := expressionType.(HasMembers); ok {
		member = ty.GetMember(identifier)
	}

	if _, isArrayType := expressionType.(ArrayType); isArrayType && member != nil {
		// TODO: implement Equatable interface: https://github.com/dapperlabs/bamboo-node/issues/78
		if identifier == "contains" {
			functionType := member.Type.(*FunctionType)

			if !IsEquatableType(functionType.ParameterTypeAnnotations[0].Type) {
				checker.report(
					&NotEquatableTypeError{
						Type:     expressionType,
						StartPos: identifierStartPosition,
						EndPos:   identifierEndPosition,
					},
				)

				return nil
			}
		}
	}

	if member == nil {
		if !IsInvalidType(expressionType) {
			checker.report(
				&NotDeclaredMemberError{
					Type:     expressionType,
					Name:     identifier,
					StartPos: identifierStartPosition,
					EndPos:   identifierEndPosition,
				},
			)
		}
	} else {
		origin := origins[identifier]
		checker.Occurrences.Put(
			identifierStartPosition,
			identifierEndPosition,
			origin,
		)
	}

	checker.Elaboration.MemberExpressionMembers[expression] = member

	return member
}

func (checker *Checker) VisitIndexExpression(expression *ast.IndexExpression) ast.Repr {
	return checker.visitIndexingExpression(expression.Expression, expression.Index, false)
}

func (checker *Checker) VisitConditionalExpression(expression *ast.ConditionalExpression) ast.Repr {

	thenType, elseType := checker.visitConditional(expression.Test, expression.Then, expression.Else)

	if thenType == nil || elseType == nil {
		panic(&errors.UnreachableError{})
	}

	// TODO: improve
	resultType := thenType

	if !IsSubType(elseType, resultType) {
		checker.report(
			&TypeMismatchError{
				ExpectedType: resultType,
				ActualType:   elseType,
				StartPos:     expression.Else.StartPosition(),
				EndPos:       expression.Else.EndPosition(),
			},
		)
	}

	return resultType
}

func (checker *Checker) VisitInvocationExpression(invocationExpression *ast.InvocationExpression) ast.Repr {
	inCreate := checker.inCreate
	checker.inCreate = false
	defer func() {
		checker.inCreate = inCreate
	}()

	// check the invoked expression can be invoked

	invokedExpression := invocationExpression.InvokedExpression
	expressionType := invokedExpression.Accept(checker).(Type)

	invokableType, ok := expressionType.(InvokableType)
	if !ok {
		if !IsInvalidType(expressionType) {
			checker.report(
				&NotCallableError{
					Type:     expressionType,
					StartPos: invokedExpression.StartPosition(),
					EndPos:   invokedExpression.EndPosition(),
				},
			)
		}
		return &InvalidType{}
	}

	functionType := invokableType.InvocationFunctionType()

	var returnType Type = &InvalidType{}

	// invoked expression has function type

	argumentTypes := checker.checkInvocationArguments(invocationExpression, functionType)

	// if the invocation refers directly to the name of the function as stated in the declaration,
	// or the invocation refers to a function of a composite (member),
	// check that the correct argument labels are supplied in the invocation

	if identifierExpression, ok := invokedExpression.(*ast.IdentifierExpression); ok {
		checker.checkIdentifierInvocationArgumentLabels(
			invocationExpression,
			identifierExpression,
		)
	} else if memberExpression, ok := invokedExpression.(*ast.MemberExpression); ok {
		checker.checkMemberInvocationArgumentLabels(
			invocationExpression,
			memberExpression,
		)
	}

	parameterTypeAnnotations := functionType.ParameterTypeAnnotations
	if len(argumentTypes) == len(parameterTypeAnnotations) &&
		functionType.GetReturnType != nil {

		returnType = functionType.GetReturnType(argumentTypes)
	} else {
		returnType = functionType.ReturnTypeAnnotation.Type
	}

	checker.Elaboration.InvocationExpressionArgumentTypes[invocationExpression] = argumentTypes

	var parameterTypes []Type
	for _, parameterTypeAnnotation := range parameterTypeAnnotations {
		parameterTypes = append(parameterTypes, parameterTypeAnnotation.Type)
	}
	checker.Elaboration.InvocationExpressionParameterTypes[invocationExpression] = parameterTypes

	checker.checkConstructorInvocationWithResourceResult(
		invocationExpression,
		invokableType,
		returnType,
		inCreate,
	)

	return returnType
}

func (checker *Checker) checkConstructorInvocationWithResourceResult(
	invocationExpression *ast.InvocationExpression,
	invokableType InvokableType,
	returnType Type,
	inCreate bool,
) {
	if _, ok := invokableType.(*ConstructorFunctionType); !ok {
		return
	}

	// NOTE: not using `isResourceType`,
	// as only direct resource types can be constructed

	if compositeReturnType, ok := returnType.(*CompositeType); !ok ||
		compositeReturnType.Kind != common.CompositeKindResource {

		return
	}

	if inCreate {
		return
	}

	checker.report(
		&MissingCreateError{
			StartPos: invocationExpression.StartPosition(),
			EndPos:   invocationExpression.EndPosition(),
		},
	)
}

func (checker *Checker) checkIdentifierInvocationArgumentLabels(
	invocationExpression *ast.InvocationExpression,
	identifierExpression *ast.IdentifierExpression,
) {
	variable := checker.findAndCheckVariable(identifierExpression.Identifier, false)

	if variable == nil || len(variable.ArgumentLabels) == 0 {
		return
	}

	checker.checkInvocationArgumentLabels(
		invocationExpression.Arguments,
		variable.ArgumentLabels,
	)
}

func (checker *Checker) checkMemberInvocationArgumentLabels(
	invocationExpression *ast.InvocationExpression,
	memberExpression *ast.MemberExpression,
) {
	member := checker.visitMember(memberExpression)

	if member == nil || len(member.ArgumentLabels) == 0 {
		return
	}

	checker.checkInvocationArgumentLabels(
		invocationExpression.Arguments,
		member.ArgumentLabels,
	)
}

func (checker *Checker) checkInvocationArgumentLabels(
	arguments []*ast.Argument,
	argumentLabels []string,
) {
	argumentCount := len(arguments)

	for i, argumentLabel := range argumentLabels {
		if i >= argumentCount {
			break
		}

		argument := arguments[i]
		providedLabel := argument.Label
		if argumentLabel == ArgumentLabelNotRequired {
			// argument label is not required,
			// check it is not provided

			if providedLabel != "" {
				checker.report(
					&IncorrectArgumentLabelError{
						ActualArgumentLabel:   providedLabel,
						ExpectedArgumentLabel: "",
						StartPos:              *argument.LabelStartPos,
						EndPos:                *argument.LabelEndPos,
					},
				)
			}
		} else {
			// argument label is required,
			// check it is provided and correct
			if providedLabel == "" {
				checker.report(
					&MissingArgumentLabelError{
						ExpectedArgumentLabel: argumentLabel,
						StartPos:              argument.Expression.StartPosition(),
						EndPos:                argument.Expression.EndPosition(),
					},
				)
			} else if providedLabel != argumentLabel {
				checker.report(
					&IncorrectArgumentLabelError{
						ActualArgumentLabel:   providedLabel,
						ExpectedArgumentLabel: argumentLabel,
						StartPos:              *argument.LabelStartPos,
						EndPos:                *argument.LabelEndPos,
					},
				)
			}
		}
	}
}

func (checker *Checker) checkInvocationArguments(
	invocationExpression *ast.InvocationExpression,
	functionType *FunctionType,
) (
	argumentTypes []Type,
) {
	argumentCount := len(invocationExpression.Arguments)

	// check the invocation's argument count matches the function's parameter count
	parameterCount := len(functionType.ParameterTypeAnnotations)
	if argumentCount != parameterCount {

		// TODO: improve
		if functionType.RequiredArgumentCount == nil ||
			argumentCount < *functionType.RequiredArgumentCount {

			checker.report(
				&ArgumentCountError{
					ParameterCount: parameterCount,
					ArgumentCount:  argumentCount,
					StartPos:       invocationExpression.StartPosition(),
					EndPos:         invocationExpression.EndPosition(),
				},
			)
		}
	}

	minCount := argumentCount
	if parameterCount < argumentCount {
		minCount = parameterCount
	}

	for i := 0; i < minCount; i++ {
		// ensure the type of the argument matches the type of the parameter

		parameterType := functionType.ParameterTypeAnnotations[i].Type
		argument := invocationExpression.Arguments[i]

		argumentType := argument.Expression.Accept(checker).(Type)

		argumentTypes = append(argumentTypes, argumentType)

		if !IsInvalidType(parameterType) &&
			!checker.IsTypeCompatible(argument.Expression, argumentType, parameterType) {

			checker.report(
				&TypeMismatchError{
					ExpectedType: parameterType,
					ActualType:   argumentType,
					StartPos:     argument.Expression.StartPosition(),
					EndPos:       argument.Expression.EndPosition(),
				},
			)
		}

		checker.checkResourceMoveOperation(argument.Expression, argumentType)
	}

	return argumentTypes
}

func (checker *Checker) VisitFunctionExpression(expression *ast.FunctionExpression) ast.Repr {

	// TODO: infer
	functionType := checker.functionType(expression.Parameters, expression.ReturnTypeAnnotation)

	checker.Elaboration.FunctionExpressionFunctionType[expression] = functionType

	checker.checkFunction(
		expression.Parameters,
		expression.ReturnTypeAnnotation.StartPos,
		functionType,
		expression.FunctionBlock,
		true,
	)

	// function expressions are not allowed in conditions

	if checker.inCondition {
		checker.report(
			&FunctionExpressionInConditionError{
				StartPos: expression.StartPosition(),
				EndPos:   expression.EndPosition(),
			},
		)
	}

	return functionType
}

// ConvertType converts an AST type representation to a sema type
func (checker *Checker) ConvertType(t ast.Type) Type {
	switch t := t.(type) {
	case *ast.NominalType:
		identifier := t.Identifier.Identifier
		result := checker.typeActivations.Find(identifier)
		if result == nil {
			checker.report(
				&NotDeclaredError{
					ExpectedKind: common.DeclarationKindType,
					Name:         identifier,
					Pos:          t.Pos,
				},
			)
			return &InvalidType{}
		}
		return result

	case *ast.VariableSizedType:
		elementType := checker.ConvertType(t.Type)
		return &VariableSizedType{
			Type: elementType,
		}

	case *ast.ConstantSizedType:
		elementType := checker.ConvertType(t.Type)
		return &ConstantSizedType{
			Type: elementType,
			Size: t.Size,
		}

	case *ast.FunctionType:
		var parameterTypeAnnotations []*TypeAnnotation
		for _, parameterTypeAnnotation := range t.ParameterTypeAnnotations {
			parameterTypeAnnotation := checker.ConvertTypeAnnotation(parameterTypeAnnotation)
			parameterTypeAnnotations = append(parameterTypeAnnotations,
				parameterTypeAnnotation,
			)
		}

		returnTypeAnnotation := checker.ConvertTypeAnnotation(t.ReturnTypeAnnotation)

		return &FunctionType{
			ParameterTypeAnnotations: parameterTypeAnnotations,
			ReturnTypeAnnotation:     returnTypeAnnotation,
		}

	case *ast.OptionalType:
		result := checker.ConvertType(t.Type)
		return &OptionalType{result}

	case *ast.DictionaryType:
		keyType := checker.ConvertType(t.KeyType)
		valueType := checker.ConvertType(t.ValueType)

		return &DictionaryType{
			KeyType:   keyType,
			ValueType: valueType,
		}
	}

	panic(&astTypeConversionError{invalidASTType: t})
}

// ConvertTypeAnnotation converts an AST type annotation representation
// to a sema type annotation
//
func (checker *Checker) ConvertTypeAnnotation(typeAnnotation *ast.TypeAnnotation) *TypeAnnotation {
	convertedType := checker.ConvertType(typeAnnotation.Type)
	return &TypeAnnotation{
		Move: typeAnnotation.Move,
		Type: convertedType,
	}
}

func (checker *Checker) enterFunction(functionType *FunctionType) {
	checker.functionContexts = append(checker.functionContexts,
		&functionContext{
			returnType: functionType.ReturnTypeAnnotation.Type,
		})
}

func (checker *Checker) leaveFunction() {
	lastIndex := len(checker.functionContexts) - 1
	checker.functionContexts = checker.functionContexts[:lastIndex]
}

func (checker *Checker) currentFunction() *functionContext {
	lastIndex := len(checker.functionContexts) - 1
	if lastIndex < 0 {
		return nil
	}
	return checker.functionContexts[lastIndex]
}

func (checker *Checker) functionType(
	parameters ast.Parameters,
	returnTypeAnnotation *ast.TypeAnnotation,
) *FunctionType {
	convertedParameterTypeAnnotations :=
		checker.parameterTypeAnnotations(parameters)

	convertedReturnTypeAnnotation :=
		checker.ConvertTypeAnnotation(returnTypeAnnotation)

	return &FunctionType{
		ParameterTypeAnnotations: convertedParameterTypeAnnotations,
		ReturnTypeAnnotation:     convertedReturnTypeAnnotation,
	}
}

func (checker *Checker) parameterTypeAnnotations(parameters ast.Parameters) []*TypeAnnotation {

	parameterTypeAnnotations := make([]*TypeAnnotation, len(parameters))

	for i, parameter := range parameters {
		convertedParameterType := checker.ConvertType(parameter.TypeAnnotation.Type)
		parameterTypeAnnotations[i] = &TypeAnnotation{
			Move: parameter.TypeAnnotation.Move,
			Type: convertedParameterType,
		}
	}

	return parameterTypeAnnotations
}

// visitConditional checks a conditional. the test expression must be a boolean.
// the then and else elements may be expressions, in which case the types are returned.
func (checker *Checker) visitConditional(
	test ast.Expression,
	thenElement ast.Element,
	elseElement ast.Element,
) (
	thenType, elseType Type,
) {
	testType := test.Accept(checker).(Type)

	if !IsSubType(testType, &BoolType{}) {
		checker.report(
			&TypeMismatchError{
				ExpectedType: &BoolType{},
				ActualType:   testType,
				StartPos:     test.StartPosition(),
				EndPos:       test.EndPosition(),
			},
		)
	}

	thenResult := thenElement.Accept(checker)
	if thenResult != nil {
		thenType = thenResult.(Type)
	}

	elseResult := elseElement.Accept(checker)
	if elseResult != nil {
		elseType = elseResult.(Type)
	}

	return
}

func (checker *Checker) VisitCompositeDeclaration(declaration *ast.CompositeDeclaration) ast.Repr {

	compositeType := checker.Elaboration.CompositeDeclarationTypes[declaration]

	// TODO: also check nested composite members

	// TODO: also check nested composite members' identifiers

	// TODO: also check nested composite fields' type annotations

	checker.checkMemberIdentifiers(
		declaration.Members.Fields,
		declaration.Members.Functions,
	)

	checker.checkInitializers(
		declaration.Members.Initializers,
		declaration.Members.Fields,
		compositeType,
		declaration.DeclarationKind(),
		declaration.Identifier.Identifier,
		compositeType.ConstructorParameterTypeAnnotations,
		initializerKindComposite,
	)

	checker.checkFieldsInitialized(declaration, compositeType)

	checker.checkCompositeFunctions(declaration.Members.Functions, compositeType)

	// check composite conforms to interfaces.
	// NOTE: perform after completing composite type (e.g. setting constructor parameter types)

	for i, interfaceType := range compositeType.Conformances {
		conformance := declaration.Conformances[i]

		checker.checkCompositeConformance(
			compositeType,
			interfaceType,
			declaration.Identifier.Pos,
			conformance.Identifier,
		)
	}

	// TODO: support non-structure composites, such as contracts and resources

	if declaration.CompositeKind != common.CompositeKindStructure {
		checker.report(
			&UnsupportedDeclarationError{
				DeclarationKind: declaration.DeclarationKind(),
				StartPos:        declaration.Identifier.StartPosition(),
				EndPos:          declaration.Identifier.EndPosition(),
			},
		)
	}

	// TODO: support nested declarations for contracts and contract interfaces

	// report error for first nested composite declaration, if any
	if len(declaration.Members.CompositeDeclarations) > 0 {
		firstNestedCompositeDeclaration := declaration.Members.CompositeDeclarations[0]

		checker.report(
			&UnsupportedDeclarationError{
				DeclarationKind: firstNestedCompositeDeclaration.DeclarationKind(),
				StartPos:        firstNestedCompositeDeclaration.Identifier.StartPosition(),
				EndPos:          firstNestedCompositeDeclaration.Identifier.EndPosition(),
			},
		)
	}

	return nil
}

func (checker *Checker) declareCompositeDeclaration(declaration *ast.CompositeDeclaration) {

	// NOTE: fields and functions might already refer to declaration itself.
	// insert a dummy type for now, so lookup succeeds during conversion,
	// then fix up the type reference

	compositeType := &CompositeType{}

	identifier := declaration.Identifier

	err := checker.typeActivations.Declare(identifier, compositeType)
	checker.report(err)
	checker.recordVariableDeclarationOccurrence(
		identifier.Identifier,
		&Variable{
			Kind:       declaration.DeclarationKind(),
			IsConstant: true,
			Type:       compositeType,
			Pos:        &identifier.Pos,
		},
	)

	conformances := checker.conformances(declaration)

	members, origins := checker.membersAndOrigins(
		declaration.Members.Fields,
		declaration.Members.Functions,
		true,
	)

	*compositeType = CompositeType{
		Kind:         declaration.CompositeKind,
		Identifier:   identifier.Identifier,
		Members:      members,
		Conformances: conformances,
	}

	checker.memberOrigins[compositeType] = origins

	// TODO: support multiple overloaded initializers

	var parameterTypeAnnotations []*TypeAnnotation
	initializerCount := len(declaration.Members.Initializers)
	if initializerCount > 0 {
		firstInitializer := declaration.Members.Initializers[0]
		parameterTypeAnnotations = checker.parameterTypeAnnotations(firstInitializer.Parameters)

		if initializerCount > 1 {
			secondInitializer := declaration.Members.Initializers[1]

			checker.report(
				&UnsupportedOverloadingError{
					DeclarationKind: common.DeclarationKindInitializer,
					StartPos:        secondInitializer.StartPosition(),
					EndPos:          secondInitializer.EndPosition(),
				},
			)
		}
	}

	compositeType.ConstructorParameterTypeAnnotations = parameterTypeAnnotations

	checker.Elaboration.CompositeDeclarationTypes[declaration] = compositeType

	// declare constructor

	checker.declareCompositeConstructor(declaration, compositeType, parameterTypeAnnotations)
}

func (checker *Checker) conformances(declaration *ast.CompositeDeclaration) []*InterfaceType {

	var interfaceTypes []*InterfaceType
	seenConformances := map[string]bool{}

	compositeIdentifier := declaration.Identifier.Identifier

	for _, conformance := range declaration.Conformances {
		convertedType := checker.ConvertType(conformance)

		if interfaceType, ok := convertedType.(*InterfaceType); ok {
			interfaceTypes = append(interfaceTypes, interfaceType)

		} else if !IsInvalidType(convertedType) {
			checker.report(
				&InvalidConformanceError{
					Type: convertedType,
					Pos:  conformance.Pos,
				},
			)
		}

		conformanceIdentifier := conformance.Identifier.Identifier

		if seenConformances[conformanceIdentifier] {
			checker.report(
				&DuplicateConformanceError{
					CompositeIdentifier: compositeIdentifier,
					Conformance:         conformance,
				},
			)

		}
		seenConformances[conformanceIdentifier] = true
	}
	return interfaceTypes
}

func (checker *Checker) checkCompositeConformance(
	compositeType *CompositeType,
	interfaceType *InterfaceType,
	compositeIdentifierPos ast.Position,
	interfaceIdentifier ast.Identifier,
) {
	var missingMembers []*Member
	var memberMismatches []MemberMismatch
	var initializerMismatch *InitializerMismatch

	// ensure the composite kinds match, e.g. a structure shouldn't be able
	// to conform to a resource interface

	if interfaceType.CompositeKind != compositeType.Kind {
		checker.report(
			&CompositeKindMismatchError{
				ExpectedKind: compositeType.Kind,
				ActualKind:   interfaceType.CompositeKind,
				StartPos:     interfaceIdentifier.StartPosition(),
				EndPos:       interfaceIdentifier.EndPosition(),
			},
		)
	}

	if interfaceType.InitializerParameterTypeAnnotations != nil {

		initializerType := &FunctionType{
			ParameterTypeAnnotations: compositeType.ConstructorParameterTypeAnnotations,
			ReturnTypeAnnotation:     NewTypeAnnotation(&VoidType{}),
		}
		interfaceInitializerType := &FunctionType{
			ParameterTypeAnnotations: interfaceType.InitializerParameterTypeAnnotations,
			ReturnTypeAnnotation:     NewTypeAnnotation(&VoidType{}),
		}

		// TODO: subtype?
		if !initializerType.Equal(interfaceInitializerType) {
			initializerMismatch = &InitializerMismatch{
				CompositeParameterTypes: compositeType.ConstructorParameterTypeAnnotations,
				InterfaceParameterTypes: interfaceType.InitializerParameterTypeAnnotations,
			}
		}
	}

	for name, interfaceMember := range interfaceType.Members {

		compositeMember, ok := compositeType.Members[name]
		if !ok {
			missingMembers = append(missingMembers, interfaceMember)
			continue
		}

		if !checker.memberSatisfied(compositeMember, interfaceMember) {
			memberMismatches = append(memberMismatches,
				MemberMismatch{
					CompositeMember: compositeMember,
					InterfaceMember: interfaceMember,
				},
			)
		}
	}

	if len(missingMembers) > 0 ||
		len(memberMismatches) > 0 ||
		initializerMismatch != nil {

		checker.report(
			&ConformanceError{
				CompositeType:       compositeType,
				InterfaceType:       interfaceType,
				Pos:                 compositeIdentifierPos,
				InitializerMismatch: initializerMismatch,
				MissingMembers:      missingMembers,
				MemberMismatches:    memberMismatches,
			},
		)
	}
}

func (checker *Checker) memberSatisfied(compositeMember, interfaceMember *Member) bool {
	// TODO: subtype?
	if !compositeMember.Type.Equal(interfaceMember.Type) {
		return false
	}

	if interfaceMember.VariableKind != ast.VariableKindNotSpecified &&
		compositeMember.VariableKind != interfaceMember.VariableKind {

		return false
	}

	return true
}

func (checker *Checker) checkFieldsInitialized(
	declaration *ast.CompositeDeclaration,
	compositeType *CompositeType,
) {
	for _, initializer := range declaration.Members.Initializers {
		unassigned, errs := self_field_analyzer.CheckSelfFieldInitializations(
			declaration.Members.Fields,
			initializer.FunctionBlock,
		)

		for _, field := range unassigned {
			checker.report(
				&FieldUninitializedError{
					Name:          field.Identifier.Identifier,
					Pos:           field.Identifier.Pos,
					CompositeType: compositeType,
					Initializer:   initializer,
				},
			)
		}

		checker.report(errs...)
	}
}

func (checker *Checker) declareCompositeConstructor(
	compositeDeclaration *ast.CompositeDeclaration,
	compositeType *CompositeType,
	parameterTypeAnnotations []*TypeAnnotation,
) {
	functionType := &ConstructorFunctionType{
		&FunctionType{
			ReturnTypeAnnotation: NewTypeAnnotation(
				compositeType,
			),
		},
	}

	var argumentLabels []string

	// TODO: support multiple overloaded initializers

	if len(compositeDeclaration.Members.Initializers) > 0 {
		firstInitializer := compositeDeclaration.Members.Initializers[0]

		argumentLabels = firstInitializer.Parameters.ArgumentLabels()

		functionType = &ConstructorFunctionType{
			FunctionType: &FunctionType{
				ParameterTypeAnnotations: parameterTypeAnnotations,
				ReturnTypeAnnotation:     NewTypeAnnotation(compositeType),
			},
		}

		checker.Elaboration.InitializerFunctionTypes[firstInitializer] = functionType
	}

	_, err := checker.valueActivations.DeclareFunction(
		compositeDeclaration.Identifier,
		functionType,
		argumentLabels,
	)
	checker.report(err)
}

func (checker *Checker) membersAndOrigins(
	fields []*ast.FieldDeclaration,
	functions []*ast.FunctionDeclaration,
	requireVariableKind bool,
) (
	members map[string]*Member,
	origins map[string]*Origin,
) {
	memberCount := len(fields) + len(functions)
	members = make(map[string]*Member, memberCount)
	origins = make(map[string]*Origin, memberCount)

	// declare a member for each field
	for _, field := range fields {
		fieldTypeAnnotation := checker.ConvertTypeAnnotation(field.TypeAnnotation)

		fieldType := fieldTypeAnnotation.Type

		checker.checkTypeAnnotation(fieldTypeAnnotation, field.TypeAnnotation.StartPos)

		identifier := field.Identifier.Identifier

		members[identifier] = &Member{
			Type:          fieldType,
			VariableKind:  field.VariableKind,
			IsInitialized: false,
		}

		origins[identifier] =
			checker.recordFieldDeclarationOrigin(field, fieldType)

		if requireVariableKind &&
			field.VariableKind == ast.VariableKindNotSpecified {

			checker.report(
				&InvalidVariableKindError{
					Kind:     field.VariableKind,
					StartPos: field.Identifier.Pos,
					EndPos:   field.Identifier.Pos,
				},
			)
		}
	}

	// declare a member for each function
	for _, function := range functions {
		functionType := checker.functionType(function.Parameters, function.ReturnTypeAnnotation)

		argumentLabels := function.Parameters.ArgumentLabels()

		identifier := function.Identifier.Identifier

		members[identifier] = &Member{
			Type:           functionType,
			VariableKind:   ast.VariableKindConstant,
			IsInitialized:  true,
			ArgumentLabels: argumentLabels,
		}

		origins[identifier] =
			checker.recordFunctionDeclarationOrigin(function, functionType)
	}

	return members, origins
}

func (checker *Checker) recordFieldDeclarationOrigin(
	field *ast.FieldDeclaration,
	fieldType Type,
) *Origin {
	startPosition := field.Identifier.StartPosition()
	endPosition := field.Identifier.EndPosition()

	origin := &Origin{
		Type:            fieldType,
		DeclarationKind: common.DeclarationKindField,
		StartPos:        &startPosition,
		EndPos:          &endPosition,
	}

	checker.Occurrences.Put(
		field.StartPos,
		field.EndPos,
		origin,
	)

	return origin
}

func (checker *Checker) recordFunctionDeclarationOrigin(
	function *ast.FunctionDeclaration,
	functionType *FunctionType,
) *Origin {
	startPosition := function.Identifier.StartPosition()
	endPosition := function.Identifier.EndPosition()

	origin := &Origin{
		Type:            functionType,
		DeclarationKind: common.DeclarationKindFunction,
		StartPos:        &startPosition,
		EndPos:          &endPosition,
	}

	checker.Occurrences.Put(
		startPosition,
		endPosition,
		origin,
	)
	return origin
}

func (checker *Checker) checkInitializers(
	initializers []*ast.InitializerDeclaration,
	fields []*ast.FieldDeclaration,
	containerType Type,
	containerDeclarationKind common.DeclarationKind,
	typeIdentifier string,
	initializerParameterTypeAnnotations []*TypeAnnotation,
	initializerKind initializerKind,
) {
	count := len(initializers)

	if count == 0 {
		checker.checkNoInitializerNoFields(fields, initializerKind, typeIdentifier)
		return
	}

	// TODO: check all initializers:
	//  parameter initializerParameterTypeAnnotations needs to be a slice

	initializer := initializers[0]
	checker.checkInitializer(
		initializer,
		fields,
		containerType,
		containerDeclarationKind,
		typeIdentifier,
		initializerParameterTypeAnnotations,
		initializerKind,
	)
}

// checkNoInitializerNoFields checks that if there are no initializers
// there are also no fields – otherwise the fields will be uninitialized.
// In interfaces this is allowed.
//
func (checker *Checker) checkNoInitializerNoFields(
	fields []*ast.FieldDeclaration,
	initializerKind initializerKind,
	typeIdentifier string,
) {
	if len(fields) == 0 || initializerKind == initializerKindInterface {
		return
	}

	// report error for first field
	firstField := fields[0]

	checker.report(
		&MissingInitializerError{
			TypeIdentifier: typeIdentifier,
			FirstFieldName: firstField.Identifier.Identifier,
			FirstFieldPos:  firstField.Identifier.Pos,
		},
	)
}

func (checker *Checker) checkInitializer(
	initializer *ast.InitializerDeclaration,
	fields []*ast.FieldDeclaration,
	containerType Type,
	containerDeclarationKind common.DeclarationKind,
	typeIdentifier string,
	initializerParameterTypeAnnotations []*TypeAnnotation,
	initializerKind initializerKind,
) {
	// NOTE: new activation, so `self`
	// is only visible inside initializer

	checker.valueActivations.Enter()
	defer checker.valueActivations.Leave()

	checker.declareSelfValue(containerType)

	// check the initializer is named properly
	identifier := initializer.Identifier.Identifier
	if identifier != InitializerIdentifier {
		checker.report(
			&InvalidInitializerNameError{
				Name: identifier,
				Pos:  initializer.StartPos,
			},
		)
	}

	functionType := &FunctionType{
		ParameterTypeAnnotations: initializerParameterTypeAnnotations,
		ReturnTypeAnnotation:     NewTypeAnnotation(&VoidType{}),
	}

	checker.checkFunction(
		initializer.Parameters,
		ast.Position{},
		functionType,
		initializer.FunctionBlock,
		true,
	)

	if initializerKind == initializerKindInterface &&
		initializer.FunctionBlock != nil {

		checker.checkInterfaceFunctionBlock(
			initializer.FunctionBlock,
			containerDeclarationKind,
			common.DeclarationKindInitializer,
		)
	}
}

func (checker *Checker) checkCompositeFunctions(
	functions []*ast.FunctionDeclaration,
	selfType *CompositeType,
) {
	for _, function := range functions {
		// NOTE: new activation, as function declarations
		// shouldn't be visible in other function declarations,
		// and `self` is is only visible inside function

		checker.valueActivations.Scoped(func() {

			checker.declareSelfValue(selfType)

			function.Accept(checker)
		})
	}
}

func (checker *Checker) declareSelfValue(selfType Type) {

	// NOTE: declare `self` one depth lower ("inside" function),
	// so it can't be re-declared by the function's parameters

	depth := checker.valueActivations.Depth() + 1

	self := &Variable{
		Kind:       common.DeclarationKindSelf,
		Type:       selfType,
		IsConstant: true,
		Depth:      depth,
		Pos:        nil,
	}
	checker.valueActivations.Set(SelfIdentifier, self)
	checker.recordVariableDeclarationOccurrence(SelfIdentifier, self)
}

// checkMemberIdentifiers checks the fields and functions are unique and aren't named `init`
//
func (checker *Checker) checkMemberIdentifiers(
	fields []*ast.FieldDeclaration,
	functions []*ast.FunctionDeclaration,
) {

	positions := map[string]ast.Position{}

	for _, field := range fields {
		checker.checkMemberIdentifier(
			field.Identifier,
			common.DeclarationKindField,
			positions,
		)
	}

	for _, function := range functions {
		checker.checkMemberIdentifier(
			function.Identifier,
			common.DeclarationKindFunction,
			positions,
		)
	}
}

func (checker *Checker) checkMemberIdentifier(
	identifier ast.Identifier,
	kind common.DeclarationKind,
	positions map[string]ast.Position,
) {
	name := identifier.Identifier
	pos := identifier.Pos

	if name == InitializerIdentifier {
		checker.report(
			&InvalidNameError{
				Name: name,
				Pos:  pos,
			},
		)
	}

	if previousPos, ok := positions[name]; ok {
		checker.report(
			&RedeclarationError{
				Name:        name,
				Pos:         pos,
				Kind:        kind,
				PreviousPos: &previousPos,
			},
		)
	} else {
		positions[name] = pos
	}
}

func (checker *Checker) VisitFieldDeclaration(field *ast.FieldDeclaration) ast.Repr {

	// NOTE: field type is already checked when determining composite function in `compositeType`

	panic(&errors.UnreachableError{})
}

func (checker *Checker) VisitInitializerDeclaration(initializer *ast.InitializerDeclaration) ast.Repr {

	// NOTE: already checked in `checkInitializer`

	panic(&errors.UnreachableError{})
}

func (checker *Checker) VisitInterfaceDeclaration(declaration *ast.InterfaceDeclaration) ast.Repr {

	interfaceType := checker.Elaboration.InterfaceDeclarationTypes[declaration]

	// TODO: also check nested composite members

	// TODO: also check nested composite members' identifiers

	checker.checkMemberIdentifiers(
		declaration.Members.Fields,
		declaration.Members.Functions,
	)

	members, origins := checker.membersAndOrigins(
		declaration.Members.Fields,
		declaration.Members.Functions,
		false,
	)

	interfaceType.Members = members

	checker.memberOrigins[interfaceType] = origins

	checker.checkMemberIdentifiers(
		declaration.Members.Fields,
		declaration.Members.Functions,
	)

	checker.checkInitializers(
		declaration.Members.Initializers,
		declaration.Members.Fields,
		interfaceType,
		declaration.DeclarationKind(),
		declaration.Identifier.Identifier,
		interfaceType.InitializerParameterTypeAnnotations,
		initializerKindInterface,
	)

	checker.checkInterfaceFunctions(
		declaration.Members.Functions,
		interfaceType,
		declaration.DeclarationKind(),
	)

	// TODO: support non-structure interfaces, such as contracts and resources

	if declaration.CompositeKind != common.CompositeKindStructure {
		checker.report(
			&UnsupportedDeclarationError{
				DeclarationKind: declaration.DeclarationKind(),
				StartPos:        declaration.Identifier.StartPosition(),
				EndPos:          declaration.Identifier.EndPosition(),
			},
		)
	}

	// TODO: support nested declarations for contracts and contract interfaces

	// report error for first nested composite declaration, if any
	if len(declaration.Members.CompositeDeclarations) > 0 {
		firstNestedCompositeDeclaration := declaration.Members.CompositeDeclarations[0]

		checker.report(
			&UnsupportedDeclarationError{
				DeclarationKind: firstNestedCompositeDeclaration.DeclarationKind(),
				StartPos:        firstNestedCompositeDeclaration.Identifier.StartPosition(),
				EndPos:          firstNestedCompositeDeclaration.Identifier.EndPosition(),
			},
		)
	}

	return nil
}

func (checker *Checker) checkInterfaceFunctions(
	functions []*ast.FunctionDeclaration,
	interfaceType Type,
	declarationKind common.DeclarationKind,
) {
	for _, function := range functions {
		// NOTE: new activation, as function declarations
		// shouldn't be visible in other function declarations,
		// and `self` is is only visible inside function

		checker.valueActivations.Scoped(func() {
			// NOTE: required for
			checker.declareSelfValue(interfaceType)

			checker.visitFunctionDeclaration(function, false)

			if function.FunctionBlock != nil {
				checker.checkInterfaceFunctionBlock(
					function.FunctionBlock,
					declarationKind,
					common.DeclarationKindFunction,
				)
			}
		})
	}
}

func (checker *Checker) declareInterfaceDeclaration(declaration *ast.InterfaceDeclaration) {

	// NOTE: fields and functions might already refer to interface itself.
	// insert a dummy type for now, so lookup succeeds during conversion,
	// then fix up the type reference

	interfaceType := &InterfaceType{}

	identifier := declaration.Identifier

	err := checker.typeActivations.Declare(identifier, interfaceType)
	checker.report(err)
	checker.recordVariableDeclarationOccurrence(
		identifier.Identifier,
		&Variable{
			Kind:       declaration.DeclarationKind(),
			IsConstant: true,
			Type:       interfaceType,
			Pos:        &identifier.Pos,
		},
	)

	// NOTE: members are added in `VisitInterfaceDeclaration` –
	//   left out for now, as field and function requirements could refer to e.g. composites
	*interfaceType = InterfaceType{
		CompositeKind: declaration.CompositeKind,
		Identifier:    identifier.Identifier,
	}

	// TODO: support multiple overloaded initializers

	var parameterTypeAnnotations []*TypeAnnotation
	initializerCount := len(declaration.Members.Initializers)
	if initializerCount > 0 {
		firstInitializer := declaration.Members.Initializers[0]
		parameterTypeAnnotations = checker.parameterTypeAnnotations(firstInitializer.Parameters)

		if initializerCount > 1 {
			secondInitializer := declaration.Members.Initializers[1]

			checker.report(
				&UnsupportedOverloadingError{
					DeclarationKind: common.DeclarationKindInitializer,
					StartPos:        secondInitializer.StartPosition(),
					EndPos:          secondInitializer.EndPosition(),
				},
			)
		}
	}

	interfaceType.InitializerParameterTypeAnnotations = parameterTypeAnnotations

	checker.Elaboration.InterfaceDeclarationTypes[declaration] = interfaceType

	// declare value

	checker.declareInterfaceMetaType(declaration, interfaceType)
}

func (checker *Checker) checkInterfaceFunctionBlock(
	block *ast.FunctionBlock,
	containerKind common.DeclarationKind,
	implementedKind common.DeclarationKind,
) {

	if len(block.Statements) > 0 {
		checker.report(
			&InvalidImplementationError{
				Pos:             block.Statements[0].StartPosition(),
				ContainerKind:   containerKind,
				ImplementedKind: implementedKind,
			},
		)
	} else if len(block.PreConditions) == 0 &&
		len(block.PostConditions) == 0 {

		checker.report(
			&InvalidImplementationError{
				Pos:             block.StartPos,
				ContainerKind:   containerKind,
				ImplementedKind: implementedKind,
			},
		)
	}
}

func (checker *Checker) declareInterfaceMetaType(
	declaration *ast.InterfaceDeclaration,
	interfaceType *InterfaceType,
) {
	metaType := &InterfaceMetaType{
		InterfaceType: interfaceType,
	}

	_, err := checker.valueActivations.Declare(
		declaration.Identifier.Identifier,
		metaType,
		// TODO: check
		declaration.DeclarationKind(),
		declaration.Identifier.Pos,
		true,
		nil,
	)
	checker.report(err)
}

func (checker *Checker) recordVariableReferenceOccurrence(startPos, endPos ast.Position, variable *Variable) {
	origin, ok := checker.variableOrigins[variable]
	if !ok {
		origin = &Origin{
			Type:            variable.Type,
			DeclarationKind: variable.Kind,
			StartPos:        variable.Pos,
			// TODO:
			EndPos: variable.Pos,
		}
		checker.variableOrigins[variable] = origin
	}
	checker.Occurrences.Put(startPos, endPos, origin)
}

func (checker *Checker) recordVariableDeclarationOccurrence(name string, variable *Variable) {
	if variable.Pos == nil {
		return
	}
	startPos := *variable.Pos
	endPos := variable.Pos.Shifted(len(name) - 1)
	checker.recordVariableReferenceOccurrence(startPos, endPos, variable)
}

func (checker *Checker) VisitImportDeclaration(declaration *ast.ImportDeclaration) ast.Repr {

	imports := checker.Program.Imports()
	imported := imports[declaration.Location]
	if imported == nil {
		checker.report(
			&UnresolvedImportError{
				ImportLocation: declaration.Location,
				StartPos:       declaration.LocationPos,
				EndPos:         declaration.LocationPos,
			},
		)
		return nil
	}

	if checker.seenImports[declaration.Location] {
		checker.report(
			&RepeatedImportError{
				ImportLocation: declaration.Location,
				StartPos:       declaration.LocationPos,
				EndPos:         declaration.LocationPos,
			},
		)
		return nil
	}
	checker.seenImports[declaration.Location] = true

	importChecker, ok := checker.ImportCheckers[declaration.Location]
	var checkerErr *CheckerError
	if !ok || importChecker == nil {
		var err error
		importChecker, err = NewChecker(
			imported,
			checker.PredeclaredValues,
			checker.PredeclaredTypes,
		)
		if err == nil {
			checker.ImportCheckers[declaration.Location] = importChecker
		}
	}

	// NOTE: ignore generic `error` result, get internal *CheckerError
	_ = importChecker.Check()
	checkerErr = importChecker.checkerError()

	if checkerErr != nil {
		checker.report(
			&ImportedProgramError{
				CheckerError:   checkerErr,
				ImportLocation: declaration.Location,
				Pos:            declaration.LocationPos,
			},
		)
		return nil
	}

	missing := make(map[ast.Identifier]bool, len(declaration.Identifiers))
	for _, identifier := range declaration.Identifiers {
		missing[identifier] = true
	}

	checker.importValues(declaration, importChecker, missing)
	checker.importTypes(declaration, importChecker, missing)

	for identifier, _ := range missing {
		checker.report(
			&NotExportedError{
				Name:           identifier.Identifier,
				ImportLocation: declaration.Location,
				Pos:            identifier.Pos,
			},
		)

		// NOTE: declare constant variable with invalid type to silence rest of program
		_, err := checker.valueActivations.Declare(
			identifier.Identifier,
			&InvalidType{},
			common.DeclarationKindValue,
			identifier.Pos,
			true,
			nil,
		)
		checker.report(err)

		// NOTE: declare type with invalid type to silence rest of program
		err = checker.typeActivations.Declare(identifier, &InvalidType{})
		checker.report(err)
	}

	return nil
}

func (checker *Checker) importValues(
	declaration *ast.ImportDeclaration,
	importChecker *Checker,
	missing map[ast.Identifier]bool,
) {
	// TODO: consider access modifiers

	// determine which identifiers are imported /
	// which variables need to be declared

	var variables map[string]*Variable
	identifierLength := len(declaration.Identifiers)
	if identifierLength > 0 {
		variables = make(map[string]*Variable, identifierLength)
		for _, identifier := range declaration.Identifiers {
			name := identifier.Identifier
			variable := importChecker.GlobalValues[name]
			if variable == nil {
				continue
			}
			variables[name] = variable
			delete(missing, identifier)
		}
	} else {
		variables = importChecker.GlobalValues
	}

	for name, variable := range variables {

		// TODO: improve position
		// TODO: allow cross-module variables?

		// don't import predeclared values
		if _, ok := importChecker.PredeclaredValues[name]; ok {
			continue
		}

		_, err := checker.valueActivations.Declare(
			name,
			variable.Type,
			variable.Kind,
			declaration.LocationPos,
			true,
			variable.ArgumentLabels,
		)
		checker.report(err)
	}
}

func (checker *Checker) importTypes(
	declaration *ast.ImportDeclaration,
	importChecker *Checker,
	missing map[ast.Identifier]bool,
) {
	// TODO: consider access modifiers

	// determine which identifiers are imported /
	// which types need to be declared

	var types map[string]Type
	identifierLength := len(declaration.Identifiers)
	if identifierLength > 0 {
		types = make(map[string]Type, identifierLength)
		for _, identifier := range declaration.Identifiers {
			name := identifier.Identifier
			ty := importChecker.GlobalTypes[name]
			if ty == nil {
				continue
			}
			types[name] = ty
			delete(missing, identifier)
		}
	} else {
		types = importChecker.GlobalTypes
	}

	for name, ty := range types {

		// TODO: improve position
		// TODO: allow cross-module types?

		// don't import predeclared values
		if _, ok := importChecker.PredeclaredValues[name]; ok {
			continue
		}

		identifier := ast.Identifier{
			Identifier: name,
			Pos:        declaration.LocationPos,
		}
		err := checker.typeActivations.Declare(identifier, ty)
		checker.report(err)
	}
}

func (checker *Checker) VisitFailableDowncastExpression(expression *ast.FailableDowncastExpression) ast.Repr {

	leftHandExpression := expression.Expression
	leftHandType := leftHandExpression.Accept(checker).(Type)

	rightHandTypeAnnotation := checker.ConvertTypeAnnotation(expression.TypeAnnotation)
	checker.checkTypeAnnotation(rightHandTypeAnnotation, expression.TypeAnnotation.StartPos)

	rightHandType := rightHandTypeAnnotation.Type

	checker.Elaboration.FailableDowncastingTypes[expression] = rightHandType

	// TODO: non-Any types (interfaces, wrapped (e.g Any?, [Any], etc.)) are not supported for now

	if _, ok := leftHandType.(*AnyType); !ok {

		checker.report(
			&UnsupportedTypeError{
				Type:     leftHandType,
				StartPos: leftHandExpression.StartPosition(),
				EndPos:   leftHandExpression.EndPosition(),
			},
		)
	}

	return &OptionalType{Type: rightHandType}
}

func (checker *Checker) VisitCreateExpression(expression *ast.CreateExpression) ast.Repr {
	// TODO: check invocation expression of resources is in creation

	checker.inCreate = true
	defer func() {
		checker.inCreate = false
	}()

	ty := expression.InvocationExpression.Accept(checker)

	// NOTE: not using `isResourceType`,
	// as only direct resource types can be constructed
	if compositeType, ok := ty.(*CompositeType); !ok ||
		compositeType.Kind != common.CompositeKindResource {

		checker.report(
			&InvalidConstructionError{
				StartPos: expression.InvocationExpression.StartPosition(),
				EndPos:   expression.InvocationExpression.EndPosition(),
			},
		)
	}

	return ty
}

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

	if _, ok := expression.Expression.(*ast.IdentifierExpression); ok {
		// TODO: record destruction of resource
	}

	return
}
