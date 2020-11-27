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

// DescribeSnapshots invokes the ecs.DescribeSnapshots API synchronously
func (client *Client) DescribeSnapshots(request *DescribeSnapshotsRequest) (response *DescribeSnapshotsResponse, err error) {
	response = CreateDescribeSnapshotsResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeSnapshotsWithChan invokes the ecs.DescribeSnapshots API asynchronously
func (client *Client) DescribeSnapshotsWithChan(request *DescribeSnapshotsRequest) (<-chan *DescribeSnapshotsResponse, <-chan error) {
	responseChan := make(chan *DescribeSnapshotsResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeSnapshots(request)
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

// DescribeSnapshotsWithCallback invokes the ecs.DescribeSnapshots API asynchronously
func (client *Client) DescribeSnapshotsWithCallback(request *DescribeSnapshotsRequest, callback func(response *DescribeSnapshotsResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeSnapshotsResponse
		var err error
		defer close(result)
		response, err = client.DescribeSnapshots(request)
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

// DescribeSnapshotsRequest is the request struct for api DescribeSnapshots
type DescribeSnapshotsRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer        `position:"Query" name:"ResourceOwnerId"`
	Filter2Value         string                  `position:"Query" name:"Filter.2.Value"`
	SnapshotIds          string                  `position:"Query" name:"SnapshotIds"`
	Usage                string                  `position:"Query" name:"Usage"`
	SnapshotLinkId       string                  `position:"Query" name:"SnapshotLinkId"`
	SnapshotName         string                  `position:"Query" name:"SnapshotName"`
	PageNumber           requests.Integer        `position:"Query" name:"PageNumber"`
	ResourceGroupId      string                  `position:"Query" name:"ResourceGroupId"`
	Filter1Key           string                  `position:"Query" name:"Filter.1.Key"`
	PageSize             requests.Integer        `position:"Query" name:"PageSize"`
	DiskId               string                  `position:"Query" name:"DiskId"`
	Tag                  *[]DescribeSnapshotsTag `position:"Query" name:"Tag"  type:"Repeated"`
	DryRun               requests.Boolean        `position:"Query" name:"DryRun"`
	ResourceOwnerAccount string                  `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string                  `position:"Query" name:"OwnerAccount"`
	SourceDiskType       string                  `position:"Query" name:"SourceDiskType"`
	Filter1Value         string                  `position:"Query" name:"Filter.1.Value"`
	Filter2Key           string                  `position:"Query" name:"Filter.2.Key"`
	OwnerId              requests.Integer        `position:"Query" name:"OwnerId"`
	InstanceId           string                  `position:"Query" name:"InstanceId"`
	Encrypted            requests.Boolean        `position:"Query" name:"Encrypted"`
	SnapshotType         string                  `position:"Query" name:"SnapshotType"`
	KMSKeyId             string                  `position:"Query" name:"KMSKeyId"`
	Category             string                  `position:"Query" name:"Category"`
	Status               string                  `position:"Query" name:"Status"`
}

// DescribeSnapshotsTag is a repeated param struct in DescribeSnapshotsRequest
type DescribeSnapshotsTag struct {
	Value string `name:"Value"`
	Key   string `name:"Key"`
}

// DescribeSnapshotsResponse is the response struct for api DescribeSnapshots
type DescribeSnapshotsResponse struct {
	*responses.BaseResponse
	RequestId  string    `json:"RequestId" xml:"RequestId"`
	TotalCount int       `json:"TotalCount" xml:"TotalCount"`
	PageNumber int       `json:"PageNumber" xml:"PageNumber"`
	PageSize   int       `json:"PageSize" xml:"PageSize"`
	Snapshots  Snapshots `json:"Snapshots" xml:"Snapshots"`
}

// CreateDescribeSnapshotsRequest creates a request to invoke DescribeSnapshots API
func CreateDescribeSnapshotsRequest() (request *DescribeSnapshotsRequest) {
	request = &DescribeSnapshotsRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Ecs", "2014-05-26", "DescribeSnapshots", "", "")
	request.Method = requests.POST
	return
}

// CreateDescribeSnapshotsResponse creates a response to parse from DescribeSnapshots response
func CreateDescribeSnapshotsResponse() (response *DescribeSnapshotsResponse) {
	response = &DescribeSnapshotsResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
