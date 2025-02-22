all: build run

build:
	go build -o dist/StartPage cmd/main.go

run:
	./dist/StartPage
