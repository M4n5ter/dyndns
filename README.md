# dnspod-ddns

## 说明

dnspod-ddns 是一个用于自动更新 DNSPod 域名解析记录的工具，可以用于动态域名解析。
开发它原因是电信运营商为我提供了动态的公网IP地址，但是我的环境是光猫拨号并将软路由作为 DMZ 主机，并不是软路由等其它路由设备拨号(并且没有改变的打算)
常见的 DDNS 服务对这种情况不是很友好所以就有了这个工具。

## 使用方法
修改 `dnspod-ddns.toml.example` 文件后并去除`.example`即可使用，也可以稍微修改下代码，推荐配合 crontab 来工作。
加载配置文件的顺序为:
1. 环境变量 `DNSPOD_DDNS_CONFIG` 指定的配置文件路径
2. 当前目录下的 `dnspod-ddns.toml` 文件
3. ~/.dnspod-ddns/dnspod-ddns.toml
4. /etc/dnspod-ddns/dnspod-ddns.toml

先执行如下命令编译代码得到 `dnspod-ddns`
```bash
$ go build -ldflags="-s -w" -o dnspod-ddns
```
确定配置文件存在且正确后，运行 `dnspod-ddns` 即可。