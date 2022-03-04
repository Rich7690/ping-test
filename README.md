# ping-test

Application that allows you to call an ip check endpoint (checkip.amazonaws.com) to get your public ip address and measure the latency of that call.
The data is then exporter prometheus style on `/metrics`. By default listens on port `2112`. Eventually these parameters may be configurable.

This application is packaged as a container you can run via docker, kubernetes, or anything else that accepts these kinds of containers:

`docker pull ghcr.io/rtdev7690/ping-test:latest`