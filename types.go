package teamcity

import "github.com/sethgrid/pester"

type Client struct {
	HTTPClient *pester.Client
	username   string
	password   string
	host       string
}

type BuildLocator struct {
	BuildType string `yaml:"build_type"`
	Branch    string `yaml:"branch"`
	Status    string `yaml:"status"`
	Running   string `yaml:"running"`
	Canceled  string `yaml:"canceled"`
	Count     string `yaml:"count"`
}

type BuildStatistics struct {
	Count    int `json:"count"`
	Property []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"property"`
	UsedFilter BuildLocator
}

type Build struct {
	ID          int    `json:"id"`
	BuildTypeID string `json:"buildTypeId"`
	Number      string `json:"number"`
	Status      string `json:"status"`
	State       string `json:"state"`
	BranchName  string `json:"branchName"`
	Href        string `json:"href"`
	WebURL      string `json:"webUrl"`
}

type Builds struct {
	Count    int     `json:"count"`
	Href     string  `json:"href"`
	NextHref string  `json:"nextHref"`
	Build    []Build `json:"build"`
}

type BuildConfigurations struct {
	Count     int    `json:"count"`
	Href      string `json:"href"`
	BuildType []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		ProjectName string `json:"projectName"`
		ProjectID   string `json:"projectId"`
		Href        string `json:"href"`
		WebURL      string `json:"webUrl"`
		Description string `json:"description,omitempty"`
		Paused      bool   `json:"paused,omitempty"`
	} `json:"buildType"`
}
