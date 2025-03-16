all: build run

build:
	go build -o dist/StartPage.bin cmd/main.go

run:
	./dist/StartPage.bin --port 8080 --config ./config/config.yml

windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o dist/StartPage.exe cmd/main.go

clean:
	rm ./dist/StartPage ./dist/StartPage.exe
