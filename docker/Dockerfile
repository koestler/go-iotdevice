# build backend
FROM golang:1.19 as go-builder

RUN apt-get update && apt-get install -y git

WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN  VERSION=`git describe --always --tags`; \
     BUILD_TIME=`date -Is`; \
     go build -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=$BUILD_TIME" -o /go-iotdevice

# used to generate /etc/passwd /etc/group
RUN groupadd -g 1000 app && \
    useradd -r -u 1000 -g app app
RUN chown app:app /go-iotdevice

# use already compiled frontend
FROM koestler/js-iotdevice:v1.2.3 AS js-builder

# build final image
FROM scratch
USER app
COPY --from=go-builder /go-iotdevice         /go-iotdevice
COPY --from=go-builder /etc/group            /etc/group
COPY --from=go-builder /etc/passwd           /etc/passwd
COPY --from=js-builder /usr/share/nginx/html /frontend-build
CMD ["/go-iotdevice", "-c", "/config.yaml"]
