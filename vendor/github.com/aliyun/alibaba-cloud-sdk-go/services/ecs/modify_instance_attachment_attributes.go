package ecs

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// ModifyInstanceAttachmentAttributes invokes the ecs.ModifyInstanceAttachmentAttributes API synchronously
func (client *Client) ModifyInstanceAttachmentAttributes(request *ModifyInstanceAttachmentAttributesRequest) (response *ModifyInstanceAttachmentAttributesResponse, err error) {
	response = CreateModifyInstanceAttachmentAttributesResponse()
	err = client.DoAction(request, response)
	return
}

// ModifyInstanceAttachmentAttributesWithChan invokes the ecs.ModifyInstanceAttachmentAttributes API asynchronously
func (client *Client) ModifyInstanceAttachmentAttributesWithChan(request *ModifyInstanceAttachmentAttributesRequest) (<-chan *ModifyInstanceAttachmentAttributesResponse, <-chan error) {
	responseChan := make(chan *ModifyInstanceAttachmentAttributesResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.ModifyInstanceAttachmentAttributes(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// ModifyInstanceAttachmentAttributesWithCallback invokes the ecs.ModifyInstanceAttachmentAttributes API asynchronously
func (client *Client) ModifyInstanceAttachmentAttributesWithCallback(request *ModifyInstanceAttachmentAttributesRequest, callback func(response *ModifyInstanceAttachmentAttributesResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *ModifyInstanceAttachmentAttributesResponse
		var err error
		defer close(result)
		response, err = client.ModifyInstanceAttachmentAttributes(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// ModifyInstanceAttachmentAttributesRequest is the request struct for api ModifyInstanceAttachmentAttributes
type ModifyInstanceAttachmentAttributesRequest struct {
	*requests.RpcRequest
	ResourceOwnerId                 requests.Integer `position:"Query" name:"ResourceOwnerId"`
	PrivatePoolOptionsMatchCriteria string           `position:"Query" name:"PrivatePoolOptions.MatchCriteria"`
	PrivatePoolOptionsId            string           `position:"Query" name:"PrivatePoolOptions.Id"`
	ResourceOwnerAccount            string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount                    string           `position:"Query" name:"OwnerAccount"`
	OwnerId                         requests.Integer `position:"Query" name:"OwnerId"`
	InstanceId                      string           `position:"Query" name:"InstanceId"`
}

// ModifyInstanceAttachmentAttributesResponse is the response struct for api ModifyInstanceAttachmentAttributes
type ModifyInstanceAttachmentAttributesResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateModifyInstanceAttachmentAttributesRequest creates a request to invoke ModifyInstanceAttachmentAttributes API
func CreateModifyInstanceAttachmentAttributesRequest() (request *ModifyInstanceAttachmentAttributesRequest) {
	request = &ModifyInstanceAttachmentAttributesRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ecs", "2014-05-26", "ModifyInstanceAttachmentAttributes", "", "")
	request.Method = requests.POST
	return
}

// CreateModifyInstanceAttachmentAttributesResponse creates a response to parse from ModifyInstanceAttachmentAttributes response
func CreateModifyInstanceAttachmentAttributesResponse() (response *ModifyInstanceAttachmentAttributesResponse) {
	response = &ModifyInstanceAttachmentAttributesResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
