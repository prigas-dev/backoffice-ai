package operations

import "errors"

type IOperationStore interface {
	GetOperation(operationName string) (*Operation, error)
}

var ErrOperationNotFound = errors.New("operation not found")

type InMemoryOperationStore struct {
	operations map[string]*Operation
}

func NewInMemoryOperationStore() *InMemoryOperationStore {
	return &InMemoryOperationStore{}
}

func (s *InMemoryOperationStore) GetOperation(operationName string) (*Operation, error) {
	operation, operationExists := s.operations[operationName]
	if !operationExists {
		return nil, ErrOperationNotFound
	}

	return operation, nil
}

func (s *InMemoryOperationStore) AddOperation(operation *Operation) {
	s.operations[operation.Name] = operation
}
