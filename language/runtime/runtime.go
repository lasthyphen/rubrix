package runtime

import (
	"errors"
	"fmt"

	"github.com/dapperlabs/flow-go/language/runtime/ast"
	runtimeErrors "github.com/dapperlabs/flow-go/language/runtime/errors"
	"github.com/dapperlabs/flow-go/language/runtime/interpreter"
	"github.com/dapperlabs/flow-go/language/runtime/parser"
	"github.com/dapperlabs/flow-go/language/runtime/sema"
	"github.com/dapperlabs/flow-go/language/runtime/stdlib"
	"github.com/dapperlabs/flow-go/language/runtime/trampoline"
	"github.com/dapperlabs/flow-go/sdk/abi/values"
)

type Interface interface {
	// ResolveImport resolves an import of a program.
	ResolveImport(Location) ([]byte, error)
	// GetValue gets a value for the given key in the storage, controlled and owned by the given accounts.
	GetValue(owner, controller, key []byte) (value []byte, err error)
	// SetValue sets a value for the given key in the storage, controlled and owned by the given accounts.
	SetValue(owner, controller, key, value []byte) (err error)
	// CreateAccount creates a new account with the given public keys and code.
	CreateAccount(publicKeys []values.Bytes, code values.Bytes) (address values.Address, err error)
	// AddAccountKey appends a key to an account.
	AddAccountKey(address values.Address, publicKey values.Bytes) error
	// RemoveAccountKey removes a key from an account by index.
	RemoveAccountKey(address values.Address, index values.Int) (publicKey values.Bytes, err error)
	// UpdateAccountCode updates the code associated with an account.
	UpdateAccountCode(address values.Address, code values.Bytes) (err error)
	// GetSigningAccounts returns the signing accounts.
	GetSigningAccounts() []values.Address
	// Log logs a string.
	Log(string)
	// EmitEvent is called when an event is emitted by the runtime.
	EmitEvent(values.Event)
}

// Runtime is a runtime capable of executing the Flow programming language.
type Runtime interface {
	// ExecuteScript executes the given script.
	//
	// This function returns an error if the program has errors (e.g syntax errors, type errors),
	// or if the execution fails.
	ExecuteScript(script []byte, runtimeInterface Interface, location Location) (values.Value, error)

	// ExecuteTransaction executes the given transaction.
	//
	// This function returns an error if the program has errors (e.g syntax errors, type errors),
	// or if the execution fails.
	ExecuteTransaction(script []byte, runtimeInterface Interface, location Location) error

	// ParseAndCheckProgram parses and checks the given code without executing the program.
	//
	// This function returns an error if the program contains any syntax or semantic errors.
	ParseAndCheckProgram(code []byte, runtimeInterface Interface, location Location) error
}

var typeDeclarations = stdlib.BuiltinTypes.ToTypeDeclarations()

type ImportResolver = func(astLocation ast.Location) (program *ast.Program, e error)

// interpreterRuntime is a interpreter-based version of the Flow runtime.
type interpreterRuntime struct{}

// NewInterpreterRuntime returns a interpreter-based version of the Flow runtime.
func NewInterpreterRuntime() Runtime {
	return &interpreterRuntime{}
}

func (r *interpreterRuntime) ExecuteScript(script []byte, runtimeInterface Interface, location Location) (values.Value, error) {
	functions := r.standardLibraryFunctions(runtimeInterface)

	checker, err := r.parseAndCheckProgram(script, runtimeInterface, location, functions)
	if err != nil {
		return nil, newError(err)
	}

	_, ok := checker.GlobalValues["main"]
	if !ok {
		// TODO: error because no main?
		return nil, nil
	}

	runtimeStorage := newInterpreterRuntimeStorage(runtimeInterface)

	inter, err := r.newInterpreter(checker, functions, runtimeInterface, runtimeStorage)
	if err != nil {
		return nil, newError(err)
	}

	if err := inter.Interpret(); err != nil {
		return nil, newError(err)
	}

	value, err := inter.Invoke("main")
	if err != nil {
		return nil, newError(err)
	}

	// Write back all stored values, which were actually just cached, back into storage
	runtimeStorage.writeCached()

	return value.(interpreter.ExportableValue).Export(), nil
}

