package teamcity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/fatih/structs"
	"github.com/orcaman/concurrent-map"
	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"
)

func New(host, username, password string) *Client {
	client := &Client{}
	cookieJar, _ := cookiejar.New(nil)

	client.HTTPClient = pester.New()
	client.HTTPClient.MaxRetries = 5
	client.HTTPClient.Backoff = pester.ExponentialBackoff
	client.HTTPClient.KeepLog = true
	client.HTTPClient.Jar = cookieJar
	client.HTTPClient.Timeout = time.Duration(60 * time.Second)

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
		fieldName := []rune(k)
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
	req.Close = true
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
		req.Close = true
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
				re := regexp.MustCompile(`build\.vcs\.branch`)
				if re.MatchString(d.Property[v].Name) {
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
	req.Close = true
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

	url := c.host + "/app/rest/buildTypes?locator=paused:false"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return statistics, err
	}
	req.Header.Add("Accept", "application/json")
	req.Close = true
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

	// log.Println(string(body))
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
	req.Close = true
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return branches, err
	}
	if res.StatusCode == 401 {
		log.Println("Have to reauth...")
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

func (c *Client) GetLatestBuild(bl BuildLocator) /*(Builds, error)*/ {
	// start := time.Now()

	nBranches := cmap.New()
	nBuilds := cmap.New()
	// builds := Builds{}

	// type BuildLocator struct {
	// 	BuildType string `yaml:"build_type"`
	// 	Branch    string `yaml:"branch"`
	// 	Status    string `yaml:"status"`
	// 	Running   string `yaml:"running"`
	// 	Canceled  string `yaml:"canceled"`
	// 	Count     string `yaml:"count"`
	// }
	bc := make(chan BuildTypeID)
	chBtb := make(chan map[BuildTypeID][]Branch)
	chL := make(chan BuildLocator, 10000)

	wg := new(sync.WaitGroup)

	wg.Add(1)
	totalBuildFilters := 0
	mutex2 := new(sync.Mutex)
	go func() {
		defer wg.Done()

		for m := range chBtb {
			for k, v := range m {
				if v == nil {
					f := BuildLocator{
						BuildType: k,
						Branch:    "",
						Count:     "1",
					}
					chL <- f
					fmt.Println("Filter: %s", f)
					mutex2.Lock()
					totalBuildFilters++
					mutex2.Unlock()
				} else {
					for z := range v {
						f := BuildLocator{
							BuildType: k,
							Branch:    v[z].Name,
							Count:     "1",
						}
						chL <- f
						fmt.Println("Filter: %s", f)
						mutex2.Lock()
						totalBuildFilters++
						mutex2.Unlock()
					}
				}
			}
		}
		close(chL)
		// log.Printf("Combined filters: %d", counter2)
	}()

	wg.Add(1)
	totalBranches := 0
	go func() {
		defer wg.Done()

		wg1 := new(sync.WaitGroup)
		for btID := range bc {
			wg1.Add(1)
			mutex := new(sync.Mutex)
			go func(bt BuildTypeID) {
				defer wg1.Done()
				// log.Printf("Working with build type '%s'", bt)
				// time.Sleep(1 * time.Second)
				br, err := c.GetAllBranches(bt)
				if err != nil {
					log.Errorf("Failed to get branches for %s: %v", bt, err)
					return
				}
				// log.Println(c.HTTPClient.LogString())
				if br.Count == 1 {
					chBtb <- map[BuildTypeID][]Branch{bt: nil}
					mutex.Lock()
					// totalBranches += br.Count
					mutex.Unlock()
				} else {
					chBtb <- map[BuildTypeID][]Branch{bt: br.Branch}
					mutex.Lock()
					totalBranches += br.Count
					mutex.Unlock()
				}
				log.Printf("BuildType: %s, branches: %v", bt, br)
				// counter++
			}(btID)
		}
		wg1.Wait()
		close(chBtb)
	}()

	totalBuildConfigs := 0
	mutex1 := new(sync.Mutex)
	if bl.BuildType == "" {
		bcAll, err := c.GetAllBuildConfigurations()
		if err != nil {
			log.Println(err)
		}
		// log.Println(bcAll)
		for _, bt := range bcAll.BuildTypes {
			bc <- bt.ID
			mutex1.Lock()
			totalBuildConfigs++
			mutex1.Unlock()
			log.Printf("'%s' added to 'bc' channel\n", bt.ID)
		}
	} else {
		bc <- bl.BuildType
		mutex1.Lock()
		totalBuildConfigs++
		mutex1.Unlock()
	}
	close(bc)

	// log.Println(bc)

	wg.Wait()
	// log.Printf("Added to channel: %d, got branches for: %d, duration: %v", counter1, counter, time.Since(start))
	iter := nBranches.IterBuffered()
	res := 0
	for v := range iter {
		s, _ := strconv.ParseInt(v.Val.(string), 10, 0)
		res += int(s)
	}
	log.Printf("nBuilds: %d, nBranches (keys): %d, nBranches (values): %d", nBuilds.Count(), nBranches.Count(), res)
	log.Printf("Total branches: %d, total build configs: %d, total build filters: %d", totalBranches, totalBuildConfigs, totalBuildFilters)
	// return builds, nil
}
