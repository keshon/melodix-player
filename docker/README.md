# Docker Deployment

The deployment is organized with Docker Compose, allowing for simple and extended deployment options. The extended deployment includes Traefik for proxy support.

## Prerequisites

Before you proceed, ensure you have the following prerequisites:

- Docker and Docker Compose installed
- Git (if you intend to pull the source code from a Git repository)
- A Discord bot token acquired from the Discord Developer Portal
- Traefik proxy is set up and ready (optional).

## Configuration

### Environment Variables

The deployment relies on environment variables, which can be configured in the `.env` file.

#### Deployment Key Environment Variables

- `ALIAS`: Docker container name.
- `HOST`: Hostname for the API gateway (only usable with `docker-compose.traefik.yml`).

### Traefik Configuration (Optional)

If you intend to use Traefik for proxy support, make sure that Traefik is properly set up and the `docker-compose.traefik.yml` file is configured with the desired settings.

## Deployment

### Simple Deployment

To deploy the app without Traefik and with the simple configuration, run:

```bash
docker-compose -f docker-compose.yml up -d
```

### Extended Deployment with Traefik

To deploy the app with Traefik for proxy support, use the `docker-compose.traefik.yml` file:

```bash
docker-compose -f docker-compose.yml -f docker-compose.traefik.yml up -d
```

## Build and Deploy Script

For easy deployment and updates, you can use the `build-and-deploy.sh` script. This script reads the environment variables from the `.env` file and automates the build and deployment process. Run it as follows:

```bash
./build-and-deploy.sh
```

Alternatively, you can use the `build-and-deploy.sh` script with the "traefik" argument to trigger the extended deployment:

```bash
./build-and-deploy.sh traefik
```