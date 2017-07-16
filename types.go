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

type BuildTypeLocator struct {
	Id string
}

type BuildStatistics struct {
	Count    int `json:"count"`
	Property []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"property"`
	UsedFilter BuildLocator
}

type Builds struct {
	ID            int    `json:"id"`
	BuildTypeID   string `json:"buildTypeId"`
	Number        string `json:"number"`
	Status        string `json:"status"`
	State         string `json:"state"`
	BranchName    string `json:"branchName"`
	DefaultBranch bool   `json:"defaultBranch"`
	Href          string `json:"href"`
	WebURL        string `json:"webUrl"`
	StatusText    string `json:"statusText"`
	BuildType     struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ProjectName string `json:"projectName"`
		ProjectID   string `json:"projectId"`
		Href        string `json:"href"`
		WebURL      string `json:"webUrl"`
	} `json:"buildType"`
	QueuedDate string `json:"queuedDate"`
	StartDate  string `json:"startDate"`
	FinishDate string `json:"finishDate"`
	Triggered  struct {
		Type    string `json:"type"`
		Details string `json:"details"`
		Date    string `json:"date"`
	} `json:"triggered"`
	LastChanges struct {
		Change []struct {
			ID       int    `json:"id"`
			Version  string `json:"version"`
			Username string `json:"username"`
			Date     string `json:"date"`
			Href     string `json:"href"`
			WebURL   string `json:"webUrl"`
		} `json:"change"`
		Count int `json:"count"`
	} `json:"lastChanges"`
	Changes struct {
		Href string `json:"href"`
	} `json:"changes"`
	Revisions struct {
		Count    int `json:"count"`
		Revision []struct {
			Version         string `json:"version"`
			VcsBranchName   string `json:"vcsBranchName"`
			VcsRootInstance struct {
				ID        string `json:"id"`
				VcsRootID string `json:"vcs-root-id"`
				Name      string `json:"name"`
				Href      string `json:"href"`
			} `json:"vcs-root-instance"`
		} `json:"revision"`
	} `json:"revisions"`
	Agent struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		TypeID int    `json:"typeId"`
		Href   string `json:"href"`
	} `json:"agent"`
	TestOccurrences struct {
		Count     int    `json:"count"`
		Href      string `json:"href"`
		Default   bool   `json:"default"`
		Passed    int    `json:"passed"`
		Failed    int    `json:"failed"`
		NewFailed int    `json:"newFailed"`
		Ignored   int    `json:"ignored"`
	} `json:"testOccurrences"`
	ProblemOccurrences struct {
		Count   int    `json:"count"`
		Href    string `json:"href"`
		Default bool   `json:"default"`
	} `json:"problemOccurrences"`
	Artifacts struct {
		Href string `json:"href"`
	} `json:"artifacts"`
	RelatedIssues struct {
		Href string `json:"href"`
	} `json:"relatedIssues"`
	Properties struct {
		Count    int `json:"count"`
		Property []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Inherited bool   `json:"inherited"`
		} `json:"property"`
	} `json:"properties"`
	Statistics struct {
		Href string `json:"href"`
	} `json:"statistics"`
	SnapshotDependencies struct {
		Count int `json:"count"`
		Build []struct {
			ID            int    `json:"id"`
			BuildTypeID   string `json:"buildTypeId"`
			Number        string `json:"number"`
			Status        string `json:"status"`
			State         string `json:"state"`
			BranchName    string `json:"branchName"`
			DefaultBranch bool   `json:"defaultBranch"`
			Href          string `json:"href"`
			WebURL        string `json:"webUrl"`
		} `json:"build"`
	} `json:"snapshot-dependencies"`
	ArtifactDependencies struct {
		Count int `json:"count"`
		Build []struct {
			ID            int    `json:"id"`
			BuildTypeID   string `json:"buildTypeId"`
			Number        string `json:"number"`
			Status        string `json:"status"`
			State         string `json:"state"`
			BranchName    string `json:"branchName"`
			DefaultBranch bool   `json:"defaultBranch"`
			Href          string `json:"href"`
			WebURL        string `json:"webUrl"`
		} `json:"build"`
	} `json:"artifact-dependencies"`
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
