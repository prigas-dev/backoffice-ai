package operations

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/afero"
)

type IOperationStore interface {
	GetOperation(operationName string) (*Operation, error)
	AddOperation(operation *Operation) error
}

var ErrOperationNotFound = errors.New("operation not found")

type InMemoryOperationStore struct {
	operations map[string]*Operation
}

func NewInMemoryOperationStore() IOperationStore {
	return &InMemoryOperationStore{
		operations: map[string]*Operation{},
	}
}

func (s *InMemoryOperationStore) GetOperation(operationName string) (*Operation, error) {
	operation, operationExists := s.operations[operationName]
	if !operationExists {
		return nil, ErrOperationNotFound
	}

	return operation, nil
}

func (s *InMemoryOperationStore) AddOperation(operation *Operation) error {
	s.operations[operation.Name] = operation
	return nil
}

type FsOperationStore struct {
	fs afero.Fs
}

func NewFsOperationStore(fs afero.Fs) IOperationStore {
	return &FsOperationStore{
		fs: fs,
	}
}

func (s *FsOperationStore) GetOperation(operationName string) (*Operation, error) {
	fileName := fmt.Sprintf("%s.json", operationName)

	file, err := s.fs.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", fileName, err)
	}
	defer file.Close()

	operation := Operation{}
	err = json.NewDecoder(file).Decode(&operation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operation json from file %s: %w", fileName, err)
	}

	return &operation, nil
}

func (s *FsOperationStore) AddOperation(operation *Operation) error {
	fileName := fmt.Sprintf("%s.json", operation.Name)

	file, err := s.fs.Create(fileName)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", fileName, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(operation)
	if err != nil {
		return fmt.Errorf("failed to write operation json to file %s: %w", fileName, err)
	}

	return nil
}
