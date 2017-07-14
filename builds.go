package teamcity

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	// "fmt"
	"unicode"
	"github.com/fatih/structs"
	"github.com/sethgrid/pester"
	// "errors"
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

func processParams(str interface{}) string {
	conv := structs.Map(str)
	res := ""

	for k, v := range conv {
		if v == ""  {
			continue
		}
		fieldName := []rune(string(k))
		fieldName[0] = unicode.ToLower(fieldName[0])
		res += "," + string(fieldName) + ":" + v.(string)
	}

	//debug
	// fmt.Println("Trimmed line: ", strings.TrimLeft(res, ","))

	return strings.TrimLeft(res, ",")
}

func (c *Client) GetBuildStatistics(bl BuildLocator) (BuildStatistics, error) {
  statistics := BuildStatistics{}
	statistics.UsedFilter = bl

	url := c.host + "/app/rest/builds/" + processParams(bl) + "/statistics"
	// fmt.Printf("url %s\n", url)
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
		// debug
		// log.Println("Add basic auth to GetBuildStatistics")
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
	// fmt.Println(string(body))
  if err != nil {
    return statistics, err
  }

	// debug
	// fmt.Println(res)
	// fmt.Println(string(body))

	if err := json.Unmarshal(body, &statistics); err != nil {
		return statistics, err
	}

	// debug
	// fmt.Println("\n")
	// fmt.Println(statistics)
  return statistics, nil
}

func (c *Client) GetBuildsByParams(btl BuildTypeLocator, bl BuildLocator) (Builds, error) {
  statistics := Builds{}

  url := c.host + "/app/rest/buildTypes/" + processParams(btl) + "/builds/" + processParams(bl)
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
		// debug
		// log.Println("Add basic auth to GetBuildsByParams")
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
	// fmt.Println(string(body))
  if err != nil {
    return statistics, err
  }

	// debug
	// fmt.Println(res)
	// fmt.Println(string(body))

	if err := json.Unmarshal(body, &statistics); err != nil {
		return statistics, err
	}

	// debug
	// fmt.Println("\n")
	// fmt.Println(statistics)

  return statistics, nil
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
		// debug
		// log.Println("Add basic auth to GetBuildsByParams")
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

	// debug
	// fmt.Println(res)
	// fmt.Println(string(body))

	if err := json.Unmarshal(body, &statistics); err != nil {
		return statistics, err
	}

	// debug
	// fmt.Println("\n")
	// fmt.Println(statistics)

  return statistics, nil

}
