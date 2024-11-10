set -eu

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
  malloc
  free

  ks_open
  ks_option
  ks_asm
  ks_free
  ks_close
  ks_arch_supported
  ks_errno
  ks_strerror
  ks_version
)
EXPORTED_FUNCTIONS=$(echo -n "${EXPORTED_FUNCTIONS[*]}" | jq -cR 'split(" ") | map("_" + .)')

EMSCRIPTEN_SETTINGS=(
  -s EXPORT_NAME=$OUTPUT_NAME
  -s EXPORTED_FUNCTIONS=$EXPORTED_FUNCTIONS
  -s EXPORTED_RUNTIME_METHODS=ccall,cwrap,getValue,UTF8ToString
  -s EXPORT_ES6=1
  -s MODULARIZE=1
  -s WASM_BIGINT=1
  -s FILESYSTEM=0
  -s DETERMINISTIC=1
  -s ALLOW_MEMORY_GROWTH=1
)

cd emsdk
source ./emsdk_env.sh
cd ..

cd keystone
emcmake cmake -B build ${BUILD_FLAGS[*]} -DCMAKE_BUILD_TYPE=Release

cd build
cmake --build . -j --target $OUTPUT_NAME
emcc llvm/lib/lib$OUTPUT_NAME.a -Os --minify 0 ${EMSCRIPTEN_SETTINGS[*]} -o $OUTPUT_NAME.mjs

cp ./keystone.wasm $CURRENT_DIR/wasm/keystone.wasm
cd ../..
