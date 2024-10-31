#!/bin/bash

go build .
scp versequick-users-api berinaniesh.xyz:/home/berinaniesh/tmp/
ssh berinaniesh.xyz deploy-versequick-users-api.sh
