#!/bin/bash

# BUILD

# Get Go version
GO_VERSION=$(go version | awk '{print $3}')

# Get the build date
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Navigate to the root of the project from the scripts folder
cd ..

# Build command
go build -o melodix -ldflags "-s -X github.com/keshon/melodix-player/internal/version.BuildDate=$BUILD_DATE -X github.com/keshon/melodix-player/internal/version.GoVersion=$GO_VERSION" cmd/melodix/melodix.go

# Return to the scripts folder after execution
cd scripts