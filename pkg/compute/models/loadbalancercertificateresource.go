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

type SLoadbalancerCertificateResourceBase struct {
	// 本地负载均衡证书ID
	CertificateId string `width:"36" charset:"ascii" nullable:"true" list:"user" create:"optional" update:"user"`
}

type SLoadbalancerCertificateResourceBaseManager struct{}

func ValidateLoadbalancerCertificateResourceInput(userCred mcclient.TokenCredential, input api.LoadbalancerCertificateResourceInput) (*SLoadbalancerCertificate, api.LoadbalancerCertificateResourceInput, error) {
	lbcertObj, err := LoadbalancerCertificateManager.FetchByIdOrName(userCred, input.CertificateId)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, input, errors.Wrapf(httperrors.ErrResourceNotFound, "%s %s", LoadbalancerCertificateManager.Keyword(), input.CertificateId)
		} else {
			return nil, input, errors.Wrap(err, "LoadbalancerCertificateManager.FetchByIdOrName")
		}
	}
	input.CertificateId = lbcertObj.GetId()
	return lbcertObj.(*SLoadbalancerCertificate), input, nil
}

func (self *SLoadbalancerCertificateResourceBase) GetCertificate() *SLoadbalancerCertificate {
	cert, err := LoadbalancerCertificateManager.FetchById(self.CertificateId)
	if err != nil {
		log.Errorf("failed to find certificate %s error: %v", self.CertificateId, err)
		return nil
	}
	return cert.(*SLoadbalancerCertificate)
}

func (self *SLoadbalancerCertificateResourceBase) GetExtraDetails(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	isList bool,
) api.LoadbalancerCertificateResourceInfo {
	return api.LoadbalancerCertificateResourceInfo{}
}

func (manager *SLoadbalancerCertificateResourceBaseManager) FetchCustomizeColumns(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	objs []interface{},
	fields stringutils2.SSortedStrings,
	isList bool,
) []api.LoadbalancerCertificateResourceInfo {
	rows := make([]api.LoadbalancerCertificateResourceInfo, len(objs))
	certIds := make([]string, len(objs))
	for i := range objs {
		var base *SLoadbalancerCertificateResourceBase
		err := reflectutils.FindAnonymouStructPointer(objs[i], &base)
		if err != nil {
			log.Errorf("Cannot find SCloudregionResourceBase in object %s", objs[i])
			continue
		}
		certIds[i] = base.CertificateId
	}
	certs := make(map[string]SLoadbalancerCertificate)
	err := db.FetchStandaloneObjectsByIds(LoadbalancerCertificateManager, certIds, &certs)
	if err != nil {
		log.Errorf("FetchStandaloneObjectsByIds fail %s", err)
		return rows
	}
	for i := range rows {
		rows[i] = api.LoadbalancerCertificateResourceInfo{}
		if cert, ok := certs[certIds[i]]; ok {
			rows[i].Certificate = cert.Name
		}
	}
	return rows
}

func (manager *SLoadbalancerCertificateResourceBaseManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.LoadbalancerCertificateFilterListInput,
) (*sqlchemy.SQuery, error) {
	if len(query.CertificateId) > 0 {
		certObj, _, err := ValidateLoadbalancerCertificateResourceInput(userCred, query.LoadbalancerCertificateResourceInput)
		if err != nil {
			return nil, errors.Wrap(err, "ValidateLoadbalancerCertificateResourceInput")
		}
		q = q.Equals("certificate_id", certObj.GetId())
	}
	return q, nil
}

func (manager *SLoadbalancerCertificateResourceBaseManager) OrderByExtraFields(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.LoadbalancerCertificateFilterListInput,
) (*sqlchemy.SQuery, error) {
	q, orders, fields := manager.GetOrderBySubQuery(q, userCred, query)
	if len(orders) > 0 {
		q = db.OrderByFields(q, orders, fields)
	}
	return q, nil
}

func (manager *SLoadbalancerCertificateResourceBaseManager) QueryDistinctExtraField(q *sqlchemy.SQuery, field string) (*sqlchemy.SQuery, error) {
	if field == "certificate" {
		certQuery := LoadbalancerCertificateManager.Query("name", "id").Distinct().SubQuery()
		q.AppendField(certQuery.Field("name", field))
		q = q.Join(certQuery, sqlchemy.Equals(q.Field("certificate_id"), certQuery.Field("id")))
		q.GroupBy(certQuery.Field("name"))
		return q, nil
	}
	return q, httperrors.ErrNotFound
}

func (manager *SLoadbalancerCertificateResourceBaseManager) GetOrderBySubQuery(
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.LoadbalancerCertificateFilterListInput,
) (*sqlchemy.SQuery, []string, []sqlchemy.IQueryField) {
	certQ := LoadbalancerCertificateManager.Query("id", "name")
	var orders []string
	var fields []sqlchemy.IQueryField
	if db.NeedOrderQuery(manager.GetOrderByFields(query)) {
		subq := certQ.SubQuery()
		q = q.LeftJoin(subq, sqlchemy.Equals(q.Field("certificate_id"), subq.Field("id")))
		orders = append(orders, query.OrderByCertificate)
		fields = append(fields, subq.Field("name"))
	}
	return q, orders, fields
}

func (manager *SLoadbalancerCertificateResourceBaseManager) GetOrderByFields(query api.LoadbalancerCertificateFilterListInput) []string {
	return []string{query.OrderByCertificate}
}

func (manager *SLoadbalancerCertificateResourceBaseManager) ListItemExportKeys(ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	keys stringutils2.SSortedStrings,
) (*sqlchemy.SQuery, error) {
	if keys.ContainsAny(manager.GetExportKeys()...) {
		subq := LoadbalancerCertificateManager.Query("id", "name").SubQuery()
		q = q.LeftJoin(subq, sqlchemy.Equals(q.Field("certificate_id"), subq.Field("id")))
		if keys.Contains("certificate") {
			q = q.AppendField(subq.Field("name", "certificate"))
		}
	}
	return q, nil
}

func (manager *SLoadbalancerCertificateResourceBaseManager) GetExportKeys() []string {
	return []string{"certificate"}
}
