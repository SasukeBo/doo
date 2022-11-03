package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sasukebo/doo/utils"
	"sync"

	"github.com/urfave/cli/v2"
	"github.com/xanzy/go-gitlab"
)

type Client struct {
	*gitlab.Client
	host  string
	token string
}

var (
	c     *Client
	cOnce sync.Once
)

func Init(ctx *cli.Context) error {
	var (
		err   error
		host  string
		token string
	)
	host, err = utils.MustGetStringArg(ctx, "host", "DOO_GITLAB_HOST")
	if err != nil {
		return err
	}
	token, err = utils.MustGetStringArg(ctx, "access_token", "DOO_GITLAB_ACCESS_TOKEN")
	if err != nil {
		return err
	}
	cOnce.Do(func() {
		_c, _err := gitlab.NewClient(token, gitlab.WithBaseURL(fmt.Sprintf("https://%s/api/v4", host)))
		c = &Client{Client: _c, host: host, token: token}
		err = _err

	})
	return err
}

func GetGroups() ([]*gitlab.Group, error) {
	var (
		outs []*gitlab.Group
		size = 100
		page = 1
	)
	for {
		gs, _, err := c.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{PerPage: size, Page: page},
		})
		if err != nil {
			return nil, err
		}
		outs = append(outs, gs...)
		if len(gs) < size {
			break
		}
		page++
	}

	return outs, nil
}

func GetGroupProjects(groupId int) ([]*gitlab.Project, error) {
	var (
		outs []*gitlab.Project
		size = 100
		page = 1
	)

	for {
		ps, _, err := c.Groups.ListGroupProjects(groupId, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{PerPage: size, Page: page},
		})
		if err != nil {
			return nil, err
		}
		outs = append(outs, ps...)
		if len(ps) < size {
			break
		}
		page++
	}

	return outs, nil
}

func GetGroupByName(name string) (*gitlab.Group, error) {
	var (
		size = 100
		page = 1
	)

	for {
		gs, _, err := c.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{PerPage: size, Page: page},
			Search:      func(s string) *string { return &s }(name),
		})
		if err != nil {
			return nil, err
		}
		for _, g := range gs {
			if g.FullName == name {
				return g, nil
			}
		}
		if len(gs) == 0 {
			return nil, nil
		}
		page++
	}
}

func GetProjectWithGroupId(id int, name string) (*gitlab.Project, error) {
	var (
		size = 100
		page = 1
	)

	for {
		ps, _, err := c.Groups.ListGroupProjects(id, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{PerPage: size, Page: page},
			Search:      pString(name),
		})
		if err != nil {
			return nil, err
		}
		for _, p := range ps {
			if p.Name == name {
				return p, nil
			}
		}
		if len(ps) == 0 {
			return nil, nil
		}
		page++
	}
}

func pString(s string) *string { return &s }

func ProtectProjectBranch(id int, branch string, levels []int) error {
	_, _, _ = c.Branches.CreateBranch(id, &gitlab.CreateBranchOptions{
		Branch: pString(branch),
		Ref:    pString("master"),
	})
	req1, err := http.NewRequest(http.MethodDelete, "https://"+c.host+fmt.Sprintf("/api/v4/projects/%v/protected_branches/%s", id, branch), nil)
	if err != nil {
		return fmt.Errorf("[ERROR] delete protect request failed for id %v: %v", id, err)
	}
	req1.Header.Add("PRIVATE-TOKEN", c.token)
	_, _ = http.DefaultClient.Do(req1)

	level := fmt.Sprintf("push_access_level=%v&merge_access_level=%v&unprotect_access_level=%v", levels[0], levels[1], levels[2])
	req, err := http.NewRequest(http.MethodPost, "https://"+c.host+fmt.Sprintf("/api/v4/projects/%v/protected_branches?name=%v&%s", id, branch, level), nil)
	if err != nil {
		return fmt.Errorf("[ERROR] create protect request failed for id %v: %v", id, err)
	}
	req.Header.Add("PRIVATE-TOKEN", c.token)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("[ERROR] do protect request failed for id %v: %v", id, err)
	}
	if rsp.StatusCode != http.StatusOK && rsp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return fmt.Errorf("[ERROR] protect branch failed for id %v, read error body failed: %v", id, err)
		}
		return fmt.Errorf("[ERROR] protect branch failed for id %v: %v", id, string(body))
	}
	return nil
}

func DeleteProjectTag(id int, tag string) error {
	_, err := c.Tags.DeleteTag(id, tag)
	return err
}
