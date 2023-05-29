package vultr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sasukebo/doo/utils"

	"github.com/urfave/cli/v2"
)

var Cmd = &cli.Command{
	Name:    "vultr",
	Usage:   "执行一些vultr的辅助功能",
	Aliases: []string{"v"},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "key",
			Usage:   "vultr Personal Access Token, same as env `DOO_VULTR_API_KEY`",
			Aliases: []string{"k"},
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:   "show",
			Usage:  "show instance by id",
			Action: showInstance,
		},
		{
			Name:   "delete",
			Usage:  "delete instance by id",
			Action: deleteInstance,
		},
		{
			Name:   "status",
			Usage:  "show status of all instances",
			Action: showStatus,
		},
		{
			Name:  "new",
			Usage: "new vultr instance for v2ray docker container",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "scale",
					Usage: "how many instances to deploy",
					Value: 1,
				},
				&cli.StringFlag{
					Name:  "label",
					Usage: "label for instance",
				},
			},
			Action: newInstance,
		},
	},
}

const host = "https://api.vultr.com/"

func newRequest(accessToken, uri, method string, data interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if data != nil {
		err := json.NewEncoder(&buf).Encode(&data)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, host+uri, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)
	return req, nil
}

func newInstance(ctx *cli.Context) error {
	var (
		err         error
		accessToken string
	)

	accessToken, err = utils.MustGetStringArg(ctx, "key", "DOO_VULTR_API_KEY")
	if err != nil {
		return err
	}

	scale := ctx.Int("scale")
	if scale <= 0 {
		return nil
	}

	label := ctx.String("label")
	if label == "" {
		label = "instance"
	}

	for i := 0; i < scale; i++ {
		var _label = label
		if i > 0 {
			_label = fmt.Sprintf("%s_%v", label, i)
		}
		var data = map[string]interface{}{
			"region":      "sea",
			"plan":        "vc2-1c-1gb",
			"label":       _label,
			"os_id":       1946,
			"script_id":   "e2b38130-ca00-4903-8dc3-4ea0d153e374",
			"enable_ipv6": false,
			"backups":     "disabled",
			"sshkey_id":   []string{"108296cf-6b8b-4c0f-9121-1a77883afacc"},
		}
		req, err := newRequest(accessToken, "v2/instances", http.MethodPost, data)
		if err != nil {
			return err
		}
		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		content, _ := ioutil.ReadAll(rsp.Body)
		rsp.Body.Close()
		if rsp.StatusCode != http.StatusAccepted {
			fmt.Printf("StatusCode:%v, rsp: %s\n", rsp.StatusCode, string(content))
			return nil
		}
		fmt.Printf("create instance %v successful\n", _label)
	}
	return nil
}

type Instance struct {
	Id     string `json:"id"`
	Label  string `json:"label"`
	Status string `json:"power_status"`
	IP     string `json:"main_ip"`
}

type showStatusRsp struct {
	Instances []*Instance `json:"instances"`
}

func showStatus(ctx *cli.Context) error {
	var (
		err         error
		accessToken string
	)

	accessToken, err = utils.MustGetStringArg(ctx, "key", "DOO_VULTR_API_KEY")
	if err != nil {
		return err
	}
	req, err := newRequest(accessToken, "v2/instances", http.MethodGet, nil)
	if err != nil {
		return err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		fmt.Printf("StatusCode:%v, rsp: %s\n", rsp.StatusCode, string(content))
		return nil
	}
	var _rsp showStatusRsp
	if err = json.Unmarshal(content, &_rsp); err != nil {
		return err
	}
	fmt.Printf("%-36s %-10s %-7s %-15s\n", "ID", "Label", "Status", "IP")
	for _, i := range _rsp.Instances {
		fmt.Printf("%-36s %-10s %s %-15s\n", i.Id, i.Label, i.Status, i.IP)
	}

	return nil
}

func deleteInstance(ctx *cli.Context) error {
	var (
		err         error
		accessToken string
		id          string
	)
	accessToken, err = utils.MustGetStringArg(ctx, "key", "DOO_VULTR_API_KEY")
	if err != nil {
		return err
	}
	if ctx.NArg() != 1 {
		return fmt.Errorf("missing instance id")
	}
	id = ctx.Args().First()
	req, err := newRequest(accessToken, "v2/instances/"+id, http.MethodDelete, nil)
	if err != nil {
		return err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	rsp.Body.Close()
	if rsp.StatusCode != http.StatusNoContent {
		fmt.Printf("StatusCode:%v, rsp: %s\n", rsp.StatusCode, string(content))
		return nil
	}
	fmt.Printf("delete instance %s successful\n", id)
	return nil
}

func showInstance(ctx *cli.Context) error {
	var (
		err         error
		accessToken string
	)
	accessToken, err = utils.MustGetStringArg(ctx, "key", "DOO_VULTR_API_KEY")
	if err != nil {
		return err
	}
	if ctx.NArg() != 1 {
		return fmt.Errorf("missing instance id")
	}
	req, err := newRequest(accessToken, "v2/instances/"+ctx.Args().First(), http.MethodGet, nil)
	if err != nil {
		return err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	content, _ := ioutil.ReadAll(rsp.Body)
	rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		fmt.Printf("StatusCode:%v, rsp: %s\n", rsp.StatusCode, string(content))
		return nil
	}
	var data = make(map[string]interface{})
	if err = json.Unmarshal(content, &data); err != nil {
		return err
	}
	_mc, _ := json.MarshalIndent(data, "", " ")
	fmt.Println(string(_mc))

	return nil
}
