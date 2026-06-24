package cmd

import (
	"fmt"
	"maps"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/unhealme/lakehouse-admin-tools/internal"
	das "github.com/unhealme/lakehouse-admin-tools/internal/dataarts-studio"
	"github.com/unhealme/lakehouse-admin-tools/internal/iam"
	"github.com/unhealme/lakehouse-admin-tools/internal/obs"
	"github.com/unhealme/lakehouse-admin-tools/internal/uam"
	"github.com/unhealme/lakehouse-admin-tools/internal/yarn"
)

type DataArtsArguments struct {
	CreateHetuConnection *DataArtsCreateHetuConnectionArgs `arg:"subcommand:create-hetu-connection"`

	InstanceId string `arg:"-i,--" placeholder:"INSTANCE_ID"`
}

type ObsArguments struct {
	Analyze              *ObsAnalyzeArgs              `arg:"subcommand:analyze"`
	BatchRename          *ObsBatchRenameArgs          `arg:"subcommand:batch-rename"`
	BatchSetStorageClass *ObsBatchSetStorageClassArgs `arg:"subcommand:batch-set-storage-class"`

	Endpoint string `arg:"-e,--endpoint" placeholder:"ENDPOINT"`
}

type UamArguments struct {
	DescribeUser *UamDescribeUserArgs `arg:"subcommand:describe-user"`
	ListMembers  *UamListMembersArgs  `arg:"subcommand:list-members"`

	Url       string `arg:"-u,--,env:LDAP_URL" placeholder:"LDAP_URL"`
	User      string `arg:"-,--user,env:LDAP_BIND_USER" placeholder:"LDAP_BIND_USER"`
	Password  string `arg:"-,--password,env:LDAP_BIND_PASSWORD" placeholder:"LDAP_BIND_PASSWORD"`
	BaseDN    string `arg:"-b,--base-dn" placeholder:"LDAP_BASE_DN"`
	GroupBase string `arg:"-g,--group-base" placeholder:"LDAP_GROUP_BASE"`
	Domain    string `arg:"-d,--domain" placeholder:"LDAP_DOMAIN"`
	Realm     string `arg:"-r,--realm" placeholder:"REALM"`
}

type YarnArguments struct {
	AutoKillApps *YarnAutoKillAppsArgs `arg:"subcommand:auto-kill"`

	RMAddress string `arg:"-u,--rm-url" placeholder:"RM_ADDRESS"`
}

type Arguments struct {
	DataArts *DataArtsArguments `arg:"subcommand:dataarts"`
	Obs      *ObsArguments      `arg:"subcommand:obs"`
	Uam      *UamArguments      `arg:"subcommand:uam"`
	Yarn     *YarnArguments     `arg:"subcommand:yarn"`

	ConfigFile   string `arg:"-c,--config,env:LHAT_CONFIG" help:"load config from FILE" placeholder:"FILE"`
	AccessKey    string `arg:"-,--ak,env:HW_ACCESS_KEY" placeholder:"ACCESS_KEY"`
	SecretKey    string `arg:"-,--sk,env:HW_SECRET_KEY" placeholder:"SECRET_KEY"`
	SessionToken string `arg:"-,--token,env:HW_SECURITY_TOKEN" placeholder:"SECURITY_TOKEN"`
	DomainId     string `arg:"-d,--domain-id" placeholder:"DOMAIN_ID"`
	Region       string `arg:"-r,--region" placeholder:"REGION"`
	Verbose      bool   `arg:"-v,--verbose" help:"enable debug logging"`
}

func (Arguments) Epilogue() string {
	var b strings.Builder
	fmt.Fprintln(&b, "Components:")
	for _, comp := range slices.Sorted(maps.Keys(compVer)) {
		fmt.Fprintf(&b, "  %s %s\n", comp, compVer[comp])
	}
	return strings.TrimRight(b.String(), "\n")
}

func (Arguments) Version() string {
	return fmt.Sprintf("%s %s (%s-%s)", os.Args[0], Version, runtime.GOOS, runtime.GOARCH)
}

// DataArts Create Hetu Connection
type DataArtsCreateHetuConnectionArgs struct {
	UserPairs []string `arg:"positional,required" placeholder:"USER:WORKSPACE"`
	AgentId   string   `arg:"-,--agent-id" help:"connection agent id" placeholder:"AGENT_ID"`
	AgentName string   `arg:"-,--agent-name" help:"connection agent name" placeholder:"AGENT_NAME"`

	DomainId       string              `arg:"-"`
	InstanceId     string              `arg:"-"`
	DataArtsClient *das.DataArtsClient `arg:"-"`
	IamClient      *iam.IamClient      `arg:"-"`
	HetuConfig     *das.DwConfig       `arg:"-"`
}

// Hive Backup Table
type HiveBackupTableArgs struct {
	Tables     []string
	Database   string
	InputFile  string
	OutputFile string
}

// OBS Analyze
type ObsAnalyzeArgs struct {
	Paths   []string `arg:"positional,required" placeholder:"PATH"`
	DirOnly bool     `arg:"-d,--dir-only" help:"only analyze directories"`

	ObsClient *obs.ObsClient `arg:"-"`
}

// OBS Batch Rename
type ObsBatchRenameArgs struct {
	Path        string `arg:"positional,required" placeholder:"PATH"`
	Prefix      string `arg:"positional,required" placeholder:"PREFIX"`
	Concurrency int    `arg:"-j,--" default:"2" help:"max job concurrency" placeholder:"NUM"`
	DirOnly     bool   `arg:"-d,--dir-only" help:"only rename directories"`
	DryRun      bool   `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg      bool   `arg:"-,--no-progress" help:"disable progress bar"`

	ObsClient *obs.ObsClient `arg:"-"`
}

// OBS Batch Set Storage Class
type ObsBatchSetStorageClassArgs struct {
	InputFiles  []string `arg:"positional,required" placeholder:"FILE"`
	Concurrency int      `arg:"-j,--" default:"2" help:"max job concurrency" placeholder:"NUM"`
	DryRun      bool     `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg      bool     `arg:"-,--no-progress" help:"disable progress bar"`

	ObsClient *obs.ObsClient `arg:"-"`
}

// UAM Describe User
type UamDescribeUserArgs struct {
	Users  []string `arg:"positional,required" placeholder:"USER"`
	Format string   `arg:"-f,--format" default:"default" help:"output format" placeholder:"{default,csv}"`

	UamClient *uam.UamClient `arg:"-"`
}

// UAM List Members
type UamListMembersArgs struct {
	Groups []string `arg:"positional,required" placeholder:"GROUP"`

	UamClient *uam.UamClient `arg:"-"`
}

// YARN Auto Kill Apps
type YarnAutoKillAppsArgs struct {
	LongerThan internal.Duration `arg:"-,--longer-than,required" placeholder:"DUR" help:"kill yarn applications that running longer than DUR"`
	DryRun     bool              `arg:"-,--dry-run" help:"simulate action without doing anything"`
	NoProg     bool              `arg:"-,--no-progress" help:"disable progress bar"`

	YarnClient *yarn.YarnRMClient `arg:"-"`
}
