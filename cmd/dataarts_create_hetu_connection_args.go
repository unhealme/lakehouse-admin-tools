package cmd

import (
	"github.com/unhealme/lakehouse-admin-tools/internal/dataarts"
	"github.com/unhealme/lakehouse-admin-tools/internal/iam"
)

type DataArtsCreateHetuConnectionArgs struct {
	UserPairs []string `arg:"positional,required" placeholder:"USER:WORKSPACE"`
	AgentId   string   `arg:"-,--agent-id" help:"connection agent id" placeholder:"AGENT_ID"`
	AgentName string   `arg:"-,--agent-name" help:"connection agent name" placeholder:"AGENT_NAME"`

	DomainId       string                   `arg:"-"`
	InstanceId     string                   `arg:"-"`
	DataArtsClient *dataarts.DataArtsClient `arg:"-"`
	IamClient      *iam.IamClient           `arg:"-"`
	HetuConfig     *dataarts.DwConfig       `arg:"-"`
}
