# build backend
FROM golang:1.20-alpine as go-builder

RUN apk add build-base git

WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN  VERSION=`git describe --always --tags`; \
     BUILD_TIME=`date -Is`; \
     echo "version="$VERSION; \
     go build -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=$BUILD_TIME" -o /go-iotdevice

# used to generate /etc/passwd /etc/group
RUN addgroup -S app && adduser -S app -G app
RUN chown app:app /go-iotdevice

# use already compiled frontend
FROM ghcr.io/koestler/js-iotdevice:v2.0.9 AS js-builder

# build final image
FROM scratch
USER app
COPY --from=go-builder /go-iotdevice         /go-iotdevice
COPY --from=go-builder /etc/group            /etc/group
COPY --from=go-builder /etc/passwd           /etc/passwd
COPY --from=js-builder /usr/share/nginx/html /frontend-build
CMD ["/go-iotdevice", "-c", "/config.yaml"]
