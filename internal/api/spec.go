package api

import _ "embed"

//go:embed openapi.yaml
var OpenAPISpec []byte

//go:embed docs.html
var OpenAPIHTML []byte
