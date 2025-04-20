#!/bin/bash

go build .
scp users-api berinaniesh.xyz:/home/berinaniesh/tmp/
ssh berinaniesh.xyz deploy-users-api.sh
