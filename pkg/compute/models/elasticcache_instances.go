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
	"regexp"
	"strings"
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/tristate"
	"yunion.io/x/pkg/util/compare"
	"yunion.io/x/pkg/utils"
	"yunion.io/x/sqlchemy"

	"yunion.io/x/onecloud/pkg/apis"
	"yunion.io/x/onecloud/pkg/apis/billing"
	billing_api "yunion.io/x/onecloud/pkg/apis/billing"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/lockman"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/quotas"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/cloudcommon/policy"
	"yunion.io/x/onecloud/pkg/cloudcommon/validators"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/options"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	bc "yunion.io/x/onecloud/pkg/util/billing"
	"yunion.io/x/onecloud/pkg/util/choices"
	"yunion.io/x/onecloud/pkg/util/logclient"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
	"yunion.io/x/onecloud/pkg/util/seclib2"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

type SElasticcacheManager struct {
	db.SVirtualResourceBaseManager
	db.SExternalizedResourceBaseManager
	SDeletePreventableResourceBaseManager
	SVpcResourceBaseManager
	SZoneResourceBaseManager
	SNetworkResourceBaseManager
}

var ElasticcacheManager *SElasticcacheManager

func init() {
	ElasticcacheManager = &SElasticcacheManager{
		SVirtualResourceBaseManager: db.NewVirtualResourceBaseManager(
			SElasticcache{},
			"elasticcacheinstances_tbl",
			"elasticcache",
			"elasticcaches",
		),
	}
	ElasticcacheManager.SetVirtualObject(ElasticcacheManager)
}

