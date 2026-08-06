package config

import (
	"fmt"
	"maps"
	"runtime"
	"slices"
	"strings"

	"github.com/unhealme/lakehouse-admin-tools/args"
	"github.com/unhealme/lakehouse-admin-tools/internal/dataarts"
)

type DataArtsArguments struct {
	CreateHetuConnection *args.DataArtsCreateHetuConnectionArgs `arg:"subcommand:create-hetu-connection" yaml:"-"`
	UpdateHetuConnection *args.DataArtsUpdateHetuConnectionArgs `arg:"subcommand:update-hetu-connection" yaml:"-"`

	Agent *struct {
		Id   string
		Name string
	} `arg:"-"`
	HetuConfig dataarts.DwConfig `arg:"-" yaml:"hetu_config"`
	InstanceId string            `arg:"-i,--" placeholder:"INSTANCE_ID" yaml:"instance_id"`
}

type FimArguments struct {
	Address  string `arg:"-u,--url" placeholder:"FIM_ADDRESS" yaml:"address"`
	User     string `arg:"-,--user,env:FIM_USER" placeholder:"FIM_USER" yaml:"user"`
	Password string `arg:"-,--password,env:FIM_PASSWORD" placeholder:"FIM_PASSWORD" yaml:"password"`
}

type IamArguments struct {
	ListGroups *args.IamListGroupsArgs `arg:"subcommand:list-groups" yaml:"-"`
	ListUsers  *args.IamListUsersArgs  `arg:"subcommand:list-users" yaml:"-"`
}

type MrsArguments struct {
	ListHetuTenants  *args.MrsListHetuTenantsArgs  `arg:"subcommand:list-hetu-tenants" yaml:"-"`
	DumpHetuClusters *args.MrsDumpHetuClustersArgs `arg:"subcommand:dump-hetu-clusters" yaml:"-"`

	ClusterId    string `arg:"-,--cluster-id" placeholder:"CLUSTER_ID" yaml:"cluster_id"`
	ProxyAddress string `arg:"-,--mrs-proxy,env:MRS_PROXY" placeholder:"MRS_PROXY" yaml:"proxy"`
}

type ObsArguments struct {
	Analyze              *args.ObsAnalyzeArgs              `arg:"subcommand:analyze" yaml:"-"`
	BatchRename          *args.ObsBatchRenameArgs          `arg:"subcommand:batch-rename" yaml:"-"`
	BatchSetStorageClass *args.ObsBatchSetStorageClassArgs `arg:"subcommand:batch-set-storage-class" yaml:"-"`

	Endpoint string `arg:"-e,--endpoint" placeholder:"ENDPOINT"`
}

type PsArguments struct {
	AutoKill *args.PsAutoKillArgs `arg:"subcommand:auto-kill"`
}

type UamArguments struct {
	DescribeUser *args.UamDescribeUserArgs `arg:"subcommand:describe-user" yaml:"-"`
	ListMembers  *args.UamListMembersArgs  `arg:"subcommand:list-members" yaml:"-"`

	Url        string `arg:"-u,--,env:LDAP_URL" placeholder:"LDAP_URL"`
	User       string `arg:"-,--user,env:LDAP_BIND_USER" placeholder:"LDAP_BIND_USER"`
	Password   string `arg:"-,--password,env:LDAP_BIND_PASSWORD" placeholder:"LDAP_BIND_PASSWORD"`
	BaseDN     string `arg:"-b,--base-dn" placeholder:"LDAP_BASE_DN" yaml:"base_dn"`
	GroupBase  string `arg:"-g,--group-base" placeholder:"LDAP_GROUP_BASE" yaml:"group_base"`
	MailDomain string `arg:"-d,--mail-domain" placeholder:"LDAP_MAIL_DOMAIN" yaml:"mail_domain"`
	Realm      string `arg:"-r,--realm" placeholder:"REALM"`
}

type YarnArguments struct {
	AutoKillApps *args.YarnAutoKillAppsArgs `arg:"subcommand:auto-kill" yaml:"-"`

	RMAddress CommaSeparatedString `arg:"-u,--rm-url" placeholder:"RM_ADDRESS" yaml:"rm_address"`
}

type Arguments struct {
	DataArts *DataArtsArguments `arg:"subcommand:dataarts"`
	Fim      *FimArguments      `arg:"subcommand:fim"`
	Iam      *IamArguments      `arg:"subcommand:iam"`
	Mrs      *MrsArguments      `arg:"subcommand:mrs"`
	Obs      *ObsArguments      `arg:"subcommand:obs"`
	Ps       *PsArguments       `arg:"subcommand:ps" yaml:"-"`
	Uam      *UamArguments      `arg:"subcommand:uam"`
	Yarn     *YarnArguments     `arg:"subcommand:yarn"`

	ConfigFile   string `arg:"-c,--config,env:LHAT_CONFIG" help:"load config from FILE" placeholder:"FILE" yaml:"-"`
	AccessKey    string `arg:"-,--ak,env:HW_ACCESS_KEY" placeholder:"ACCESS_KEY" yaml:"access_key"`
	SecretKey    string `arg:"-,--sk,env:HW_SECRET_KEY" placeholder:"SECRET_KEY" yaml:"secret_key"`
	SessionToken string `arg:"-,--token,env:HW_SECURITY_TOKEN" placeholder:"SECURITY_TOKEN" yaml:"session_token"`
	DomainId     string `arg:"-d,--domain-id" placeholder:"DOMAIN_ID" yaml:"domain_id"`
	Region       string `arg:"-r,--region" placeholder:"REGION"`
	NoColor      bool   `arg:"-,--no-color" help:"disable colorized output" yaml:"no_color"`
	Verbose      bool   `arg:"-v,--verbose" help:"enable debug logging"`
}

func (Arguments) Epilogue() string {
	b := &strings.Builder{}
	fmt.Fprintln(b, "Components:")
	for _, comp := range slices.Sorted(maps.Keys(compVer)) {
		fmt.Fprintf(b, "  %-32s %s\n", comp, compVer[comp])
	}
	return strings.TrimRight(b.String(), "\n")
}

func (Arguments) Version() string {
	return fmt.Sprintf("lakehouse-admin-tools %s (%s-%s)\n", Version, runtime.GOOS, runtime.GOARCH)
}
