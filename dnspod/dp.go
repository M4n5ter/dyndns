package dnspod

import (
	"net"

	"github.com/m4n5ter/log"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

var Logger *log.Logger

type Config struct {
	// 域名
	Domain string `toml:"domain"`
	// 子域名
	SubDomain string `toml:"subdomain"`
	// dnspod 的 SecretId
	SecretId string `toml:"secret_id"`
	// dnspod 的 SecretKey
	SecretKey string `toml:"secret_key"`
}

func DDNS(publicIp net.Addr, conf Config) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error("DNSPOD DDNS 失败", "error", err)
		}
	}()
	Logger.Info("开始更新 DNSPOD DNS 记录")
	// 实例化一个认证对象，入参需要传入腾讯云账户secretId，secretKey,此处还需注意密钥对的保密
	// 密钥可前往https://console.cloud.tencent.com/cam/capi网站进行获取
	credential := common.NewCredential(
		conf.SecretId,
		conf.SecretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	// 获取记录详情，并得到目标记录的ID
	var recordId uint64
	var needUpdate = true

	response, err := get(credential, cpf, conf.Domain, conf.SubDomain)
	if err != nil {
		if err, ok := err.(*errors.TencentCloudSDKError); ok {
			switch err.GetCode() {
			case dnspod.RESOURCENOTFOUND_NODATAOFRECORD:
				// 未找到记录，需要添加该记录
				resp, err := create(credential, cpf, conf.Domain, conf.SubDomain, "A", publicIp.String())
				if err != nil {
					Logger.Error("发生错误", "error", err)
					return
				}

				needUpdate = false
				recordId = *resp.Response.RecordId
				Logger.Info("添加记录成功", "content", publicIp)
				return
			default:
				Logger.Error("发送错误", "error", err)
				return
			}
		}
	} else {
		recordList := response.Response.RecordList
		// 目前最多只会有一个记录，所以直接取第一个
		recordId = *recordList[0].RecordId
		// 判断是否需要更新
		oldRecordValue := recordList[0].Value
		if *oldRecordValue == publicIp.String() {
			needUpdate = false
			Logger.Info("无需更新!")
			return
		}

		// 更新记录
		if needUpdate {
			_, err := update(credential, cpf, conf.Domain, conf.SubDomain, recordId, publicIp.String())
			if err != nil {
				Logger.Error("发送错误", "error", err)
				return
			}
			Logger.Info("更新成功!")
		}
	}
}

// CheckConfig 检查配置文件是否提供了所有必须的配置
func (c *Config) CheckConfig() {
	if c == nil ||
		c.Domain == "" ||
		c.SubDomain == "" ||
		c.SecretId == "" ||
		c.SecretKey == "" {
		Logger.Panic("缺少必须字段")
	}
}
