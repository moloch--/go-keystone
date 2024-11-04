package keystone

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// just for prevent [import _ "embed"] :)
var fs embed.FS

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
	ksASM      api.Function
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
	// load keystone wasm module
	cm, err := runtime.CompileModule(ctx, module)
	if err != nil {
		panic(fmt.Sprintf("failed to compile keystone wasm module: %s", err))
	}

	envBuilder := runtime.NewHostModuleBuilder("a")

	fb := envBuilder.NewFunctionBuilder()
	hello := func(int32, int32, int32) {
		fmt.Println("throw")
	}
	fb.WithFunc(hello).Export("a")

	hello2 := func(int32, int32) int32 {

		return 1
	}
	fb.WithFunc(hello2).Export("b")

	hello3 := func(int32) int32 {

		return 1
	}
	fb.WithFunc(hello3).Export("c")

	hello4 := func(int32, int32) int32 {
		fmt.Println("___syscall_lstat64")
		return 0
	}
	fb.WithFunc(hello4).Export("d")

	hello5 := func(int32, int32, int32, int32) int32 {
		fmt.Println("___syscall_newfstatat")
		return 0
	}
	fb.WithFunc(hello5).Export("e")

	hello6 := func(int32, int32) int32 {
		fmt.Println("syscall_stat64")
		return 0
	}
	fb.WithFunc(hello6).Export("f")

	hello7 := func(int32, int32) int32 {
		fmt.Println("syscall_fstat64")
		return 0
	}
	fb.WithFunc(hello7).Export("g")

	hello8 := func(int32) {
		fmt.Println("exit")
		return
	}
	fb.WithFunc(hello8).Export("h")

	hello9 := func(int32, int32, int32, int64, int32) int32 {

		return 1
	}
	fb.WithFunc(hello9).Export("i")

	hello10 := func(int32, int64, int32, int32) int32 {

		return 1
	}
	fb.WithFunc(hello10).Export("j")

	hello11 := func(int32, int32, int32, int32, int64, int32, int32) int32 {

		fmt.Println("mmap")

		return 1
	}
	fb.WithFunc(hello11).Export("k")

	hello12 := func(int32, int32, int32, int32, int32, int64) int32 {
		fmt.Println("munmap")

		return 1
	}
	fb.WithFunc(hello12).Export("l")

	hello13 := func(int32, int32) int32 {

		return 1
	}
	fb.WithFunc(hello13).Export("m")

	hello14 := func(int32, int32) int32 {

		return 1
	}
	fb.WithFunc(hello14).Export("n")

	hello15 := func(buf int32, size int32) int32 {
		fmt.Println("getcwd")
		fmt.Println(buf, size)

		return 1
	}
	fb.WithFunc(hello15).Export("o")

	hello16 := func(int32, int32, int32, int32) int32 {
		fmt.Println("openat")
		return 0
	}
	fb.WithFunc(hello16).Export("p")

	hello17 := func(int32, int32, int32, int32) int32 {

		return 1
	}
	fb.WithFunc(hello17).Export("q")

	hello18 := func(int32, int32, int32, int32) int32 {

		return 1
	}
	fb.WithFunc(hello18).Export("r")

	hello19 := func() {
		fmt.Println("abort")
	}
	fb.WithFunc(hello19).Export("s")

	hello20 := func(v int32) int32 {

		fmt.Println("resize heap", v)

		return 0
	}
	fb.WithFunc(hello20).Export("t")

	_, err = envBuilder.Instantiate(ctx)
	if err != nil {
		return nil, errors.New("9999")
	}

	mc := wazero.NewModuleConfig()
	mod, err := runtime.InstantiateModule(ctx, cm, mc)
	if err != nil {
		return nil, err
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

	rets, err = malloc.Call(ctx, 4096)
	fmt.Println(rets, err)
	asm := rets[0]

	ok = memory.WriteString(uint32(asm), "xor eax, eax\nret\x00")
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
	fmt.Println(memory.Read(instAddr, 3))

	engine := Engine{
		arch: arch,
		mode: mode,

		context: ctx,
		runtime: runtime,
		module:  mod,

		malloc: mod.ExportedFunction("malloc"),
		free:   mod.ExportedFunction("free"),

		ksOpen:     mod.ExportedFunction("ks_open"),
		ksASM:      mod.ExportedFunction("ks_asm"),
		ksFree:     mod.ExportedFunction("ks_free"),
		ksClose:    mod.ExportedFunction("ks_close"),
		ksOption:   mod.ExportedFunction("ks_option"),
		ksErrno:    mod.ExportedFunction("ks_errno"),
		ksVersion:  mod.ExportedFunction("ks_version"),
		ksStrerror: mod.ExportedFunction("ks_strerror"),
	}
	return &engine, nil
}

func (e *Engine) Assemble() error {
	return nil
}

func (e *Engine) Close() error {
	return e.runtime.Close(e.context)
}
