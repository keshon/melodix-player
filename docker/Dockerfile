# First stage: build the Go application
FROM golang:1.23.0-alpine3.20 AS build

RUN apk add --no-cache build-base
ADD src /usr/project
WORKDIR /usr/project

# Combine commands for better layer caching
RUN GO_VERSION=$(go version | awk '{print $3}') && \
    BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    CGO_ENABLED=1 && \
    rm -f .env && \
    go build -o app -ldflags "-X 'github.com/keshon/melodix-player/internal/version.BuildDate=${BUILD_DATE}' -X 'github.com/keshon/melodix-player/internal/version.GoVersion=${GO_VERSION}'" cmd/melodix/melodix.go && \
    touch database.db && \
    apk del build-base && \
    rm -rf /var/cache/apk/* && \
    rm -rf /usr/share/man

# Second stage: create the final image
FROM alpine:3.20

RUN apk update && \
    apk add --no-cache ffmpeg

COPY --from=build /usr/project /usr/project

WORKDIR /usr/project

ENTRYPOINT ["/usr/project/app"]