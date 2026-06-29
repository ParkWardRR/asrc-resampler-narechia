.PHONY: all build test fuzz clean libs

all: libs build

libs:
	cd asrc-rs && cargo build --release

build: libs
	go build -v ./...

test: libs
	go test -v ./...

fuzz: libs
	go test -fuzz=Fuzz -fuzztime=10s

clean:
	cd asrc-rs && cargo clean
	go clean
