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

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/tristate"
	"yunion.io/x/pkg/util/compare"
	"yunion.io/x/sqlchemy"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/lockman"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/options"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/bitmap"
	"yunion.io/x/onecloud/pkg/util/validate"
)

type SSnapshotPolicyManager struct {
	db.SVirtualResourceBaseManager
}

type SSnapshotPolicy struct {
	db.SVirtualResourceBase
	//db.SExternalizedResourceBase
	//
	//SManagedResourceBase
	//SCloudregionResourceBase

	RetentionDays int `nullable:"false" list:"user" get:"user" create:"required"`

	// 0~6, 0 is Monday
	RepeatWeekdays uint8 `charset:"utf8" create:"required"`
	// 0~23
	TimePoints  uint32            `charset:"utf8" create:"required"`
	IsActivated tristate.TriState `list:"user" get:"user" create:"optional" default:"true"`
}

var SnapshotPolicyManager *SSnapshotPolicyManager

func init() {
	SnapshotPolicyManager = &SSnapshotPolicyManager{
		SVirtualResourceBaseManager: db.NewVirtualResourceBaseManager(
			SSnapshotPolicy{},
			"snapshotpolicies_tbl",
			"snapshotpolicy.yaml",
			"snapshotpolicies",
		),
	}
	SnapshotPolicyManager.SetVirtualObject(SnapshotPolicyManager)
}

