package operations

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dop251/goja"
)

func ExecuteJavascript[T any](name string, script string) (T, error) {
	// TODO cache compiled scripts
	vm := goja.New()
	defer vm.Interrupt("halt")

	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	var zero T

	v, err := vm.RunString("Error")
	if err != nil {
		return zero, fmt.Errorf("failed to parse Error class")
	}
	Error := v.ToObject(vm)

	runResult, err := vm.RunString(script)
	if err != nil {
		return zero, err
	}
	err, isResultError := unwrapError(vm, Error, runResult)
	if isResultError {
		return zero, err
	}

	promise, isPromise := runResult.Export().(*goja.Promise)
	if isPromise {
		promiseResult := promise.Result()

		if promise.State() == goja.PromiseStateRejected {
			err, isResultError := unwrapError(vm, Error, promiseResult)
			if isResultError {
				return zero, err
			}

			errorResult, err := castJSValue[any](vm, promiseResult)
			if err != nil {
				return zero, err
			}

			return zero, fmt.Errorf("%+v", errorResult)
		}

		result, err := castJSValue[T](vm, promiseResult)
		if err != nil {
			return zero, err
		}

		return result, nil
	}

	result, err := castJSValue[T](vm, runResult)
	if err != nil {
		return zero, err
	}

	return result, nil
}

func unwrapError(vm *goja.Runtime, Error *goja.Object, value goja.Value) (error, bool) {
	if vm.InstanceOf(value, Error) {
		errorObject := value.ToObject(vm)

		message := errorObject.Get("message").Export().(string)
		stack := errorObject.Get("stack").Export().(string)
		errorClassName := errorObject.ClassName()

		return fmt.Errorf("%s: %s\n%s", errorClassName, message, stack), true
	}

	return nil, false
}

var ErrInvalidCast = errors.New("invalid cast")

func castJSValue[T any](vm *goja.Runtime, jsValue goja.Value) (T, error) {

	var zero T

	castType := reflect.TypeFor[T]()
	if castType == reflect.TypeFor[string]() {
		goValue := jsValue.Export()
		_, isString := goValue.(string)
		if !isString {
			return zero, ErrInvalidCast
		}
		return goValue.(T), nil
	}

	result := new(T)

	err := vm.ExportTo(jsValue, result)
	if err != nil {
		return zero, err
	}

	return *result, nil
}
