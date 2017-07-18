#!/usr/bin/env bash

if [ $1 = "development" ]; then
	curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"/development
elif [ $1 = "testing" ]; then
	curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"/testing
elif [ $1 = "stage" ]; then
	curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"/stage
elif [ $1 = "production" ]; then
	curl -X GET --header "Accept: application/json" --header "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_HEADER" "$RASPI_DEPLOYMENT_MICROSERVICE_WSO2_ADDRESS"/production
else
	echo "Either 'development', 'testing', 'stage', or 'production' must be passed as the sole argument"
fi
