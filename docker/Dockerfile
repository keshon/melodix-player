# First stage: build the Go application
FROM golang:1.21.1-alpine3.17 AS mybuildstage

RUN apk add --no-cache build-base
ADD src /usr/project
WORKDIR /usr/project

# Run a single command to set variables, remove .env file, and build the Go application
RUN GO_VERSION=$(go version | awk '{print $3}') && \
    BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    CGO_ENABLED=1 && \
    rm -f .env && \
    go build -o app -ldflags "-X 'app/internal/version.BuildDate=${BUILD_DATE}' -X 'app/internal/version.GoVersion=${GO_VERSION}'" cmd/main.go && \
    mv melodix.db.empty melodix.db

# Second stage: create the final image
FROM alpine:3.17

RUN apk update && \
    apk add --no-cache ffmpeg

COPY --from=mybuildstage /usr/project /usr/project

# Set the entrypoint to run the app
ENTRYPOINT ["/usr/project/app"]