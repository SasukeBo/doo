package gitlab

import (
	"fmt"
	"os"
	"sasukebo/doo/gitlab/client"
	"sasukebo/doo/utils"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/urfave/cli/v2"
	"github.com/xanzy/go-gitlab"
)

// GenerateLocalDirectories 从远程gitlab同步当前用户的分组目录结构至本地
func GenerateLocalDirectories(ctx *cli.Context) error {
	var (
		err         error
		accessToken string
		root        string
		gs          = make(map[string]struct{})
	)

	accessToken, err = utils.MustGetStringArg(ctx, "access_token", "DOO_GITLAB_ACCESS_TOKEN")
	if err != nil {
		return err
	}

	err = client.Init(ctx)
	if err != nil {
		return err
	}
	root, err = utils.MustGetStringArg(ctx, "root", "DOO_GITLAB_SYNC_ROOT")
	if err != nil {
		return err
	}

	{
		groups := strings.Split(ctx.String("groups"), ",")
		for _, group := range groups {
			if group == "" {
				continue
			}
			gs[group] = struct{}{}
		}
	}

	groups, err := client.GetGroups()
	if err != nil {
		return err
	}

	if len(gs) == 0 {
		fmt.Println("--- [INFO] Total groups:", len(groups))
	}

	for _, group := range groups {
		// fmt.Println("--- [debug]", group.FullName, group.Name, group.FullPath)
		if _, ok := gs[group.FullPath]; len(gs) > 0 && !ok {
			continue
		}
		dir := fmt.Sprintf("%s/%s", root, group.FullPath)
		fmt.Printf("--- [INFO] Make dir for group %s: %s\n", group.Name, dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("--- [ERROR] Make dir for group %s failed: %v\n", group.Name, dir)
		}

		if utils.IsDir(dir) {
			fmt.Printf("--- [INFO] Clone or update projects for group %s\n", group.Name)
			cloneOrUpdateProjects(group, accessToken, root)
		}
	}
	return nil
}

func cloneOrUpdateProjects(group *gitlab.Group, accessToken, root string) {
	var auth = http.BasicAuth{Username: "thingyouwe", Password: accessToken}
	projects, err := client.GetGroupProjects(group.ID)
	if err != nil {
		fmt.Printf("--- [ERROR] Get group projects failed: %v\n", err)
		return
	}
	for _, project := range projects {
		projectDir := fmt.Sprintf("%s/%s/%s", root, group.FullPath, project.Path)
		if utils.IsDir(projectDir) {
			fmt.Printf("--- [INFO] Pull project %s\n", project.Name)
			if repo, err := git.PlainOpen(projectDir); err == nil {
				if w, err := repo.Worktree(); err == nil {
					if err := w.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(project.DefaultBranch)}); err == nil {
						_ = w.Pull(&git.PullOptions{RemoteName: "origin", Auth: &auth})
					}
				}
			}
			if err != nil {
				fmt.Printf("--- [ERROR] Pull project %s for branch %s failed: %v\n", project.Name, project.DefaultBranch, err)
			}
		} else {
			fmt.Printf("--- [INFO] Clone project %s\n", project.Name)
			_, err := git.PlainClone(projectDir, false, &git.CloneOptions{
				URL:               project.HTTPURLToRepo,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
				Auth:              &auth,
			})
			if err != nil {
				fmt.Printf("--- [ERROR] Clone project %s failed: %v\n", project.Name, err)
			}
		}
	}
}
