VERSION=`git describe --always --tags`
BUILD_TIME=`date -Is`
GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -trimpath -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=$BUILD_TIME"
