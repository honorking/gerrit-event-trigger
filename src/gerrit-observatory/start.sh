#!/bin/bash

chmod 400 .id_rsa
test -d ./logs || mkdir ./logs
nohup ./gerrit-observatory > ./logs/gerrit-observatory.log 2>&1 &
