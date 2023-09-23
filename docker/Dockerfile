# build backend
FROM golang:1.21-alpine as go-builder

RUN apk add build-base git

WORKDIR /app
COPY . /app

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN  ./build.sh

# used to generate /etc/passwd /etc/group
RUN addgroup -S app && adduser -S app -G app
RUN chown app:app /app/go-iotdevice

# use already compiled frontend
FROM ghcr.io/koestler/js-iotdevice:v2.0.9 AS js-builder

# build final image
FROM scratch
USER app
COPY --from=go-builder /app/go-iotdevice     /go-iotdevice
COPY --from=go-builder /etc/group            /etc/group
COPY --from=go-builder /etc/passwd           /etc/passwd
COPY --from=js-builder /usr/share/nginx/html /frontend-build
CMD ["/go-iotdevice", "-c", "/config.yaml"]
