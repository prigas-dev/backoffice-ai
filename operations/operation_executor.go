package operations

import (
	"database/sql"
	"fmt"
)

type IOperationExecutor interface {
	Execute(operationName string, arguments map[string]any) (any, error)
}

func NewOperationExecutor(db *sql.DB, store IOperationStore) IOperationExecutor {
	return &OperationExecutor{
		store: store,
		db:    db,
	}
}

type OperationExecutor struct {
	store IOperationStore
	db    *sql.DB
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

	globals := map[string]any{
		"query": func(query string, parameters ...any) ([][]any, error) {
			rows, err := o.db.Query(query)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				return nil, err
			}

			scannedRows := [][]any{}
			for rows.Next() {
				values := make([]any, len(columns))
				scanArgs := make([]any, len(columns))
				for i := range values {
					scanArgs[i] = &values[i]
				}

				err := rows.Scan(scanArgs...)
				if err != nil {
					return nil, err
				}

				for i, val := range values {
					// Ensuring byte slices are converted to string
					if b, ok := val.([]byte); ok {
						values[i] = string(b)
					}
				}

				scannedRows = append(scannedRows, values)
			}
			err = rows.Err()
			if err != nil {
				return nil, err
			}

			return scannedRows, nil

		},
	}

	result, err := ExecuteJavascript[any](operationName, operation.JavascriptCode, arguments, globals)
	if err != nil {
		return nil, err
	}

	resultValidationResult := operation.Return.Spec.Validate(result)
	if !resultValidationResult.Success {
		return nil, fmt.Errorf("invalid result: %s", resultValidationResult.Message)
	}

	return result, nil
}
