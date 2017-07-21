package teamcity

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/sethgrid/pester"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
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
		Branch: "default:any",
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

		if statistics.NextHref != "" {
			url = c.host + statistics.NextHref
		} else {
			break
		}
	}
	return result, nil
}

func (c *Client) GetAllBuildConfigurations() (BuildConfigurations, error) {
	statistics := BuildConfigurations{}

	url := c.host + "/app/rest/buildTypes?="
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
