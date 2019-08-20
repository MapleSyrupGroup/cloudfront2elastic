# cloudfront2elastic
A lambda script for migrating Cloudfront logs from S3 to an elastic cluster.

## Purpose
We wanted a way to easily analyse our Cloudfront logs and decided to utilise an Elastic cluster to do this.
The solution we wanted needed to be simple and automated, we have a large volume of logs so we also need to 
be able to control the quantity of logs held on the elastic cluster.

## Solution
We created a Lambda script which is triggered when a new file is created in the S3 bucket used to store Cloudfront logs.
The log files are gzipped so the code needs to decompress them first.
The data is then converted into a Json string in the format used for bulk inserts into Elastic.
In order to be able to efficiently purge logs we use a template file at Elastic to create a new index for each day.
We can then purge indexes after N days if the volume becomes to great.
We chose not to import every field in the logs only those that were of interest but it can be easily extended to add other fields.
There is a check to see if traffic comes from a crawler bot and those are highlighted.
The list of bots is the ones we most commonly see but again this could be extended.

## Parameters
The code requires 4 environmental variable to be set on the Lambda function.
elastic_url - The location of your Elastic instance
elastic_user - The username to log into your Elastic instance
elastic_password - The password needed to log into your Elastic instance
aws_s3_region - The region name of the S3 bucket where your logs reside.

## Build
Amazon requires Go packages for Lambda to be built to a specific environment.
env GOOS=linux GOARCH=amd64 go build -o main lambda.go 
Note our package name here 'main' must match the value of the function name entered when setting up the Lambda function in AWS
The package needs to be zipped before upload
zip -j main.zip main

## Elastic security
Suggest creating a user who only has rights to create indexes and write data to minimise the risks
Don't leave an Elastic insistence on the internet with out security.
Elastic.co have cloud based instances you can use for work like this.

## What's in here
This repository contains the Lambda script and the template for Elastic.

## Helpful Amazon Documentation
Using Lambda with S3: https://docs.aws.amazon.com/lambda/latest/dg/with-s3.html
Lambda permissions: https://docs.aws.amazon.com/lambda/latest/dg/lambda-intro-execution-role.html
