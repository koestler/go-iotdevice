VERSION=`git describe --always --tags`
BUILD_TIME=`date -Is`
echo "version="$VERSION;

go build -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=$BUILD_TIME" \
    -tags=jsoniter