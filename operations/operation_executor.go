package operations

import (
	"fmt"
)

type IOperationExecutor interface {
	Execute(operationName string, arguments map[string]any) (any, error)
}

func NewOperationExecutor(store IOperationStore) IOperationExecutor {
	return &OperationExecutor{
		store: store,
	}
}

type OperationExecutor struct {
	store IOperationStore
}

func (o *OperationExecutor) Execute(operationName string, arguments map[string]any) (any, error) {
	operation, err := o.store.GetOperation(operationName)
	if err != nil {
		return nil, err
	}
	for parameterName, parameter := range operation.Parameters {
		value, hasValue := arguments[parameterName]
		if !hasValue {
			return nil, fmt.Errorf("argument not provided: %s", parameterName)
		}
		validationResult := parameter.Spec.Validate(value)
		if !validationResult.Success {
			return nil, fmt.Errorf("invalid argument %s: %s", parameterName, validationResult.Message)
		}
	}

	result, err := ExecuteJavascript[any](operationName, operation.JavascriptCode, arguments)
	if err != nil {
		return nil, err
	}

	resultValidationResult := operation.Return.Spec.Validate(result)
	if !resultValidationResult.Success {
		return nil, fmt.Errorf("invalid result: %s", resultValidationResult.Message)
	}

	return result, nil
}
