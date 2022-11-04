package utils

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/urfave/cli/v2"
)

func Now(_ *cli.Context) error {
	t := time.Now()
	fmt.Println("---", t.In(time.FixedZone("Asia/Shanghai", 8*3600)).Format("2006年01月02号 15:04:05"))
	fmt.Println("---", t.In(time.FixedZone("Asia/Shanghai", 8*3600)).Format(time.RFC3339))
	fmt.Println("---", t.UnixMilli())
	return nil
}

func CD(ctx *cli.Context) error {
	root := ctx.String("root")
	if !IsDir(root) {
		fmt.Printf("%s is not a path\n", root)
		return nil
	}

	cmd := exec.Command("find", root, "-name", ctx.String("dir"))
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
