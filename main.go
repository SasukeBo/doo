package main

import (
	"log"
	"os"
	"sasukebo/gitlab-helper/gitlab"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "dmf",
		Usage: "Do me a favor. 集成一些小工具来提高工作效率",
		Commands: []*cli.Command{
			{
				Name:  "gitlab",
				Usage: "some convenient tools to handle your gitlab with your access key.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "host",
						Usage: "gitlab service host, same as env `DMF_GITLAB_HOST`",
					},
					&cli.StringFlag{
						Name:    "access_token",
						Usage:   "gitlab personal access_token, same as env `DMF_GITLAB_ACCESS_TOKEN`",
						Aliases: []string{"t"},
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:  "sync",
						Usage: "sync your gitlab groups and projects to local",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "root",
								Usage:   "root path, same as env `DMF_GITLAB_SYNC_ROOT`",
								Aliases: []string{"r"},
							},
							&cli.StringFlag{
								Name:    "groups",
								Usage:   "only sync target groups, seperated by comma",
								Aliases: []string{"g"},
							},
						},
						Action: gitlab.GenerateLocalDirectories,
					},
					{
						Name:  "analyze",
						Usage: "analyze your gitlab groups and projects code lines",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "root",
								Usage:   "root path, same as env `DMF_GITLAB_SYNC_ROOT`",
								Aliases: []string{"r"},
							},
							&cli.StringFlag{
								Name:    "groups",
								Usage:   "only analyze target groups, seperated by comma",
								Aliases: []string{"g"},
							},
						},
						Action: gitlab.CodeLineSummary,
					},
					{
						Name:  "init",
						Usage: "init your gitlab project with protected branches",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "group", Usage: "`GROUP` of project", Aliases: []string{"g"}, Required: true},
							&cli.StringFlag{
								Name:    "branches",
								Usage:   "`BRANCHES` need protect, 0 3 4",
								Aliases: []string{"b"},
							},
						},
						Action: gitlab.InitProject,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
