#!/bin/bash

eval set GOARCH=amd64
eval set GOOS=windows

eval go build -o goub-win-amd64

eval set GOOS=linux

eval go build -o goub-linux-amd64