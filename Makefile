# perl -pi -e 's/^  */\t/' Makefile
BINARY_NAME=termleek
# export CGO_ENABLED=0

$(shell if [ ! -d "./bin" ]; then mkdir -p "./bin"; fi)

all: clean build_small

build:
	go build -o bin/${BINARY_NAME}.bin *.go

build_small:
	go build -ldflags "-s -w" -o bin/${BINARY_NAME}.bin *.go
	upx bin/${BINARY_NAME}.bin

run: ./bin/${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm -f bin/${BINARY_NAME}.bin
