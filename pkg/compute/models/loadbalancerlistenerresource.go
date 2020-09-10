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
	"database/sql"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/util/reflectutils"
	"yunion.io/x/sqlchemy"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

type SLoadbalancerListenerResourceBase struct {
	// 负载均衡监听器ID
	ListenerId string `width:"36" charset:"ascii" nullable:"true" list:"user" create:"optional" json:"listener_id"`
}

type SLoadbalancerListenerResourceBaseManager struct {
	SLoadbalancerResourceBaseManager
}

func ValidateLoadbalancerListenerResourceInput(userCred mcclient.TokenCredential, input api.LoadbalancerListenerResourceInput) (*SLoadbalancerListener, api.LoadbalancerListenerResourceInput, error) {
	listenerObj, err := LoadbalancerListenerManager.FetchByIdOrName(userCred, input.ListenerId)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, input, errors.Wrapf(httperrors.ErrResourceNotFound, "%s %s", LoadbalancerListenerManager.Keyword(), input.ListenerId)
		} else {
			return nil, input, errors.Wrap(err, "LoadbalancerListenerManager.FetchByIdOrName")
		}
	}
	input.ListenerId = listenerObj.GetId()
	return listenerObj.(*SLoadbalancerListener), input, nil
}

func (self *SLoadbalancerListenerResourceBase) GetLoadbalancerListener() *SLoadbalancerListener {
	listener, err := LoadbalancerListenerManager.FetchById(self.ListenerId)
	if err != nil {
		log.Errorf("failed to find LoadbalancerCluster %s error: %v", self.ListenerId, err)
		return nil
	}
	return listener.(*SLoadbalancerListener)
}

func (self *SLoadbalancerListenerResourceBase) GetCloudproviderId() string {
	cloudprovider := self.GetCloudprovider()
	if cloudprovider != nil {
		return cloudprovider.Id
	}
	return ""
}

func (self *SLoadbalancerListenerResourceBase) GetCloudprovider() *SCloudprovider {
	listener := self.GetLoadbalancerListener()
	if listener != nil {
		return listener.GetCloudprovider()
	}
	return nil
}

func (self *SLoadbalancerListenerResourceBase) GetProviderName() string {
	listener := self.GetLoadbalancerListener()
	if listener != nil {
		return listener.GetProviderName()
	}
	return ""
}

func (self *SLoadbalancerListenerResourceBase) GetExtraDetails(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	isList bool,
) api.LoadbalancerListenerResourceInfo {
	return api.LoadbalancerListenerResourceInfo{}
}

func (manager *SLoadbalancerListenerResourceBaseManager) FetchCustomizeColumns(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	objs []interface{},
	fields stringutils2.SSortedStrings,
	isList bool,
) []api.LoadbalancerListenerResourceInfo {
	rows := make([]api.LoadbalancerListenerResourceInfo, len(objs))
	listenerIds := make([]string, len(objs))
	for i := range objs {
		var base *SLoadbalancerListenerResourceBase
		err := reflectutils.FindAnonymouStructPointer(objs[i], &base)
		if err != nil {
			log.Errorf("Cannot find SLoadbalancerListenerResourceBase in object %#v: %s", objs[i], err)
			continue
		}
		listenerIds[i] = base.ListenerId
	}
	listeners := make(map[string]SLoadbalancerListener)
	err := db.FetchStandaloneObjectsByIds(LoadbalancerListenerManager, listenerIds, listeners)
	if err != nil {
		log.Errorf("FetchStandaloneObjectsByIds fail %s", err)
		return rows
	}
	lbs := make([]interface{}, len(rows))
	for i := range rows {
		rows[i] = api.LoadbalancerListenerResourceInfo{}
		if listener, ok := listeners[listenerIds[i]]; ok {
			rows[i].Listener = listener.Name
			rows[i].LoadbalancerId = listener.LoadbalancerId
		}
		lbs[i] = &SLoadbalancerResourceBase{rows[i].LoadbalancerId}
	}

	lbRows := manager.SLoadbalancerResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, lbs, fields, isList)

	for i := range rows {
		rows[i].LoadbalancerResourceInfo = lbRows[i]
	}

	return rows
}

func (manager *SLoadbalancerListenerResourceBaseManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.LoadbalancerListenerFilterListInput,
) (*sqlchemy.SQuery, error) {
	if len(query.ListenerId) > 0 {
		listenerObj, _, err := ValidateLoadbalancerListenerResourceInput(userCred, query.LoadbalancerListenerResourceInput)
		if err != nil {
			return nil, errors.Wrap(err, "ValidateLoadbalancerListenerResourceInput")
		}
		q = q.Equals("listener_id", listenerObj.GetId())
	}
	subq := LoadbalancerListenerManager.Query("id").Snapshot()
	subq, err := manager.SLoadbalancerResourceBaseManager.ListItemFilter(ctx, subq, userCred, query.LoadbalancerFilterListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SLoadbalancerResourceBaseManager.ListItemFilter")
	}
	if subq.IsAltered() {
		q = q.Filter(sqlchemy.In(q.Field("listener_id"), subq.SubQuery()))
	}
	return q, nil
}

