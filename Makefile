.PHONY: all build test lint clean install-deps

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOLINT=$(GOCMD) vet
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Binary names
COLLECTOR=collector
NORMALIZER=normalizer
DEDUPER=deduper
SCORER=scorer
DIGESTOR=digestor
RENDERER=renderer
PUBLISHER=publisher

# Directories
CMD_DIR=cmd
INTERNAL_DIR=internal
BUILD_DIR=build
BIN_DIR=$(BUILD_DIR)/bin

# Astro site
SITE_DIR=site
SITE_SRC=$(SITE_DIR)/src

# Data directories
DATA_DIR=data
DATA_RAW=$(DATA_DIR)/raw
DATA_ITEMS=$(DATA_DIR)/items
DATA_TOPICS=$(DATA_DIR)/topics
DATA_DIGESTS=$(DATA_DIR)/digests
DATA_INDEXES=$(DATA_DIR)/indexes
DATA_STATE=$(DATA_DIR)/state

all: build

build: build-collector build-normalizer build-deduper build-scorer build-digestor build-renderer build-publisher

build-collector:
	$(GOBUILD) -o $(BIN_DIR)/$(COLLECTOR) $(CMD_DIR)/collector

build-normalizer:
	$(GOBUILD) -o $(BIN_DIR)/$(NORMALIZER) $(CMD_DIR)/normalizer

build-deduper:
	$(GOBUILD) -o $(BIN_DIR)/$(DEDUPER) $(CMD_DIR)/deduper

build-scorer:
	$(GOBUILD) -o $(BIN_DIR)/$(SCORER) $(CMD_DIR)/scorer

build-digestor:
	$(GOBUILD) -o $(BIN_DIR)/$(DIGESTOR) $(CMD_DIR)/digestor

build-renderer:
	$(GOBUILD) -o $(BIN_DIR)/$(RENDERER) $(CMD_DIR)/renderer

build-publisher:
	$(GOBUILD) -o $(BIN_DIR)/$(PUBLISHER) $(CMD_DIR)/publisher

test:
	$(GOTEST) ./...

lint:
	$(GOLINT) ./...

fmt:
	$(GOFMT) -s -w .

clean:
	rm -rf $(BIN_DIR)

install-deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Site commands
site-build:
	cd $(SITE_DIR) && npm install && npm run build

site-dev:
	cd $(SITE_DIR) && npm run dev

# Data commands
data-init:
	mkdir -p $(DATA_RAW) $(DATA_ITEMS) $(DATA_TOPICS) $(DATA_DIGESTS) $(DATA_INDEXES) $(DATA_STATE)/logs

# CI commands
ci-collect:
	$(BIN_DIR)/$(COLLECTOR) --date=$(DATE) --batch=$(BATCH)

ci-normalize:
	$(BIN_DIR)/$(NORMALIZER) --date=$(DATE)

ci-dedupe:
	$(BIN_DIR)/$(DEDUPER) --date=$(DATE)

ci-score:
	$(BIN_DIR)/$(SCORER) --date=$(DATE)

ci-digest:
	$(BIN_DIR)/$(DIGESTOR) --date=$(DATE)

ci-render:
	$(BIN_DIR)/$(RENDERER) --date=$(DATE)

ci-publish:
	$(BIN_DIR)/$(PUBLISHER) --date=$(DATE)
