package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"

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
	conn := connectUrl(Config.CheckIpUrl)
	defer conn.Close()

	// 组装一个使用自定义连接的 http client
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return conn, nil
		},
	}
	client := &http.Client{
		Timeout:   time.Second * 15,
		Transport: transport,
	}

	response, err := client.Get(Config.CheckIpUrl)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("请求 %s 失败: %v\n", Config.CheckIpUrl, err))
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("获取公网IP失败: %s\n", err))
	}

	re := regexp.MustCompile(`(?m)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ipStr := re.FindString(string(body))
	pip, err := net.ResolveIPAddr("ip", ipStr)
	if err != nil {

		log.LogPanic(Logger, fmt.Sprintf("解析公网IP失败: %s\n", err))
	}
	Logger.Info("公网IP获取成功:", "ip", pip)
	return pip
}

// connectUrl 生成一个优先使用 IPv4 的连接
func connectUrl(uri string) net.Conn {
	parsedUrl, err := url.Parse(uri)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("解析 %s 失败: %v\n", uri, err))
	}

	host := parsedUrl.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("Lookup %s 失败: %v\n", host, err))
	}

	var ip net.IP
	// v4
	for _, i := range ips {
		if i.To4() != nil {
			ip = i
			break
		}
	}
	// v6
	if ip == nil {
		ip = ips[0].To16()
	}

	conn, err := net.Dial("tcp", ip.String()+":443")
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("连接 %s 失败: %v\n", Config.CheckIpUrl, err))
	}

	return conn
}