type SElasticcache struct {
	db.SVirtualResourceBase
	db.SExternalizedResourceBase
	SBillingResourceBase
	SDeletePreventableResourceBase
	SVpcResourceBase `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional"`
	SZoneResourceBase

	// 备可用区
	SlaveZones string `width:"128" charset:"ascii" nullable:"false" list:"user" create:"optional" json:"slave_zones"`

	// 实例规格
	// example: redis.master.micro.default
	InstanceType string `width:"96" charset:"ascii" nullable:"true" list:"user" create:"optional" json:"instance_type"`

	// 容量
	// example: 1024
	CapacityMB int `nullable:"false" list:"user" create:"optional" json:"capacity_mb"`

	// 对应Sku
	LocalCategory string `width:"16" charset:"ascii" nullable:"false" list:"user" create:"optional" json:"local_category"`

	// 类型
	// single（单副本） | double（双副本) | readone (单可读) | readthree （3可读） | readfive（5只读）
	NodeType string `width:"16" charset:"ascii" nullable:"false" list:"user" create:"optional" json:"node_type"`

	// 后端存储引擎
	// Redis | Memcache
	// example: redis
	Engine string `width:"16" charset:"ascii" nullable:"false" list:"user" create:"required" json:"engine"`

	// 后端存储引擎版本
	// example: 4.0
	EngineVersion string `width:"16" charset:"ascii" nullable:"false" list:"user" create:"required" json:"engine_version"`

	// VpcId           string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional"`

	// 网络类型, CLASSIC（经典网络）  VPC（专有网络）
	// example: CLASSIC
	NetworkType string `width:"16" charset:"ascii" nullable:"false" list:"user" create:"optional" json:"network_type"`

	// 所属网络ID
	NetworkId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional" json:"network_id"`

	// 安全组
	SecurityGroupId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"optional" json:"security_group_id"`

	//  内网DNS
	PrivateDNS string `width:"256" charset:"ascii" nullable:"true" list:"user" create:"optional" json:"private_dns"`

	//  内网IP地址
	PrivateIpAddr string `width:"17" charset:"ascii" list:"user" create:"optional" json:"private_ip_addr"`

	// 内网访问端口
	PrivateConnectPort int `nullable:"true" list:"user" create:"optional" json:"private_connect_port"`

	// 公网DNS
	PublicDNS string `width:"256" charset:"ascii" nullable:"true" list:"user" create:"optional" json:"public_dns"`

	// 公网IP地址
	PublicIpAddr string `width:"17" charset:"ascii" list:"user" create:"optional" json:"public_ip_addr"`

	// 外网访问端口
	PublicConnectPort int `nullable:"true" list:"user" create:"optional" json:"public_connect_port"`

	// 维护开始时间，格式为HH:mmZ
	// example: 02:00Z
	MaintainStartTime string `width:"8" charset:"ascii" nullable:"true" list:"user" create:"optional" json:"maintain_start_time"`

	// 维护结束时间
	MaintainEndTime string `width:"8" charset:"ascii" nullable:"true" list:"user" create:"optional" json:"maintain_end_time"`

	// 访问密码？ on （开启密码）|off （免密码访问）
	AuthMode string `width:"8" charset:"ascii" nullable:"false" list:"user" create:"optional" json:"auth_mode"`

	// AutoRenew // 自动续费
	// AutoRenewPeriod // 自动续费周期
}

// elastic cache 子资源获取owner id
func elasticcacheSubResourceFetchOwnerId(ctx context.Context, data jsonutils.JSONObject) (mcclient.IIdentityProvider, error) {
	parentId := jsonutils.GetAnyString(data, []string{"elasticcache_id", "elasticcache"})
	if len(parentId) > 0 {
		userCred := policy.FetchUserCredential(ctx)
		ec, err := db.FetchByIdOrName(ElasticcacheManager, userCred, parentId)
		if err != nil {
			log.Errorf("elasticcache sub resource FetchOwnerId %s", err)
			return nil, nil
		}

		return ec.(*SElasticcache).GetOwnerId(), nil
	}

	return db.FetchProjectInfo(ctx, data)
}

// elastic cache 子资源获取owner query
func elasticcacheSubResourceFetchOwner(q *sqlchemy.SQuery, userCred mcclient.IIdentityProvider, scope rbacutils.TRbacScope) *sqlchemy.SQuery {
	if userCred != nil {
		var subq *sqlchemy.SSubQuery

		q1 := ElasticcacheManager.Query()
		switch scope {
		case rbacutils.ScopeProject:
			subq = q1.Equals("tenant_id", userCred.GetProjectId()).SubQuery()
		case rbacutils.ScopeDomain:
			subq = q1.Equals("domain_id", userCred.GetProjectDomainId()).SubQuery()
		}

		if subq != nil {
			q = q.Join(subq, sqlchemy.Equals(q.Field("elasticcache_id"), subq.Field("id")))
		}
	}

	return q
}

func (self *SElasticcache) getCloudProviderInfo() SCloudProviderInfo {
	region := self.GetRegion()
	provider := self.GetCloudprovider()
	zone := self.GetZone()
	return MakeCloudProviderInfo(region, zone, provider)
}

func (self *SElasticcache) GetExtraDetails(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, isList bool) (api.ElasticcacheDetails, error) {
	return api.ElasticcacheDetails{}, nil
}

func (manager *SElasticcacheManager) FetchCustomizeColumns(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	objs []interface{},
	fields stringutils2.SSortedStrings,
	isList bool,
) []api.ElasticcacheDetails {
	rows := make([]api.ElasticcacheDetails, len(objs))

	virtRows := manager.SVirtualResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, objs, fields, isList)
	vpcRows := manager.SVpcResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, objs, fields, isList)
	zoneRows := manager.SZoneResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, objs, fields, isList)

	netIds := make([]string, len(objs))
	for i := range rows {
		rows[i] = api.ElasticcacheDetails{
			VirtualResourceDetails: virtRows[i],
			VpcResourceInfo:        vpcRows[i],
			ZoneResourceInfoBase:   zoneRows[i].ZoneResourceInfoBase,
		}
		netIds[i] = objs[i].(*SElasticcache).NetworkId
	}

	networks := make(map[string]SNetwork)
	err := db.FetchStandaloneObjectsByIds(NetworkManager, netIds, &networks)
	if err != nil {
		log.Errorf("FetchStandaloneObjectsByIds fail %s", err)
		return rows
	}

	for i := range rows {
		if net, ok := networks[netIds[i]]; ok {
			rows[i].Network = net.Name
		}
	}

	return rows
}

func (self *SElasticcache) GetElasticcacheParameters() ([]SElasticcacheParameter, error) {
	ret := []SElasticcacheParameter{}
	q := ElasticcacheParameterManager.Query().Equals("elasticcache_id", self.Id)
	err := db.FetchModelObjects(ElasticcacheParameterManager, q, &ret)
	if err != nil {
		return nil, errors.Wrapf(err, "GetElasticcacheParameters.FetchModelObjects for elastic cache %s", self.Id)
	}
	return ret, nil
}

func (self *SElasticcache) GetElasticcacheAccounts() ([]SElasticcacheAccount, error) {
	ret := []SElasticcacheAccount{}
	q := ElasticcacheAccountManager.Query().Equals("elasticcache_id", self.Id)
	err := db.FetchModelObjects(ElasticcacheAccountManager, q, &ret)
	if err != nil {
		return nil, errors.Wrapf(err, "GetElasticcacheAccounts.FetchModelObjects for elastic cache %s", self.Id)
	}
	return ret, nil
}

func (self *SElasticcache) GetElasticcacheAcls() ([]SElasticcacheAcl, error) {
	ret := []SElasticcacheAcl{}
	q := ElasticcacheAclManager.Query().Equals("elasticcache_id", self.Id)
	err := db.FetchModelObjects(ElasticcacheAclManager, q, &ret)
	if err != nil {
		return nil, errors.Wrapf(err, "GetElasticcacheAcls.FetchModelObjects for elastic cache %s", self.Id)
	}
	return ret, nil
}

func (self *SElasticcache) GetElasticcacheBackups() ([]SElasticcacheBackup, error) {
	ret := []SElasticcacheBackup{}
	q := ElasticcacheBackupManager.Query().Equals("elasticcache_id", self.Id)
	err := db.FetchModelObjects(ElasticcacheBackupManager, q, &ret)
	if err != nil {
		return nil, errors.Wrapf(err, "GetElasticcacheBackups.FetchModelObjects for elastic cache %s", self.Id)
	}
	return ret, nil
}

func (self *SElasticcache) AllowGetDetailsLoginInfo(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowGetSpec(userCred, self, "login-info")
}

func (self *SElasticcache) GetDetailsLoginInfo(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	account, err := self.GetAdminAccount()
	if err != nil {
		return nil, err
	}

	ret := jsonutils.NewDict()
	ret.Add(jsonutils.NewString(account.Id), "account_id")
	ret.Add(jsonutils.NewString(account.Name), "username")
	ret.Add(jsonutils.NewString(account.Password), "password")
	return ret, nil
}

func (manager *SElasticcacheManager) GetOwnerIdByElasticcacheId(elasticcacheId string) mcclient.IIdentityProvider {
	ec, err := db.FetchById(ElasticcacheManager, elasticcacheId)
	if err != nil {
		log.Errorf("SElasticcacheManager.GetOwnerIdByElasticcacheId %s", err)
		return nil
	}

	return ec.(*SElasticcache).GetOwnerId()
}

// 列出弹性缓存（redis等）
func (manager *SElasticcacheManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.ElasticcacheListInput,
) (*sqlchemy.SQuery, error) {
	var err error
	q, err = manager.SVirtualResourceBaseManager.ListItemFilter(ctx, q, userCred, query.VirtualResourceListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVirtualResourceBaseManager.ListItemFilter")
	}
	q, err = manager.SExternalizedResourceBaseManager.ListItemFilter(ctx, q, userCred, query.ExternalizedResourceBaseListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SExternalizedResourceBaseManager.ListItemFilter")
	}
	q, err = manager.SDeletePreventableResourceBaseManager.ListItemFilter(ctx, q, userCred, query.DeletePreventableResourceBaseListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SDeletePreventableResourceBaseManager.ListItemFilter")
	}
	q, err = manager.SVpcResourceBaseManager.ListItemFilter(ctx, q, userCred, query.VpcFilterListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVpcResourceBaseManager.ListItemFilter")
	}
	zoneQuery := api.ZonalFilterListInput{
		ZonalFilterListBase: query.ZonalFilterListBase,
	}
	q, err = manager.SZoneResourceBaseManager.ListItemFilter(ctx, q, userCred, zoneQuery)
	if err != nil {
		return nil, errors.Wrap(err, "SZoneResourceBaseManager.ListItemFilter")
	}

	if len(query.InstanceType) > 0 {
		q = q.In("instance_type", query.InstanceType)
	}
	if len(query.LocalCategory) > 0 {
		q = q.In("local_category", query.LocalCategory)
	}
	if len(query.NodeType) > 0 {
		q = q.In("node_type", query.NodeType)
	}
	if len(query.Engine) > 0 {
		q = q.In("engine", query.Engine)
	}
	if len(query.EngineVersion) > 0 {
		q = q.In("engine_version", query.EngineVersion)
	}
	if len(query.NetworkType) > 0 {
		q = q.In("network_type", query.NetworkType)
	}
	netQuery := api.NetworkFilterListInput{
		NetworkFilterListBase: query.NetworkFilterListBase,
	}
	q, err = manager.SNetworkResourceBaseManager.ListItemFilter(ctx, q, userCred, netQuery)
	if err != nil {
		return nil, errors.Wrap(err, "SNetworkResourceBaseManager.ListItemFilter")
	}

	if len(query.PrivateDNS) > 0 {
		q = q.In("private_dns", query.PrivateDNS)
	}
	if len(query.PrivateIpAddr) > 0 {
		q = q.In("private_ip_addr", query.PrivateIpAddr)
	}
	if len(query.PrivateConnectPort) > 0 {
		q = q.In("private_connect_port", query.PrivateConnectPort)
	}
	if len(query.PublicDNS) > 0 {
		q = q.In("public_dns", query.PublicDNS)
	}
	if len(query.PublicIpAddr) > 0 {
		q = q.In("public_ip_addr", query.PublicIpAddr)
	}
	if len(query.PublicConnectPort) > 0 {
		q = q.In("public_connect_port", query.PublicConnectPort)
	}
	if len(query.AuthMode) > 0 {
		q = q.In("auth_mode", query.AuthMode)
	}

	return q, nil
}

func (manager *SElasticcacheManager) OrderByExtraFields(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.ElasticcacheListInput,
) (*sqlchemy.SQuery, error) {
	var err error

	q, err = manager.SVirtualResourceBaseManager.OrderByExtraFields(ctx, q, userCred, query.VirtualResourceListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVirtualResourceBaseManager.OrderByExtraFields")
	}
	q, err = manager.SVpcResourceBaseManager.OrderByExtraFields(ctx, q, userCred, query.VpcFilterListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVpcResourceBaseManager.OrderByExtraFields")
	}
	zoneQuery := api.ZonalFilterListInput{
		ZonalFilterListBase: query.ZonalFilterListBase,
	}
	q, err = manager.SZoneResourceBaseManager.OrderByExtraFields(ctx, q, userCred, zoneQuery)
	if err != nil {
		return nil, errors.Wrap(err, "SZoneResourceBaseManager.OrderByExtraFields")
	}

	return q, nil
}

func (manager *SElasticcacheManager) QueryDistinctExtraField(q *sqlchemy.SQuery, field string) (*sqlchemy.SQuery, error) {
	var err error

	q, err = manager.SVirtualResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}
	q, err = manager.SVpcResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}
	q, err = manager.SZoneResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}

	return q, httperrors.ErrNotFound
}

func (manager *SElasticcacheManager) SyncElasticcaches(ctx context.Context, userCred mcclient.TokenCredential, syncOwnerId mcclient.IIdentityProvider, provider *SCloudprovider, region *SCloudregion, cloudElasticcaches []cloudprovider.ICloudElasticcache) ([]SElasticcache, []cloudprovider.ICloudElasticcache, compare.SyncResult) {
	lockman.LockClass(ctx, manager, db.GetLockClassKey(manager, provider.GetOwnerId()))
	defer lockman.ReleaseClass(ctx, manager, db.GetLockClassKey(manager, provider.GetOwnerId()))

	localElasticcaches := []SElasticcache{}
	remoteElasticcaches := []cloudprovider.ICloudElasticcache{}
	syncResult := compare.SyncResult{}

	dbInstances, err := region.GetElasticcaches(provider)
	if err != nil {
		syncResult.Error(err)
		return nil, nil, syncResult
	}

	for i := range dbInstances {
		if taskman.TaskManager.IsInTask(&dbInstances[i]) {
			syncResult.Error(fmt.Errorf("ElasticCacheInstance %s(%s)in task", dbInstances[i].Name, dbInstances[i].Id))
			return nil, nil, syncResult
		}
	}

	removed := make([]SElasticcache, 0)
	commondb := make([]SElasticcache, 0)
	commonext := make([]cloudprovider.ICloudElasticcache, 0)
	added := make([]cloudprovider.ICloudElasticcache, 0)
	if err := compare.CompareSets(dbInstances, cloudElasticcaches, &removed, &commondb, &commonext, &added); err != nil {
		syncResult.Error(err)
		return nil, nil, syncResult
	}

	for i := 0; i < len(removed); i++ {
		err := removed[i].syncRemoveCloudElasticcache(ctx, userCred)
		if err != nil {
			syncResult.DeleteError(err)
		} else {
			syncResult.Delete()
		}
	}

	for i := 0; i < len(commondb); i++ {
		err := commondb[i].SyncWithCloudElasticcache(ctx, userCred, provider, commonext[i])
		if err != nil {
			syncResult.UpdateError(err)
			continue
		}
		syncVirtualResourceMetadata(ctx, userCred, &commondb[i], commonext[i])
		localElasticcaches = append(localElasticcaches, commondb[i])
		remoteElasticcaches = append(remoteElasticcaches, commonext[i])
		syncResult.Update()
	}

	for i := 0; i < len(added); i++ {
		instance, err := manager.newFromCloudElasticcache(ctx, userCred, syncOwnerId, provider, region, added[i])
		if err != nil {
			syncResult.AddError(err)
			continue
		}
		syncVirtualResourceMetadata(ctx, userCred, instance, added[i])
		localElasticcaches = append(localElasticcaches, *instance)
		remoteElasticcaches = append(remoteElasticcaches, added[i])
		syncResult.Add()
	}
	return localElasticcaches, remoteElasticcaches, syncResult
}

func (self *SElasticcache) syncRemoveCloudElasticcache(ctx context.Context, userCred mcclient.TokenCredential) error {
	lockman.LockObject(ctx, self)
	defer lockman.ReleaseObject(ctx, self)

	self.SetDisableDelete(userCred, false)

	self.DeleteSubResources(ctx, userCred)

	err := self.ValidateDeleteCondition(ctx)
	if err != nil {
		self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_ERROR, "sync to delete")
		return errors.Wrap(err, "ValidateDeleteCondition")
	}
	return self.SVirtualResourceBase.Delete(ctx, userCred)
}

func (self *SElasticcache) SyncWithCloudElasticcache(ctx context.Context, userCred mcclient.TokenCredential, provider *SCloudprovider, extInstance cloudprovider.ICloudElasticcache) error {
	diff, err := db.UpdateWithLock(ctx, self, func() error {
		self.Status = extInstance.GetStatus()
		self.InstanceType = extInstance.GetInstanceType()
		self.CapacityMB = extInstance.GetCapacityMB()
		self.LocalCategory = extInstance.GetArchType()
		self.NodeType = extInstance.GetNodeType()
		self.Engine = extInstance.GetEngine()
		self.EngineVersion = extInstance.GetEngineVersion()

		self.NetworkType = extInstance.GetNetworkType()
		self.PrivateDNS = extInstance.GetPrivateDNS()
		self.PrivateIpAddr = extInstance.GetPrivateIpAddr()
		self.PrivateConnectPort = extInstance.GetPrivateConnectPort()
		self.PublicDNS = extInstance.GetPublicDNS()
		self.PublicIpAddr = extInstance.GetPublicIpAddr()
		self.PublicConnectPort = extInstance.GetPublicConnectPort()
		self.MaintainStartTime = extInstance.GetMaintainStartTime()
		self.MaintainEndTime = extInstance.GetMaintainEndTime()
		self.AuthMode = extInstance.GetAuthMode()

		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "syncWithCloudElasticcache.Update")
	}

	db.OpsLog.LogSyncUpdate(self, diff, userCred)
	return nil
}

func (manager *SElasticcacheManager) newFromCloudElasticcache(ctx context.Context, userCred mcclient.TokenCredential, ownerId mcclient.IIdentityProvider, provider *SCloudprovider, region *SCloudregion, extInstance cloudprovider.ICloudElasticcache) (*SElasticcache, error) {
	lockman.LockClass(ctx, manager, db.GetLockClassKey(manager, userCred))
	defer lockman.ReleaseClass(ctx, manager, db.GetLockClassKey(manager, userCred))

	instance := SElasticcache{}
	instance.SetModelManager(manager, &instance)

	newName, err := db.GenerateName(manager, ownerId, extInstance.GetName())
	if err != nil {
		return nil, err
	}
	instance.Name = newName
	instance.ExternalId = extInstance.GetGlobalId()
	// instance.CloudregionId = region.Id
	// instance.ManagerId = provider.Id
	instance.IsEmulated = extInstance.IsEmulated()
	instance.Status = extInstance.GetStatus()

	instance.InstanceType = extInstance.GetInstanceType()
	instance.CapacityMB = extInstance.GetCapacityMB()
	instance.LocalCategory = extInstance.GetArchType()
	instance.NodeType = extInstance.GetNodeType()
	instance.Engine = extInstance.GetEngine()
	instance.EngineVersion = extInstance.GetEngineVersion()

	instance.PrivateDNS = extInstance.GetPrivateDNS()
	instance.PrivateIpAddr = extInstance.GetPrivateIpAddr()
	instance.PrivateConnectPort = extInstance.GetPrivateConnectPort()
	instance.PublicDNS = extInstance.GetPublicDNS()
	instance.PublicIpAddr = extInstance.GetPublicIpAddr()
	instance.PublicConnectPort = extInstance.GetPublicConnectPort()
	instance.MaintainStartTime = extInstance.GetMaintainStartTime()
	instance.MaintainEndTime = extInstance.GetMaintainEndTime()
	instance.AuthMode = extInstance.GetAuthMode()

	var zone *SZone
	if zoneId := extInstance.GetZoneId(); len(zoneId) > 0 {
		_zone, err := db.FetchByExternalId(ZoneManager, zoneId)
		if err != nil {
			return nil, errors.Wrapf(err, "newFromCloudElasticcache.FetchZoneId")
		}
		instance.ZoneId = _zone.GetId()
		zone = _zone.(*SZone)
	}

	instance.NetworkType = extInstance.GetNetworkType()
	if instance.NetworkType == api.LB_NETWORK_TYPE_CLASSIC {
		vpc, err := VpcManager.GetOrCreateVpcForClassicNetwork(ctx, provider, region)
		if err != nil {
			return nil, errors.Wrap(err, "NewVpcForClassicNetwork")
		}
		instance.VpcId = vpc.GetId()

		wire, err := WireManager.GetOrCreateWireForClassicNetwork(ctx, vpc, zone)
		if err != nil {
			return nil, errors.Wrap(err, "NewWireForClassicNetwork")
		}
		network, err := NetworkManager.GetOrCreateClassicNetwork(ctx, wire)
		if err != nil {
			return nil, errors.Wrap(err, "GetOrCreateClassicNetwork")
		}
		instance.NetworkId = network.GetId()
	} else {
		if vpcId := extInstance.GetVpcId(); len(vpcId) > 0 {
			vpc, err := db.FetchByExternalIdAndManagerId(VpcManager, vpcId, func(q *sqlchemy.SQuery) *sqlchemy.SQuery {
				return q.Equals("manager_id", provider.Id)
			})
			if err != nil {
				return nil, errors.Wrapf(err, "newFromCloudElasticcache.FetchVpcId")
			}
			instance.VpcId = vpc.GetId()
		}

		if networkId := extInstance.GetNetworkId(); len(networkId) > 0 {
			network, err := db.FetchByExternalIdAndManagerId(NetworkManager, networkId, func(q *sqlchemy.SQuery) *sqlchemy.SQuery {
				wire := WireManager.Query().SubQuery()
				vpc := VpcManager.Query().SubQuery()
				return q.Join(wire, sqlchemy.Equals(wire.Field("id"), q.Field("wire_id"))).
					Join(vpc, sqlchemy.Equals(vpc.Field("id"), wire.Field("vpc_id"))).
					Filter(sqlchemy.Equals(vpc.Field("manager_id"), provider.Id))
			})
			if err != nil {
				return nil, errors.Wrapf(err, "newFromCloudElasticcache.FetchNetworkId")
			}
			instance.NetworkId = network.GetId()
		}
	}

	if createdAt := extInstance.GetCreatedAt(); !createdAt.IsZero() {
		instance.CreatedAt = createdAt
	}

	factory, err := provider.GetProviderFactory()
	if err != nil {
		return nil, errors.Wrap(err, "newFromCloudElasticcache.GetProviderFactory")
	}

	if factory.IsSupportPrepaidResources() {
		instance.BillingType = extInstance.GetBillingType()
		if expired := extInstance.GetExpiredAt(); !expired.IsZero() {
			instance.ExpiredAt = expired
		}
		instance.AutoRenew = extInstance.IsAutoRenew()
	}

	err = manager.TableSpec().Insert(ctx, &instance)
	if err != nil {
		return nil, errors.Wrapf(err, "newFromCloudElasticcache.Insert")
	}

	SyncCloudProject(userCred, &instance, ownerId, extInstance, provider.Id)
	db.OpsLog.LogEvent(&instance, db.ACT_CREATE, instance.GetShortDesc(ctx), userCred)

	return &instance, nil
}

func (manager *SElasticcacheManager) getElasticcachesByProviderId(providerId string) ([]SElasticcache, error) {
	instances := []SElasticcache{}
	err := fetchByVpcManagerId(manager, providerId, &instances)
	if err != nil {
		return nil, errors.Wrapf(err, "getElasticcachesByProviderId.fetchByManagerId")
	}
	return instances, nil
}

func (manager *SElasticcacheManager) AllowCreateItem(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return db.IsAdminAllowCreate(userCred, manager)
}

func (manager *SElasticcacheManager) BatchCreateValidateCreateData(ctx context.Context, userCred mcclient.TokenCredential, ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	input, err := manager.ValidateCreateData(ctx, userCred, ownerId, query, data)
	if err != nil {
		return nil, err
	}

	return input, nil
}

func (manager *SElasticcacheManager) ValidateCreateData(ctx context.Context, userCred mcclient.TokenCredential, ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {
	var region *SCloudregion
	if id, _ := data.GetString("network"); len(id) > 0 {
		network, err := db.FetchByIdOrName(NetworkManager, userCred, strings.Split(id, ",")[0])
		if err != nil {
			return nil, fmt.Errorf("getting network failed")
		}
		region = network.(*SNetwork).GetRegion()
	}

	if region == nil {
		return nil, fmt.Errorf("getting region failed")
	}

	// postpiad billing cycle
	billingType, _ := data.GetString("billing_type")
	if billingType == billing_api.BILLING_TYPE_POSTPAID {
		billingCycle, _ := data.GetString("duration")
		if len(billingCycle) > 0 {
			cycle, err := bc.ParseBillingCycle(billingCycle)
			if err != nil {
				return nil, httperrors.NewInputParameterError("invalid billing_cycle %s", billingCycle)
			}

			tm := time.Time{}
			data.Set("billing_cycle", jsonutils.NewString(cycle.String()))
			data.Set("expired_at", jsonutils.NewString(cycle.EndAt(tm).Format("2006-01-02 15:04:05")))
		}
	}

	input := apis.VirtualResourceCreateInput{}
	var err error
	err = data.Unmarshal(&input)
	if err != nil {
		return nil, httperrors.NewInternalServerError("unmarshal VirtualResourceCreateInput fail %s", err)
	}
	input, err = manager.SVirtualResourceBaseManager.ValidateCreateData(ctx, userCred, ownerId, query, input)
	if err != nil {
		return nil, err
	}
	data.Update(jsonutils.Marshal(input))

	return region.GetDriver().ValidateCreateElasticcacheData(ctx, userCred, nil, data)
}

func (self *SElasticcache) PostCreate(ctx context.Context, userCred mcclient.TokenCredential, ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject, data jsonutils.JSONObject) {
	self.SVirtualResourceBase.PostCreate(ctx, userCred, ownerId, query, data)

	password, _ := data.GetString("password")
	if reset, _ := data.Bool("reset_password"); reset && len(password) == 0 {
		password = seclib2.RandomPassword2(12)
	}

	params := jsonutils.NewDict()
	params.Set("password", jsonutils.NewString(password))
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_DEPLOYING, "")
	if err := self.StartElasticcacheCreateTask(ctx, userCred, params, ""); err != nil {
		log.Errorf("Failed to create elastic cache error: %v", err)
	}
}

func (self *SElasticcache) StartElasticcacheCreateTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheCreateTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}
	task.ScheduleRun(nil)
	return nil
}

/*func (self *SElasticcache) GetIRegion() (cloudprovider.ICloudRegion, error) {
	provider, err := self.GetDriver()
	if err != nil {
		return nil, fmt.Errorf("No cloudprovider for elastic cache %s: %s", self.Name, err)
	}
	region := self.GetRegion()
	if region == nil {
		return nil, fmt.Errorf("failed to find region for elastic cache %s", self.Name)
	}
	return provider.GetIRegionById(region.ExternalId)
}*/

func (self *SElasticcache) GetCreateAliyunElasticcacheParams(data *jsonutils.JSONDict) (*cloudprovider.SCloudElasticCacheInput, error) {
	input := &cloudprovider.SCloudElasticCacheInput{}
	iregion, err := self.GetIRegion()
	if err != nil {
		return nil, fmt.Errorf("elastic cache %s(%s) region not found", self.Name, self.Id)
	} else {
		input.RegionId = iregion.GetId()
	}

	input.InstanceType = self.InstanceType
	input.InstanceName = self.GetName()

	if password, _ := data.GetString("password"); len(password) > 0 {
		input.Password = password
	}

	input.Engine = strings.Title(self.Engine)
	input.EngineVersion = self.EngineVersion
	input.PrivateIpAddress = self.PrivateIpAddr

	zone := self.GetZone()
	if zone != nil {
		izone, err := iregion.GetIZoneById(zone.ExternalId)
		if err != nil {
			return nil, errors.Wrap(err, "elasticcache.GetCreateAliyunElasticcacheParams.Zone")
		}
		input.ZoneIds = []string{izone.GetId()}
	}

	switch self.BillingType {
	case billing.BILLING_TYPE_PREPAID:
		input.ChargeType = "PrePaid"
		billingCycle, err := bc.ParseBillingCycle(self.BillingCycle)
		if err != nil {
			return nil, errors.Wrap(err, "elasticcache.GetCreateAliyunElasticcacheParams.BillingCycle")
		}
		input.BC = &billingCycle
	default:
		input.ChargeType = "PostPaid"
	}

	// todo: fix me
	if len(self.NodeType) > 0 {
		switch self.NodeType {
		case "single":
			input.NodeType = "STAND_ALONE"
		case "double":
			input.NodeType = "MASTER_SLAVE"
		default:
			input.NodeType = ""
		}
	}

	switch self.NetworkType {
	case api.LB_NETWORK_TYPE_CLASSIC:
		input.NetworkType = "CLASSIC"
	default:
		input.NetworkType = "VPC"
	}

	if ivpc, err := db.FetchById(VpcManager, self.VpcId); err != nil {
		return nil, errors.Wrap(err, "elasticcache.GetCreateAliyunElasticcacheParams.Vpc")
	} else {
		if ivpc != nil {
			vpc := ivpc.(*SVpc)
			input.VpcId = vpc.ExternalId
		}
	}

	if inetwork, err := db.FetchById(NetworkManager, self.NetworkId); err != nil {
		return nil, errors.Wrap(err, "elasticcache.GetCreateAliyunElasticcacheParams.Network")
	} else {
		if inetwork != nil {
			network := inetwork.(*SNetwork)
			input.NetworkId = network.ExternalId
		}
	}

	return input, nil
}

func (self *SElasticcache) GetCreateHuaweiElasticcacheParams(data *jsonutils.JSONDict) (*cloudprovider.SCloudElasticCacheInput, error) {
	input := &cloudprovider.SCloudElasticCacheInput{}
	iregion, err := self.GetIRegion()
	if err != nil {
		return nil, fmt.Errorf("elastic cache %s(%s) region not found", self.Name, self.Id)
	} else {
		input.RegionId = iregion.GetId()
	}

	if self.CapacityMB > 0 {
		input.CapacityGB = int64(self.CapacityMB / 1024)
	}

	input.InstanceType = self.InstanceType
	input.InstanceName = self.GetName()

	if password, _ := data.GetString("password"); len(password) > 0 {
		input.Password = password
	}

	switch self.Engine {
	case "redis":
		input.Engine = "Redis"
	case "memcache":
		input.Engine = "Memcached"
	}
	input.EngineVersion = self.EngineVersion
	input.PrivateIpAddress = self.PrivateIpAddr

	zone := self.GetZone()
	if zone != nil {
		izone, err := iregion.GetIZoneById(zone.ExternalId)
		if err != nil {
			return nil, errors.Wrap(err, "elasticcache.GetCreateHuaweiElasticcacheParams.Zone")
		}
		input.ZoneIds = []string{izone.GetId()}
	}

	switch self.BillingType {
	case billing.BILLING_TYPE_PREPAID:
		input.ChargeType = "PrePaid"
		billingCycle, err := bc.ParseBillingCycle(self.BillingCycle)
		if err != nil {
			return nil, errors.Wrap(err, "elasticcache.GetCreateHuaweiElasticcacheParams.BillingCycle")
		}
		input.BC = &billingCycle
	default:
		input.ChargeType = "PostPaid"
	}

	if len(self.NodeType) > 0 {
		input.NodeType = self.NodeType
	}

	switch self.NetworkType {
	case api.LB_NETWORK_TYPE_CLASSIC:
		input.NetworkType = "CLASSIC"
	default:
		input.NetworkType = "VPC"
	}

	if ivpc, err := db.FetchById(VpcManager, self.VpcId); err != nil {
		return nil, errors.Wrap(err, "elasticcache.GetCreateHuaweiElasticcacheParams.Vpc")
	} else {
		if ivpc != nil {
			vpc := ivpc.(*SVpc)
			input.VpcId = vpc.ExternalId
		}
	}

	if inetwork, err := db.FetchById(NetworkManager, self.NetworkId); err != nil {
		return nil, errors.Wrap(err, "elasticcache.GetCreateHuaweiElasticcacheParams.Network")
	} else {
		if inetwork != nil {
			network := inetwork.(*SNetwork)
			input.NetworkId = network.ExternalId
		}
	}

	// fill security group here
	if len(self.SecurityGroupId) > 0 {
		sgCache, err := SecurityGroupCacheManager.GetSecgroupCache(context.Background(), nil, self.SecurityGroupId, self.VpcId, self.GetRegion().Id, self.GetCloudprovider().Id, "")
		if err != nil {
			return nil, errors.Wrap(err, "elasticcache.GetCreateHuaweiElasticcacheParams.SecurityGroup")
		}

		if sgCache == nil {
			return nil, errors.Wrap(fmt.Errorf("cached security group not found"), "elasticcache.GetCreateHuaweiElasticcacheParams.SecurityGroup")
		}

		input.SecurityGroupId = sgCache.GetExternalId()
	}

	if len(self.MaintainEndTime) > 0 {
		input.MaintainBegin = self.MaintainStartTime
		input.MaintainEnd = self.MaintainEndTime
	}
	return input, nil
}

func (self *SElasticcache) AllowPerformRestart(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "restart")
}

func (self *SElasticcache) PerformRestart(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	if utils.IsInStringArray(self.Status, []string{api.ELASTIC_CACHE_STATUS_RUNNING, api.ELASTIC_CACHE_STATUS_INACTIVE}) {
		self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_RESTARTING, "")
		return nil, self.StartRestartTask(ctx, userCred, "", data)
	} else {
		return nil, httperrors.NewInvalidStatusError("Cannot do restart elasticcache instance in status %s", self.Status)
	}
}

func (self *SElasticcache) StartRestartTask(ctx context.Context, userCred mcclient.TokenCredential, parentTaskId string, data jsonutils.JSONObject) error {
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_RESTARTING, "")
	if task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheRestartTask", self, userCred, data.(*jsonutils.JSONDict), parentTaskId, "", nil); err != nil {
		log.Errorln(err)
		return err
	} else {
		task.ScheduleRun(nil)
	}

	return nil
}

func (self *SElasticcache) ValidateDeleteCondition(ctx context.Context) error {
	if self.DisableDelete.IsTrue() {
		return httperrors.NewInvalidStatusError("Elastic cache is locked, cannot delete")
	}

	if self.GetChargeType() == billing.BILLING_TYPE_PREPAID && self.ExpiredAt.Sub(time.Now()).Seconds() > 0 {
		return httperrors.NewInvalidStatusError("Elastic cache is not expired, cannot delete")
	}

	return self.ValidatePurgeCondition(ctx)
}

func (self *SElasticcache) ValidatePurgeCondition(ctx context.Context) error {
	return self.SVirtualResourceBase.ValidateDeleteCondition(ctx)
}

func (self *SElasticcache) CustomizeDelete(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) error {
	return self.StartDeleteElasticcacheTask(ctx, userCred, jsonutils.NewDict(), "")
}

func (self *SElasticcache) StartDeleteElasticcacheTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheDeleteTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}
	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformChangeSpec(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "change_spec")
}

func (self *SElasticcache) ValidatorChangeSpecData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	skuV := validators.NewModelIdOrNameValidator("sku", "elasticcachesku", self.GetOwnerId())
	if err := skuV.Optional(false).Validate(data.(*jsonutils.JSONDict)); err != nil {
		return nil, err
	}

	sku := skuV.Model.(*SElasticcacheSku)
	if sku.Provider != self.GetProviderName() {
		return nil, httperrors.NewInputParameterError("provider mismatch: %s instance can't use %s sku", self.GetProviderName(), sku.Provider)
	}

	region := self.GetRegion()
	if sku.CloudregionId != region.Id {
		return nil, httperrors.NewInputParameterError("region mismatch: instance region %s, sku region %s", region.Id, sku.CloudregionId)
	}

	if sku.ZoneId != "" && sku.ZoneId != self.ZoneId {
		return nil, httperrors.NewInputParameterError("zone mismatch: instance zone %s, sku zone %s", self.ZoneId, sku.ZoneId)
	}

	if self.EngineVersion != "" && sku.EngineVersion != self.EngineVersion {
		return nil, httperrors.NewInputParameterError("engine version mismatch: instance version %s, sku version %s", self.EngineVersion, sku.EngineVersion)
	}

	data.(*jsonutils.JSONDict).Set("sku_ext_id", jsonutils.NewString(skuV.Model.(*SElasticcacheSku).GetName()))
	return data, nil
}

func (self *SElasticcache) PerformChangeSpec(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	if !utils.IsInStringArray(self.Status, []string{api.ELASTIC_CACHE_STATUS_RUNNING}) {
		return nil, httperrors.NewResourceNotReadyError("can not change specification in status %s", self.Status)
	}

	data, err := self.ValidatorChangeSpecData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	params := jsonutils.NewDict()
	sku, _ := data.GetString("sku_ext_id")
	params.Set("sku_ext_id", jsonutils.NewString(sku))
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_CHANGING, "")
	return nil, self.StartChangeSpecTask(ctx, userCred, params, "")
}

func (self *SElasticcache) StartChangeSpecTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheChangeSpecTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformUpdateAuthMode(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "update_auth_mode")
}

func (self *SElasticcache) ValidatorUpdateAuthModeData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	region := self.GetRegion()
	if region == nil {
		return nil, fmt.Errorf("fail to found region for elastic cache")
	}

	driver := region.GetDriver()
	if driver == nil {
		return nil, fmt.Errorf("fail to found driver for elastic cache")
	}

	err := driver.AllowUpdateElasticcacheAuthMode(ctx, userCred, self.GetOwnerId(), self)
	if err != nil {
		return nil, err
	}

	authModeV := validators.NewStringChoicesValidator("auth_mode", choices.NewChoices("on", "off"))
	if err := authModeV.Optional(false).Validate(data.(*jsonutils.JSONDict)); err != nil {
		return nil, err
	}

	if authModeV.Value == self.AuthMode {
		return nil, httperrors.NewConflictError("auth mode aready in status %s", self.AuthMode)
	}

	return data, nil
}

func (self *SElasticcache) PerformUpdateAuthMode(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	data, err := self.ValidatorUpdateAuthModeData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	params := jsonutils.NewDict()
	authMode, _ := data.GetString("auth_mode")
	params.Set("auth_mode", jsonutils.NewString(authMode))
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_CHANGING, "")
	return nil, self.StartUpdateAuthModeTask(ctx, userCred, params, "")
}

func (self *SElasticcache) StartUpdateAuthModeTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheUpdateAuthModeTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformResetPassword(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "reset-password")
}

func (self *SElasticcache) ValidatorResetPasswordData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	if reset, _ := data.Bool("reset_password"); reset {
		if _, err := data.GetString("password"); err != nil {
			randomPasswd := seclib2.RandomPassword2(12)
			data.(*jsonutils.JSONDict).Set("password", jsonutils.NewString(randomPasswd))
		}
	}

	if password, err := data.GetString("password"); err != nil || len(password) == 0 {
		return nil, httperrors.NewMissingParameterError("password")
	} else {
		if !seclib2.MeetComplxity(password) {
			return nil, httperrors.NewWeakPasswordError()
		}
	}

	return data, nil
}

func (self *SElasticcache) PerformResetPassword(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	data, err := self.ValidatorResetPasswordData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_CHANGING, "")
	return nil, self.StartResetPasswordTask(ctx, userCred, data.(*jsonutils.JSONDict), "")
}

func (self *SElasticcache) GetAdminAccount() (*SElasticcacheAccount, error) {
	accounts, err := self.GetElasticcacheAccounts()
	if err != nil {
		return nil, err
	}

	for i := range accounts {
		if accounts[i].AccountType == api.ELASTIC_CACHE_ACCOUNT_TYPE_ADMIN {
			return &accounts[i], nil
		}
	}

	return nil, fmt.Errorf("no admin account found for elastic cache %s", self.Id)
}

func (self *SElasticcache) StartResetPasswordTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	account, err := self.GetAdminAccount()
	if err != nil {
		return err
	}

	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheAccountResetPasswordTask", account, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformSetMaintainTime(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "set_maintain_time")
}

func (self *SElasticcache) ValidatorSetMaintainTimeData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	timeReg, _ := regexp.Compile("^(0[0-9]|1[0-9]|2[0-3]|[0-9]):[0-5][0-9]Z$")
	startTimeV := validators.NewRegexpValidator("maintain_start_time", timeReg)
	endTimeV := validators.NewRegexpValidator("maintain_end_time", timeReg)
	keyV := map[string]validators.IValidator{
		"maintain_start_time": startTimeV.Optional(false),
		"maintain_end_time":   endTimeV.Optional(false),
	}

	for _, v := range keyV {
		if err := v.Validate(data.(*jsonutils.JSONDict)); err != nil {
			return nil, err
		}
	}

	if startTimeV.Value == self.MaintainStartTime && endTimeV.Value == self.MaintainEndTime {
		return nil, httperrors.NewInputParameterError("maintain time has no change")
	}

	return data, nil
}

func (self *SElasticcache) PerformSetMaintainTime(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	data, err := self.ValidatorSetMaintainTimeData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	params := jsonutils.NewDict()
	startTime, _ := data.GetString("maintain_start_time")
	endTime, _ := data.GetString("maintain_end_time")
	params.Set("maintain_start_time", jsonutils.NewString(startTime))
	params.Set("maintain_end_time", jsonutils.NewString(endTime))
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_CHANGING, "")
	return nil, self.StartSetMaintainTimeTask(ctx, userCred, params, "")
}

func (self *SElasticcache) StartSetMaintainTimeTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheSetMaintainTimeTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformAllocatePublicConnection(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "allocate_public_connection")
}

func (self *SElasticcache) ValidatorAllocatePublicConnectionData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	if self.PublicDNS != "" || self.PublicIpAddr != "" {
		return nil, httperrors.NewConflictError("public connection aready allocated")
	}

	portV := validators.NewRangeValidator("port", 1024, 65535)
	portV.Default(6379).Optional(true)
	if err := portV.Validate(data.(*jsonutils.JSONDict)); err != nil {
		return nil, err
	}

	return data, nil
}

func (self *SElasticcache) PerformAllocatePublicConnection(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	data, err := self.ValidatorAllocatePublicConnectionData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	params := jsonutils.NewDict()
	port, _ := data.Int("port")
	params.Set("port", jsonutils.NewInt(port))
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_NETWORKMODIFYING, "")
	return nil, self.StartAllocatePublicConnectionTask(ctx, userCred, params, "")
}

func (self *SElasticcache) StartAllocatePublicConnectionTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheAllocatePublicConnectionTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformReleasePublicConnection(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "release_public_connection")
}

func (self *SElasticcache) ValidatorReleasePublicConnectionData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	if self.PublicIpAddr == "" && self.PublicDNS == "" {
		return nil, httperrors.NewConflictError("release public connection aready released")
	}

	return data, nil
}

func (self *SElasticcache) PerformReleasePublicConnection(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	data, err := self.ValidatorReleasePublicConnectionData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_NETWORKMODIFYING, "")
	return nil, self.StartReleasePublicConnectionTask(ctx, userCred, jsonutils.NewDict(), "")
}

func (self *SElasticcache) StartReleasePublicConnectionTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheReleasePublicConnectionTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformFlushInstance(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "flush_instance")
}

func (self *SElasticcache) PerformFlushInstance(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_FLUSHING, "")
	return nil, self.StartFlushInstanceTask(ctx, userCred, jsonutils.NewDict(), "")
}

func (self *SElasticcache) StartFlushInstanceTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheFlushInstanceTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformUpdateInstanceParameters(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "update_instance_parameters")
}

func (self *SElasticcache) ValidatorUpdateInstanceParametersData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	parameters, err := data.Get("parameters")
	if err != nil {
		return nil, httperrors.NewMissingParameterError("parameters")
	}

	_, ok := parameters.(*jsonutils.JSONDict)
	if !ok {
		return nil, httperrors.NewInputParameterError("invalid parameter format. json dict required")
	}

	return data, nil
}

func (self *SElasticcache) PerformUpdateInstanceParameters(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	data, err := self.ValidatorUpdateInstanceParametersData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	params := jsonutils.NewDict()
	parameters, _ := data.Get("parameters")
	params.Set("parameters", parameters)
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_CHANGING, "")
	return nil, self.StartUpdateInstanceParametersTask(ctx, userCred, params, "")
}

func (self *SElasticcache) StartUpdateInstanceParametersTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheUpdateInstanceParametersTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformUpdateBackupPolicy(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "update_backup_policy")
}

func (self *SElasticcache) ValidatorUpdateBackupPolicyData(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	timeReg, _ := regexp.Compile("^(0[0-9]|1[0-9]|2[0-3]|[0-9]):[0-5][0-9]Z-(0[0-9]|1[0-9]|2[0-3]|[0-9]):[0-5][0-9]Z$")
	backupTypeV := validators.NewStringChoicesValidator("backup_type", choices.NewChoices(api.BACKUP_MODE_AUTOMATED, api.ELASTIC_CACHE_BACKUP_MODE_MANUAL))
	BackupReservedDaysV := validators.NewRangeValidator("backup_reserved_days", 1, 7).Default(7)
	PreferredBackupPeriodV := validators.NewStringChoicesValidator("preferred_backup_period", choices.NewChoices("Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"))
	PreferredBackupTimeV := validators.NewRegexpValidator("preferred_backup_time", timeReg)

	keyV := map[string]validators.IValidator{
		"backup_type":             backupTypeV.Optional(true),
		"backup_reserved_days":    BackupReservedDaysV.Optional(true),
		"preferred_backup_period": PreferredBackupPeriodV.Optional(false),
		"preferred_backup_time":   PreferredBackupTimeV.Optional(false),
	}

	for _, v := range keyV {
		if err := v.Validate(data.(*jsonutils.JSONDict)); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (self *SElasticcache) PerformUpdateBackupPolicy(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	data, err := self.ValidatorUpdateBackupPolicyData(ctx, userCred, query, data)
	if err != nil {
		return nil, err
	}

	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_CHANGING, "")
	return nil, self.StartUpdateBackupPolicyTask(ctx, userCred, data.(*jsonutils.JSONDict), "")
}

func (self *SElasticcache) StartUpdateBackupPolicyTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheUpdateBackupPolicyTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

func (self *SElasticcache) AllowPerformSyncstatus(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "syncstatus")
}

// 同步弹性缓存状态
func (self *SElasticcache) PerformSyncstatus(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, input api.ElasticcacheSyncstatusInput) (jsonutils.JSONObject, error) {
	var openTask = true
	count, err := taskman.TaskManager.QueryTasksOfObject(self, time.Now().Add(-3*time.Minute), &openTask).CountWithError()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, httperrors.NewBadRequestError("Elasticcache has %d task active, can't sync status", count)
	}

	return nil, StartResourceSyncStatusTask(ctx, userCred, self, "ElasticcacheSyncstatusTask", "")
}

func (self *SElasticcache) AllowPerformSync(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "sync")
}

func (self *SElasticcache) PerformSync(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	self.SetStatus(userCred, api.ELASTIC_CACHE_STATUS_SYNCING, "")
	return nil, self.StartSyncTask(ctx, userCred, data.(*jsonutils.JSONDict), "")
}

func (self *SElasticcache) StartSyncTask(ctx context.Context, userCred mcclient.TokenCredential, params *jsonutils.JSONDict, parentTaskId string) error {
	task, err := taskman.TaskManager.NewTask(ctx, "ElasticcacheSyncTask", self, userCred, params, parentTaskId, "", nil)
	if err != nil {
		return err
	}

	task.ScheduleRun(nil)
	return nil
}

// 清理所有关联资源记录
func (self *SElasticcache) DeleteSubResources(ctx context.Context, userCred mcclient.TokenCredential) {
	ms := []db.IResourceModelManager{
		ElasticcacheAccountManager,
		ElasticcacheAclManager,
		ElasticcacheBackupManager,
		ElasticcacheParameterManager,
	}

	ownerId := self.GetOwnerId()
	for _, m := range ms {
		func(man db.IResourceModelManager) {
			lockman.LockClass(ctx, man, db.GetLockClassKey(man, ownerId))
			defer lockman.ReleaseClass(ctx, man, db.GetLockClassKey(man, ownerId))
			q := man.Query().IsFalse("deleted").Equals("elasticcache_id", self.GetId())

			models := make([]interface{}, 0)
			err := db.FetchModelObjects(man, q, &models)
			if err != nil {
				log.Errorf("elasticcache.DeleteSubResources.FetchModelObjects %s", err)
			}

			for i := range models {
				var imodel db.IModel
				switch models[i].(type) {
				case SElasticcacheAccount:
					_m := models[i].(SElasticcacheAccount)
					imodel = &_m
				case SElasticcacheAcl:
					_m := models[i].(SElasticcacheAcl)
					imodel = &_m
				case SElasticcacheBackup:
					_m := models[i].(SElasticcacheBackup)
					imodel = &_m
				case SElasticcacheParameter:
					_m := models[i].(SElasticcacheParameter)
					imodel = &_m
				default:
					log.Errorf("elasticcache.DeleteSubResources.UnknownModelType %s", models[i])
				}

				err = db.DeleteModel(ctx, userCred, imodel)
				if err != nil {
					log.Errorf("elasticcache.DeleteSubResources.DeleteModel %s", err)
				}
			}
		}(m)
	}
}

func (man *SElasticcacheManager) TotalCount(
	scope rbacutils.TRbacScope,
	ownerId mcclient.IIdentityProvider,
	rangeObjs []db.IStandaloneModel,
	providers []string, brands []string, cloudEnv string,
) (int, error) {
	q := man.Query()
	vpcs := VpcManager.Query().SubQuery()
	q = q.Join(vpcs, sqlchemy.Equals(q.Field("vpc_id"), vpcs.Field("id")))
	q = scopeOwnerIdFilter(q, scope, ownerId)
	q = CloudProviderFilter(q, vpcs.Field("manager_id"), providers, brands, cloudEnv)
	q = RangeObjectsFilter(q, rangeObjs, vpcs.Field("cloudregion_id"), nil, vpcs.Field("manager_id"), nil, nil)
	return q.CountWithError()
}

func (cache *SElasticcache) GetQuotaKeys() quotas.IQuotaKeys {
	return fetchRegionalQuotaKeys(
		rbacutils.ScopeProject,
		cache.GetOwnerId(),
		cache.GetRegion(),
		cache.GetCloudprovider(),
	)
}

func (cache *SElasticcache) GetUsages() []db.IUsage {
	if cache.PendingDeleted || cache.Deleted {
		return nil
	}
	usage := SRegionQuota{Cache: 1}
	keys := cache.GetQuotaKeys()
	usage.SetKeys(keys)
	return []db.IUsage{
		&usage,
	}
}

func (manager *SElasticcacheManager) ListItemExportKeys(ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	keys stringutils2.SSortedStrings,
) (*sqlchemy.SQuery, error) {
	var err error

	q, err = manager.SVirtualResourceBaseManager.ListItemExportKeys(ctx, q, userCred, keys)
	if err != nil {
		return nil, errors.Wrap(err, "SVirtualResourceBaseManager.ListItemExportKeys")
	}
	if keys.ContainsAny(manager.SVpcResourceBaseManager.GetExportKeys()...) {
		q, err = manager.SVpcResourceBaseManager.ListItemExportKeys(ctx, q, userCred, keys)
		if err != nil {
			return nil, errors.Wrap(err, "SVpcResourceBaseManager.ListItemExportKeys")
		}
	}
	if keys.Contains("zone") {
		q, err = manager.SZoneResourceBaseManager.ListItemExportKeys(ctx, q, userCred, stringutils2.NewSortedStrings([]string{"zone"}))
		if err != nil {
			return nil, errors.Wrap(err, "SZoneResourceBaseManager.ListItemExportKeys")
		}
	}
	if keys.ContainsAny("network", "wire") {
		q, err = manager.SNetworkResourceBaseManager.ListItemExportKeys(ctx, q, userCred, stringutils2.NewSortedStrings([]string{"network", "wire"}))
		if err != nil {
			return nil, errors.Wrap(err, "SNetworkResourceBaseManager.ListItemExportKeys")
		}
	}
	return q, nil
}

func (manager *SElasticcacheManager) getExpiredPostpaids() []SElasticcache {
	q := ListExpiredPostpaidResources(manager.Query(), options.Options.ExpiredPrepaidMaxCleanBatchSize)
	q = q.IsFalse("pending_deleted")

	ecs := make([]SElasticcache, 0)
	err := db.FetchModelObjects(ElasticcacheManager, q, &ecs)
	if err != nil {
		log.Errorf("fetch elasitc cache instances error %s", err)
		return nil
	}

	return ecs
}

func (cache *SElasticcache) SaveRenewInfo(
	ctx context.Context, userCred mcclient.TokenCredential,
	bcycle *bc.SBillingCycle, expireAt *time.Time, billingType string,
) error {
	_, err := db.Update(cache, func() error {
		if billingType == "" {
			billingType = billing_api.BILLING_TYPE_PREPAID
		}
		if cache.BillingType == "" {
			cache.BillingType = billingType
		}
		if expireAt != nil && !expireAt.IsZero() {
			cache.ExpiredAt = *expireAt
		} else if bcycle != nil {
			cache.BillingCycle = bcycle.String()
			cache.ExpiredAt = bcycle.EndAt(cache.ExpiredAt)
		}
		return nil
	})
	if err != nil {
		log.Errorf("Update error %s", err)
		return err
	}
	db.OpsLog.LogEvent(cache, db.ACT_RENEW, cache.GetShortDesc(ctx), userCred)
	return nil
}

func (cache *SElasticcache) SetDisableDelete(userCred mcclient.TokenCredential, val bool) error {
	diff, err := db.Update(cache, func() error {
		if val {
			cache.DisableDelete = tristate.True
		} else {
			cache.DisableDelete = tristate.False
		}
		return nil
	})
	if err != nil {
		return err
	}
	db.OpsLog.LogEvent(cache, db.ACT_UPDATE, diff, userCred)
	logclient.AddSimpleActionLog(cache, logclient.ACT_UPDATE, diff, userCred, true)
	return err
}

func (self *SElasticcache) doExternalSync(ctx context.Context, userCred mcclient.TokenCredential) error {
	provider := self.GetCloudprovider()
	if provider != nil {
		return fmt.Errorf("no cloud provider???")
	}

	iregion, err := self.GetIRegion()
	if err != nil || iregion == nil {
		return fmt.Errorf("no cloud region??? %s", err)
	}

	iecs, err := iregion.GetIElasticcacheById(self.ExternalId)
	if err != nil {
		return err
	}
	return self.SyncWithCloudElasticcache(ctx, userCred, provider, iecs)
}

func (model *SElasticcacheManager) Delete(ctx context.Context, userCred mcclient.TokenCredential) error {
	log.Infof("elasticcache delete do nothing. Do delete in StartDeleteElasticcacheTask")
	return nil
}

func (manager *SElasticcacheManager) DeleteExpiredPostpaids(ctx context.Context, userCred mcclient.TokenCredential, isStart bool) {
	ecs := manager.getExpiredPostpaids()
	if ecs == nil {
		return
	}
	for i := 0; i < len(ecs); i += 1 {
		if len(ecs[i].ExternalId) > 0 {
			err := ecs[i].doExternalSync(ctx, userCred)
			if err == nil && ecs[i].IsValidPrePaid() {
				continue
			}
		}
		ecs[i].SetDisableDelete(userCred, false)
		ecs[i].StartDeleteElasticcacheTask(ctx, userCred, jsonutils.NewDict(), "")
	}
}

func (self *SElasticcache) AllowPerformPostpaidExpire(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "postpaid-expire")
}

func (self *SElasticcache) PerformPostpaidExpire(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	if self.BillingType != billing_api.BILLING_TYPE_POSTPAID {
		return nil, httperrors.NewBadRequestError("elasticcache billing type is %s", self.BillingType)
	}

	bc, err := ParseBillingCycleInput(&self.SBillingResourceBase, data)
	if err != nil {
		return nil, err
	}

	err = self.SaveRenewInfo(ctx, userCred, bc, nil, billing_api.BILLING_TYPE_POSTPAID)
	return nil, err
}

func (self *SElasticcache) AllowPerformCancelExpire(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "cancel-expire")
}

func (self *SElasticcache) PerformCancelExpire(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	if err := self.CancelExpireTime(ctx, userCred); err != nil {
		return nil, err
	}

	return nil, nil
}

func (self *SElasticcache) CancelExpireTime(ctx context.Context, userCred mcclient.TokenCredential) error {
	if self.BillingType != billing_api.BILLING_TYPE_POSTPAID {
		return httperrors.NewBadRequestError("elasticcache billing type %s not support cancel expire", self.BillingType)
	}

	_, err := sqlchemy.GetDB().Exec(
		fmt.Sprintf(
			"update %s set expired_at = NULL and billing_cycle = NULL where id = ?",
			ElasticcacheManager.TableSpec().Name(),
		), self.Id,
	)
	if err != nil {
		return errors.Wrap(err, "elasticcache cancel expire time")
	}
	db.OpsLog.LogEvent(self, db.ACT_RENEW, "elasticcache cancel expire time", userCred)
	return nil
}
