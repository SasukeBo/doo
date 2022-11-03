package gitlab

import (
	"fmt"
	"sasukebo/gitlab-helper/gitlab/client"
	"sasukebo/gitlab-helper/utils"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

func InitProject(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		fmt.Println("project name is required")
		return nil
	}

	var (
		groupName, branches string
		err                 error
		name                = ctx.Args().Get(0)
	)

	groupName, err = utils.MustGetStringArg(ctx, "group", "")
	if err != nil {
		return err
	}
	branches, err = utils.MustGetStringArg(ctx, "branches", "")
	if err != nil {
		return err
	}

	err = client.Init(ctx)
	if err != nil {
		return err
	}

	group, err := client.GetGroupByName(groupName)
	if err != nil {
		return err
	}
	if group == nil {
		fmt.Printf("group %s not exist", groupName)
		return nil
	}

	project, err := client.GetProjectWithGroupId(group.ID, name)
	if err != nil {
		return err
	}
	if project == nil {
		fmt.Printf("project %s not exist", name)
		return nil
	}

	configs := strings.Split(branches, ",")
	for _, config := range configs {
		items := strings.Split(config, ":")
		if len(items) == 1 {
			items = append(items, "044")
		}

		var levels = []int{0, 0, 0}
		for i, item := range []rune(items[1]) {
			if i >= 3 {
				break
			}
			v, err := strconv.Atoi(string([]rune{item}))
			if err != nil {
				fmt.Println("unexpected level config", items[1])
				return nil
			}
			levels[i] = v * 10
		}

		if err = client.ProtectProjectBranch(project.ID, items[0], levels); err != nil {
			return err
		}
		fmt.Printf(
			"protect branch %s with push_access_level=%v merge_access_level=%v unprotect_access_level=%v\n",
			items[0], levels[0], levels[1], levels[2],
		)
	}

	return nil
}
