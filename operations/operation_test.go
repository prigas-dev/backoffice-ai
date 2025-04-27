package operations_test

import (
	"encoding/json"
	"testing"

	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/stretchr/testify/assert"
)

func TestValueSchema(t *testing.T) {
	t.Parallel()

	t.Run("json serialization", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc        string
			valueSchema *operations.ValueSchema
		}{
			{
				desc: "string",
				valueSchema: &operations.ValueSchema{
					Type: operations.String,
					Spec: &operations.StringSpec{},
				},
			},
			{
				desc: "integer",
				valueSchema: &operations.ValueSchema{
					Type: operations.Number,
					Spec: &operations.NumberSpec{},
				},
			},
			{
				desc: "boolean",
				valueSchema: &operations.ValueSchema{
					Type: operations.Boolean,
					Spec: &operations.BooleanSpec{},
				},
			},
			{
				desc: "object",
				valueSchema: &operations.ValueSchema{
					Type: operations.Object,
					Spec: &operations.ObjectSpec{
						Properties: map[string]*operations.ValueSchema{
							"prop": {
								Type: operations.Boolean,
								Spec: &operations.BooleanSpec{},
							},
						},
					},
				},
			},
			{
				desc: "array",
				valueSchema: &operations.ValueSchema{
					Type: operations.Array,
					Spec: &operations.ArraySpec{
						Items: &operations.ValueSchema{
							Type: operations.Boolean,
							Spec: &operations.BooleanSpec{},
						},
					},
				},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				t.Parallel()

				encoded, err := json.Marshal(&tC.valueSchema)
				assert.NoError(t, err)

				var decoded operations.ValueSchema
				err = json.Unmarshal(encoded, &decoded)
				assert.NoError(t, err)

				assert.Equal(t, tC.valueSchema, &decoded)

			})
		}
	})

	t.Run("validations", func(t *testing.T) {
		testCases := []struct {
			desc                     string
			spec                     operations.Spec
			value                    any
			expectedValidationResult operations.ValidationResult
		}{
			{
				desc:                     "valid string",
				spec:                     &operations.StringSpec{},
				value:                    "prigas",
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc:                     "valid string: nil",
				spec:                     &operations.StringSpec{Nullable: true},
				value:                    nil,
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc:                     "invalid string",
				spec:                     &operations.StringSpec{},
				value:                    12,
				expectedValidationResult: operations.ValidationResult{Success: false, Message: "value is not a string"},
			},
			{
				desc:                     "valid number: float64",
				spec:                     &operations.NumberSpec{},
				value:                    12.5,
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc:                     "valid number: int64",
				spec:                     &operations.NumberSpec{},
				value:                    int64(12),
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc:                     "valid number: nil",
				spec:                     &operations.NumberSpec{Nullable: true},
				value:                    nil,
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc:                     "invalid number",
				spec:                     &operations.NumberSpec{},
				value:                    "",
				expectedValidationResult: operations.ValidationResult{Success: false, Message: "value is not a float64 or int64"},
			},
			{
				desc:                     "valid boolean",
				spec:                     &operations.BooleanSpec{},
				value:                    false,
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc:                     "valid boolean: nil",
				spec:                     &operations.BooleanSpec{Nullable: true},
				value:                    nil,
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc:                     "invalid boolean",
				spec:                     &operations.BooleanSpec{},
				value:                    "",
				expectedValidationResult: operations.ValidationResult{Success: false, Message: "value is not a bool"},
			},
			{
				desc: "valid object",
				spec: &operations.ObjectSpec{
					Properties: map[string]*operations.ValueSchema{
						"strProp": {
							Type: operations.String,
							Spec: &operations.StringSpec{},
						},
						"boolProp": {
							Type: operations.Boolean,
							Spec: &operations.BooleanSpec{},
						},
					},
				},
				value: map[string]any{
					"strProp":  "prigas",
					"boolProp": false,
				},
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc: "valid object: nil",
				spec: &operations.ObjectSpec{
					Properties: map[string]*operations.ValueSchema{},
					Nullable:   true,
				},
				value:                    nil,
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc: "invalid object: not a map",
				spec: &operations.ObjectSpec{
					Properties: map[string]*operations.ValueSchema{},
				},
				value:                    "",
				expectedValidationResult: operations.ValidationResult{Success: false, Message: "value is not a map"},
			},
			{
				desc: "invalid object: missing property",
				spec: &operations.ObjectSpec{
					Properties: map[string]*operations.ValueSchema{
						"prigas": {
							Type: operations.Boolean,
							Spec: &operations.BooleanSpec{},
						},
					},
				},
				value:                    map[string]any{},
				expectedValidationResult: operations.ValidationResult{Success: false, Message: "missing property prigas"},
			},
			{
				desc: "invalid object: invalid property",
				spec: &operations.ObjectSpec{
					Properties: map[string]*operations.ValueSchema{
						"prigas": {
							Type: operations.Boolean,
							Spec: &operations.BooleanSpec{},
						},
					},
				},
				value: map[string]any{
					"prigas": "non bool",
				},
				expectedValidationResult: operations.ValidationResult{Success: false, Message: "invalid property prigas: value is not a bool"},
			},
			{
				desc: "valid array",
				spec: &operations.ArraySpec{
					Items: &operations.ValueSchema{
						Type: operations.Number,
						Spec: &operations.NumberSpec{},
					},
				},
				value:                    []any{12.0},
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc: "valid array: nil",
				spec: &operations.ArraySpec{
					Nullable: true,
					Items: &operations.ValueSchema{
						Type: operations.Number,
						Spec: &operations.NumberSpec{},
					},
				},
				value:                    nil,
				expectedValidationResult: operations.ValidationResult{Success: true},
			},
			{
				desc: "invalid array: invalid item",
				spec: &operations.ArraySpec{
					Items: &operations.ValueSchema{
						Type: operations.Number,
						Spec: &operations.NumberSpec{},
					},
				},
				value:                    []any{12.0, "not a number"},
				expectedValidationResult: operations.ValidationResult{Success: false, Message: "invalid item 1: value is not a float64 or int64"},
			},
		}
		for _, tC := range testCases {
			t.Run(tC.desc, func(t *testing.T) {
				t.Parallel()

				actualValidationResult := tC.spec.Validate(tC.value)

				assert.Equal(t, tC.expectedValidationResult, actualValidationResult)
			})
		}
	})
}
