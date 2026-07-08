package hetu

import (
	"net/http"
	"net/url"
)

type HetuAuth struct {
	SessionId *http.Cookie
	Url       *url.URL
}

type HetuToken struct {
	Token string
}

type TenantInfo struct {
	Content TenantContent
}

type TenantContent struct {
	Tenants []Tenant
	Total   int
}

type Tenant struct {
	Tenant       string
	TotalMemory  int // Megabyte
	TotalVcores  int
	ClusterIds   []string
	RunningCount int
	StoppedCount int
	ErrorCount   int
}
