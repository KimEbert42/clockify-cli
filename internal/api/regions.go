package api

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
	r := GetRegion(region)
	return "https://" + r.Host + r.APIPath
}
