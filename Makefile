#
# Project: TGDP - Traffic Generator for Diameter Protocol
# Description: Simple tool for testing and debugging the Diameter protocol
#
# Author: Alexander Kefeli <alexander.kefeli@gmail.com>
#
# File: Makefile
# Description: build the project
#

MAKEFLAGS += --silent

TARGET     = tgdp
VERSION    = $(shell git describe --tags $(git rev-list --tags --max-count=1))
BUILD_DATE = $(shell date +"%Y-%m-%d")
VER_INFO   = -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)

GOLANG = go
GO_FLAGS = build -o $(TARGET) -gcflags '-N -l'
LD_FLAGS = -ldflags '$(VER_INFO)'
LD_FLAGS_REL = -ldflags '$(VER_INFO) -s -w -linkmode external -extldflags -static'

PKL2GO = pkl-gen-go
PKL2GO_FLAGS = --output-path ../../pkg

PKLDIR = ./configs/pkl
PKLSRC = $(shell find $(PKLDIR) -type f -name '*.pkl')
PKL_GO = $(shell find ./pkg -type f -name '*.pkl.go')
DIAPKL = Diameter.pkl

SRCMAIN = ./cmd/$(TARGET)/main.go
SOURCES = $(shell find . -type f -name '*.go')

$(TARGET): $(SOURCES)
	$(GOLANG) $(GO_FLAGS) $(LD_FLAGS) $(SRCMAIN)

release: $(SOURCES)
	$(GOLANG) $(GO_FLAGS) $(LD_FLAGS_REL) $(SRCMAIN)

pkl2go: $(PKLSRC)
	@cd $(PKLDIR); $(PKL2GO) $(PKL2GO_FLAGS) $(DIAPKL)

run: $(TARGET)
	$(GOLANG) $(RUN_FLAGS)

mod:
	$(GOLANG) mod tidy
	$(GOLANG) mod vendor

clean:
	rm -f $(TARGET)
	rm -f *.pcap

rmxattr:
	find . -name '._*' -delete