func (r *interpreterRuntime) ExecuteTransaction(
	script []byte,
	runtimeInterface Interface,
	location Location,
) error {
	functions := r.standardLibraryFunctions(runtimeInterface)

	checker, err := r.parseAndCheckProgram(script, runtimeInterface, location, functions)
	if err != nil {
		return newError(err)
	}

	transactions := checker.TransactionTypes
	transactionCount := len(transactions)
	if transactionCount != 1 {
		return newError(InvalidTransactionCountError{Count: transactionCount})
	}

	transactionType := transactions[0]
	transactionFunctionType := transactionType.EntryPointFunctionType()

	signingAccountAddresses := runtimeInterface.GetSigningAccounts()

	// check parameter count

	signingAccountsCount := len(signingAccountAddresses)
	transactionFunctionParameterCount := len(transactionFunctionType.ParameterTypeAnnotations)
	if signingAccountsCount != transactionFunctionParameterCount {
		return newError(InvalidTransactionParameterCountError{
			Expected: transactionFunctionParameterCount,
			Actual:   signingAccountsCount,
		})
	}

	// check parameter types

	for _, parameterTypeAnnotation := range transactionFunctionType.ParameterTypeAnnotations {
		parameterType := parameterTypeAnnotation.Type

		if !parameterType.Equal(&sema.AccountType{}) {
			return newError(InvalidTransactionParameterTypeError{
				Actual: parameterType,
			})
		}
	}

	runtimeStorage := newInterpreterRuntimeStorage(runtimeInterface)

	inter, err := r.newInterpreter(checker, functions, runtimeInterface, runtimeStorage)
	if err != nil {
		return newError(err)
	}

	if err := inter.Interpret(); err != nil {
		return newError(err)
	}

	signingAccounts := make([]interface{}, signingAccountsCount)

	for i, address := range signingAccountAddresses {
		signingAccounts[i] = interpreter.NewAccountValue(interpreter.AddressValue(address))
	}

	err = inter.InvokeTransaction(0, signingAccounts...)
	if err != nil {
		return newError(err)
	}

	// Write back all stored values, which were actually just cached, back into storage
	runtimeStorage.writeCached()

	return nil
}

func (r *interpreterRuntime) ParseAndCheckProgram(script []byte, runtimeInterface Interface, location Location) error {
	functions := r.standardLibraryFunctions(runtimeInterface)
	_, err := r.parseAndCheckProgram(script, runtimeInterface, location, functions)
	if err != nil {
		return newError(err)
	}

	return nil
}

func (r *interpreterRuntime) parseAndCheckProgram(
	script []byte,
	runtimeInterface Interface,
	location Location,
	functions stdlib.StandardLibraryFunctions,
) (*sema.Checker, error) {
	program, err := r.parse(script)
	if err != nil {
		return nil, err
	}

	importResolver := r.importResolver(runtimeInterface)
	err = program.ResolveImports(importResolver)
	if err != nil {
		return nil, err
	}

	valueDeclarations := functions.ToValueDeclarations()

	checker, err := sema.NewChecker(
		program,
		location,
		sema.WithPredeclaredValues(valueDeclarations),
		sema.WithPredeclaredTypes(typeDeclarations),
	)
	if err != nil {
		return nil, err
	}

	if err := checker.Check(); err != nil {
		return nil, err
	}

	return checker, nil
}

func (r *interpreterRuntime) newInterpreter(
	checker *sema.Checker,
	functions stdlib.StandardLibraryFunctions,
	runtimeInterface Interface,
	runtimeStorage *interpreterRuntimeStorage,
) (*interpreter.Interpreter, error) {
	return interpreter.NewInterpreter(
		checker,
		interpreter.WithPredefinedValues(functions.ToValues()),
		interpreter.WithOnEventEmittedHandler(func(inter *interpreter.Interpreter, event interpreter.EventValue) {
			r.emitEvent(inter, runtimeInterface, event)
		}),
		interpreter.WithStorageReadHandler(runtimeStorage.readValue),
		interpreter.WithStorageWriteHandler(runtimeStorage.writeValue),
		interpreter.WithStorageKeyHandlerFunc(func(_ *interpreter.Interpreter, _ string, indexingType sema.Type) string {
			return indexingType.String()
		}),
	)
}

