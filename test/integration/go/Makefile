clean:
	rm -rf test/integration/go/bin/*

precross: clean
	go mod tidy
	go build -o bin/testclient src/bin/testclient/main.go
	go build -o bin/testserver src/bin/testserver/main.go
	go build -o bin/testpublisher src/bin/testpublisher/main.go
	go build -o bin/testsubscriber src/bin/testsubscriber/main.go
