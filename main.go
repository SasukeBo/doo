package main

import (
	"log"
	"os"
	"sasukebo/doo/gitlab"
	"sasukebo/doo/utils"

	"github.com/urfave/cli/v2"
)

var _gitlab = &cli.Command{
	Name:  "gitlab",
	Usage: "some convenient tools to handle your gitlab with your access key.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "host",
			Usage: "gitlab service host, same as env `DOO_GITLAB_HOST`",
		},
		&cli.StringFlag{
			Name:    "access_token",
			Usage:   "gitlab personal access_token, same as env `DOO_GITLAB_ACCESS_TOKEN`",
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
					Usage:   "root path, same as env `DOO_GITLAB_SYNC_ROOT`",
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
					Usage:   "root path, same as env `DOO_GITLAB_SYNC_ROOT`",
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
		{
			Name:  "force_delete_tag",
			Usage: "delete tag for project, ignore protect",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Usage: "target `PROJECT`", Aliases: []string{"p"}, Required: true},
				&cli.StringFlag{Name: "group", Usage: "target `GROUP`", Aliases: []string{"g"}, Required: true},
			},
			Action: gitlab.ForceDeleteTag,
		},
	},
}

var _now = &cli.Command{Name: "now", Action: utils.Now}

var _find = &cli.Command{
	Name:  "find",
	Usage: "find workdir to target directory, doo find -r / my_directory",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "root", Usage: "find target dir inside root path", Aliases: []string{"r"}, Required: false, Value: "/Users/sasukebo/workspace"},
	},
	Action: utils.CD,
}

func main() {
	app := &cli.App{
		Name:  "doo",
		Usage: "Do me a favor. ??????????????????????????????????????????",
		Commands: []*cli.Command{
			_gitlab,
			_now,
			_find,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
