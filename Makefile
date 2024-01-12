# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOINSTALL := $(GOCMD) install
GOCLEAN := $(GOCMD) clean
GOGET := $(GOCMD) get

# Binary name
BINARY_NAME := kubestat

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME)

install:
	sudo cp $(BINARY_NAME) /usr/local/bin

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: all build install clean