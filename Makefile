build: main.go
	go build -o ./test/lit

install: main.go
	go install

uninstall:
	rm -f $(GOPATH)/bin/lit
