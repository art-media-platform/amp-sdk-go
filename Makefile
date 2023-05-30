SHELL = /bin/bash -o nounset -o errexit -o pipefail
.DEFAULT_GOAL = build
BUILD_PATH  := $(patsubst %/,%,$(abspath $(dir $(lastword $(MAKEFILE_LIST)))))
PARENT_PATH := $(patsubst %/,%,$(dir $(BUILD_PATH)))
UNITY_PROJ := ${PARENT_PATH}/arcspace.unity-app
ARC_LIBS = ${UNITY_PROJ}/Assets/Plugins/Arcspace/Plugins
ARC_UNITY_PATH = ${UNITY_PROJ}/Assets/Arcspace
grpc_csharp_exe="${GOPATH}/bin/grpc_csharp_plugin"
LIB_PROJ := ${BUILD_PATH}/cmd/libarchost

#UNITY_PATH = "${HOME}/Applications/2021.3.16f1"
UNITY_PATH := $(shell python3 ${UNITY_PROJ}/arc-utils.py UNITY_PATH "${UNITY_PROJ}")


CAPNP_DIST := "capnproto-c++-0.10.4"

## display this help message
help:
	@echo -e "\033[32m"
	@echo "go-archost"
	@echo "  PARENT_PATH:     ${PARENT_PATH}"
	@echo "  BUILD_PATH:      ${BUILD_PATH}"
	@echo "  UNITY_PROJ:      ${UNITY_PROJ}"
	@echo "  ARC_LIBS:        ${ARC_LIBS}"
	@echo "  UNITY_PATH:      ${UNITY_PATH}"
	@echo
	@awk '/^##.*$$/,/[a-zA-Z_-]+:/' $(MAKEFILE_LIST) | awk '!(NR%2){print $$0p}{p=$$0}' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-32s\033[0m %s\n", $$1, $$2}' | sort

# ----------------------------------------
# build

GOFILES = $(shell find . -type f -name '*.go')
	
.PHONY: build protos tools

## build archost and libarchost
build:  arc-sdk



## build arc-sdk
arc-sdk:
	echo "Ship it!"


## build and install cap'n proto tools -- https://capnproto.org/install.html
tools-capnp-nix:
	curl -O --insecure https://capnproto.org/${CAPNP_DIST}.tar.gz \
	&& tar zxf ${CAPNP_DIST}.tar.gz \
	&& cd ${CAPNP_DIST} \
	&& ./configure \
	&& make -j6 check \
	&& sudo make install \
	&& cd ..  \
	&& rm -rf ${CAPNP_DIST}

## install protobufs tools needed to turn a .proto file into Go and C# files
tools-proto:
	go install github.com/gogo/protobuf/protoc-gen-gogoslick
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go get -d  github.com/gogo/protobuf/proto

capnp:
	capnp compile -I capnproto.org/go/capnp/v3/std -ogo apis/arc/arc.capnp

## generate .cs and .go from proto files
protos:
#   GrpcTools (2.49.1)
#   Install protoc & grpc_csharp_plugin:
#      - Download latest Grpc.Tools from https://nuget.org/packages/Grpc.Tools
#      - Extract .nupkg as .zip, move protoc and grpc_csharp_plugin to ${GOPATH}/bin 
#   Or, just protoc: https://github.com/protocolbuffers/protobuf/releases
#   Links: https://grpc.io/docs/languages/csharp/quickstart/
	protoc \
	    --gogoslick_out=plugins=grpc:. --gogoslick_opt=paths=source_relative \
	    --csharp_out "${ARC_UNITY_PATH}/Arc" \
	    --grpc_out   "${ARC_UNITY_PATH}/Arc" \
	    --plugin=protoc-gen-grpc="${grpc_csharp_exe}" \
	    --proto_path=. \
		apis/arc/arc.proto

	protoc \
	    --gogoslick_out=plugins=grpc:. --gogoslick_opt=paths=source_relative \
	    --csharp_out "${ARC_UNITY_PATH}/Crates" \
	    --proto_path=. \
		apis/crates/crates.proto



