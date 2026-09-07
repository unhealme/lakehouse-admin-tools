package config

import (
	"fmt"
	"maps"
	"runtime"
	"slices"
	"strings"

	"github.com/unhealme/lakehouse-admin-tools/internal/clients/dataarts"
	"github.com/unhealme/lakehouse-admin-tools/pkg/arguments"
)

type DataArtsArguments struct {
	CreateHetuConnection *arguments.DataArtsCreateHetuConnectionArgs `arg:"subcommand:create-hetu-connection" yaml:"-"`
	UpdateHetuConnection *arguments.DataArtsUpdateHetuConnectionArgs `arg:"subcommand:update-hetu-connection" yaml:"-"`

	Agent *struct {
		Id   string
		Name string
	} `arg:"-"`
	HetuConfig dataarts.DwConfig `arg:"-" yaml:"hetu_config"`
	InstanceId string            `arg:"-i,--" placeholder:"INSTANCE_ID" yaml:"instance_id"`
}

type FimArguments struct {
	ResetUserPassword *arguments.FimResetUserPasswordArgs `arg:"subcommand:reset-user-password" yaml:"-"`

	Addresses       map[string]string `arg:"-" yaml:"addresses"`
	Address         string            `arg:"-u,--url,required" help:"can use mapping name instead instead of url address" placeholder:"FIM_ADDRESS" yaml:"-"`
	User            string            `arg:"-,--user,env:FIM_USER" placeholder:"FIM_USER" yaml:"user"`
	Password        string            `arg:"-,--password,env:FIM_PASSWORD" placeholder:"FIM_PASSWORD" yaml:"password"`
	DefaultPassword string            `arg:"-" yaml:"default_password"`
}

type HiveArguments struct {
	BackupTable *arguments.HiveBackupTableArgs `arg:"subcommand:backup-table" yaml:"-"`

	Url          string `arg:"-u,--url" placeholder:"HIVE_URL" yaml:"url"`
	HostQualName string `arg:"-,--host-qn" placeholder:"NAME" yaml:"host_qual_name"`
}

type IamArguments struct {
	ListGroups *arguments.IamListGroupsArgs `arg:"subcommand:list-groups" yaml:"-"`
	ListUsers  *arguments.IamListUsersArgs  `arg:"subcommand:list-users" yaml:"-"`
}

type MrsArguments struct {
	ListHetuTenants  *arguments.MrsListHetuTenantsArgs  `arg:"subcommand:list-hetu-tenants" yaml:"-"`
	DumpHetuClusters *arguments.MrsDumpHetuClustersArgs `arg:"subcommand:dump-hetu-clusters" yaml:"-"`

	ClusterId    string `arg:"-,--cluster-id" placeholder:"CLUSTER_ID" yaml:"cluster_id"`
	ProxyAddress string `arg:"-,--mrs-proxy,env:MRS_PROXY" placeholder:"MRS_PROXY" yaml:"proxy"`
}

type ObsArguments struct {
	Analyze              *arguments.ObsAnalyzeArgs              `arg:"subcommand:analyze" yaml:"-"`
	BatchRename          *arguments.ObsBatchRenameArgs          `arg:"subcommand:batch-rename" yaml:"-"`
	BatchSetStorageClass *arguments.ObsBatchSetStorageClassArgs `arg:"subcommand:batch-set-storage-class" yaml:"-"`

	Endpoint string `arg:"-e,--endpoint" placeholder:"ENDPOINT"`
}

type PsArguments struct {
	AutoKill *arguments.PsAutoKillArgs `arg:"subcommand:auto-kill"`
}

type UamArguments struct {
	DescribeUser *arguments.UamDescribeUserArgs `arg:"subcommand:describe-user" yaml:"-"`
	ListMembers  *arguments.UamListMembersArgs  `arg:"subcommand:list-members" yaml:"-"`

	BaseDN     string `arg:"-b,--base-dn" placeholder:"LDAP_BASE_DN" yaml:"base_dn"`
	Url        string `arg:"-u,--,env:LDAP_URL" placeholder:"LDAP_URL"`
	User       string `arg:"-,--user,env:LDAP_BIND_USER" placeholder:"LDAP_BIND_USER"`
	Password   string `arg:"-,--password,env:LDAP_BIND_PASSWORD" placeholder:"LDAP_BIND_PASSWORD"`
	GroupBase  string `arg:"-,--group-base" placeholder:"LDAP_GROUP_BASE" yaml:"group_base"`
	MailDomain string `arg:"-,--mail-domain" placeholder:"LDAP_MAIL_DOMAIN" yaml:"mail_domain"`
	Realm      string `arg:"-,--realm" placeholder:"REALM"`
}

type YarnArguments struct {
	AutoKillApps *arguments.YarnAutoKillAppsArgs `arg:"subcommand:auto-kill" yaml:"-"`

	RMAddress CommaSeparatedString `arg:"-u,--rm-url" placeholder:"RM_ADDRESS" yaml:"rm_address"`
}

type Arguments struct {
	DataArts *DataArtsArguments `arg:"subcommand:dataarts"`
	Fim      *FimArguments      `arg:"subcommand:fim"`
	Hive     *HiveArguments     `arg:"subcommand:hive"`
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
	DomainId     string `arg:"-,--domain-id" placeholder:"DOMAIN_ID" yaml:"domain_id"`
	Region       string `arg:"-,--region" placeholder:"REGION"`
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
	return fmt.Sprintf("lakehouse-admin-tools %s (%s-%s)", Version, runtime.GOOS, runtime.GOARCH)
}
