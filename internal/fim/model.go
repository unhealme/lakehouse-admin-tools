package fim

import "github.com/goccy/go-json"

type AuthToken struct {
	Token string
}

type Clusters []Cluster

type Cluster struct {
	Id            int
	Name          string
	HostsNum      int
	ServiceNum    int
	BadHostsNum   int
	BadServiceNum int
	Packs         []ClusterPack
}

type ClusterPack struct {
	Type    string
	Version string
}

type ServiceSummary struct {
	Properties []SummaryProperty
}

type SummaryProperty struct {
	Key   string
	Name  string
	Value json.RawMessage
	Type  string
}

func (p SummaryProperty) LinkValues() (rawLinks []string, err error) {
	var links SummaryLinks
	if err = json.Unmarshal(p.Value, &links); err != nil {
		return
	}
	for _, link := range links.Links {
		rawLinks = append(rawLinks, link.Url)
	}
	return
}

type SummaryLinks struct {
	Links []struct {
		Title string
		Url   string
	}
}
