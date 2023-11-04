package keystone

import (
	"context"
	_ "embed"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

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
}

func (k *Keystone) Close() {
	k.runtime.Close(k.ctx)
}

func NewKeystone(ctx context.Context, arch Architecture, mode Mode) (*Keystone, error) {
	runtime := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	mod, err := runtime.Instantiate(ctx, keystoneWasm)
	if err != nil {
		return nil, err
	}

	keystone := &Keystone{
		arch: arch,
		mode: mode,

		ctx:     ctx,
		runtime: runtime,
		module:  mod,
	}

	return keystone, nil
}
