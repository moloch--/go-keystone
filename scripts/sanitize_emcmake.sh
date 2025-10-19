#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Strip host-specific flags that break Emscripten sanity checks on macOS.
unset CMAKE_OSX_ARCHITECTURES CMAKE_OSX_DEPLOYMENT_TARGET SDKROOT MACOSX_DEPLOYMENT_TARGET
export CFLAGS= CXXFLAGS= LDFLAGS= CPPFLAGS= OBJCFLAGS= OBJCXXFLAGS=

export EMCC_REAL="${EMCC_REAL:-$(command -v emcc)}"
export EMXX_REAL="${EMXX_REAL:-$(command -v em++)}"

export EMCC="${ROOT_DIR}/scripts/strip_macos_flags_emcc.sh"
export EMXX="${ROOT_DIR}/scripts/strip_macos_flags_emxx.sh"

exec emcmake "$@"
