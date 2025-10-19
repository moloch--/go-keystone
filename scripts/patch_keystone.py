#!/usr/bin/env python3

import pathlib
import sys


def patch_cmake_minimum(src_dir: pathlib.Path) -> None:
    """Ensure Keystone CMake files require at least CMake 3.5."""
    for rel in ("CMakeLists.txt", "llvm/CMakeLists.txt"):
        path = src_dir / rel
        text = path.read_text()
        updated = text.replace(
            "cmake_minimum_required(VERSION 2.8.7)",
            "cmake_minimum_required(VERSION 3.5)",
        )
        if updated != text:
            path.write_text(updated)


def main(argv: list[str]) -> int:
    if len(argv) != 3:
        print(f"usage: {argv[0]} <keystone_src_dir> <patch_mark>", file=sys.stderr)
        return 1

    src_dir = pathlib.Path(argv[1])
    patch_mark = pathlib.Path(argv[2])

    patch_cmake_minimum(src_dir)
    patch_mark.touch()
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
