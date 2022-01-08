# go-victron-to-mqtt
This deamon reads various values from Victron Energy devices and publishes them on an MQTT-Server.
In addition, a http endpoint and a web-frontend is provided to show the values.

## Development

### Compile and run inside docker
```bash
docker build -f docker/Dockerfile -t go-victron-to-mqtt .
docker run --rm --name go-victron-to-mqtt -p 127.0.0.1:8000:8000 \
  -v "$(pwd)"/documentation/config.yaml:/config.yaml:ro \
  go-victron-to-mqtt
```

### run tests
```bash
go install github.com/golang/mock/mockgen@v1.6.0
go genreate ./...
go test ./...
```

### Update README.md
```bash
npx embedme README.md
```
