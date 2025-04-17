package operations_test

import (
	"testing"

	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/stretchr/testify/assert"
)

func TestOperationExecutor(t *testing.T) {
	t.Parallel()

	t.Run("operation not found", func(t *testing.T) {
		t.Parallel()

		store := operations.NewInMemoryOperationStore()
		executor := operations.NewOperationExecutor(store)

		_, err := executor.Execute("op", map[string]any{})

		assert.EqualError(t, err, "operation not found")
	})

	t.Run("simple value returns", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc                string
			jsCode              string
			returnSchema        *operations.ValueSchema
			expectedReturnValue any
		}{
			{
				desc:   "string",
				jsCode: `function run() { return "banana" }`,
				returnSchema: &operations.ValueSchema{
					Type: operations.String,
					Spec: &operations.StringSpec{},
				},
				expectedReturnValue: "banana",
			},
			{
				desc:   "int64",
				jsCode: `function run() { return 1 }`,
				returnSchema: &operations.ValueSchema{
					Type: operations.Number,
					Spec: &operations.NumberSpec{},
				},
				expectedReturnValue: int64(1),
			},
			{
				desc:   "float64",
				jsCode: `function run() { return 12.1 }`,
				returnSchema: &operations.ValueSchema{
					Type: operations.Number,
					Spec: &operations.NumberSpec{},
				},
				expectedReturnValue: float64(12.1),
			},
			{
				desc:   "bool",
				jsCode: `function run() { return false }`,
				returnSchema: &operations.ValueSchema{
					Type: operations.Boolean,
					Spec: &operations.BooleanSpec{},
				},
				expectedReturnValue: false,
			},
			{
				desc:   "object",
				jsCode: `function run() { return { prop: "prigas" } }`,
				returnSchema: &operations.ValueSchema{
					Type: operations.Object,
					Spec: &operations.ObjectSpec{
						Properties: map[string]*operations.ValueSchema{
							"prop": {
								Type: operations.String,
								Spec: &operations.StringSpec{},
							},
						},
					},
				},
				expectedReturnValue: map[string]any{
					"prop": "prigas",
				},
			},
			{
				desc:   "array",
				jsCode: `function run() { return ["1", "😏😏"] }`,
				returnSchema: &operations.ValueSchema{
					Type: operations.Array,
					Spec: &operations.ArraySpec{
						Items: &operations.ValueSchema{
							Type: operations.String,
							Spec: &operations.StringSpec{},
						},
					},
				},
				expectedReturnValue: []any{"1", "😏😏"},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				t.Parallel()

				store := operations.NewInMemoryOperationStore()
				store.AddOperation(&operations.Operation{
					Name:           "simple_return",
					Parameters:     map[string]*operations.ValueSchema{},
					JavascriptCode: tC.jsCode,
					Return:         tC.returnSchema,
				})
				executor := operations.NewOperationExecutor(store)

				value, err := executor.Execute("simple_return", map[string]any{})
				assert.NoError(t, err)

				assert.Equal(t, tC.expectedReturnValue, value)
			})
		}
	})

	t.Run("argument not provided", func(t *testing.T) {
		t.Parallel()

		store := operations.NewInMemoryOperationStore()
		store.AddOperation(&operations.Operation{
			Name: "argument_not_provided",
			Parameters: map[string]*operations.ValueSchema{
				"stuff": {
					Type: operations.String,
					Spec: &operations.StringSpec{},
				},
			},
		})
		executor := operations.NewOperationExecutor(store)

		_, err := executor.Execute("argument_not_provided", map[string]any{})

		assert.EqualError(t, err, "argument not provided: stuff")
	})

	t.Run("invalid argument", func(t *testing.T) {
		t.Parallel()

		store := operations.NewInMemoryOperationStore()
		store.AddOperation(&operations.Operation{
			Name: "invalid_argument",
			Parameters: map[string]*operations.ValueSchema{
				"stuff": {
					Type: operations.String,
					Spec: &operations.StringSpec{},
				},
			},
		})
		executor := operations.NewOperationExecutor(store)

		_, err := executor.Execute("invalid_argument", map[string]any{
			"stuff": 12,
		})

		assert.EqualError(t, err, "invalid argument stuff: value is not a string")
	})

	t.Run("arguments are passed", func(t *testing.T) {
		store := operations.NewInMemoryOperationStore()
		store.AddOperation(&operations.Operation{
			Name: "arguments_are_passed",
			Parameters: map[string]*operations.ValueSchema{
				"prigas": {
					Type: operations.String,
					Spec: &operations.StringSpec{},
				},
			},
			Return: &operations.ValueSchema{
				Type: operations.Number,
				Spec: &operations.NumberSpec{},
			},
			JavascriptCode: `function run({ prigas }) { return prigas.length }`,
		})

		executor := operations.NewOperationExecutor(store)

		result, err := executor.Execute("arguments_are_passed", map[string]any{
			"prigas": "prigas",
		})
		assert.NoError(t, err)

		assert.Equal(t, int64(6), result)
	})
}
