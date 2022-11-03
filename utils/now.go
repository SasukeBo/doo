package utils

import (
	"fmt"
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
