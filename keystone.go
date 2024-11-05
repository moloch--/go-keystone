package keystone

import (
	"context"
	"embed"
	"fmt"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// just for prevent [import _ "embed"] :)
var _ embed.FS

//go:embed wasm/keystone.wasm
var module []byte

type Engine struct {
	arch Arch
	mode Mode

	context context.Context
	runtime wazero.Runtime
	module  api.Module

	malloc api.Function
	free   api.Function

	ksOpen     api.Function
	ksOption   api.Function
	ksAsm      api.Function
	ksFree     api.Function
	ksClose    api.Function
	ksErrno    api.Function
	ksVersion  api.Function
	ksStrerror api.Function
}

// NewEngine is used to create keystone engine above wasm interpreter.
func NewEngine(arch Arch, mode Mode) (*Engine, error) {
	ctx := context.Background()
	// prevent generate RWX memory
	rc := wazero.NewRuntimeConfigInterpreter()
	runtime := wazero.NewRuntimeWithConfig(ctx, rc)
	// if failed to create engine, close the wasm runtime
	var ok bool
	defer func() {
		if !ok {
			_ = runtime.Close(ctx)
		}
	}()
	// load keystone wasm module
	cm, err := runtime.CompileModule(ctx, module)
	if err != nil {
		panic(fmt.Sprintf("failed to load keystone wasm module: %s", err))
	}
	err = processImport(runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to process wasm module import: %s", err)
	}
	mc := wazero.NewModuleConfig()
	mod, err := runtime.InstantiateModule(ctx, cm, mc)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %s", err)
	}

	malloc := mod.ExportedFunction("F")
	rets, err := malloc.Call(ctx, 4096)
	fmt.Println(rets, err)
	engineHandle := rets[0]

	ksOpen := mod.ExportedFunction("A")

	fmt.Println(ksOpen == nil)

	rets, err = ksOpen.Call(ctx, uint64(arch), uint64(mode), engineHandle)
	fmt.Println(rets, err)

	memory := mod.Memory()
	engineHd, ok := memory.ReadUint32Le(uint32(engineHandle))
	fmt.Println("read handle", ok)
	engineHH := uint64(engineHd)

	ksOption := mod.ExportedFunction("C")
	rets, err = ksOption.Call(ctx, engineHH, uint64(OPT_SYNTAX), uint64(OPT_SYNTAX_INTEL))
	fmt.Println(rets, err)

	rets, err = malloc.Call(ctx, 4096*1024)
	fmt.Println(rets, err)
	asm := rets[0]

	src := strings.Repeat("xor eax, eax\nret\n", 2048) + "\x00"
	ok = memory.WriteString(uint32(asm), src)
	fmt.Println("write asm", ok)

	fmt.Println(memory.Read(uint32(engineHandle), 16))
	fmt.Println(memory.Read(uint32(asm), 16))

	rets, err = malloc.Call(ctx, 1024*1024)
	fmt.Println(rets, err)
	inst := rets[0]

	rets, err = malloc.Call(ctx, 4096)
	fmt.Println(rets, err)
	instSize := rets[0]

	rets, err = malloc.Call(ctx, 4096)
	fmt.Println(rets, err)
	statCount := rets[0]

	ksASM := mod.ExportedFunction("E")
	rets, err = ksASM.Call(ctx,
		engineHH, asm, 0, inst, instSize, statCount,
	)
	fmt.Println(rets, err)

	fmt.Println(memory.Read(uint32(inst), 16))
	fmt.Println(memory.Read(uint32(instSize), 16))

	instAddr, ok := memory.ReadUint32Le(uint32(inst))
	fmt.Println("read inst", ok)
	fmt.Println(memory.Read(instAddr, 3*2048))

	engine := Engine{
		arch: arch,
		mode: mode,

		context: ctx,
		runtime: runtime,
		module:  mod,

		malloc: mod.ExportedFunction("malloc"),
		free:   mod.ExportedFunction("free"),

		ksOpen:     mod.ExportedFunction("ks_open"),
		ksAsm:      mod.ExportedFunction("ks_asm"),
		ksFree:     mod.ExportedFunction("ks_free"),
		ksClose:    mod.ExportedFunction("ks_close"),
		ksOption:   mod.ExportedFunction("ks_option"),
		ksErrno:    mod.ExportedFunction("ks_errno"),
		ksVersion:  mod.ExportedFunction("ks_version"),
		ksStrerror: mod.ExportedFunction("ks_strerror"),
	}
	return &engine, nil
}

func processImport(runtime wazero.Runtime) error {
	builder := runtime.NewHostModuleBuilder(importModuleName)
	fb := builder.NewFunctionBuilder()

	padFn1 := func(int32, int32, int32) {
	}
	fb.WithFunc(padFn1).Export(___cxa_throw)

	padFn2 := func(int32, int32) int32 {
		return 0
	}
	fb.WithFunc(padFn2).Export(___syscall_fstat64)

	padFn3 := func(buf int32, size int32) int32 {
		return 1
	}
	fb.WithFunc(padFn3).Export(___syscall_getcwd)

	padFn4 := func(int32, int32) int32 {
		return 0
	}
	fb.WithFunc(padFn4).Export(___syscall_lstat64)

	padFn5 := func(int32, int32, int32, int32) int32 {
		return 0
	}
	fb.WithFunc(padFn5).Export(___syscall_newfstatat)

	padFn6 := func(int32, int32, int32, int32) int32 {
		return 0
	}
	fb.WithFunc(padFn6).Export(___syscall_openat)

	padFn7 := func(int32, int32) int32 {
		return 0
	}
	fb.WithFunc(padFn7).Export(___syscall_stat64)

	padFn8 := func() {
	}
	fb.WithFunc(padFn8).Export(__abort_js)

	padFn9 := func(int32, int32, int32, int32, int64, int32, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn9).Export(__mmap_js)

	padFn10 := func(int32, int32, int32, int32, int32, int64) int32 {
		return 1
	}
	fb.WithFunc(padFn10).Export(__munmap_js)

	padFn11 := func(v int32) int32 {
		return 0
	}
	fb.WithFunc(padFn11).Export(_emscripten_resize_heap)

	padFn12 := func(int32, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn12).Export(_environ_get)

	padFn13 := func(int32, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn13).Export(_environ_sizes_get)

	padFn14 := func(int32) {
	}
	fb.WithFunc(padFn14).Export(_exit)

	padFn15 := func(int32) int32 {
		return 1
	}
	fb.WithFunc(padFn15).Export(_fd_close)

	padFn16 := func(int32, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn16).Export(_fd_fdstat_get)

	padFn17 := func(int32, int32, int32, int64, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn17).Export(_fd_pread)

	padFn18 := func(int32, int32, int32, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn18).Export(_fd_read)

	padFn19 := func(int32, int64, int32, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn19).Export(_fd_seek)

	padFn20 := func(int32, int32, int32, int32) int32 {
		return 1
	}
	fb.WithFunc(padFn20).Export(_fd_write)

	_, err := builder.Instantiate(context.Background())
	return err
}

func (e *Engine) Assemble() error {
	return nil
}

func (e *Engine) Close() error {
	return e.runtime.Close(e.context)
}
