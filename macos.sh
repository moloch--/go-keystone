#!/usr/bin/env bash
set -euo pipefail

if ! command -v brew >/dev/null 2>&1; then
  echo "Homebrew is required but not found. Install it from https://brew.sh/ and re-run this script." >&2
  exit 1
fi

PACKAGES=(
  go
  cmake
  git
  jq
  gcc
  emscripten
)

echo "Updating Homebrew..."
brew update

echo "Installing packages: ${PACKAGES[*]}"
brew install "${PACKAGES[@]}"

BREW_PREFIX="$(brew --prefix)"
EMSDK_ENV="${BREW_PREFIX}/opt/emscripten/libexec/emsdk_env.sh"

if [[ -f "${EMSDK_ENV}" ]]; then
  echo ""
  echo "Emscripten environment script detected at:"
  echo "  ${EMSDK_ENV}"
  echo "Source it before running 'make wasm', for example:"
  echo "  source \"${EMSDK_ENV}\""
else
  echo ""
  echo "Warning: Unable to locate emsdk environment script. Verify the emscripten installation."
fi

LIBATOMIC_DIR="$(brew --prefix gcc)/lib/gcc/current"
if [[ -d "${LIBATOMIC_DIR}" ]]; then
  echo ""
  echo "libatomic detected under:"
  echo "  ${LIBATOMIC_DIR}"
  echo "To help CMake find it, you can export:"
  echo "  export LIBRARY_PATH=\"${LIBATOMIC_DIR}:\${LIBRARY_PATH:-}\""
  echo "  export LDFLAGS=\"-L${LIBATOMIC_DIR} \${LDFLAGS:-}\""
  echo "or pass LIBATOMIC_LIB=${LIBATOMIC_DIR}/libatomic.a to 'make wasm'."
fi

echo ""
echo "Dependencies installed. You can now run:"
echo "  source \"${EMSDK_ENV}\"  # ensure emcc/emcmake are on PATH"
echo "  make all                # build wasm and Go artifacts"
