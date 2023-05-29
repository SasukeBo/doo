package main

import (
	"log"
	"os"
	"sasukebo/doo/gitlab"
	"sasukebo/doo/harbor"
	"sasukebo/doo/utils"
	"sasukebo/doo/vultr"

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

var _harbor = &cli.Command{
	Name:  "harbor",
	Usage: "执行一些Harbor的辅助功能",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "config", Usage: "指定Harbor api v2.0的Authorization信息文件地址", Aliases: []string{"c"}, Required: true},
	},
	Subcommands: []*cli.Command{
		{
			Name:  "delete_artifact_by_reference",
			Usage: "删除指定的镜像版本",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Usage: "指定项目名称", Aliases: []string{"p"}},
				&cli.StringFlag{Name: "repository", Usage: "指定仓库名称", Aliases: []string{"r"}},
			},
			Action: harbor.DeleteTagetArtifacts,
		},
		{
			Name:  "clean_artifacts",
			Usage: "清理N天未被拉取过的，且无标签的镜像文件",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Usage: "指定项目名称", Aliases: []string{"p"}},
				&cli.StringFlag{Name: "repository", Usage: "指定仓库名称，如果不指定则清理整个项目", Aliases: []string{"r"}},
				&cli.StringFlag{Name: "days_not_pulled", Usage: "清理N天内没有被拉取过的镜像", Aliases: []string{"d"}},
				&cli.BoolFlag{Name: "disable_ignore", Usage: "取消对有标签的镜像过滤", Aliases: []string{"i"}},
			},
			Action: harbor.CleanArtifacts,
		},
		{
			Name:  "delete_project",
			Usage: "删除项目",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "project", Usage: "指定项目名称", Aliases: []string{"p"}},
			},
			Action: harbor.DeleteProject,
		},
	},
}

func main() {
	app := &cli.App{
		Name:  "doo",
		Usage: "Do me a favor. 集成一些小工具来提高工作效率",
		Commands: []*cli.Command{
			_gitlab,
			_now,
			_find,
			_harbor,
			vultr.Cmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
