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

package service

import (
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"yunion.io/x/log"

	api "yunion.io/x/onecloud/pkg/apis/notify"
	"yunion.io/x/onecloud/pkg/cloudcommon"
	"yunion.io/x/onecloud/pkg/cloudcommon/app"
	"yunion.io/x/onecloud/pkg/cloudcommon/cronman"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	common_options "yunion.io/x/onecloud/pkg/cloudcommon/options"
	"yunion.io/x/onecloud/pkg/notify/models"
	"yunion.io/x/onecloud/pkg/notify/options"
	_ "yunion.io/x/onecloud/pkg/notify/policy"
	"yunion.io/x/onecloud/pkg/notify/rpc"
	_ "yunion.io/x/onecloud/pkg/notify/tasks"
)

func StartService() {
	// parse options
	opts := &options.Options
	commonOpts := &options.Options.CommonOptions
	dbOpts := &options.Options.DBOptions
	baseOpts := &options.Options.BaseOptions
	common_options.ParseOptions(opts, os.Args, "notify.conf", api.SERVICE_TYPE)

	// init auth
	app.InitAuth(commonOpts, func() {
		log.Infof("Auth complete!")
	})

	common_options.StartOptionManager(opts, opts.ConfigSyncPeriodSeconds, api.SERVICE_TYPE, api.SERVICE_VERSION, options.OnOptionsChange)

	// init handler
	applicaion := app.InitApp(baseOpts, true)
	InitHandlers(applicaion)

	// init database
	db.EnsureAppInitSyncDB(applicaion, dbOpts, models.InitDB)
	defer cloudcommon.CloseDB()

	err := models.ReceiverManager.StartWatchUserInKeystone()
	if err != nil {
		log.Logger().Panic(err.Error())
	}

	err = db.RoleCacheManager.StartWatchRoleInKeystone()
	if err != nil {
		log.Fatalf("unable to StartWatchRoleInKeystone: %v", err)
	}

	// init notify service
	models.NotifyService = rpc.NewSRpcService(opts.SocketFileDir, models.ConfigManager, models.TemplateManager)
	models.NotifyService.InitAll()
	defer models.NotifyService.StopAll()

	cron := cronman.InitCronJobManager(true, 2)
	// update service
	cron.AddJobAtIntervals("UpdateServices", time.Duration(opts.UpdateInterval)*time.Minute, models.NotifyService.UpdateServices)

	// wrapped func to resend notifications
	cron.AddJobAtIntervals("ReSendNotifications", time.Duration(opts.ReSendScope)*time.Second, models.NotificationManager.ReSend)
	cron.Start()

	app.ServeForever(applicaion, baseOpts)
}
