package utils

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func MustGetStringArg(ctx *cli.Context, name, env string) (v string, err error) {
	v = ctx.String(name)
	if v == "" && env != "" {
		v = os.Getenv(env)
	}
	if v == "" {
		err = fmt.Errorf("%s is required", name)
	}
	return
}

func IsDir(dir string) bool {
	f, err := os.Stat(dir)
	if err != nil {
		return false
	}

	return f.IsDir()
}
