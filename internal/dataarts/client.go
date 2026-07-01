package dataarts

import (
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	das "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dataartsstudio/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dataartsstudio/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dataartsstudio/v1/region"
	iam_model "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
)

type DataArtsClient struct{ *das.DataArtsStudioClient }

func (c DataArtsClient) CreateHetuConnection(
	workspaceId, connectionName, agentId, agentName string,
	hetuConfig *DwConfig,
) error {
	dwConf := any(hetuConfig)
	voDev := model.ApigDataSourceVo{
		DwName:         connectionName,
		DwType:         "HETUENGINE",
		DwConfig:       &dwConf,
		AgentId:        &agentId,
		AgentName:      &agentName,
		EnvType:        &EnvTypeDev,
		SupportService: &SupportServiceDefault,
	}
	voProd := voDev
	voProd.EnvType = &EnvTypeProd

	req := &model.CreateConnectionsRequest{}
	req.Workspace = workspaceId
	req.Body = &model.ApigDataSourcesVo{}
	req.Body.Mode = &CreateConnectionsRequestBodyModeDefault
	req.Body.Visible = &CreateConnectionsRequestBodyVisibleDefault
	req.Body.DataSourceVos = &[]model.ApigDataSourceVo{voDev, voProd}
	_, err := c.CreateConnections(req)
	return err
}

func (c DataArtsClient) CreateResourcePermission(workspaceId string, resource *model.ApigDataSourceView, member *iam_model.KeystoneListUsersResult) error {
	req := &model.CreateSecurityResourcePermissionPolicyRequest{}
	req.Workspace = workspaceId
	req.Body = &model.PermissionResourcePolicyCreateDto{}
	req.Body.PolicyName = member.Name
	memberType := model.GetMemberPolicyItemMemberTypeEnum()
	req.Body.Members = append(req.Body.Members, model.MemberPolicyItem{
		MemberId:   member.Id,
		MemberName: member.Name,
		MemberType: &memberType.USER,
	})
	req.Body.Resources = append(req.Body.Resources, model.ResourcePolicyItem{
		ResourceId:   *resource.DwId,
		ResourceName: *resource.DwName,
		ResourceType: model.GetResourcePolicyItemResourceTypeEnum().DATA_CONNECTION,
	})
	_, err := c.CreateSecurityResourcePermissionPolicy(req)
	return err
}

func (c DataArtsClient) GetConnectionFromName(workspaceId, connectionName string, refreshCache bool) (*model.ApigDataSourceView, error) {
	index := fmt.Sprintf("%s/%s", workspaceId, connectionName)
	if connectionCache == nil || refreshCache {
		req := &model.ListDataconnectionsRequest{}
		req.Workspace = workspaceId
		req.Name = &connectionName
		resp, err := c.ListDataconnections(req)
		if err != nil {
			return nil, err
		}
		connectionCache = make(map[string]model.ApigDataSourceView)
		for _, conn := range *resp.DataConnectionLists {
			if conn.DwName != nil {
				connectionCache[index] = conn
			}
		}
	}
	if conn, ok := connectionCache[index]; ok {
		return &conn, nil
	}
	return nil, fmt.Errorf("connection name '%s' not found", connectionName)
}

func (c DataArtsClient) GetWorkspaceFromName(instanceId, workspaceName string, refreshCache bool) (*model.Workspacebody, error) {
	if workspaceCache == nil || refreshCache {
		req := &model.ListManagerWorkSpacesRequest{}
		req.InstanceId = instanceId
		resp, err := c.ListManagerWorkSpaces(req)
		if err != nil {
			return nil, err
		}
		workspaceCache = make(map[string]model.Workspacebody)
		for _, workspace := range *resp.Data {
			if workspace.Name != nil {
				workspaceCache[*workspace.Name] = workspace
			}
		}
	}
	if workspace, ok := workspaceCache[workspaceName]; ok {
		return &workspace, nil
	}
	return nil, fmt.Errorf("workspace name '%s' not found", workspaceName)
}

func NewClient(ak, sk, token, regionId string) (*DataArtsClient, error) {
	credsBuilder := auth.NewBasicCredentialsBuilder().WithAk(ak).WithSk(sk)
	if token != "" {
		credsBuilder = credsBuilder.WithSecurityToken(token)
	}
	creds, err := credsBuilder.SafeBuild()
	if err != nil {
		return nil, err
	}
	region, err := region.SafeValueOf(regionId)
	if err != nil {
		return nil, err
	}
	base, err := das.DataArtsStudioClientBuilder().WithCredential(creds).WithRegion(region).SafeBuild()
	if err != nil {
		return nil, err
	}
	return &DataArtsClient{das.NewDataArtsStudioClient(base)}, nil
}
