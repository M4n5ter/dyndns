package config

import (
	"fmt"
	// 第三方
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/exp/slog"

	// 本地
	"dyndns/cloudflare"
	"dyndns/dnspod"
	"dyndns/utils/log"

	// 标准库
	"io"
	"os"
	"path/filepath"
	"runtime"
)

var Logger *slog.Logger

type Config struct {
	// 用于检查 IP 的 URL
	CheckIpUrl       string            `toml:"check_ip_url"`
	DnspodConfig     dnspod.Config     `toml:"dnspod"`
	CloudflareConfig cloudflare.Config `toml:"cloudflare"`
}

// Load 加载配置文件
func Load(conf *Config) {
	configPath := getConfigPath("dyndns.toml", "dyndns", "DYNDNS_CONFIG")
	fs, err := os.Open(configPath)
	defer fs.Close()
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("配置文件打开失败,请检查%s文件的状态\n", configPath))
	}
	config, err := io.ReadAll(fs)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("配置文件读取失败,检查%s文件\n", configPath))
	}

	// 解析配置文件到 Conf
	err = toml.Unmarshal(config, &conf)
	if err != nil {
		log.LogPanic(Logger, fmt.Sprintf("配置文件解析失败: %s\n", err))
	}

	conf.check()
}

// 获取配置文件路径: 优先级: 环境变量 > 默认路径(当前目录,用户目录,系统目录)
// configName 为配置文件名（例如 "dyndns.toml"）,
// configDir 为期望的包含配置文件的目录（例如 "dyndns"，在本函数中具体会被指定为 "~/.dyndns" 和 "/etc/dyndns"或者 "%APPDATA%/dyndns"）,
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

func (c *Config) check() {
	c.DnspodConfig.CheckConfig()
	c.CloudflareConfig.CheckConfig()
}
