package gitlab

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sasukebo/doo/gitlab/client"
	"sasukebo/doo/utils"
	"strings"

	"github.com/urfave/cli/v2"
)

var ignoreGroups = map[string]struct{}{
	//"uino-thingclub-mp-ussjs":  {},
	//"thingyouwe-github-mirror": {},
	//"thingyouwe-micro":         {},
}

var ignoreProjects = map[string]struct{}{
	//"traefik":  {},
	//"micro":    {},
	//"go-micro": {},
	//"gengine":  {},
	//"go-druid": {},
	//"doccano":  {},
}

// CodeLineSummary 代码行数统计
func CodeLineSummary(ctx *cli.Context) error {
	var (
		err  error
		root string
		gs   = make(map[string]struct{})
	)
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

	var total int

	for _, group := range groups {
		if _, ok := gs[group.FullName]; len(gs) > 0 && !ok {
			continue
		}
		if _, ok := ignoreGroups[group.FullName]; ok {
			continue
		}
		fmt.Printf("*** Gitlab 项目组： %s ***\n", group.FullName)
		fmt.Printf("%s\n\n", group.Description)
		projects, err := client.GetGroupProjects(group.ID)
		if err != nil {
			//fmt.Printf("--- [ERROR] Get projects for group %s failed: %v\n", group.Name, err)
			continue
		}

		var groupC int

		for _, project := range projects {
			if _, ok := ignoreProjects[project.Name]; ok {
				continue
			}
			dir := fmt.Sprintf("%s/%s/%s", root, group.FullPath, project.Path)

			line, err := countProject(dir)
			if err != nil {
				//fmt.Printf("--- [ERROR] Count line for project %s failed: %v\n", project.Name, err)
				continue
			}
			if line < 10 {
				continue
			}
			fmt.Printf("  %s %v行\n", project.Name, line)
			groupC += line
		}
		if groupC < 10 {
			continue
		}
		fmt.Printf("\n  总计: %v行\n\n", groupC)
		total += groupC
	}

	fmt.Printf("\n最终统计： %v行\n", total)
	return nil
}

var fileExtInCount = map[string]struct{}{
	".md":    {},
	".yaml":  {},
	".js":    {},
	".py":    {},
	".wxml":  {},
	".json":  {},
	".sh":    {},
	".mod":   {},
	".bat":   {},
	".yml":   {},
	".sql":   {},
	".vue":   {},
	".proto": {},
	".txt":   {},
	".lua":   {},
	".css":   {},
	".toml":  {},
	".wxss":  {},
	".conf":  {},
	".scss":  {},
	".bash":  {},
	".zsh":   {},
	".styl":  {},
	".html":  {},
	".xml":   {},
	".go":    {},
}

func countProject(dir string) (int, error) {
	var buf []byte
	b := bytes.NewBuffer(buf)

	cmd := exec.Command("git", "ls-files")
	cmd.Dir = dir
	cmd.Stdout = b
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	var total int
	files := strings.Split(b.String(), "\n")
	for _, f := range files {
		if f == "" {
			continue
		}

		if _, ok := fileExtInCount[filepath.Ext(f)]; !ok {
			continue
		}
		if strings.Contains(f, ".pb.") {
			continue
		}

		c, err := countLine(dir, f)
		if err != nil {
			return 0, err
		}
		total += c
	}

	return total, nil
}

func countLine(dir, path string) (int, error) {
	f, err := os.Open(dir + "/" + path)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = f.Close()
	}()

	r := bufio.NewReader(f)
	var c int
	for {
		_, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		c++
	}
	return c, nil
}
