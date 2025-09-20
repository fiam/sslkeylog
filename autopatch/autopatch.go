// Package autopatch automatically patches http.DefaultTransport using sslkeylogfile.PatchDefaultTransport.
package autopatch

import (
	"fmt"

	"github.com/fiam/sslkeylogfile"
)

func init() {
	if err := sslkeylogfile.PatchDefaultTransport(); err != nil {
		panic(fmt.Errorf("autopatch: failed to patch http.DefaultTransport: %w", err))
	}
}
