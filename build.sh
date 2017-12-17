cd httpServer
go generate
cd -

CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w'
