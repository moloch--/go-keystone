package main

import (
	"context"

	"github.com/moloch--/go-keystone/keystone"
)

func main() {
	keystone.NewKeystone(context.Background(), keystone.ARCH_X86, keystone.MODE_32)
}
