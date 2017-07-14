# TeamCity API bindings

Usage example:

```
func main() {
	// put your login and password here
	client := tc.New("https://gwre-devexp-ci-production-devci.gwre-devops.net", "login", "password")

	bl := tc.BuildLocator{
		BuildType: "BillingCenter_DailyTests",
		Branch:    "g-master",
		Status:    "failure",
	}
	buildStats, err := client.GetBuildStatistics(bl)
    if err != nil {
      log.Fatal(err)
    }
    fmt.Println(buildStats)

    // --------- //

	btl := tc.BuildTypeLocator{
	  Id: "BillingCenter_DailyTests",
	}
	buildParams, err := client.GetBuildsByParams(btl, bl)
    if err != nil {
      log.Fatal(err)
    }
    fmt.Println(buildParams)
}

```
