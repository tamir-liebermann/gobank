#!/bin/sh

# Start the Go application
./docker-gobank &

# Start Ngrok and make it tunnel to port 5252
ngrok http 5252 --authtoken $NGROK_AUTHTOKEN