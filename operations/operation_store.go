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
	fs OperationsFs
}
type OperationsFs afero.Fs

func NewFsOperationStore(fs OperationsFs) IOperationStore {
	return &FsOperationStore{
		fs: fs,
	}
}

func (s *FsOperationStore) GetOperation(operationName string) (*Operation, error) {
	manifestFileName := fmt.Sprintf("%s/operation_manifest.json", operationName)

	manifestFile, err := s.fs.Open(manifestFileName)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", manifestFileName, err)
	}
	defer manifestFile.Close()

	operationManifest := OperationManifest{}
	err = json.NewDecoder(manifestFile).Decode(&operationManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse operation manifest json from file %s: %w", manifestFileName, err)
	}

	javscriptCodeFileName := fmt.Sprintf("%s/operation.js", operationName)
	javascriptCode, err := afero.ReadFile(s.fs, javscriptCodeFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read operation javascript coe from file %s: %w", javscriptCodeFileName, err)
	}

	operation := &Operation{
		Name:           operationManifest.Name,
		JavascriptCode: string(javascriptCode),
		Parameters:     operationManifest.Parameters,
		Return:         operationManifest.Return,
	}

	return operation, nil
}

func (s *FsOperationStore) AddOperation(operation *Operation) error {

	err := s.fs.MkdirAll(operation.Name, 0755)
	if err != nil {
		return fmt.Errorf("failed to create operation %s folder: %w", operation.Name, err)
	}

	manifestFileName := fmt.Sprintf("%s/operation_manifest.json", operation.Name)

	file, err := s.fs.Create(manifestFileName)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", manifestFileName, err)
	}
	defer file.Close()

	operationManifest := OperationManifest{
		Name:       operation.Name,
		Parameters: operation.Parameters,
		Return:     operation.Return,
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(operationManifest)
	if err != nil {
		return fmt.Errorf("failed to write operation manifest to file %s: %w", manifestFileName, err)
	}

	javscriptCodeFileName := fmt.Sprintf("%s/operation.js", operation.Name)
	err = afero.WriteFile(s.fs, javscriptCodeFileName, []byte(operation.JavascriptCode), 0755)
	if err != nil {
		return fmt.Errorf("failed to write operation javascript code to file %s: %w", javscriptCodeFileName, err)
	}

	return nil
}
