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

package tasks

import (
	"context"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/util/logclient"
)

type RouteTableSyncStatusTask struct {
	taskman.STask
}

func init() {
	taskman.RegisterTask(RouteTableSyncStatusTask{})
}

func (self *RouteTableSyncStatusTask) taskFailed(ctx context.Context, routeTable *models.SRouteTable, err error) {
	routeTable.SetStatus(self.GetUserCred(), api.ROUTE_TABLE_UNKNOWN, err.Error())
	db.OpsLog.LogEvent(routeTable, db.ACT_CREATE, routeTable.GetShortDesc(ctx), self.GetUserCred())
	logclient.AddActionLogWithContext(ctx, routeTable, logclient.ACT_UPDATE, err, self.UserCred, false)
	self.SetStageFailed(ctx, jsonutils.NewString(err.Error()))
}

func (self *RouteTableSyncStatusTask) taskComplete(ctx context.Context, routeTable *models.SRouteTable) {
	routeTable.SetStatus(self.GetUserCred(), api.ROUTE_TABLE_AVAILABLE, "")
	self.SetStageComplete(ctx, nil)
}

func (self *RouteTableSyncStatusTask) OnInit(ctx context.Context, obj db.IStandaloneModel, data jsonutils.JSONObject) {
	routeTable := obj.(*models.SRouteTable)
	vpc := routeTable.GetVpc()
	iRouteTable, err := routeTable.GetICloudRouteTable()
	if err != nil {
		self.taskFailed(ctx, routeTable, errors.Wrapf(err, "routeTable.GetICloudRouteTable()"))
		return
	}
	err = routeTable.SyncWithCloudRouteTable(ctx, self.GetUserCred(), vpc, iRouteTable, nil)
	if err != nil {
		self.taskFailed(ctx, routeTable, errors.Wrapf(err, "routeTable.GetICloudRouteTable()"))
		return
	}
	result := routeTable.SyncRouteTableRouteSets(ctx, self.GetUserCred(), iRouteTable, nil)
	if result.IsError() {
		self.taskFailed(ctx, routeTable, errors.Wrapf(result.AllError(), "routeTable.SyncRouteTableRouteSets()"))
		return
	}

	result = routeTable.SyncRouteTableAssociations(ctx, self.GetUserCred(), iRouteTable, nil)
	if result.IsError() {
		self.taskFailed(ctx, routeTable, errors.Wrapf(result.AllError(), "routeTable.SyncRouteTableAssociations()"))
		return
	}
	logclient.AddActionLogWithStartable(self, routeTable, logclient.ACT_SYNC_STATUS, nil, self.UserCred, true)
	self.SetStageComplete(ctx, nil)
}
