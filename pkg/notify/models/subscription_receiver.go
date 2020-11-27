// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/tristate"

	"yunion.io/x/onecloud/pkg/apis/notify"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/mcclient/auth"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
)

var SubscriptionReceiverManager *SSubscriptionReceiverManager

func init() {
	db.InitManager(func() {
		SubscriptionReceiverManager = &SSubscriptionReceiverManager{
			SResourceBaseManager: db.NewResourceBaseManager(
				SSubscriptionReceiver{},
				"subscriptionreceiver_tbl",
				"subscriptionreceiver",
				"subscriptionreceivers",
			),
		}
		SubscriptionReceiverManager.SetVirtualObject(ReceiverNotificationManager)
	})
}

type SSubscriptionReceiverManager struct {
	db.SResourceBaseManager
}

type SSubscriptionReceiver struct {
	db.SResourceBase

	SubscriptionID string `width:"128" charset:"ascii" nullable:"fase"`
	Receiver       string `width:"128" charset:"ascii" nullable:"false"`
	ReceiverType   string `width:"16" charset:"ascii" nullable:"false" index:"true"`
	Invalid        tristate.TriState

	ExtendedInfo SSRExtendedInfo
}

type SSRExtendedInfo struct {
	RoleScope string `width:"8" charset:"ascii" nullable:"false"`
}

const (
	ReceiverRole          = "role"
	ReceiverNormal        = "normal"
	ReceiverFeishuRobot   = notify.FEISHU_ROBOT
	ReceiverDingtalkRobot = notify.DINGTALK_ROBOT
	ReceiverWorkwxRobot   = notify.WORKWX_ROBOT
	ReceiverWeebhook      = "webhook"

	ScopeSystem  = "system"
	ScopeDomain  = "domain"
	ScopeProject = "project"
)

func (srm *SSubscriptionReceiverManager) robot(ssid string) (string, error) {
	return srm.findSingleReceiver(ReceiverFeishuRobot, ReceiverDingtalkRobot, ReceiverWorkwxRobot)
}

func (srm *SSubscriptionReceiverManager) webhook(ssid string) (string, error) {
	return srm.findSingleReceiver(ssid, ReceiverWeebhook)
}

func (srm *SSubscriptionReceiverManager) findSingleReceiver(ssid string, receiverTypes ...string) (string, error) {
	srs, err := srm.findReceivers(ssid, receiverTypes...)
	if err != nil {
		return "", err
	}
	if len(srs) > 1 {
		return "", errors.Error("multi receiver")
	}
	if len(srs) == 0 {
		return "", errors.ErrNotFound
	}
	return srs[0].ReceiverType, nil
}

func (srm *SSubscriptionReceiverManager) findReceivers(ssid string, receiverTypes ...string) ([]SSubscriptionReceiver, error) {
	if len(receiverTypes) == 0 {
		return nil, nil
	}
	q := srm.Query().Equals("subscription_id", ssid)
	if len(receiverTypes) == 1 {
		q = q.Equals("receiver_type", receiverTypes[0])
	} else {
		q = q.In("receiver_type", receiverTypes)
	}
	srs := make([]SSubscriptionReceiver, 0)
	err := db.FetchModelObjects(srm, q, &srs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to FetchModelObjects")
	}
	return srs, nil
}

// TODO: Use cache to increase speed
func (srm *SSubscriptionReceiverManager) getReceivers(ctx context.Context, ssid string, projectDomainId string, projectId string) ([]string, error) {
	srs, err := srm.findReceivers(ssid, ReceiverRole, ReceiverNormal)
	if err != nil {
		return nil, err
	}
	receivers := make([]string, 0, len(srs))
	projectOrDomainRole := make([]string, 0)
	var scope string
	for _, sr := range srs {
		if sr.ReceiverType == ReceiverNormal && sr.Invalid.IsTrue() {
			receivers = append(receivers, sr.Receiver)
		} else if sr.Invalid.IsTrue() {
			projectOrDomainRole = append(projectOrDomainRole, sr.Receiver)
			if len(scope) == 0 {
				scope = sr.ExtendedInfo.RoleScope
			}
			if scope != sr.ExtendedInfo.RoleScope {
				return nil, fmt.Errorf("There cannot be role receivers with different scopes under the same subscription")
			}
		}
	}
	query := jsonutils.NewDict()
	query.Set("roles", jsonutils.NewStringArray(projectOrDomainRole))
	switch scope {
	case ScopeSystem:
		query.Set("include_system", jsonutils.JSONNull)
	case ScopeDomain:
		query.Set("project_domains", jsonutils.NewStringArray([]string{projectDomainId}))
	case ScopeProject:
		query.Set("projects", jsonutils.NewStringArray([]string{projectId}))
	}
	s := auth.GetAdminSession(ctx, "", "")
	listRet, err := modules.RoleAssignments.List(s, query)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list RoleAssignments")
	}
	groups := make([]string, 0)
	for i := range listRet.Data {
		ras := listRet.Data[i]
		user, err := ras.Get("user")
		if err == nil {
			id, err := user.GetString("id")
			if err != nil {
				return nil, errors.Wrap(err, "unable to get user.id from result of RoleAssignments.List")
			}
			receivers = append(receivers, id)
		}
		group, err := ras.Get("group")
		if err == nil {
			id, err := group.GetString("id")
			if err != nil {
				return nil, errors.Wrap(err, "unale to get group.id from result of RoleAssignments.List")
			}
			groups = append(groups, id)
		}
	}
	if len(groups) == 0 {
		return receivers, nil
	}
	rungo, _ := errgroup.WithContext(ctx)
	param := jsonutils.NewDict()
	param.Set("scope", jsonutils.NewString("system"))
	userList2 := make([][]string, len(groups))
	for i := range groups {
		id := groups[i]
		index := i
		rungo.Go(func() error {
			list, err := modules.Groups.GetUsers(s, id, param)
			if err != nil {
				return err
			}
			for j := range list.Data {
				id, err := list.Data[j].GetString("user_id")
				if err != nil {
					return errors.Wrap(err, "can't find user_id from the result of Groups.GetUsers")
				}
				userList2[index] = append(userList2[index], id)
			}
			return nil
		})
	}
	err = rungo.Wait()
	if err != nil {
		return nil, errors.Wrap(err, "unable get users of groups")
	}

	for i := range userList2 {
		receivers = append(receivers, userList2[i]...)
	}
	return receivers, nil
}

func (srm *SSubscriptionReceiverManager) create(ctx context.Context, ssid, receiver, receiverType string, extendedInfo SSRExtendedInfo) (*SSubscriptionReceiver, error) {
	sr := &SSubscriptionReceiver{
		SubscriptionID: ssid,
		Receiver:       receiver,
		ReceiverType:   receiverType,
		ExtendedInfo:   extendedInfo,
	}
	err := srm.TableSpec().Insert(ctx, sr)
	if err != nil {
		return nil, errors.Wrap(err, "unable to insert")
	}
	return sr, nil
}
