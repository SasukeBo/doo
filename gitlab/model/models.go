package model

type Group struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	FullPath string `json:"full_path"`
}

type Project struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	HttpUrlToRepo string `json:"http_url_to_repo"`
}
