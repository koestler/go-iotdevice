# build backend
FROM golang:1.23-alpine AS go-builder

RUN apk add build-base git npm

WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go generate docs.go
RUN  ./build.sh

# used to generate /etc/passwd /etc/group
RUN addgroup -S app && adduser -S app -G app
RUN chown app:app /app/go-iotdevice

# add gpio group
RUN addgroup -g 993 gpio

# use already compiled frontend
FROM ghcr.io/koestler/js-iotdevice:v2.1.0 AS js-builder

# build final image
FROM scratch
USER app
COPY --from=go-builder /app/go-iotdevice     /go-iotdevice
COPY --from=go-builder /app/docs             /docs
COPY --from=go-builder /etc/group            /etc/group
COPY --from=go-builder /etc/passwd           /etc/passwd
COPY --from=js-builder /usr/share/nginx/html /frontend-build
CMD ["/go-iotdevice", "-c", "/config.yaml"]
