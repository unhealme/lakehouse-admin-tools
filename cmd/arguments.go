package cmd

import (
	"fmt"
	"maps"
	"os"
	"runtime"
	"slices"
	"strings"
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

type PsArguments struct {
	AutoKill *PsAutoKillArgs `arg:"subcommand:auto-kill"`
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
	Ps       *PsArguments       `arg:"subcommand:ps"`
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
