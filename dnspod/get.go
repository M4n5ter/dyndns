package dnspod

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// 获取记录详情
func get(credential *common.Credential, cpf *profile.ClientProfile, domain string, subdomain string) (response *dnspod.DescribeRecordListResponse, err error) {
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewDescribeRecordListRequest()

	request.Domain = common.StringPtr(domain)
	request.Subdomain = common.StringPtr(subdomain)

	// 返回的resp是一个DescribeRecordListResponse的实例，与请求对象对应
	response, err = client.DescribeRecordList(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return
	}
	if err != nil {
		Logger.Info("发生未知错误", "error", err)
		return nil, err
	}
	return response, nil
}
