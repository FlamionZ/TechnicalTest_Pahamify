# Go parameters
GOCMD=go
DEVCMD=bra
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=jobqueue
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build
build: 
	make bind-static
	$(GOBUILD) -o $(BINARY_NAME) -v main.go
bind-static:
	$(GOCMD) generate ./delivery/graphql/schema
test:
	@echo "\n\n==================== Start unit test and Integration Test ...... ====================\n\n"
	$(GOTEST) ./... -cover -race -count=1
	@echo "\n\n==================== Unit test and Integration Test Done ====================\n\n"
unittest:
	@echo "\n\n==================== Start unit test ...... ====================\n\n"
	@go test ./... --short -cover -race -count=1
	@echo "\n\n==================== Unit test done ====================\n\n"
lint:
	@golangci-lint run
clean: 
	$(GOCLEAN) cmd/$(BINARY_NAME)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
dev:
	$(DEVCMD) run
tool:
	# $(GOGET) github.com/Unknwon/bra
	$(GOGET) -u github.com/go-bindata/go-bindata/...
build-linux:
	make bind-static
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -o $(BINARY_UNIX) -a -installsuffix cgo -v main.go
mysql-suite:
	@git checkout repository/mysql/mysql_suite.go