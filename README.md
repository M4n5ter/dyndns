# dyndns

## 说明

dyndns 是一个用于自动更新域名解析记录的工具，可以用于动态域名解析。
开发它原因是电信运营商为我提供了动态的公网IP地址，但是我的环境是光猫拨号并将软路由作为 DMZ 主机，并不是软路由等其它路由设备拨号(并且没有改变的打算)
常见的 DDNS 服务对这种情况不是很友好所以就有了这个工具。

## 使用方法

修改 `dyndns.toml.example` 文件后并去除`.example`即可使用，也可以稍微修改下代码，推荐配合 crontab 来工作。
加载配置文件的顺序为:
1. 环境变量 `DYNDNS_CONFIG` 指定的配置文件路径
2. 当前目录下的 `dyndns.toml` 文件
3. `~/.dyndns/dyndns.toml`
4. `/etc/dyndns/dyndns.toml`

先执行如下命令编译代码得到 `dyndns`
```bash
$ go build -ldflags="-s -w" -o dyndns
```
确定配置文件存在且正确后，运行 `dyndns` 即可。

## DNS 提供商支持

目前支持的 DNS 提供商有:
- [DNSPod](https://www.dnspod.cn/)
- [Cloudflare](https://www.cloudflare.com/)