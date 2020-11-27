// Copyright 2019 Yunion
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

package options

import "yunion.io/x/jsonutils"

type SubscriptionListOptions struct {
	BaseListOptions
}

func (opts *SubscriptionListOptions) Params() (jsonutils.JSONObject, error) {
	return ListStructToParams(opts)
}

type SubscriptionSetReceiverOptions struct {
	ReceiverRoles     []string `help:"role nam or id"`
	ReceiverRoleScope string   `help:"the scope of role"`
	Receivers         []string `help:"receivers"`
	ID                string   `help:"id of subscription"`
}

func (opts *SubscriptionSetReceiverOptions) Params() (jsonutils.JSONObject, error) {
	return ListStructToParams(opts)
}

type SubscriptionSetRobotOptions struct {
	ID    string `help:"id of subscription"`
	ROBOT string `choices:"feishu-robot|dingtalk-robot|workwx-robot"`
}

func (opts *SubscriptionSetRobotOptions) Params() (jsonutils.JSONObject, error) {
	return ListStructToParams(opts)
}

type SubscriptionSetWebhookOptions struct {
	ID      string `help:"id of subscription"`
	WEBHOOK string `choices:"webhook"`
}

func (opts *SubscriptionSetWebhookOptions) Params() (jsonutils.JSONObject, error) {
	return ListStructToParams(opts)
}
