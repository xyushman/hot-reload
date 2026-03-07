.PHONY: build run-demo test clean

build:
	go build -o hotreload.exe ./cmd/hotreload

run-demo: build
	hotreload.exe --root ./testserver --build "go build -o testserver/server.exe ./testserver/main.go" --exec "testserver\server.exe"

test:
	go test ./...

clean:
	cmd /c "if exist hotreload.exe del hotreload.exe"
	cmd /c "if exist testserver\server.exe del testserver\server.exe"
