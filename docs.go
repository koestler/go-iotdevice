package main

//go:generate swag init -g docs.go -d internal/httpServer/ --outputTypes json,yaml
//go:generate npx @redocly/cli@2.69.0 build-docs docs/swagger.yaml -o docs/swagger.html
