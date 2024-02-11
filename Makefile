APP_NAME=gravity

build-wasm:
	GOOS=js GOARCH=wasm go build -o bin/$(APP_NAME).wasm ./main.go

build:
	go build -o bin/$(APP_NAME) main.go

run: build
	./bin/$(APP_NAME)

