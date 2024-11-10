package keystone

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEngine(t *testing.T) {
	t.Run("x86 with 32bit", func(t *testing.T) {
		engine, err := NewEngine(ARCH_X86, MODE_32)
		require.NoError(t, err)
		err = engine.Option(OPT_SYNTAX, OPT_SYNTAX_INTEL)
		require.NoError(t, err)

		src := strings.Repeat(".code32\nxor eax, eax\nret\n", 5000)
		inst, err := engine.Assemble(src, 0)
		require.NoError(t, err)
		expected := bytes.Repeat([]byte{0x31, 0xC0, 0xC3}, 5000)
		require.Equal(t, expected, inst)

		err = engine.Close()
		require.NoError(t, err)
	})

	t.Run("x86 with 64bit", func(t *testing.T) {
		engine, err := NewEngine(ARCH_X86, MODE_64)
		require.NoError(t, err)
		err = engine.Option(OPT_SYNTAX, OPT_SYNTAX_INTEL)
		require.NoError(t, err)

		src := strings.Repeat(".code64\nxor rax, rax\nret\n", 5000)
		inst, err := engine.Assemble(src, 0)
		require.NoError(t, err)
		expected := bytes.Repeat([]byte{0x48, 0x31, 0xC0, 0xC3}, 5000)
		require.Equal(t, expected, inst)

		err = engine.Close()
		require.NoError(t, err)
	})
}

func TestEngine_Option(t *testing.T) {
	engine, err := NewEngine(ARCH_X86, MODE_32)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		err = engine.Option(OPT_SYNTAX, OPT_SYNTAX_INTEL)
		require.NoError(t, err)
	})

	t.Run("invalid option type", func(t *testing.T) {
		err = engine.Option(123, OPT_SYNTAX_INTEL)
		errStr := "failed to set keystone option: Invalid option (KS_ERR_OPT_INVALID)"
		require.EqualError(t, err, errStr)
	})

	t.Run("invalid option value", func(t *testing.T) {
		err = engine.Option(OPT_SYNTAX, 123)
		errStr := "failed to set keystone option: Invalid option (KS_ERR_OPT_INVALID)"
		require.EqualError(t, err, errStr)
	})

	err = engine.Close()
	require.NoError(t, err)
}

func TestEngine_Assemble(t *testing.T) {
	engine, err := NewEngine(ARCH_X86, MODE_32)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		src := "xor eax, eax\nret\n"
		inst, err := engine.Assemble(src, 0)
		require.NoError(t, err)
		expected := []byte{0x31, 0xC0, 0xC3}
		require.Equal(t, expected, inst)
	})

	t.Run("invalid source", func(t *testing.T) {
		src := "invalid\n"
		inst, err := engine.Assemble(src, 0)
		errStr := "failed to assemble: Invalid mnemonic (KS_ERR_ASM_MNEMONICFAIL)"
		require.EqualError(t, err, errStr)
		require.Nil(t, inst)
	})

	err = engine.Close()
	require.NoError(t, err)
}

func TestEngine_Version(t *testing.T) {
	engine, err := NewEngine(ARCH_X86, MODE_32)
	require.NoError(t, err)

	version := engine.Version()
	require.Equal(t, "0.9", version)

	err = engine.Close()
	require.NoError(t, err)
}
