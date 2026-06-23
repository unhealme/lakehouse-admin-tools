package iam

import (
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
)

type IamClient struct{ *iam.IamClient }

func (c IamClient) GetUser(domainId, userName string, refreshCache bool) (*model.KeystoneListUsersResult, error) {
	if userCache == nil || refreshCache {
		req := &model.KeystoneListUsersRequest{}
		req.DomainId = &domainId
		resp, err := c.KeystoneListUsers(req)
		if err != nil {
			return nil, err
		}
		userCache = make(map[string]model.KeystoneListUsersResult)
		for _, user := range *resp.Users {
			userCache[user.Name] = user
		}
	}
	if user, ok := userCache[userName]; ok {
		return &user, nil
	}
	return nil, fmt.Errorf("user name '%s' not found", userName)
}

func NewClient(ak, sk, token, regionId string) (*IamClient, error) {
	credsBuilder := auth.NewGlobalCredentialsBuilder().WithAk(ak).WithSk(sk)
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
	base, err := iam.IamClientBuilder().WithCredential(creds).WithRegion(region).SafeBuild()
	if err != nil {
		return nil, err
	}
	return &IamClient{iam.NewIamClient(base)}, nil
}
