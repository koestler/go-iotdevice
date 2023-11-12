package main

//go:generate swag init
//go:generate rm docs/docs.go
//go:generate npx @redocly/cli build-docs docs/swagger.yaml -o docs/swagger.html