func (manager *SLoadbalancerListenerResourceBaseManager) OrderByExtraFields(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.LoadbalancerListenerFilterListInput,
) (*sqlchemy.SQuery, error) {
	q, orders, fields := manager.GetOrderBySubQuery(q, userCred, query)
	if len(orders) > 0 {
		q = db.OrderByFields(q, orders, fields)
	}
	return q, nil
}

func (manager *SLoadbalancerListenerResourceBaseManager) QueryDistinctExtraField(q *sqlchemy.SQuery, field string) (*sqlchemy.SQuery, error) {
	if field == "listener" {
		listenerQuery := LoadbalancerListenerManager.Query("name", "id").Distinct().SubQuery()
		q.AppendField(listenerQuery.Field("name", field))
		q = q.Join(listenerQuery, sqlchemy.Equals(q.Field("listener_id"), listenerQuery.Field("id")))
		q.GroupBy(listenerQuery.Field("name"))
		return q, nil
	}
	listeners := LoadbalancerListenerManager.Query("id", "loadbalancer_id").SubQuery()
	q = q.LeftJoin(listeners, sqlchemy.Equals(q.Field("listener_id"), listeners.Field("id")))
	q, err := manager.SLoadbalancerResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}
	return q, httperrors.ErrNotFound
}

func (manager *SLoadbalancerListenerResourceBaseManager) GetOrderBySubQuery(
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.LoadbalancerListenerFilterListInput,
) (*sqlchemy.SQuery, []string, []sqlchemy.IQueryField) {
	listenerQ := LoadbalancerListenerManager.Query("id", "name")
	var orders []string
	var fields []sqlchemy.IQueryField

	if db.NeedOrderQuery(manager.SLoadbalancerResourceBaseManager.GetOrderByFields(query.LoadbalancerFilterListInput)) {
		var lbOrders []string
		var lbFields []sqlchemy.IQueryField
		listenerQ, lbOrders, lbFields = manager.SLoadbalancerResourceBaseManager.GetOrderBySubQuery(listenerQ, userCred, query.LoadbalancerFilterListInput)
		if len(lbOrders) > 0 {
			orders = append(orders, lbOrders...)
			fields = append(fields, lbFields...)
		}
	}

	if db.NeedOrderQuery(manager.GetOrderByFields(query)) {
		subq := listenerQ.SubQuery()
		q = q.LeftJoin(subq, sqlchemy.Equals(q.Field("listener_id"), subq.Field("id")))
		if db.NeedOrderQuery([]string{query.OrderByListener}) {
			orders = append(orders, query.OrderByListener)
			fields = append(fields, subq.Field("name"))
		}
	}

	return q, orders, fields
}

func (manager *SLoadbalancerListenerResourceBaseManager) GetOrderByFields(query api.LoadbalancerListenerFilterListInput) []string {
	fields := make([]string, 0)
	lbFields := manager.SLoadbalancerResourceBaseManager.GetOrderByFields(query.LoadbalancerFilterListInput)
	fields = append(fields, lbFields...)
	fields = append(fields, query.OrderByListener)
	return fields
}

func (manager *SLoadbalancerListenerResourceBaseManager) ListItemExportKeys(ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	keys stringutils2.SSortedStrings,
) (*sqlchemy.SQuery, error) {
	if keys.ContainsAny(manager.GetExportKeys()...) {
		var err error
		subq := LoadbalancerListenerManager.Query("id", "name", "loadbalancer_id").SubQuery()
		q = q.LeftJoin(subq, sqlchemy.Equals(q.Field("loadbalancer_id"), subq.Field("id")))
		if keys.Contains("listener") {
			q = q.AppendField(subq.Field("name", "listener"))
		}
		if keys.ContainsAny(manager.SLoadbalancerResourceBaseManager.GetExportKeys()...) {
			q, err = manager.SLoadbalancerResourceBaseManager.ListItemExportKeys(ctx, q, userCred, keys)
			if err != nil {
				return nil, errors.Wrap(err, "SLoadbalancerResourceBaseManager.ListItemExportKeys")
			}
		}
	}

	return q, nil
}

func (manager *SLoadbalancerListenerResourceBaseManager) GetExportKeys() []string {
	keys := []string{"listener"}
	keys = append(keys, manager.SLoadbalancerResourceBaseManager.GetExportKeys()...)
	return keys
}