package examples

import _ "embed"

//go:embed username/feature.json
var FeatureJSON string

var ComponentFilename = "Component.tsx"

//go:embed username/Component.tsx
var ComponentFileContent string

var GetUsernameFilename = "get-username.js"

//go:embed username/get-username.js
var GetUsernameFileContent string

var UpdateUsernameFilename = "update-username.js"

//go:embed username/update-username.js
var UpdateUsernameFileContent string
