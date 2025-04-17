package operations_test

import (
	"testing"

	"github.com/prigas-dev/backoffice-ai/operations"
)

func TestOperationExecutor(t *testing.T) {
	t.Parallel()

	store := operations.NewInMemoryOperationStore()
	store.AddOperation(&operations.Operation{
		Name:       "single_value",
		Parameters: map[string]*operations.ValueSchema{},
		JavascriptCode: `function () {
			return 1
		}`,
	})
}
