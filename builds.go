package teamcity

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"unicode"
)

func New(host, username, password string) *Client {
	client := &Client{}
	cookieJar, _ := cookiejar.New(nil)

	client.HTTPClient = pester.New()
	client.HTTPClient.MaxRetries = 5
	client.HTTPClient.Backoff = pester.ExponentialBackoff
	client.HTTPClient.KeepLog = true
	client.HTTPClient.Jar = cookieJar

	client.username = username
	client.password = password
	client.host = host

	return client
}

func NewBuildLocator() *BuildLocator {
	return &BuildLocator{
		Branch: "",
		Count:  "1",
	}
}

func processParams(str interface{}) string {
	conv := structs.Map(str)
	res := ""

	for k, v := range conv {
		if v == "" {
			continue
		}
		fieldName := []rune(string(k))
		fieldName[0] = unicode.ToLower(fieldName[0])
		res += "," + string(fieldName) + ":" + v.(string)
	}

	return strings.TrimLeft(res, ",")
}

func (c *Client) GetBuildStat(id int) (BuildStatistics, error) {
	statistics := BuildStatistics{}
	url := fmt.Sprint(c.host, "/app/rest/builds/id:", id, "/statistics")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return statistics, err
	}
	req.Header.Add("Accept", "application/json")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return statistics, err
	}
	if res.StatusCode == 401 {
		req.SetBasicAuth(c.username, c.password)
		res, err = c.HTTPClient.Do(req)
		if err != nil {
			return statistics, err
		}
	}

	if res.StatusCode == 404 {
		return statistics, nil
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return statistics, err
	}

	if err := json.Unmarshal(body, &statistics); err != nil {
		return statistics, err
	}

	return statistics, nil
}

func (c *Client) GetBuildsByParams(bl BuildLocator) (Builds, error) {
	result := Builds{}
	url := c.host + "/app/rest/builds/?locator=" + processParams(bl)

	for {
		statistics := Builds{}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return statistics, err
		}
		req.Header.Add("Accept", "application/json")
		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return statistics, err
		}
		if res.StatusCode == 401 {
			req.SetBasicAuth(c.username, c.password)
			res, err = c.HTTPClient.Do(req)
			if err != nil {
				return statistics, err
			}
		}

		if res.StatusCode == 404 {
			return statistics, nil
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return statistics, err
		}

		if err := json.Unmarshal(body, &statistics); err != nil {
			return statistics, err
		}

		for i := range statistics.Build {
			result.Build = append(result.Build, statistics.Build[i])
		}

		if bl.Count == "" && statistics.NextHref != "" {
			url = c.host + statistics.NextHref
		} else {
			break
		}
	}

	for i := range result.Build {
		if result.Build[i].BranchName == "" {
			d, err := c.GetBuildDetails(result.Build[i].ID)
			if err != nil {
				log.Errorf("Failed to query build details for build ID %d: %v", result.Build[i].ID, err)
				continue
			}
			p := []string{}
			for v := range d.Property {
				r, _ := regexp.MatchString("build\\.vcs\\.branch", d.Property[v].Name)
				if r {
					p = append(p, strings.Replace(d.Property[v].Value, "refs/heads/", "", -1))
				}
			}
			// don't need branch name if build configuration has only one branch
			if len(uniqSlice(p)) <= 1 {
				result.Build[i].BranchName = ""
			} else {
				result.Build[i].BranchName = p[0]
			}
		}
	}
	return result, nil
}

func (c *Client) GetBuildDetails(id BuildID) (BuildDetails, error) {
	buildDetails := BuildDetails{}

	url := fmt.Sprint(c.host, "/app/rest/builds/id:", id, "/resulting-properties")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return buildDetails, err
	}
	req.Header.Add("Accept", "application/json")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return buildDetails, err
	}
	if res.StatusCode == 401 {
		req.SetBasicAuth(c.username, c.password)
		res, err = c.HTTPClient.Do(req)
		if err != nil {
			return buildDetails, err
		}
	}

	if res.StatusCode == 404 {
		return buildDetails, nil
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return buildDetails, err
	}

	if err := json.Unmarshal(body, &buildDetails); err != nil {
		return buildDetails, err
	}
	return buildDetails, nil
}

func (c *Client) GetAllBuildConfigurations() (BuildConfiguration, error) {
	statistics := BuildConfiguration{}

	url := c.host + "/app/rest/buildTypes"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return statistics, err
	}
	req.Header.Add("Accept", "application/json")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return statistics, err
	}
	if res.StatusCode == 401 {
		req.SetBasicAuth(c.username, c.password)
		res, err = c.HTTPClient.Do(req)
		if err != nil {
			return statistics, err
		}
	}

	if res.StatusCode == 404 {
		return statistics, nil
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return statistics, err
	}

	if err := json.Unmarshal(body, &statistics); err != nil {
		return statistics, err
	}

	return statistics, nil

}

func (c *Client) GetAllBranches(bt BuildTypeID) (Branches, error) {
	branches := Branches{}

	url := c.host + "/app/rest/buildTypes/id:" + string(bt) + "/branches"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return branches, err
	}
	req.Header.Add("Accept", "application/json")
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return branches, err
	}
	if res.StatusCode == 401 {
		req.SetBasicAuth(c.username, c.password)
		res, err = c.HTTPClient.Do(req)
		if err != nil {
			return branches, err
		}
	}

	if res.StatusCode == 404 {
		return branches, nil
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return branches, err
	}

	if err := json.Unmarshal(body, &branches); err != nil {
		return branches, err
	}

	return branches, nil
}
