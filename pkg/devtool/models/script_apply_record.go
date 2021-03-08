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
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/sqlchemy"

	api "yunion.io/x/onecloud/pkg/apis/devtool"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
)

type SScriptApplyRecord struct {
	db.SStatusStandaloneResourceBase
	ScriptId  string    `width:"36" charset:"ascii" nullable:"true" list:"user" index:"true"`
	ServerId  string    `width:"36" charset:"ascii" nullable:"true" list:"user"`
	StartTime time.Time `list:"user"`
	EndTime   time.Time `list:"user"`
	Reason    string    `list:"user"`
}

type SScriptApplyRecordManager struct {
	db.SStatusStandaloneResourceBaseManager
}

var ScriptApplyRecordManager *SScriptApplyRecordManager

func init() {
	ScriptApplyRecordManager = &SScriptApplyRecordManager{
		SStatusStandaloneResourceBaseManager: db.NewStatusStandaloneResourceBaseManager(
			SScriptApplyRecord{},
			"scriptapplyrecord_tbl",
			"scriptapplyrecord",
			"scriptapplyrecords",
		),
	}
	ScriptApplyRecordManager.SetVirtualObject(ScriptApplyRecordManager)
}

func (sarm *SScriptApplyRecordManager) ListItemFilter(ctx context.Context, q *sqlchemy.SQuery, userCred mcclient.TokenCredential, input api.ScriptApplyRecoredListInput) (*sqlchemy.SQuery, error) {
	q, err := sarm.SStatusStandaloneResourceBaseManager.ListItemFilter(ctx, q, userCred, input.StatusStandaloneResourceListInput)
	if err != nil {
		return q, err
	}
	if len(input.ScriptId) > 0 {
		q = q.Equals("script_id", input.ScriptId)
	}
	return q, nil
}

func (sarm *SScriptApplyRecordManager) CreateRecord(ctx context.Context, scriptId, serverId string) (*SScriptApplyRecord, error) {
	return sarm.createRecordWithResult(ctx, scriptId, serverId, nil, "")
}

func (sarm *SScriptApplyRecordManager) createRecordWithResult(ctx context.Context, scriptId, serverId string, success *bool, reason string) (*SScriptApplyRecord, error) {
	now := time.Now()
	sar := &SScriptApplyRecord{
		StartTime: now,
		ScriptId:  scriptId,
		ServerId:  serverId,
	}
	if success == nil {
		sar.Status = api.SCRIPT_APPLY_RECORD_APPLYING
	} else if *success {
		sar.Status = api.SCRIPT_APPLY_RECORD_SUCCEED
	} else if !*success {
		sar.Status = api.SCRIPT_APPLY_RECORD_FAILED
	}
	sar.Reason = reason
	err := sarm.TableSpec().Insert(ctx, sar)
	if err != nil {
		return nil, err
	}
	sar.SetModelManager(sarm, sar)
	return sar, nil
}

func (sarm *SScriptApplyRecordManager) NamespaceScope() rbacutils.TRbacScope {
	return rbacutils.ScopeProject
}

func (sarm *SScriptApplyRecordManager) ResourceScope() rbacutils.TRbacScope {
	return rbacutils.ScopeProject
}

func (sarm *SScriptApplyRecordManager) FileterByOwner(q *sqlchemy.SQuery, owner mcclient.IIdentityProvider, scope rbacutils.TRbacScope) *sqlchemy.SQuery {
	if owner != nil {
		switch scope {
		case rbacutils.ScopeProject, rbacutils.ScopeDomain:
			scriptQ := ScriptManager.Query("id", "domain_id").SubQuery()
			q = q.Join(scriptQ, sqlchemy.Equals(q.Field("script_id"), scriptQ.Field("id")))
			q = q.Filter(sqlchemy.Equals(scriptQ.Field("domain_id"), owner.GetProjectDomainId()))
		}
	}
	return q
}

func (sarm *SScriptApplyRecordManager) FetchOwnerId(ctx context.Context, data jsonutils.JSONObject) (mcclient.IIdentityProvider, error) {
	return db.FetchDomainInfo(ctx, data)
}

func (sar *SScriptApplyRecord) GetOwnerId() mcclient.IIdentityProvider {
	obj, _ := ScriptManager.FetchById(sar.ScriptId)
	if obj == nil {
		return nil
	}
	return obj.GetOwnerId()
}

func (sar *SScriptApplyRecord) SetResult(status, reason string) error {
	_, err := db.Update(sar, func() error {
		sar.Status = status
		sar.Reason = reason
		sar.EndTime = time.Now()
		return nil
	})
	return err
}

func (sar *SScriptApplyRecord) Fail(reason string) error {
	return sar.SetResult(api.SCRIPT_APPLY_RECORD_FAILED, reason)
}

func (sar *SScriptApplyRecord) Succeed(reason string) error {
	return sar.SetResult(api.SCRIPT_APPLY_RECORD_SUCCEED, reason)
}
