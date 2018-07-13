# Modify this (on the command line!) if you want to use another compiler, like gccgo
GOCMD=go

GOBUILD=$(GOCMD) build

# Modify this to change the name you use to run the program
BINARY_NAME=pagetree

# Modify this if using a non-Unix system (possibly easier to build 
# manually) or if you want to use somewhere else in your path
BIN_DIR=/usr/local/bin

build:
	# Compile and move the binary to your bin
	$(GOBUILD) -o $(BINARY_NAME) -v
	mv $(BINARY_NAME) $(BIN_DIR)

local:
	$(GOBUILD) -o $(BINARY_NAME) -v