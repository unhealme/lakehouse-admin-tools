package iam

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	iam_v3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	model_v3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	region_v3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"
	iam_v5 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v5"
	model_v5 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v5/model"
	region_v5 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v5/region"
)

type IamClient struct {
	IamV3 *iam_v3.IamClient
	IamV5 *iam_v5.IamClient
}

func (c IamClient) GetUser(domainId, userName string, refreshCache bool) (*model_v3.KeystoneListUsersResult, error) {
	if userCache == nil || refreshCache {
		req := &model_v3.KeystoneListUsersRequest{DomainId: &domainId}
		resp, err := c.IamV3.KeystoneListUsers(req)
		if err != nil {
			return nil, err
		}
		userCache = make(map[string]model_v3.KeystoneListUsersResult, len(*resp.Users))
		for _, user := range *resp.Users {
			userCache[user.Name] = user
		}
	}
	if user, ok := userCache[userName]; ok {
		return &user, nil
	}
	return nil, fmt.Errorf("user name '%s' not found", userName)
}

func (c IamClient) GetUserGroups(userId string) (groups []model_v3.KeystoneGroupResult, err error) {
	req := &model_v3.KeystoneListGroupsForUserRequest{UserId: userId}
	resp, err := c.IamV3.KeystoneListGroupsForUser(req)
	if err != nil {
		return nil, err
	}
	if resp.Groups != nil {
		groups = *resp.Groups
	}
	return
}

func (c IamClient) GetUsers(domainId string, refreshCache bool) (users []*model_v3.KeystoneListUsersResult, err error) {
	if userCache == nil || refreshCache {
		req := &model_v3.KeystoneListUsersRequest{}
		req.DomainId = &domainId
		resp, err := c.IamV3.KeystoneListUsers(req)
		if err != nil {
			return nil, err
		}
		userCache = make(map[string]model_v3.KeystoneListUsersResult, len(*resp.Users))
		for _, user := range *resp.Users {
			userCache[user.Name] = user
			users = append(users, &user)
		}
		return users, nil
	}
	for _, user := range userCache {
		users = append(users, &user)
	}
	return
}

func (c IamClient) GetUserLastLogin(userId string) (lastLogin *time.Time, err error) {
	req := &model_v5.ShowUserLastLoginV5Request{UserId: userId}
	resp, err := c.IamV5.ShowUserLastLoginV5(req)
	if err != nil {
		return nil, err
	}
	if resp.UserLastLogin.LastLoginAt != nil {
		lastLogin = new(time.Time(*resp.UserLastLogin.LastLoginAt))
	}
	return
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
	regionv3, err := region_v3.SafeValueOf(regionId)
	if err != nil {
		return nil, err
	}
	regionv5, err := region_v5.SafeValueOf(regionId)
	if err != nil {
		return nil, err
	}
	basev3, err := iam_v3.IamClientBuilder().WithCredential(creds).WithRegion(regionv3).SafeBuild()
	if err != nil {
		return nil, err
	}
	basev5, err := iam_v5.IamClientBuilder().WithCredential(creds).WithRegion(regionv5).SafeBuild()
	if err != nil {
		return nil, err
	}
	return &IamClient{IamV3: iam_v3.NewIamClient(basev3), IamV5: iam_v5.NewIamClient(basev5)}, nil
}
