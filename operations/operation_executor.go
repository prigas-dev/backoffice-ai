package operations

import (
	"fmt"

	"github.com/prigas-dev/backoffice-ai/utils"
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
	for _, parameter := range operation.Parameters {
		value, hasValue := arguments[parameter.Name]
		if !hasValue {
			return nil, fmt.Errorf("argument not provided: %s", parameter.Name)
		}
		validationResult := parameter.TypeProperties.Validate(value)
		if !validationResult.Success {
			return nil, fmt.Errorf("invalid argument %s: %s", parameter.Name, validationResult.Message)
		}
	}

	argumentsList := utils.Map(operation.Parameters,
		func(parameter *ValueSchema) any {
			value, ok := arguments[parameter.Name]
			if !ok {
				return nil
			}
			return value
		})

	result, err := ExecuteJavascript[any](operationName, operation.JavascriptCode, argumentsList)
	if err != nil {
		return nil, err
	}

	return result, nil
}
