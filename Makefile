
# go params
GOCMD=go
GOBUILD=$(GOCMD) build -buildvcs=false
GOTEST=$(GOCMD) test -run
GOPATH=/usr/local/bin
DIR=$(shell pwd)

# normal entry points
build:
	clear
	@echo "building beer cellah..."
	@$(GOBUILD) -o $(GOPATH)/beer-cellah ./cmd
	
update:
	clear
	@echo "updating dependencies..."
	@$(GOCMD) get -u -t ./...
	@$(GOCMD) mod tidy 

test:
	@clear 
	@echo "QA testing..."
	@$(GOTEST) QA ./...

run: build
run:
	$(GOPATH)/beer-cellah
