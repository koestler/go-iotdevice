# build backend
FROM golang:1.17 as go-builder

RUN apt-get update && apt-get install -y git upx
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app
COPY . /app

RUN go generate

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN  VERSION=`git describe --always --tags`; \
     BUILD_TIME=`date -Is`; \
     go build -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=$BUILD_TIME" -o /go-iotdevice

RUN upx /go-iotdevice

# used to generate /etc/passwd /etc/group
RUN groupadd -g 1000 app && \
    useradd -r -u 1000 -g app app
RUN chown app:app /go-iotdevice

# use already compiled frontend
FROM koestler/js-iotdevice:main AS js-builder
# todo: swtich version from latest to v1

# build final image
FROM scratch
USER app
COPY --from=go-builder /go-iotdevice         /go-iotdevice
COPY --from=go-builder /etc/group            /etc/group
COPY --from=go-builder /etc/passwd           /etc/passwd
COPY --from=js-builder /usr/share/nginx/html /frontend-build
CMD ["/go-iotdevice", "-c", "/config.yaml"]
