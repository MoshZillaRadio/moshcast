# Go parameters
GOCMD=go
GOBUILD=CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOCMD) build
GOMOD=$(GOCMD) mod
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
    
all: build
build:
		$(GOBUILD) -v -o "moshcast" main.go
test: 
		$(GOTEST) -v ./...
tidy:
		$(GOMOD) tidy
clean:
		rm -f moshcast 
