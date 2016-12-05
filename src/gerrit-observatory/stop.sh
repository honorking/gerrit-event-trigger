#!/bin/bash

if ps -ef | grep -v python | grep -v make | grep gerrit-observatory | grep -v grep
then
	killall gerrit-observatory
fi
