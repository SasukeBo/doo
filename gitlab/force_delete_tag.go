package gitlab

import (
	"fmt"
	"sasukebo/doo/gitlab/client"

	"github.com/urfave/cli/v2"
)

func ForceDeleteTag(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		fmt.Println("tag name required")
		return nil
	}

	if err := client.Init(ctx); err != nil {
		return err
	}

	var (
		name    = ctx.String("project")
		group   = ctx.String("group")
		tagName = ctx.Args().Get(0)
	)

	g, err := client.GetGroupByName(group)
	if err != nil {
		return err
	}
	if g == nil {
		fmt.Println("group not exist")
		return nil
	}

	p, err := client.GetProjectWithGroupId(g.ID, name)
	if err != nil {
		return err
	}
	if p == nil {
		fmt.Println("project not exist")
		return nil
	}

	return client.DeleteProjectTag(p.ID, tagName)
}
