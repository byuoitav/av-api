#!/usr/bin/env bash
# Pasted into Jenkins to build (will eventually be fleshed out to work with a Docker Hub and Amazon AWS)

echo "Stopping running application"
docker stop av-api
docker rm av-api

echo "Building container"
docker build -t byuoitav/av-api .

echo "Starting the new version"
docker run -d -e EMS_API_USERNAME=$EMS_API_USERNAME -e EMS_API_PASSWORD=$EMS_API_PASSWORD --restart=always --name av-api -p 8000:8000 byuoitav/av-api:latest
