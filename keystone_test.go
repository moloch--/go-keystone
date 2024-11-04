package keystone

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine(ARCH_X86, MODE_64)
	require.NoError(t, err)

	err = engine.Close()
	require.NoError(t, err)
}