func (r *interpreterRuntime) standardLibraryFunctions(runtimeInterface Interface) stdlib.StandardLibraryFunctions {
	return append(
		stdlib.FlowBuiltInFunctions(stdlib.FlowBuiltinImpls{
			CreateAccount:     r.newCreateAccountFunction(runtimeInterface),
			AddAccountKey:     r.addAccountKeyFunction(runtimeInterface),
			RemoveAccountKey:  r.removeAccountKeyFunction(runtimeInterface),
			UpdateAccountCode: r.newUpdateAccountCodeFunction(runtimeInterface),
			GetAccount:        r.newGetAccountFunction(runtimeInterface),
			Log:               r.newLogFunction(runtimeInterface),
		}),
		stdlib.BuiltinFunctions...,
	)
}

func (r *interpreterRuntime) importResolver(runtimeInterface Interface) ImportResolver {
	return func(astLocation ast.Location) (program *ast.Program, e error) {
		var location Location
		switch astLocation := astLocation.(type) {
		case ast.StringLocation:
			location = StringLocation(astLocation)
		case ast.AddressLocation:
			location = AddressLocation(astLocation)
		default:
			panic(runtimeErrors.NewUnreachableError())
		}
		script, err := runtimeInterface.ResolveImport(location)
		if err != nil {
			return nil, err
		}
		return r.parse(script)
	}
}

func (r *interpreterRuntime) parse(script []byte) (program *ast.Program, err error) {
	program, _, err = parser.ParseProgram(string(script))
	return
}

// emitEvent converts an event value to native Go types and emits it to the runtime interface.
func (r *interpreterRuntime) emitEvent(
	inter *interpreter.Interpreter,
	runtimeInterface Interface,
	event interpreter.EventValue,
) {
	functionType := inter.Checker.GlobalValues[event.Identifier].Type.(*sema.SpecialFunctionType)
	eventType := functionType.ReturnTypeAnnotation.Type.(*sema.EventType).Export(nil, nil)

	eventValue := event.Export().(values.Event)
	eventValue = eventValue.WithType(eventType)

	runtimeInterface.EmitEvent(eventValue)
}

func (r *interpreterRuntime) emitAccountEvent(
	eventType sema.EventType,
	runtimeInterface Interface,
	eventFields ...values.Value,
) {
	t := eventType.Export(nil, nil)
	eventValue := values.NewEvent(eventFields).WithType(t)

	runtimeInterface.EmitEvent(eventValue)
}

func (r *interpreterRuntime) newCreateAccountFunction(runtimeInterface Interface) interpreter.HostFunction {
	return func(arguments []interpreter.Value, _ interpreter.LocationPosition) trampoline.Trampoline {
		pkArray := arguments[0].(*interpreter.ArrayValue)
		pkValues := pkArray.Values
		publicKeys := make([]values.Bytes, len(pkValues))

		for i, pkVal := range pkValues {
			publicKey, err := toBytes(pkVal)
			if err != nil {
				panic(fmt.Sprintf("createAccount requires the first parameter to be an array of arrays"))
			}
			publicKeys[i] = publicKey
		}

		code, err := toBytes(arguments[1])
		if err != nil {
			panic(fmt.Sprintf("createAccount requires the third parameter to be an array"))
		}

		accountAddress, err := runtimeInterface.CreateAccount(publicKeys, code)
		if err != nil {
			panic(err)
		}

		r.emitAccountEvent(stdlib.AccountCreatedEventType, runtimeInterface, accountAddress)

		result := interpreter.AddressValue(accountAddress)
		return trampoline.Done{Result: result}
	}
}

