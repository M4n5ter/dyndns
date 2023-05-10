package dnspod

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// 添加记录
func create(credential *common.Credential, cpf *profile.ClientProfile, domain string, subdomain string, recordType string, value string) (response *dnspod.CreateRecordResponse, err error) {
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewCreateRecordRequest()

	request.Domain = common.StringPtr(domain)
	request.SubDomain = common.StringPtr(subdomain)
	request.RecordType = common.StringPtr(recordType)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(value)

	// 返回的resp是一个AddRecordResponse的实例，与请求对象对应
	response, err = client.CreateRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return
	}
	if err != nil {
		Logger.Info("发生未知错误", "error", err)
		return nil, err
	}
	return response, nil
}
