#!/bin/bash

# Check if "traefik" argument is provided
if [ "$1" == "traefik" ]; then
    DOCKER_COMPOSE_COMMAND="docker-compose -f docker-compose.yml -f docker-compose.traefik.yml up -d"
else
    DOCKER_COMPOSE_COMMAND="docker-compose -f docker-compose.yml up -d"
fi

# Read .env
if [ -f .env ]; then
    # Source the .env file to set environment variables
    source .env
else
    echo ".env file not found!"
    exit 1  # Exit with an error code
fi

# Git repo
if [ "$GIT" != "false" ]; then
    # Remove old git project
    rm -rf ./src
    # Make a new git clone
    git clone "$GIT_URL" src
else
    if [ ! -d "./src" ]; then
        echo "src dir not found!"
        exit 1  # Exit with an error code
    fi
fi

# Docker
# - Remove old container (if it exists)
docker-compose down

# - Remove old image (if it exists)
if [ "$(docker images -q "${ALIAS}-image" 2>/dev/null)" ]; then
    docker rmi "${ALIAS}-image"
fi

# - Build new docker image from Dockerfile using BuildKit for parallel builds
DOCKER_BUILDKIT=1 docker build -t "${ALIAS}-image" .

# Start new container using docker-compose based on the selected command
eval "$DOCKER_COMPOSE_COMMAND"

# Remove unused images
docker image prune -a