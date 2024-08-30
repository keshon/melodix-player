#!/bin/bash

if [ "$1" == "traefik" ]; then
    DOCKER_COMPOSE_COMMAND="docker compose -f docker-compose.yml -f docker-compose.traefik.yml up -d"
else
    DOCKER_COMPOSE_COMMAND="docker compose -f docker-compose.yml up -d"
fi

if [ -f .env ]; then
    source .env
else
    echo ".env file not found!"
    exit 1
fi

if [ "$GIT" != "false" ]; then
    rm -rf ./src
    git clone "$GIT_URL" src
else
    if [ ! -d "./src" ]; then
        echo "src dir not found!"
        exit 1
    fi
fi

docker-compose down

if [ "$(docker images -q "${ALIAS}-image" 2>/dev/null)" ]; then
    docker rmi "${ALIAS}-image"
fi

DOCKER_BUILDKIT=1 docker build -t "${ALIAS}-image" .

eval "$DOCKER_COMPOSE_COMMAND"

docker system prune --all