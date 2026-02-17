APP_NAME=gravity

build-wasm:
	GOOS=js GOARCH=wasm go build -o bin/$(APP_NAME).wasm .

build:
	go build -o bin/$(APP_NAME) .

run: build
	./bin/$(APP_NAME)

