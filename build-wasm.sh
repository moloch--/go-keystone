#!/usr/bin/env bash

set -eu

EMSCRIPTEN=/opt/homebrew/Cellar/emscripten/3.1.46/bin/

CURRENT_DIR=$(realpath .)
OUTPUT_NAME=keystone

ARCHS=(
  AArch64
  ARM
  X86
  Mips
  PowerPC
  Sparc
  SystemZ
  Hexagon
  RISCV
)

BUILD_FLAGS=(
  -D BUILD_LIBS_ONLY=ON

  -D LLVM_TARGETS_TO_BUILD=$(IFS=';'; echo "${ARCHS[*]}")
)

EXPORTED_FUNCTIONS=(
  free
  malloc

  ks_open
  ks_asm
  ks_free
  ks_close
  ks_option
  ks_errno
  ks_version
  ks_arch_supported
  ks_strerror

  __getTypeName
)
EXPORTED_FUNCTIONS=$(echo -n "${EXPORTED_FUNCTIONS[*]}" | jq -cR 'split(" ") | map("_" + .)')

EMSCRIPTEN_SETTINGS=(
  -lembind
  --no-entry
  -s STANDALONE_WASM
  # -s MODULARIZE
  # -s EXPORT_ES6
  -s WASM_BIGINT
  -s EXPORTED_FUNCTIONS=$EXPORTED_FUNCTIONS
  -s FILESYSTEM=0
  -s EXPORTED_RUNTIME_METHODS=ccall,cwrap,getValue,UTF8ToString
  -s EXPORT_NAME=$OUTPUT_NAME
  # -s ASSERTIONS
)

cd keystone-engine
$EMSCRIPTEN/emcmake cmake -B build ${BUILD_FLAGS[*]} -DCMAKE_BUILD_TYPE=Release

cd build
cmake --build . -j --target $OUTPUT_NAME
$EMSCRIPTEN/emcc llvm/lib/lib$OUTPUT_NAME.a -Os --minify 0 ${EMSCRIPTEN_SETTINGS[*]} -o $OUTPUT_NAME.mjs

cp ./keystone.wasm $CURRENT_DIR/keystone/keystone.wasm

