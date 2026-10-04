package api

import "os"

type Region struct {
	Name     string
	Host     string
	APIPath  string
}

var Regions = map[string]Region{
	"global": {Name: "global", Host: "api.clockify.me", APIPath: "/api/v1"},
	"eu-de":  {Name: "eu-de", Host: "euc1.clockify.me", APIPath: "/api/v1"},
	"us":     {Name: "us", Host: "use2.clockify.me", APIPath: "/api/v1"},
	"uk":     {Name: "uk", Host: "euw2.clockify.me", APIPath: "/api/v1"},
	"au":     {Name: "au", Host: "apse2.clockify.me", APIPath: "/api/v1"},
}

func GetRegion(name string) Region {
	if r, ok := Regions[name]; ok {
		return r
	}
	return Regions["global"]
}

func BaseURL(region string) string {
	if url := os.Getenv("CLOCKIFY_API_URL"); url != "" {
		return url
	}
	r := GetRegion(region)
	return "https://" + r.Host + r.APIPath
}

// ReportsBaseURL returns the base URL for the Clockify Reports API.
//
// Global:  https://reports.api.clockify.me/v1
// Regional: https://{host}/report/v1  (same host as the main API, /report path prefix)
//
// CLOCKIFY_REPORTS_URL overrides the computed URL (used by tests).
func ReportsBaseURL(region string) string {
	if url := os.Getenv("CLOCKIFY_REPORTS_URL"); url != "" {
		return url
	}
	r := GetRegion(region)
	if r.Name == "global" {
		return "https://reports." + r.Host + "/v1"
	}
	return "https://" + r.Host + "/report/v1"
}
