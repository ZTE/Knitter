#!/bin/bash

echo "-------> mkdir reports"
mkdir reports
echo "-------> docker pull ppodgorsek/robot-framework:latest"
docker pull ppodgorsek/robot-framework:latest
echo "-------> docker run rf test..."
docker run \
    -v $PWD/reports:/opt/robotframework/reports:Z \
    -v $PWD/tests:/opt/robotframework/tests:Z \
    -e BROWSER=chrome \
    ppodgorsek/robot-framework:latest

### show report content
for file in $PWD/reports/*
do
    echo $file
done

