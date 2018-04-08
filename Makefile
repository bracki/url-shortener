build: 
	env GOOS=linux go build -ldflags="-s -w" -o bin/create-url create-url/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/get-url get-url/main.go

clean:
	rm -fr bin