func (manager *SSnapshotPolicyManager) ValidateListConditions(ctx context.Context, userCred mcclient.TokenCredential,
	query *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {

	input := &api.SSnapshotPolicyCreateInput{}
	err := query.Unmarshal(input)
	if err != nil {
		return nil, httperrors.NewInputParameterError("Unmarshal input failed %s", err)
	}
	if query.Contains("repeat_weekdays") {
		query.Remove("repeat_weekdays")
		query.Add(jsonutils.NewInt(int64(manager.RepeatWeekdaysParseIntArray(input.RepeatWeekdays))), "repeat_weekdays")
	}
	if query.Contains("time_points") {
		query.Remove("time_points")
		query.Add(jsonutils.NewInt(int64(manager.RepeatWeekdaysParseIntArray(input.RepeatWeekdays))), "time_points")
	}
	return query, nil
}

func (sp *SSnapshotPolicy) AllowUpdateItem(ctx context.Context, userCred mcclient.TokenCredential) bool {
	return false
}

// ==================================================== fetch ==========================================================
func (manager *SSnapshotPolicyManager) GetSnapshotPoliciesAt(week, timePoint uint32) ([]string, error) {

	q := manager.Query("id")
	q = q.Filter(sqlchemy.Equals(sqlchemy.AND_Val("", q.Field("repeat_weekdays"), 1<<week), 1<<week))
	q = q.Filter(sqlchemy.Equals(sqlchemy.AND_Val("", q.Field("time_points"), 1<<timePoint), 1<<timePoint))
	q = q.Equals("is_activated", true)
	q.DebugQuery()

	sps := make([]SSnapshotPolicy, 0)
	err := q.All(&sps)
	if err != nil {
		return nil, err
	}
	if len(sps) > 0 {
		ret := make([]string, len(sps))
		for i := 0; i < len(sps); i++ {
			ret[i] = sps[i].Id
		}
		return ret, nil
	}
	return nil, nil
}

func (manager *SSnapshotPolicyManager) FetchSnapshotPolicyById(spId string) *SSnapshotPolicy {
	sp, err := manager.FetchById(spId)
	if err != nil {
		log.Errorf("FetchBId fail %s", err)
		return nil
	}
	return sp.(*SSnapshotPolicy)
}

func (manager *SSnapshotPolicyManager) FetchAllByIds(spIds []string) ([]SSnapshotPolicy, error) {
	if spIds == nil || len(spIds) == 0 {
		return []SSnapshotPolicy{}, nil
	}
	q := manager.Query().In("id", spIds)
	sps := make([]SSnapshotPolicy, 0, 1)
	err := db.FetchModelObjects(manager, q, &sps)
	if err != nil {
		return nil, err
	}
	return sps, nil
}

// ==================================================== create =========================================================

func (manager *SSnapshotPolicyManager) ValidateCreateData(ctx context.Context, userCred mcclient.TokenCredential,
	ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject, data *jsonutils.JSONDict) (*jsonutils.JSONDict, error) {

	input := &api.SSnapshotPolicyCreateInput{}
	err := data.Unmarshal(input)
	if err != nil {
		return nil, httperrors.NewInputParameterError("Unmarshal input failed %s", err)
	}
	input.ProjectId = ownerId.GetProjectId()
	input.DomainId = ownerId.GetProjectDomainId()

	err = db.NewNameValidator(manager, ownerId, input.Name)
	if err != nil {
		return nil, err
	}

	if input.RetentionDays < -1 || input.RetentionDays == 0 || input.RetentionDays > options.Options.RetentionDaysLimit {
		return nil, httperrors.NewInputParameterError("Retention days must in 1~%d or -1", options.Options.RetentionDaysLimit)
	}

	if len(input.RepeatWeekdays) == 0 {
		return nil, httperrors.NewMissingParameterError("repeat_weekdays")
	}

	if len(input.RepeatWeekdays) > options.Options.RepeatWeekdaysLimit {
		return nil, httperrors.NewInputParameterError("repeat_weekdays only contains %d days at most",
			options.Options.RepeatWeekdaysLimit)
	}
	input.RepeatWeekdays, err = validate.DaysCheck(input.RepeatWeekdays, 1, 7)
	if err != nil {
		return nil, httperrors.NewInputParameterError(err.Error())
	}

	if len(input.TimePoints) == 0 {
		return nil, httperrors.NewMissingParameterError("time_points")
	}
	if len(input.TimePoints) > options.Options.TimePointsLimit {
		return nil, httperrors.NewInputParameterError("time_points only contains %d points at most", options.Options.TimePointsLimit)
	}
	input.TimePoints, err = validate.DaysCheck(input.TimePoints, 0, 23)
	if err != nil {
		return nil, httperrors.NewInputParameterError(err.Error())
	}

	internalInput := manager.sSnapshotPolicyCreateInputToInternal(input)
	data = internalInput.JSON(internalInput)
	return data, nil
}

// ==================================================== update =========================================================

func (self *SSnapshotPolicy) AllowPerformUpdate(ctx context.Context, userCred mcclient.TokenCredential,
	query jsonutils.JSONObject, data jsonutils.JSONObject) bool {
	// no fo now
	return false
	//return self.IsOwner(userCred) || db.IsAdminAllowPerform(userCred, self, "update")
}

func (self *SSnapshotPolicy) PerformUpdate(ctx context.Context, userCred mcclient.TokenCredential,
	query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	//check param
	input := &api.SSnapshotPolicyCreateInput{}
	err := data.Unmarshal(input)
	if err != nil {
		return nil, httperrors.NewInputParameterError("Unmarshel input failed %s", err)
	}
	err = self.UpdateParamCheck(input)
	if err != nil {
		return nil, err
	}
	return nil, self.StartSnapshotPolicyUpdateTask(ctx, userCred, input)
}

func (self *SSnapshotPolicy) StartSnapshotPolicyUpdateTask(ctx context.Context, userCred mcclient.TokenCredential,
	input *api.SSnapshotPolicyCreateInput) error {

	params := jsonutils.NewDict()
	params.Add(jsonutils.Marshal(input), "input")
	self.SetStatus(userCred, api.SNAPSHOT_POLICY_UPDATING, "")
	if task, err := taskman.TaskManager.NewTask(ctx, "SnapshotPolicyUpdateTask", self, userCred, params,
		"", "", nil); err == nil {
		return err
	} else {
		task.ScheduleRun(nil)
	}
	return nil
}

// UpdateParamCheck check if update parameters are correct and need to update
func (self *SSnapshotPolicy) UpdateParamCheck(input *api.SSnapshotPolicyCreateInput) error {
	var err error
	updateNum := 0

	if input.RetentionDays != 0 {
		if input.RetentionDays < -1 || input.RetentionDays > 65535 {
			return httperrors.NewInputParameterError("Retention days must in 1~65535 or -1")
		}
		if input.RetentionDays != self.RetentionDays {
			updateNum++
		}
	}

	if input.RepeatWeekdays != nil && len(input.RepeatWeekdays) != 0 {
		input.RepeatWeekdays, err = validate.DaysCheck(input.RepeatWeekdays, 1, 7)
		if err != nil {
			return httperrors.NewInputParameterError(err.Error())
		}
		if self.RepeatWeekdays != SnapshotPolicyManager.RepeatWeekdaysParseIntArray(input.RepeatWeekdays) {
			updateNum++
		}
	}

	if input.TimePoints != nil && len(input.TimePoints) != 0 {
		input.TimePoints, err = validate.DaysCheck(input.TimePoints, 0, 23)
		if err != nil {
			return httperrors.NewInputParameterError(err.Error())
		}
		if self.TimePoints != SnapshotPolicyManager.TimePointsParseIntArray(input.TimePoints) {
			updateNum++
		}
	}

	if updateNum == 0 {
		return httperrors.NewInputParameterError("Do not need to update")
	}
	return nil
}

// ==================================================== delete =========================================================

func (self *SSnapshotPolicy) DetachAfterDelete(ctx context.Context, userCred mcclient.TokenCredential) error {
	err := SnapshotPolicyDiskManager.SyncDetachBySnapshotpolicy(ctx, userCred, nil, self)
	if err != nil {
		return errors.Wrap(err, "detach after delete failed")
	}
	return nil
}

func (self *SSnapshotPolicy) CustomizeDelete(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.
	JSONObject, data jsonutils.JSONObject) error {

	// check if self bind to some disks
	sds, err := SnapshotPolicyDiskManager.FetchAllBySnapshotpolicyID(ctx, userCred, self.GetId())
	if err != nil {
		return errors.Wrap(err, "fetch bind info failed")
	}
	if len(sds) != 0 {
		return httperrors.NewBadRequestError("Couldn't delete snapshot policy binding to disks")
	}
	self.SetStatus(userCred, api.SNAPSHOT_POLICY_DELETING, "")
	return self.StartSnapshotPolicyDeleteTask(ctx, userCred, jsonutils.NewDict(), "")
}

func (self *SSnapshotPolicy) StartSnapshotPolicyDeleteTask(ctx context.Context, userCred mcclient.TokenCredential,
	params *jsonutils.JSONDict, parentTaskId string) error {

	task, err := taskman.TaskManager.NewTask(ctx, "SnapshotPolicyDeleteTask", self, userCred, params,
		parentTaskId, "", nil)
	if err != nil {
		return err
	}
	task.ScheduleRun(nil)
	return nil
}

func (self *SSnapshotPolicy) GetCustomizeColumns(ctx context.Context, userCred mcclient.TokenCredential,
	query jsonutils.JSONObject) *jsonutils.JSONDict {

	ret := jsonutils.NewDict()
	// more
	weekdays := SnapshotPolicyManager.RepeatWeekdaysToIntArray(self.RepeatWeekdays)
	timePoints := SnapshotPolicyManager.TimePointsToIntArray(self.TimePoints)
	ret.Add(jsonutils.Marshal(weekdays), "repeat_weekdays")
	ret.Add(jsonutils.Marshal(timePoints), "time_points")
	return ret
}

func (self *SSnapshotPolicy) GetExtraDetails(ctx context.Context, userCred mcclient.TokenCredential,
	query jsonutils.JSONObject) (*jsonutils.JSONDict, error) {

	ret := jsonutils.NewDict()
	// more
	weekdays := SnapshotPolicyManager.RepeatWeekdaysToIntArray(self.RepeatWeekdays)
	timePoints := SnapshotPolicyManager.TimePointsToIntArray(self.TimePoints)
	ret.Add(jsonutils.Marshal(weekdays), "repeat_weekdays")
	ret.Add(jsonutils.Marshal(timePoints), "time_points")
	return ret, nil
}

// ==================================================== sync ===========================================================
func (manager *SSnapshotPolicyManager) SyncSnapshotPolicies(ctx context.Context, userCred mcclient.TokenCredential,
	provider *SCloudprovider, region *SCloudregion, cloudSPs []cloudprovider.ICloudSnapshotPolicy,
	syncOwnerId mcclient.IIdentityProvider) compare.SyncResult {

	lockman.LockClass(ctx, manager, db.GetLockClassKey(manager, syncOwnerId))
	defer lockman.ReleaseClass(ctx, manager, db.GetLockClassKey(manager, syncOwnerId))
	syncResult := compare.SyncResult{}

	// Fetch all snapshotpolicy.yaml caches
	spCaches, err := SnapshotPolicyCacheManager.FetchAllByRegionProvider(region.GetId(), provider.GetId())
	if err != nil {
		syncResult.Error(err)
		return syncResult
	}
	spIdSet, spIds := make(map[string]struct{}), make([]string, 0, 2)
	for _, spCache := range spCaches {
		if _, ok := spIdSet[spCache.SnapshotpolicyId]; !ok {
			spIds = append(spIds, spCache.SnapshotpolicyId)
			spIdSet[spCache.SnapshotpolicyId] = struct{}{}
		}
	}

	// Fetch all snapshotpolicy.yaml of caches above
	snapshotPolicies, err := manager.FetchAllByIds(spIds)
	if err != nil {
		syncResult.Error(err)
		return syncResult
	}

	// structure two sets (externalID, snapshotpolicyCache), (snapshotPolicyID, snapshotPolicy)
	spSet, spCacheSet := make(map[string]*SSnapshotPolicy), make(map[string]*SSnapshotPolicyCache)
	for i := range snapshotPolicies {
		spSet[snapshotPolicies[i].GetId()] = &snapshotPolicies[i]
	}
	for i := range spCaches {
		externalId := spCaches[i].ExternalId
		if len(externalId) != 0 {
			spCacheSet[spCaches[i].ExternalId] = &spCaches[i]
		}
	}

	// start sync
	// add for snapshotpolicy.yaml and cache
	// delete for snapshotpolicy.yaml cache
	added := make([]cloudprovider.ICloudSnapshotPolicy, 0, 1)
	removed := make([]*SSnapshotPolicyCache, 0, 1)
	for _, cloudSP := range cloudSPs {
		spCache, ok := spCacheSet[cloudSP.GetGlobalId()]
		if !ok {
			added = append(added, cloudSP)
			continue
		}
		snapshotPolicy := spSet[spCache.SnapshotpolicyId]
		if !snapshotPolicy.Equals(cloudSP) {
			removed = append(removed, spCache)
			added = append(added, cloudSP)
		}
	}

	for i := range removed {
		// change snapshotpolicy.yaml cache
		err := removed[i].RealDetele(ctx, userCred)
		if err != nil {
			syncResult.DeleteError(err)
		}
	}

	for i := range added {
		locol, err := manager.newFromCloudSnapshotPolicy(ctx, userCred, added[i], region, syncOwnerId, provider)
		if err != nil {
			syncResult.AddError(err)
		} else {
			syncMetadata(ctx, userCred, locol, added[i])
			syncResult.Add()
		}
	}
	return compare.SyncResult{}
}

func (manager *SSnapshotPolicyManager) newFromCloudSnapshotPolicy(
	ctx context.Context, userCred mcclient.TokenCredential,
	ext cloudprovider.ICloudSnapshotPolicy, region *SCloudregion,
	syncOwnerId mcclient.IIdentityProvider, provider *SCloudprovider,
) (*SSnapshotPolicy, error) {

	snapshotPolicy := SSnapshotPolicy{}
	snapshotPolicy.SetModelManager(manager, &snapshotPolicy)

	newName, err := db.GenerateName(manager, syncOwnerId, ext.GetName())
	if err != nil {
		return nil, err
	}
	snapshotPolicy.Name = newName
	snapshotPolicy.Status = ext.GetStatus()
	snapshotPolicy.RetentionDays = ext.GetRetentionDays()
	arw, err := ext.GetRepeatWeekdays()
	if err != nil {
		return nil, err
	}
	snapshotPolicy.RepeatWeekdays = SnapshotPolicyManager.RepeatWeekdaysParseIntArray(arw)
	atp, err := ext.GetTimePoints()
	if err != nil {
		return nil, err
	}
	snapshotPolicy.TimePoints = SnapshotPolicyManager.TimePointsParseIntArray(atp)

	snapshotPolicy.IsActivated = tristate.NewFromBool(ext.IsActivated())

	err = manager.TableSpec().Insert(&snapshotPolicy)
	if err != nil {
		log.Errorf("newFromCloudEip fail %s", err)
		return nil, err
	}

	// add cache
	_, err = SnapshotPolicyCacheManager.NewCacheWithExternalId(ctx, userCred, snapshotPolicy.GetId(),
		ext.GetGlobalId(), region.GetId(), provider.GetId())
	if err != nil {
		// snapshotpolicy.yaml has been exist so that created is successful although cache created is fail.
		// disk will be sync after snapshotpolicy.yaml sync, cache must be right so that this sync is fail
		log.Errorf("snapshotpolicy.yaml %s created successfully but corresponding cache created fail", snapshotPolicy.GetId())
		return nil, errors.Wrapf(err, "snapshotpolicy.yaml %s created successfully but corresponding cache created fail",
			snapshotPolicy.GetId())
	}

	SyncCloudProject(userCred, &snapshotPolicy, syncOwnerId, ext, provider.GetId())

	db.OpsLog.LogEvent(&snapshotPolicy, db.ACT_CREATE, snapshotPolicy.GetShortDesc(ctx), userCred)
	return &snapshotPolicy, nil
}

func (self *SSnapshotPolicy) Equals(cloudSP cloudprovider.ICloudSnapshotPolicy) bool {
	rws, err := cloudSP.GetRepeatWeekdays()
	if err != nil {
		return false
	}
	tps, err := cloudSP.GetTimePoints()
	if err != nil {
		return false
	}
	repeatWeekdays := SnapshotPolicyManager.RepeatWeekdaysParseIntArray(rws)
	timePoints := SnapshotPolicyManager.TimePointsParseIntArray(tps)

	return self.RetentionDays == cloudSP.GetRetentionDays() && self.RepeatWeekdays == repeatWeekdays && self.
		TimePoints == timePoints && self.IsActivated == self.IsActivated
}

func (manager *SSnapshotPolicyManager) getProviderSnapshotPolicies(region *SCloudregion, provider *SCloudprovider) ([]SSnapshotPolicy, error) {
	if region == nil && provider == nil {
		return nil, fmt.Errorf("Region is nil or provider is nil")
	}
	snapshotPolicies := make([]SSnapshotPolicy, 0)
	q := manager.Query().Equals("cloudregion_id", region.Id).Equals("manager_id", provider.Id)
	err := db.FetchModelObjects(manager, q, &snapshotPolicies)
	if err != nil {
		return nil, err
	}
	return snapshotPolicies, nil
}

func (self *SSnapshotPolicy) syncRemoveCloudSnapshot(ctx context.Context, userCred mcclient.TokenCredential) error {
	lockman.LockObject(ctx, self)
	defer lockman.ReleaseObject(ctx, self)

	err := self.ValidateDeleteCondition(ctx)
	if err != nil {
		err = self.SetStatus(userCred, api.SNAPSHOT_POLICY_UNKNOWN, "sync to delete")
	} else {
		err = self.RealDelete(ctx, userCred)
	}
	return err
}

func (self *SSnapshotPolicy) Delete(ctx context.Context, userCred mcclient.TokenCredential) error {
	return nil
}

func (self *SSnapshotPolicy) RealDelete(ctx context.Context, userCred mcclient.TokenCredential) error {
	return db.DeleteModel(ctx, userCred, self)
}

// ==================================================== utils ==========================================================

func (manager *SSnapshotPolicyManager) sSnapshotPolicyCreateInputToInternal(input *api.SSnapshotPolicyCreateInput,
) *api.SSnapshotPolicyCreateInternalInput {
	ret := api.SSnapshotPolicyCreateInternalInput{
		Meta:          input.Meta,
		Name:          input.Name,
		ProjectId:     input.ProjectId,
		DomainId:      input.DomainId,
		RetentionDays: input.RetentionDays,
	}

	ret.RepeatWeekdays = manager.RepeatWeekdaysParseIntArray(input.RepeatWeekdays)
	ret.TimePoints = manager.TimePointsParseIntArray(input.TimePoints)
	return &ret
}

func (manager *SSnapshotPolicyManager) sSnapshotPolicyCreateInputFromInternal(input *api.
	SSnapshotPolicyCreateInternalInput) *api.SSnapshotPolicyCreateInput {
	return nil
}

func (self *SSnapshotPolicyManager) RepeatWeekdaysParseIntArray(nums []int) uint8 {
	return uint8(bitmap.IntArray2Uint(nums))
}

func (self *SSnapshotPolicyManager) RepeatWeekdaysToIntArray(n uint8) []int {
	return bitmap.Uint2IntArray(uint32(n))
}

func (self *SSnapshotPolicyManager) TimePointsParseIntArray(nums []int) uint32 {
	return bitmap.IntArray2Uint(nums)
}

func (self *SSnapshotPolicyManager) TimePointsToIntArray(n uint32) []int {
	return bitmap.Uint2IntArray(n)
}

func (self *SSnapshotPolicy) GenerateCreateSpParams() *cloudprovider.SnapshotPolicyInput {
	intWeekdays := SnapshotPolicyManager.RepeatWeekdaysToIntArray(self.RepeatWeekdays)
	intTimePoints := SnapshotPolicyManager.TimePointsToIntArray(self.TimePoints)

	return &cloudprovider.SnapshotPolicyInput{
		RetentionDays:  self.RetentionDays,
		RepeatWeekdays: intWeekdays,
		TimePoints:     intTimePoints,
		PolicyName:     self.Name,
	}
}
