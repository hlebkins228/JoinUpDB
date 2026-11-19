package api

import _ "embed"

//go:generate go run ./generate

//go:embed openapi.yaml
var Specification []byte
