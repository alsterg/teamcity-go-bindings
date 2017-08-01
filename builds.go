package teamcity

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/sethgrid/pester"
)

func New(url, username, password string, concurrencyLimit int) *Client {
	if concurrencyLimit == 0 {
		concurrencyLimit = 1000
	}

	http := pester.New()
	http.Concurrency = concurrencyLimit
	http.MaxRetries = 5
	http.Backoff = pester.ExponentialBackoff
	http.KeepLog = true

	client := &Client{
		HTTPClient: http,
		URL:        url,
		Username:   username,
		Password:   password,
		Flow:       make(chan DataFlow, 10000),
		semaphore:  make(chan bool, concurrencyLimit),
	}

	go client.processDataFlow()

	return client
}

func (c *Client) Close() {
	close(c.Flow)
}

func (c *Client) GetBuildDetails(id BuildID) (BuildDetails, error) {
	buildDetails := BuildDetails{}
	chData := DataFlow{
		Response: make(chan *http.Response, 1),
	}

	url := fmt.Sprint(c.URL, "/app/rest/builds/id:", id, "/resulting-properties")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return buildDetails, err
	}

	chData.Request = req
	c.Flow <- chData

	for res := range chData.Response {
		body, err := processResponse(res)
		if err != nil {
			return buildDetails, err
		}

		if err := json.Unmarshal(body, &buildDetails); err != nil {
			return buildDetails, err
		}
	}
	return buildDetails, nil
}

func (c *Client) GetAllBuildConfigurations() (BuildConfigurations, error) {
	buildConfigs := BuildConfigurations{}
	chData := DataFlow{
		Response: make(chan *http.Response, 1),
	}

	url := fmt.Sprint(c.URL, "/app/rest/buildTypes")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return buildConfigs, err
	}

	chData.Request = req
	c.Flow <- chData

	for res := range chData.Response {
		body, err := processResponse(res)
		if err != nil {
			return buildConfigs, err
		}
		if err := json.Unmarshal(body, &buildConfigs); err != nil {
			return buildConfigs, err
		}
	}
	return buildConfigs, nil
}

func (c *Client) GetAllBranches(bt BuildTypeID) (Branches, error) {
	branches := Branches{}
	chData := DataFlow{
		Response: make(chan *http.Response, 1),
	}

	url := fmt.Sprint(c.URL, "/app/rest/buildTypes/id:", bt, "/branches")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return branches, err
	}

	chData.Request = req
	c.Flow <- chData

	for res := range chData.Response {
		body, err := processResponse(res)
		if err != nil {
			return branches, err
		}
		if err := json.Unmarshal(body, &branches); err != nil {
			return branches, err
		}
	}

	for _, branch := range branches.Branches {
		branch.Name = strings.Replace(branch.Name, "/refs/heads/", "", -1)
	}
	return branches, nil
}

func (c *Client) GetBuildsByParams(bl BuildLocator) (Builds, error) {
	builds := Builds{}

	url := fmt.Sprint(c.URL, "/app/rest/builds/?locator=", convertLocatorToString(bl))

	for {
		buildsIter := Builds{}
		chData := DataFlow{
			Response: make(chan *http.Response, 1),
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return builds, err
		}

		chData.Request = req
		c.Flow <- chData

		for res := range chData.Response {
			body, err := processResponse(res)
			if err != nil {
				return builds, err
			}
			if err := json.Unmarshal(body, &buildsIter); err != nil {
				return builds, err
			}
		}

		for i := range buildsIter.Builds {
			if buildsIter.Builds[i].BranchName == "" {
				buildsIter.Builds[i].BranchName = "<default>"
			}
			builds.Builds = append(builds.Builds, buildsIter.Builds[i])
		}

		if bl.Count == 0 && buildsIter.NextHref != "" {
			url = c.URL + buildsIter.NextHref
		} else {
			break
		}
	}
	return builds, nil
}

func (c *Client) GetBuildStat(id BuildID) (BuildStatistics, error) {
	stat := BuildStatistics{}
	chData := DataFlow{
		Response: make(chan *http.Response, 1),
	}

	url := fmt.Sprint(c.URL, "/app/rest/builds/id:", id, "/statistics")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return stat, err
	}

	chData.Request = req
	c.Flow <- chData

	for res := range chData.Response {
		body, err := processResponse(res)
		if err != nil {
			return stat, err
		}
		if err := json.Unmarshal(body, &stat); err != nil {
			return stat, err
		}
	}

	return stat, nil
}

func (c *Client) getBuildsByParamsPipelined(in <-chan BuildLocator, out chan<- Build) {
	wg := new(sync.WaitGroup)
	for filter := range in {
		wg.Add(1)
		go func(f BuildLocator) {
			defer wg.Done()
			build, err := c.GetBuildsByParams(f)
			if err != nil {
				log.Println(err)
				return
			}
			if len(build.Builds) > 0 {
				out <- build.Builds[0]
			} else {
				log.Printf("No builds found for build configuration '%s', branch '%s'", f.BuildType, f.Branch)
				return
			}
		}(filter)
	}
	wg.Wait()
	close(out)
}

func (c *Client) GetLatestBuild(bl BuildLocator) (Builds, error) {
	chFilters := make(chan BuildLocator)
	chBuilds := make(chan Build)
	builds := Builds{}

	go c.getBuildsByParamsPipelined(chFilters, chBuilds)

	wg1 := new(sync.WaitGroup)
	wg1.Add(1)
	go func() {
		defer wg1.Done()
		for build := range chBuilds {
			builds.Builds = append(builds.Builds, build)
		}
	}()

	// get build types
	buildTypes := []BuildTypeID{}
	if bl.BuildType == "" {
		bt, err := c.GetAllBuildConfigurations()
		if err != nil {
			log.Fatal(err)
		}
		for i := range bt.BuildTypes {
			buildTypes = append(buildTypes, bt.BuildTypes[i].ID)
		}
	} else {
		buildTypes = append(buildTypes, bl.BuildType)
	}

	// get branches and combine filters
	wg2 := new(sync.WaitGroup)
	if bl.Branch == "" {
		for _, buildType := range buildTypes {
			wg2.Add(1)
			go func(bt BuildTypeID) {
				defer wg2.Done()
				branches, err := c.GetAllBranches(bt)
				if err != nil {
					log.Fatal(err)
				}
				if len(branches.Branches) == 1 {
					branches.Branches[0].Name = ""
				}
				for _, branch := range branches.Branches {
					f := BuildLocator{
						BuildType: bt,
						Branch:    branch.Name,
						Status:    bl.Status,
						Running:   bl.Running,
						Canceled:  bl.Canceled,
						Count:     1,
					}
					chFilters <- f
				}
			}(buildType)
		}
		wg2.Wait()
		close(chFilters)
	} else {
		for _, buildType := range buildTypes {
			wg2.Add(1)
			go func(bt BuildTypeID) {
				defer wg2.Done()
				f := BuildLocator{
					BuildType: bt,
					Branch:    bl.Branch,
					Status:    bl.Status,
					Running:   bl.Running,
					Canceled:  bl.Canceled,
					Count:     1,
				}
				chFilters <- f
			}(buildType)
		}
		wg2.Wait()
		close(chFilters)
	}

	wg1.Wait()
	return builds, nil
}