func (r *interpreterRuntime) addAccountKeyFunction(runtimeInterface Interface) interpreter.HostFunction {
	return func(arguments []interpreter.Value, _ interpreter.LocationPosition) trampoline.Trampoline {
		accountAddress := arguments[0].(interpreter.AddressValue)
		publicKey, err := toBytes(arguments[1])
		if err != nil {
			panic(fmt.Sprintf("addAccountKey requires the second parameter to be an array"))
		}

		accountAddressValue := accountAddress.Export().(values.Address)

		err = runtimeInterface.AddAccountKey(accountAddressValue, publicKey)
		if err != nil {
			panic(err)
		}

		r.emitAccountEvent(stdlib.AccountKeyAddedEventType, runtimeInterface, accountAddressValue, publicKey)

		result := &interpreter.VoidValue{}
		return trampoline.Done{Result: result}
	}
}

func (r *interpreterRuntime) removeAccountKeyFunction(runtimeInterface Interface) interpreter.HostFunction {
	return func(arguments []interpreter.Value, _ interpreter.LocationPosition) trampoline.Trampoline {
		accountAddress := arguments[0].(interpreter.AddressValue)
		index := arguments[1].(interpreter.IntValue)

		accountAddressValue := accountAddress.Export().(values.Address)

		indexValue := index.Export().(values.Int)

		publicKey, err := runtimeInterface.RemoveAccountKey(accountAddressValue, indexValue)
		if err != nil {
			panic(err)
		}

		r.emitAccountEvent(stdlib.AccountKeyRemovedEventType, runtimeInterface, accountAddressValue, publicKey)

		result := &interpreter.VoidValue{}
		return trampoline.Done{Result: result}
	}
}

func (r *interpreterRuntime) newUpdateAccountCodeFunction(runtimeInterface Interface) interpreter.HostFunction {
	return func(arguments []interpreter.Value, _ interpreter.LocationPosition) trampoline.Trampoline {
		accountAddress := arguments[0].(interpreter.AddressValue)

		code, err := toBytes(arguments[1])
		if err != nil {
			panic(fmt.Sprintf("updateAccountCode requires the second parameter to be an array"))
		}

		accountAddressValue := accountAddress.Export().(values.Address)

		err = runtimeInterface.UpdateAccountCode(accountAddressValue, code)
		if err != nil {
			panic(err)
		}

		r.emitAccountEvent(stdlib.AccountCodeUpdatedEventType, runtimeInterface, accountAddressValue, code)

		result := &interpreter.VoidValue{}
		return trampoline.Done{Result: result}
	}
}

func (r *interpreterRuntime) newGetAccountFunction(runtimeInterface Interface) interpreter.HostFunction {
	return func(arguments []interpreter.Value, _ interpreter.LocationPosition) trampoline.Trampoline {
		accountAddress := arguments[0].(interpreter.AddressValue)
		publicAccount := interpreter.NewPublicAccountValue(accountAddress)
		return trampoline.Done{Result: publicAccount}
	}
}

func (r *interpreterRuntime) newLogFunction(runtimeInterface Interface) interpreter.HostFunction {
	return func(arguments []interpreter.Value, _ interpreter.LocationPosition) trampoline.Trampoline {
		runtimeInterface.Log(fmt.Sprint(arguments[0]))
		return trampoline.Done{Result: &interpreter.VoidValue{}}
	}
}

func toBytes(value interpreter.Value) (values.Bytes, error) {
	_, isNil := value.(interpreter.NilValue)
	if isNil {
		return values.Bytes{}, nil
	}

	someValue, ok := value.(*interpreter.SomeValue)
	if ok {
		value = someValue.Value
	}

	array, ok := value.(*interpreter.ArrayValue)
	if !ok {
		return nil, errors.New("value is not an array")
	}

	result := make([]byte, len(array.Values))
	for i, arrayValue := range array.Values {
		intValue, ok := arrayValue.(interpreter.IntValue)
		if !ok {
			return nil, errors.New("array value is not an Int")
		}

		j := intValue.IntValue()

		if j < 0 || j > 255 {
			return nil, errors.New("array value is not in byte range (0-255)")
		}

		result[i] = byte(j)
	}

	return result, nil
}
