package operations_test

import (
	"testing"

	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFsOperationStore(t *testing.T) {
	t.Run("store and retrieve", func(t *testing.T) {
		t.Parallel()

		files := afero.NewMemMapFs()
		store := operations.NewFsOperationStore(files)

		operation := &operations.Operation{
			Name:           "op",
			JavascriptCode: `12`,
			Parameters: map[string]*operations.ValueSchema{
				"p": {
					Type: operations.String,
					Spec: &operations.StringSpec{},
				},
			},
			Return: &operations.ValueSchema{
				Type: operations.Number,
				Spec: &operations.NumberSpec{},
			},
		}

		err := store.AddOperation(operation)
		assert.NoError(t, err)

		storeOperation, err := store.GetOperation("op")
		assert.NoError(t, err)

		assert.Equal(t, operation, storeOperation)
	})
}
