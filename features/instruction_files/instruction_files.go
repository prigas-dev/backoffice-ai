package instruction_files

import _ "embed"

type File struct {
	MarkdownLanguageIdentifier string
	Filename                   string
	Content                    string
}

//go:embed system-instructions-v3.txt
var systemInstructionsTemplateContent string
var SystemInstructionsTemplate = &File{
	MarkdownLanguageIdentifier: "plaintext",
	Filename:                   "system-instructions-v2.txt",
	Content:                    systemInstructionsTemplateContent,
}

//go:embed feature_schema.json
var featureJSONSchemaContent string
var FeatureJSONSchema = &File{
	MarkdownLanguageIdentifier: "json",
	Filename:                   "feature_schema.json",
	Content:                    featureJSONSchemaContent,
}

//go:embed llm_error_schema.json
var errorJSONSchemaContent string
var ErrorJSONSchema = &File{
	MarkdownLanguageIdentifier: "json",
	Filename:                   "llm_error_schema.json",
	Content:                    errorJSONSchemaContent,
}

//go:embed example/feature.json
var exampleFeatureJSONContent string
var ExampleFeatureJSON = &File{
	MarkdownLanguageIdentifier: "json",
	Filename:                   "feature.json",
	Content:                    exampleFeatureJSONContent,
}

//go:embed example/Component.tsx
var exampleComponentContent string

//go:embed example/get-username.js
var exampleGetUsernameOperationContent string

//go:embed example/update-username.js
var exampleUpdateUsernameOperationContent string

var ExampleFeatureFiles = []File{
	{
		MarkdownLanguageIdentifier: "typescriptreact",
		Filename:                   "Component.tsx",
		Content:                    exampleComponentContent,
	},
	{
		MarkdownLanguageIdentifier: "javascript",
		Filename:                   "get-username.js",
		Content:                    exampleGetUsernameOperationContent,
	},
	{
		MarkdownLanguageIdentifier: "javascript",
		Filename:                   "update-username.js",
		Content:                    exampleUpdateUsernameOperationContent,
	},
}
