#!/usr/bin/env bash
set -euo pipefail

if [ -z "${EMCC_REAL:-}" ]; then
	echo "EMCC_REAL is not set" >&2
	exit 1
fi

args=()
skip_next=0
for arg in "$@"; do
	if [ "${skip_next}" -eq 1 ]; then
		skip_next=0
		continue
	fi

	if [ "${arg}" = "-arch" ] || [ "${arg}" = "-isysroot" ]; then
		skip_next=1
		continue
	fi

	args+=("${arg}")
done

exec "${EMCC_REAL}" "${args[@]}"
