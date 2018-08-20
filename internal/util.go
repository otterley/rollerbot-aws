package internal

import (
	"fmt"
	"os"
)

// MustEnv returns the value of the environment variable specified by name.
// It will panic if no such variable is defined, or the value is empty.
func MustEnv(name string) string {
	val := os.Getenv(name)
	if val == "" {
		panic(fmt.Errorf("Env var %s not defined", val))
	}
	return val
}
