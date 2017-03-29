#!/usr/bin/env bash

if [ $1 = "development" ]; then
	curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"_development
elif [ $1 = "stage" ]; then
	curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"_stage
elif [ $1 = "production" ]; then
	curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"_production
else
	echo "Either 'development', 'stage', or 'production' must be passed as the sole argument"
fi
