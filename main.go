package main

import (
	"github.com/mcuadros/go-defaults"
	"github.com/pelletier/go-toml/v2"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var CFG Config

var (
	PUBLIC_IP net.Addr
	LOGGER    = log.Default()
)

type Config struct {
	// 域名
	Domain string `toml:"domain"`
	// 子域名
	SubDomain string `toml:"subdomain"`
	// dnspod 的 SecretId
	SecretId string `toml:"secret_id"`
	// dnspod 的 SecretKey
	SecretKey string `toml:"secret_key"`
	// 用于检查 IP 的 URL
	CheckIpUrl string `toml:"check_ip_url" default:"http://checkip.dyndns.com/"`
}

func init() {
	// 设置默认值
	defaults.SetDefaults(&CFG)
}

func main() {
	initConfig()

	// 实例化一个认证对象，入参需要传入腾讯云账户secretId，secretKey,此处还需注意密钥对的保密
	// 密钥可前往https://console.cloud.tencent.com/cam/capi网站进行获取
	credential := common.NewCredential(
		CFG.SecretId,
		CFG.SecretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	// 获取记录详情，并得到目标记录的ID
	var recordId uint64
	var needUpdate = true

	response, err := getRecordDesc(credential, cpf, CFG.Domain, CFG.SubDomain)
	if err != nil {
		if err, ok := err.(*errors.TencentCloudSDKError); ok {
			switch err.GetCode() {
			case dnspod.RESOURCENOTFOUND_NODATAOFRECORD:
				// 未找到记录，需要添加该记录
				resp, err := addRecord(credential, cpf, CFG.Domain, CFG.SubDomain, "A", PUBLIC_IP.String())
				if err != nil {
					LOGGER.Printf("An error occured: %s\n", err)
					return
				}

				needUpdate = false
				recordId = *resp.Response.RecordId
				LOGGER.Println("添加记录成功,记录值: ", PUBLIC_IP)
				return
			default:
				LOGGER.Printf("An error occured: %s\n", err)
				return
			}
		}
	} else {
		recordList := response.Response.RecordList
		// 目前最多只会有一个记录，所以直接取第一个
		recordId = *recordList[0].RecordId
		// 判断是否需要更新
		oldRecordValue := recordList[0].Value
		if *oldRecordValue == PUBLIC_IP.String() {
			needUpdate = false
			LOGGER.Println("无需更新!")
			return
		}

		// 更新记录
		if needUpdate {
			_, err := modifyRecord(credential, cpf, CFG.Domain, CFG.SubDomain, recordId, PUBLIC_IP.String())
			if err != nil {
				LOGGER.Printf("An error occured: %s\n", err)
				return
			}
			LOGGER.Println("更新成功!")
		}
	}
}

// 初始化配置
func initConfig() {
	loadConfig()

	// 获取公网IP
	ip, err := getPublicIp()
	if err != nil {
		LOGGER.Printf("获取公网IP失败: %s\n", err)
		return
	}
	PUBLIC_IP = ip
	LOGGER.Println("获取公网IP成功: ", PUBLIC_IP)
}

// 加载配置文件
func loadConfig() {
	configPath := getConfigPath("dnspod-ddns.toml", "dnspod-ddns", "DNSPOD_DDNS_CONFIG")
	fs, err := os.Open(configPath)
	defer fs.Close()
	if err != nil {
		LOGGER.Panicf("配置文件打开失败,请检查%s文件的状态\n", configPath)
	}
	config, err := io.ReadAll(fs)
	if err != nil {
		LOGGER.Panicf("配置文件读取失败,检查%s文件\n", configPath)
	}

	// 解析配置文件到 CFG
	err = toml.Unmarshal(config, &CFG)
	if err != nil {
		LOGGER.Panicf("配置文件解析失败: %s\n", err)
	}

	checkConfig(&CFG)
}

// 获取配置文件路径
// configName 为配置文件名（例如 "dnspod-ddns.toml"）,
// configDir 为期望的包含配置文件的目录（例如 "dnspod-ddns"，在本函数中具体会被指定为 "~/.dnspod-ddns" 和 "/etc/dnspod-ddns"或者 "%APPDATA%/dnspod-ddns"）,
// env 为环境变量名（例如 "DNSPOD_DDNS_CONFIG"）,
func getConfigPath(configName, configDir, env string) string {
	// 优先从环境变量获取配置文件路径
	// 如果没有则从当前目录下加载 configName
	// 如果当前目录下没有 configName 则加载 ~/.configDir/configName
	// 上述路径都没有则加载 /etc/configDir/configName（类 Unix）或者 %APPDATA%/configDir/configName（Windows）
	configPath := os.Getenv(env)
	if configPath == "" {
		// 当前目录下
		configPath = configName
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// 当前目录下没有则加载 ~/.configDir/configName
			configPath = filepath.Join(os.Getenv("HOME"), "."+configDir, configName)

			// ~/.configDir/configName 不存在
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// 如果是 windows 系统则加载 %APPDATA%/configDir/configName
				if runtime.GOOS == "windows" {
					configPath = filepath.Join(os.Getenv("APPDATA"), configDir, configName)
				} else {
					// 如果是类 Unix 系统则加载 /etc/configDir/configName
					configPath = filepath.Join("/etc", configDir, configName)
				}
			}
		}
	}
	return configPath
}

