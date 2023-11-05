package keystone

/*
   Go Keystone
   Copyright (C) 2023  moloch--

   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; either version 2 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License along
   with this program; if not, write to the Free Software Foundation, Inc.,
   51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

import (
	"context"
	_ "embed"

	embind "github.com/jerbob92/wazero-emscripten-embind"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/emscripten"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Update generated constants: cp ./keystone-engine/bindings/go/keystone/*_const.go ./keystone/

type Architecture uint
type Mode uint
type OptionType uint
type OptionValue uint
type Error uint32

//go:embed keystone.wasm
var keystoneWasm []byte

/*
	keystoneWasm Module Exports:
		- free
		- malloc

		- ks_open
		- ks_asm
		- ks_free
		- ks_close
		- ks_option
		- ks_errno
		- ks_version
		- ks_arch_supported
		- ks_strerror
*/

type Keystone struct {
	arch Architecture
	mode Mode

	ctx     context.Context
	runtime wazero.Runtime
	module  api.Module

	free   api.Function
	malloc api.Function

	ks_open           api.Function
	ks_asm            api.Function
	ks_free           api.Function
	ks_close          api.Function
	ks_option         api.Function
	ks_errno          api.Function
	ks_version        api.Function
	ks_arch_supported api.Function
	ks_strerror       api.Function
}

func (k *Keystone) Close() {
	k.runtime.Close(k.ctx)
}

func NewKeystone(ctx context.Context, arch Architecture, mode Mode) (*Keystone, error) {
	runtime := wazero.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		return nil, err
	}
	compiledModule, err := runtime.CompileModule(ctx, keystoneWasm)
	if err != nil {
		return nil, err
	}
	builder := runtime.NewHostModuleBuilder("env")
	emscriptenExporter, err := emscripten.NewFunctionExporterForModule(compiledModule)
	if err != nil {
		return nil, err
	}
	emscriptenExporter.ExportFunctions(builder)
	engine := embind.CreateEngine(embind.NewConfig())
	embindExporter := engine.NewFunctionExporterForModule(compiledModule)
	err = embindExporter.ExportFunctions(builder)
	if err != nil {
		return nil, err
	}
	_, err = builder.Instantiate(ctx)
	if err != nil {
		return nil, err
	}

	moduleConfig := wazero.NewModuleConfig().
		WithStartFunctions("_initialize").
		WithName("")
	ctx = engine.Attach(ctx)
	mod, err := runtime.InstantiateModule(ctx, compiledModule, moduleConfig)
	if err != nil {
		return nil, err
	}

	err = Attach(engine)
	if err != nil {
		return nil, err
	}

	keystone := &Keystone{
		arch: arch,
		mode: mode,

		ctx:     ctx,
		runtime: runtime,
		module:  mod,

		free:   mod.ExportedFunction("free"),
		malloc: mod.ExportedFunction("malloc"),

		ks_open:           mod.ExportedFunction("ks_open"),
		ks_asm:            mod.ExportedFunction("ks_asm"),
		ks_free:           mod.ExportedFunction("ks_free"),
		ks_close:          mod.ExportedFunction("ks_close"),
		ks_option:         mod.ExportedFunction("ks_option"),
		ks_errno:          mod.ExportedFunction("ks_errno"),
		ks_version:        mod.ExportedFunction("ks_version"),
		ks_arch_supported: mod.ExportedFunction("ks_arch_supported"),
		ks_strerror:       mod.ExportedFunction("ks_strerror"),
	}

	const pointerSize = 4
	_, err = keystone.ks_open.Call(ctx, uint64(keystone.arch), uint64(keystone.mode), pointerSize)
	if err != nil {
		return nil, err
	}

	return keystone, nil
}
