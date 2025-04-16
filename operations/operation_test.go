package operations_test

import (
	"encoding/json"
	"testing"

	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/stretchr/testify/assert"
)

func TestOperation(t *testing.T) {
	t.Parallel()

	t.Run("value schema json serialization", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc        string
			valueSchema *operations.ValueSchema
		}{
			{
				desc: "string",
				valueSchema: &operations.ValueSchema{
					Type:           operations.String,
					TypeProperties: &operations.StringProperties{},
				},
			},
			{
				desc: "integer",
				valueSchema: &operations.ValueSchema{
					Type:           operations.Number,
					TypeProperties: &operations.NumberProperties{},
				},
			},
			{
				desc: "boolean",
				valueSchema: &operations.ValueSchema{
					Type:           operations.Boolean,
					TypeProperties: &operations.BooleanProperties{},
				},
			},
			{
				desc: "object",
				valueSchema: &operations.ValueSchema{
					Type: operations.Object,
					TypeProperties: &operations.ObjectProperties{
						Properties: map[string]*operations.ValueSchema{
							"prop": {
								Type:           operations.Boolean,
								TypeProperties: &operations.BooleanProperties{},
							},
						},
					},
				},
			},
			{
				desc: "array",
				valueSchema: &operations.ValueSchema{
					Type: operations.Array,
					TypeProperties: &operations.ArrayProperties{
						Items: &operations.ValueSchema{
							Type:           operations.Boolean,
							TypeProperties: &operations.BooleanProperties{},
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
}