// 检查配置文件是否提供了所有必须的配置
func checkConfig(c *Config) {
	if c == nil {
		LOGGER.Panic("配置文件为空")
	}
	if c.Domain == "" {
		LOGGER.Panic("配置文件缺少域名(domain)配置")
	}
	if c.SubDomain == "" {
		LOGGER.Panic("配置文件缺少子域名(subdomain)配置")
	}
	if c.SecretId == "" {
		LOGGER.Panic("配置文件缺少 secret_id 配置")
	}
	if c.SecretKey == "" {
		LOGGER.Panic("配置文件缺少 secret_key 配置")
	}
}

// 获取 public ip TODO: 可以提供更多的 public ip 获取方式
func getPublicIp() (ip net.Addr, err error) {
	LOGGER.Printf("正在从%s获取公网IP,请稍后...\n", CFG.CheckIpUrl)
	response, err := http.Get(CFG.CheckIpUrl)
	defer response.Body.Close()
	if err != nil {
		LOGGER.Printf("请求 %s 失败: %s\n", CFG.CheckIpUrl, err)
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		LOGGER.Printf("获取公网IP失败: %s\n", err)
		return nil, err
	}

	// use regex to get ip
	re := regexp.MustCompile(`(?m)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ipStr := re.FindString(string(body))
	ip, err = net.ResolveIPAddr("ip", ipStr)
	return ip, nil
}

// 获取记录详情
func getRecordDesc(credential *common.Credential, cpf *profile.ClientProfile, domain string, subdomain string) (response *dnspod.DescribeRecordListResponse, err error) {
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
		LOGGER.Printf("An unknown error has returned: %s\n", err)
		return nil, err
	}
	return response, nil
}

// 修改记录
func modifyRecord(credential *common.Credential, cpf *profile.ClientProfile, domain string, subdomain string, recordId uint64, value string) (response *dnspod.ModifyRecordResponse, err error) {
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := dnspod.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewModifyRecordRequest()

	request.Domain = common.StringPtr(domain)
	request.SubDomain = common.StringPtr(subdomain)
	request.RecordId = common.Uint64Ptr(recordId)
	request.RecordLine = common.StringPtr("默认")
	request.RecordType = common.StringPtr("A")
	request.Value = common.StringPtr(value)

	// 返回的resp是一个ModifyRecordResponse的实例，与请求对象对应
	response, err = client.ModifyRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return
	}
	if err != nil {
		LOGGER.Printf("An unknown error has returned: %s\n", err)
		return nil, err
	}
	return response, nil
}

// 添加记录
func addRecord(credential *common.Credential, cpf *profile.ClientProfile, domain string, subdomain string, recordType string, value string) (response *dnspod.CreateRecordResponse, err error) {
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
		LOGGER.Printf("An unknown error has returned: %s\n", err)
		return nil, err
	}
	return response, nil
}
