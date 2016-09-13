#!/usr/bin/env bash

curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"
