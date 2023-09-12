SHELL = /bin/bash -o nounset -o errexit -o pipefail
.DEFAULT_GOAL = build
BUILD_PATH  := $(patsubst %/,%,$(abspath $(dir $(lastword $(MAKEFILE_LIST)))))
PARENT_PATH := $(patsubst %/,%,$(dir $(BUILD_PATH)))
UNITY_PROJ := ${PARENT_PATH}/arcspace.unity-app
UNITY_PATH := $(shell python3 ${UNITY_PROJ}/arc-utils.py UNITY_PATH "${UNITY_PROJ}")
ARC_UNITY_PATH = ${UNITY_PROJ}/Assets/ArcXR


## display this help message
help:
	@echo -e "\033[32m"
	@echo "go-archost"
	@echo "  PARENT_PATH:     ${PARENT_PATH}"
	@echo "  BUILD_PATH:      ${BUILD_PATH}"
	@echo "  UNITY_PROJ:      ${UNITY_PROJ}"
	@echo "  UNITY_PATH:      ${UNITY_PATH}"
	@echo
	@awk '/^##.*$$/,/[a-zA-Z_-]+:/' $(MAKEFILE_LIST) | awk '!(NR%2){print $$0p}{p=$$0}' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-32s\033[0m %s\n", $$1, $$2}' | sort

# ----------------------------------------
# build

GOFILES = $(shell find . -type f -name '*.go')
	
.PHONY: tools generate build

## build archost and libarchost
build:  arc-sdk



## build arc-sdk
arc-sdk:
	echo "Ship it!"


## install protobufs tools needed to turn a .proto file into Go and C# files
tools-proto:
	go install github.com/gogo/protobuf/protoc-gen-gogoslick
	go get -d  github.com/gogo/protobuf/proto


## generate .cs and .go from .proto files
generate:
#   protoc: https://github.com/protocolbuffers/protobuf/releases
	protoc \
	    --gogoslick_out=plugins:. --gogoslick_opt=paths=source_relative \
	    --csharp_out "${ARC_UNITY_PATH}/Arc.Apps" \
	    --proto_path=. \
		apis/arc/arc.proto

	protoc \
	    --gogoslick_out=plugins:. --gogoslick_opt=paths=source_relative \
	    --csharp_out "${ARC_UNITY_PATH}/Arc.Crates" \
	    --proto_path=. \
		apis/crates/crates.proto

