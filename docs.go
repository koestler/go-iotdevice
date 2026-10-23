package main

//go:generate swag init -g docs.go -d internal/httpServer/
//go:generate rm docs/docs.go
//go:generate npx @redocly/cli@2.69.0 build-docs docs/swagger.yaml -o docs/swagger.html
