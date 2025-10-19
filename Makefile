SHELL := /bin/bash

DIST_DIR ?= dist
BIN_NAME := keystone
GO_BIN ?= $(shell if [ -x /opt/homebrew/bin/go ]; then echo /opt/homebrew/bin/go; elif command -v go >/dev/null 2>&1; then command -v go; else echo go; fi)

WASM_BUILD_ROOT := build/wasm
KEYSTONE_SRC_DIR := $(WASM_BUILD_ROOT)/keystone
KEYSTONE_PATCH_MARK := $(KEYSTONE_SRC_DIR)/.cmake_patched
KEYSTONE_BUILD_DIR := $(WASM_BUILD_ROOT)/build
WASM_EXPORT_DIR := $(DIST_DIR)/wasm
WASM_MODULE := $(WASM_EXPORT_DIR)/$(BIN_NAME).wasm
WASM_BUNDLE := $(WASM_EXPORT_DIR)/$(BIN_NAME).mjs

LLVM_TARGETS := AArch64;ARM;X86;Mips;PowerPC;Sparc;SystemZ;Hexagon;RISCV
EXPORTED_FUNCTIONS := '["_malloc","_free","_ks_open","_ks_option","_ks_asm","_ks_free","_ks_close","_ks_arch_supported","_ks_errno","_ks_strerror","_ks_version"]'
EMSCRIPTEN_SETTINGS := -s EXPORT_NAME=$(BIN_NAME) \
	-s EXPORTED_FUNCTIONS=$(EXPORTED_FUNCTIONS) \
	-s EXPORTED_RUNTIME_METHODS=ccall,cwrap,getValue,UTF8ToString \
	-s EXPORT_ES6=1 \
	-s MODULARIZE=1 \
	-s WASM_BIGINT=1 \
	-s FILESYSTEM=0 \
	-s DETERMINISTIC=1 \
	-s ALLOW_MEMORY_GROWTH=1

EMSDK_ENV ?=

LIBATOMIC_LIB ?=
LIBATOMIC_CANDIDATES := /usr/lib/libatomic.so /usr/lib64/libatomic.so /usr/local/lib/libatomic.so /opt/homebrew/opt/gcc/lib/gcc/current/libatomic.a
ifeq ($(LIBATOMIC_LIB),)
LIBATOMIC_LIB := $(firstword $(foreach path,$(LIBATOMIC_CANDIDATES),$(if $(wildcard $(path)),$(path))))
endif

EM_CACHE_DIR := $(abspath $(WASM_BUILD_ROOT)/cache)
EM_PORTS_DIR := $(abspath $(WASM_BUILD_ROOT)/ports)

GO_SOURCES := $(shell find cmd -type f -name '*.go') $(shell find . -maxdepth 1 -type f -name '*.go')

.PHONY: all go wasm clean

all: go wasm

$(DIST_DIR):
	mkdir -p $@

$(DIST_DIR)/$(BIN_NAME): $(GO_SOURCES) go.mod go.sum | $(DIST_DIR)
	$(GO_BIN) build -v -trimpath -ldflags "-s -w" -o $@ ./cmd

go: $(DIST_DIR)/$(BIN_NAME)

$(KEYSTONE_SRC_DIR):
	git clone --depth 1 https://github.com/For-ACGN/keystone $@

$(KEYSTONE_PATCH_MARK): $(KEYSTONE_SRC_DIR)
	python3 scripts/patch_keystone.py "$(KEYSTONE_SRC_DIR)" "$(KEYSTONE_PATCH_MARK)"

$(WASM_EXPORT_DIR):
	mkdir -p $@

$(WASM_BUNDLE): $(KEYSTONE_PATCH_MARK) | $(WASM_EXPORT_DIR)
	@set -euo pipefail; \
	if [ -n "$(EMSDK_ENV)" ]; then \
		if [ ! -f "$(EMSDK_ENV)" ]; then \
			echo "EMSDK_ENV path '$(EMSDK_ENV)' not found" >&2; \
			exit 1; \
		fi; \
		source "$(EMSDK_ENV)"; \
	fi; \
	if [ -n "$(LIBATOMIC_LIB)" ]; then \
		LIBATOMIC_DIR=$$(dirname "$(LIBATOMIC_LIB)"); \
		export LIBRARY_PATH="$$LIBATOMIC_DIR$${LIBRARY_PATH:+:$$LIBRARY_PATH}"; \
		export LDFLAGS="-L$$LIBATOMIC_DIR $${LDFLAGS:-}"; \
	fi; \
	export EM_CACHE="$(EM_CACHE_DIR)"; \
	export EM_PORTS="$(EM_PORTS_DIR)"; \
	mkdir -p "$(EM_CACHE_DIR)" "$(EM_PORTS_DIR)"; \
	cmake -E rm -rf $(KEYSTONE_BUILD_DIR); \
	scripts/sanitize_emcmake.sh cmake -S $(KEYSTONE_SRC_DIR) -B $(KEYSTONE_BUILD_DIR) \
		-D BUILD_LIBS_ONLY=ON \
		-D LLVM_TARGETS_TO_BUILD="$(LLVM_TARGETS)" \
		-D CMAKE_BUILD_TYPE=Release \
		-D CMAKE_POLICY_VERSION=3.5 \
		-D CMAKE_POLICY_VERSION_MINIMUM=3.5 \
		-D CMAKE_OSX_ARCHITECTURES= \
		-D CMAKE_OSX_DEPLOYMENT_TARGET= \
		-D CMAKE_C_COMPILER_WORKS=ON \
		-D CMAKE_CXX_COMPILER_WORKS=ON \
		-D CMAKE_C_COMPILER=$(abspath scripts/strip_macos_flags_emcc.sh) \
		-D CMAKE_CXX_COMPILER=$(abspath scripts/strip_macos_flags_emxx.sh); \
	cmake --build $(KEYSTONE_BUILD_DIR) -j --target $(BIN_NAME); \
	emcc $(KEYSTONE_BUILD_DIR)/llvm/lib/lib$(BIN_NAME).a -Os --minify 0 $(EMSCRIPTEN_SETTINGS) -o $(WASM_BUNDLE)

$(WASM_MODULE): $(WASM_BUNDLE)
	@true

wasm: $(WASM_BUNDLE) $(WASM_MODULE)

clean:
	cmake -E rm -rf $(DIST_DIR) build
