#!/usr/bin/env bash

PROJECT_NAME=$1
SHA1=$2 # Nab the SHA1 of the desired build from a command-line argument
EB_BUCKET=elasticbeanstalk-us-west-2-194925301021

# Create new Elastic Beanstalk version
aws configure set default.region us-west-2
aws configure set region us-west-2
aws s3 cp Dockerrun.aws.json s3://$EB_BUCKET/Dockerrun.aws.json # Copy the Dockerrun file to the S3 bucket
aws elasticbeanstalk create-application-version --application-name $PROJECT_NAME --version-label $SHA1 --source-bundle S3Bucket=$EB_BUCKET,S3Key=Dockerrun.aws.json

# Update Elastic Beanstalk environment to new version
aws elasticbeanstalk update-environment --environment-name $PROJECT_NAME-env --version-label $SHA1
