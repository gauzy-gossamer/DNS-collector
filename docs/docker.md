# DNS-collector - Docker Deployment Guide

- [Docker](#docker)
- [Docker Compose](#docker-compose-recommended)

## Quick Start with Docker Compose (Recommended)

Create a directory of your choice (e.g. ./dnscollector) to hold the docker-compose.yml and .env files.

```bash
mkdir ./dnscollector
cd ./dnscollector
```

Download docker-compose.yml and docker-example.env, either by running the following commands:

```bash
wget https://github.com/dmachard/DNS-collector/releases/latest/download/docker-compose.yml
wget -O .env https://github.com/dmachard/DNS-collector/releases/latest/download/docker-example.env
```

Populate the .env file with custom values:

- Update DNSCOLLECTOR_DATA with your preferred location for storing DNS logs.

Start the containers using docker compose command

```bash
docker compose up -d
```


## Basic docker run

Docker run with a custom configuration:

```bash
docker run -d dmachard/dnscollector -v $(pwd)/config.yml:/etc/dnscollector/config.yml
```
