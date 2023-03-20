package harbor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sasukebo/doo/utils"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

const (
	apiAddress = "https://dk.uino.cn/api/v2.0"
)

type Config struct {
	Cookie   string `json:"cookie"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

func getConfig(ctx *cli.Context) (*Config, error) {
	configFilePath, err := utils.MustGetStringArg(ctx, "config", "")
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var config Config
	_ = json.Unmarshal(content, &config)
	return &config, nil
}

func DeleteTagetArtifacts(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return fmt.Errorf("the reference of the artifact required, can be digest or tag")
	}

	var (
		project    string
		repository string
		err        error
		reference  = ctx.Args().Get(0)
	)

	config, err := getConfig(ctx)
	if err != nil {
		return err
	}
	project, err = utils.MustGetStringArg(ctx, "project", "")
	if err != nil {
		return err
	}
	repository, err = utils.MustGetStringArg(ctx, "repository", "")
	if err != nil {
		return err
	}

	var url = fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts/%s", apiAddress, project, repository, reference)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	setHeader(req, config)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	var data = make(map[string]interface{})
	if json.Unmarshal(content, &data) == nil {
		_content, _ := json.MarshalIndent(&data, "", " ")
		fmt.Println(string(_content))
	}

	return nil
}

func setHeader(req *http.Request, config *Config) {
	req.SetBasicAuth(config.Username, config.Password)
	req.Header.Add("X-Harbor-CSRF-Token", config.Token)
	req.Header.Set("Cookie", config.Cookie)
	req.Header.Set("Accept", "application/json")
}

type repositoryListItem struct {
	Name string `json:"name"`
}

var ignoreTagExp = regexp.MustCompile(`([123]+\.[0-9]+\.[0-9]+)|staging|test|dev`)

func CleanArtifacts(ctx *cli.Context) error {
	var (
		project    string
		repository string
		day        int
		err        error
	)

	config, err := getConfig(ctx)
	if err != nil {
		return err
	}
	project, err = utils.MustGetStringArg(ctx, "project", "")
	if err != nil {
		return err
	}
	days, err := utils.MustGetStringArg(ctx, "days_not_pulled", "")
	if err != nil {
		return err
	}
	day, err = strconv.Atoi(days)
	if err != nil {
		return err
	}
	disableIgnore := ctx.Bool("disable_ignore")

	repository = ctx.String("repository")
	var repositories []string
	if repository != "" {
		repositories = append(repositories, strings.Split(repository, ",")...)
	} else {
		var page, limit = 1, 100
		for {
			_repositories, err := getRepositories(project, config, page, limit)
			if err != nil {
				return err
			}
			repositories = append(repositories, _repositories...)
			if len(_repositories) < limit {
				break
			} else {
				page++
			}
		}
	}

	var deleteChan = make(chan *D, 5)
	var finishChan = make(chan struct{}, 5)

	for i := 0; i < 5; i++ {
		go func(dc chan *D, fc chan struct{}) {
			for {
				select {
				case d := <-dc:
					deleteArtifact(project, d.Repo, d.Digest, config)
					fc <- struct{}{}
				}
			}
		}(deleteChan, finishChan)
	}

	for _, repo := range repositories {
		fmt.Println("[INFO] process repo:", repo)
		artifacts, err := getTotalArtifacts(project, repo, day, config)
		if err != nil {
			return err
		}

		var digests []string
		for _, a := range artifacts {
			var ignore bool
			for _, t := range a.Tags {
				// fmt.Printf("%s %v\n", t.Name, ignoreTagExp.MatchString(t.Name))
				ignore = ignore || ignoreTagExp.MatchString(t.Name)
			}
			if ignore && !disableIgnore {
				continue
			}
			digests = append(digests, a.Digest)
		}
		var total = len(digests)
		if total > 0 {
			go func(dc chan *D, _total *int) {
				for _, digest := range digests {
					dc <- &D{repo, digest}
				}
			}(deleteChan, &total)

			var finishCount int
			for {
				select {
				case <-finishChan:
					finishCount++
				}

				if total == finishCount {
					break
				}
			}
		}

		fmt.Println(repo, "find:", len(artifacts), "deleted:", total)
	}

	return nil
}

type D struct {
	Repo   string
	Digest string
}

func getRepositories(project string, config *Config, page, limit int) ([]string, error) {
	var outs []string
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(
		"%s/projects/%s/repositories?page=%v&page_size=%v",
		apiAddress, project, page, limit,
	), nil)
	if err != nil {
		return nil, err
	}
	setHeader(req, config)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(content))
	}

	var repositoryList []*repositoryListItem
	_ = json.Unmarshal(content, &repositoryList)
	for _, item := range repositoryList {
		namePieces := strings.Split(item.Name, "/")
		if len(namePieces) > 1 {
			outs = append(outs, namePieces[1])
		}
	}

	return outs, nil
}

type Tag struct {
	Name string `json:"name"`
}

type Artifact struct {
	Digest   string `json:"digest"`
	PullTime string `json:"pull_time"`
	Tags     []*Tag `json:"tags"`
}

func getArtifacts(project, repo string, config *Config, day, page, limit int) ([]*Artifact, error) {
	var outs []*Artifact
	timeRange := fmt.Sprintf(
		"[\"0001-01-01 00:00:00\"~\"%s 00:00:00\"]",
		time.Now().Add(time.Duration(-1*day*24*int(time.Hour))).Format("2006-01-02"),
	)
	timeRangeEscape := url.PathEscape(timeRange)

	uri := fmt.Sprintf(
		"%s/projects/%s/repositories/%s/artifacts?page=%v&page_size=%v&sort=pull_time&q=pull_time=%s",
		apiAddress, project, repo, page, limit, timeRangeEscape,
	)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	setHeader(req, config)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(string(content))
	}

	_ = json.Unmarshal(content, &outs)
	return outs, nil
}

func getTotalArtifacts(project, repo string, day int, config *Config) ([]*Artifact, error) {
	var outs []*Artifact
	var page, limit = 1, 100
	for {
		_outs, err := getArtifacts(project, repo, config, day, page, limit)
		if err != nil {
			return nil, err
		}
		outs = append(outs, _outs...)
		if len(_outs) < limit {
			break
		} else {
			page++
		}
	}

	return outs, nil
}

func deleteArtifact(project, repo, sha string, config *Config) error {
	uri := fmt.Sprintf(
		"%s/projects/%s/repositories/%s/artifacts/%s",
		apiAddress, project, repo, sha,
	)
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	setHeader(req, config)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf(string(content))
	}

	fmt.Printf("DELETE %s ok\n", sha)
	return nil
}

func DeleteProject(ctx *cli.Context) error {
	config, err := getConfig(ctx)
	if err != nil {
		return err
	}
	project, err := utils.MustGetStringArg(ctx, "project", "")
	if err != nil {
		return err
	}
	var repositoryNames []string
	var page, limit = 1, 100
	for {
		names, err := getRepositories(project, config, page, limit)
		if err != nil {
			return err
		}
		repositoryNames = append(repositoryNames, names...)
		if len(names) < limit {
			break
		} else {
			page++
		}
	}

	var (
		deleteChan = make(chan string, 10)
		finishChan = make(chan struct{}, 5)
	)

	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case repo := <-deleteChan:
					deleteRepository(project, repo, config)
					finishChan <- struct{}{}
				}
			}
		}()
	}

	go func() {
		for _, name := range repositoryNames {
			deleteChan <- name
		}
	}()

	var finished int
	for {
		select {
		case <-finishChan:
			finished++
		}
		if finished == len(repositoryNames) {
			break
		}
	}

	uri := fmt.Sprintf("%s/projects/%s", apiAddress, project)
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	setHeader(req, config)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf(string(content))
	}
	fmt.Printf("%s deleted\n", project)
	return nil
}

func deleteRepository(project, repository string, config *Config) error {
	uri := fmt.Sprintf("%s/projects/%s/repositories/%s", apiAddress, project, repository)
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	setHeader(req, config)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf(string(content))
	}

	fmt.Println("delete", project+"/"+repository)
	return nil
}
