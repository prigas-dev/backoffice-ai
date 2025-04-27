package schemas

import _ "embed"

//go:embed feature_schema.json
var FeatureSchema string

//go:embed llm_error_schema.json
var ErrorSchema string
