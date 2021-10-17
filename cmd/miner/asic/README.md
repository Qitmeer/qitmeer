# compile

     CGO_ENABLED=1 GOOS=linux GOARCH=arm CC=arm-linux-gnueabihf-gcc go build -ldflags="-extldflags=-static" -a