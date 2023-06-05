package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"

	"golang.org/x/exp/slog"

	cf "github.com/m4n5ter/dyndns/cloudflare"
	"github.com/m4n5ter/dyndns/config"
	dp "github.com/m4n5ter/dyndns/dnspod"
	"github.com/m4n5ter/dyndns/utils/log"
)

var (
	Config   config.Config
	PublicIp net.Addr
	Logger   = slog.New(slog.NewTextHandler(os.Stderr)).WithGroup("MAIN")
	wg       sync.WaitGroup
)

func init() {
	Config.CheckIpUrl = "http://checkip.dyndns.com/"
	dp.Logger = Logger.WithGroup("DP")
	config.Logger = Logger.WithGroup("CONFIG")
	cf.Logger = Logger.WithGroup("CF")
}

func main() {
	config.Load(&Config)
	// 获取公网 IP
	PublicIp = getPublicIp()
	// 启动 DDNS
	wg.Add(2)
	go func() {
		dp.DDNS(PublicIp, Config.DnspodConfig)
		wg.Done()
	}()
	go func() {
		cf.DDNS(PublicIp, Config.CloudflareConfig)
		wg.Done()
	}()
	wg.Wait()
}

// 获取 public ip TODO: 可以提供更多的 public ip 获取方式，并且可以在其中一个获取失败时切换到另一个
func getPublicIp() net.Addr {
	Logger.Info("正在获取公网IP,请稍后...", "check_ip_url", Config.CheckIpUrl)
	response, err := http.Get(Config.CheckIpUrl)
	defer response.Body.Close()
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("请求 %s 失败: %v\n", Config.CheckIpUrl, err))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("获取公网IP失败: %s\n", err))
	}

	// use regex to get ip
	re := regexp.MustCompile(`(?m)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ipStr := re.FindString(string(body))
	ip, err := net.ResolveIPAddr("ip", ipStr)
	if err != nil {

		log.LogPanic(Logger, fmt.Sprintf("解析公网IP失败: %s\n", err))
	}
	Logger.Info("公网IP获取成功:", "ip", ip)
	return ip
}
