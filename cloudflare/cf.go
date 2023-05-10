package cloudflare

import (
	"bytes"
	"dyndns/utils/log"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/exp/slog"
	"io"
	"net"
	"net/http"
)

var (
	ApiPre = "https://api.cloudflare.com/client/v4/"
	Logger *slog.Logger
)

type Config struct {
	ApiKey   string `toml:"api_key"`
	ZoneId   string `toml:"zone_id"`
	Name     string `toml:"name"`
	Proxied  bool   `toml:"proxied"`
	Type     string `toml:"type"`
	Comment  string `toml:"comment"`
	TTL      int    `toml:"ttl"`
	RecordId string `toml:"record_id"`
}

type Request struct {
	apiUrl string
	apiKey string
	method string
}

func DDNS(publicIp net.Addr, conf Config) {
	defer func() {
		if err := recover(); err != nil {
			Logger.Info("Cloudflare DDNS 失败: %s\n", err)
		}
	}()
	Logger.Info("开始更新 Cloudflare DNS 记录")

	// 获取记录详情，并得到目标记录的ID，如果记录不存在则创建
	var needUpdate = true
	var needCreate = true
	listResp := NewRequest(conf, http.MethodGet).Do(&ListRequestBody{
		Name: conf.Name,
		Type: conf.Type,
	}, &ListResponse{}).(*ListResponse)
	if listResp.Success {
		for _, record := range listResp.Result {
			if record.Content == publicIp.String() {
				needUpdate = false
				needCreate = false
				conf.RecordId = record.Id
				Logger.Info("无需更新", "name", record.Name, "type", record.Type, "content", record.Content)
				break
			}
			if record.Name == conf.Name {
				needCreate = false
				conf.RecordId = record.Id
				Logger.Info("需要更新为", "name", record.Name, "type", record.Type, "content", publicIp)
				break
			}
		}
	} else {
		Logger.Info("获取记录失败，尝试创建记录")
	}

	if needCreate {
		// 创建记录
		createResp := NewRequest(conf, http.MethodPost).Do(&CreateRequestBody{
			Name:    conf.Name,
			Type:    conf.Type,
			Content: publicIp.String(),
			Comment: conf.Comment,
			Proxied: conf.Proxied,
			TTL:     conf.TTL,
		}, &CreateResponse{}).(*CreateResponse)
		if createResp.Success {
			Logger.Info("记录创建成功", "name", conf.Name, "type", conf.Type)
			conf.RecordId = createResp.Result.Id
		} else {
			Logger.Info("记录创建失败", "name", conf.Name, "type", conf.Type, "error", createResp.Errors)
		}
	}

	if needUpdate {
		// 更新记录
		updateResp := NewRequest(conf, http.MethodPut).Do(&UpdateRequestBody{
			Name:    conf.Name,
			Type:    conf.Type,
			Content: publicIp.String(),
			Comment: conf.Comment,
			Proxied: conf.Proxied,
			TTL:     conf.TTL,
		}, &UpdateResponse{}).(*UpdateResponse)
		if updateResp.Success {
			Logger.Info("记录更新成功", "name", conf.Name, "type", conf.Type)
		} else {
			Logger.Info("记录更新失败", "name", conf.Name, "type", conf.Type, "error", updateResp.Errors)
		}
	}
}

func NewRequest(conf Config, method string) *Request {
	return &Request{
		apiUrl: NewApiUrl(conf.ZoneId, conf.RecordId),
		apiKey: conf.ApiKey,
		method: method,
	}
}

func NewApiUrl(zid, rid string) string {
	if zid == "" {
		panic("zone id is required")
	}

	var rUri string
	if rid != "" {
		rUri = "/" + rid
	}
	return fmt.Sprintf("%szones/%s/dns_records%s", ApiPre, zid, rUri)
}

func (c *Request) Do(body any, response any) any {
	payload, err := jsoniter.Marshal(body)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("json marshal error: %s", err))
	}

	req, err := http.NewRequest(c.method, c.apiUrl, bytes.NewReader(payload))
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("new request error: %s", err))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp == nil {
		log.LogPanic(Logger, fmt.Sprintf("do request error: %s", err))
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	err = jsoniter.Unmarshal(respBody, response)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("json unmarshal error: %s", err))
	}

	return response
}

// CheckConfig 检查配置文件是否提供了所有必须的配置
func (c *Config) CheckConfig() {
	if c == nil {
		log.LogPanic(Logger, "配置文件为空")
	}
	if c.ApiKey == "" {
		log.LogPanic(Logger, "api_key 不能为空")
	}
	if c.ZoneId == "" {
		log.LogPanic(Logger, "zone_id 不能为空")
	}
	if c.Name == "" {
		log.LogPanic(Logger, "name 不能为空")
	}
	if c.Type == "" {
		log.LogPanic(Logger, "type 不能为空")
	}
}
