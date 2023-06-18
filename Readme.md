# DDNS程序

## 安装/卸载

- Ubuntu/Debian

```bash
#从Github拉取安装包
wget https://github.com/fursion/FursionDDNS/releases/download/v1.0.0/fursion-ddns_1.0.0.deb
#安装
#安装时会自动释放配置文件至/etc/fursion/ddns/目录下,同时会自动注册service并运行
sudo dpkg -i fursion-ddns_1.0.0.deb
# 安装完成后需要需改配置文件，添加账户信息和监听
# 修改完后重启service
sudo systemctl restart fursion_ddns.service

#卸载
sudo dpkg -r fursion-ddns
#卸载时会自动清除配置文件和注销service
```

## 配置文件介绍

```yaml
#阿里云账户信息，为保证账户安全，建议在阿里云RAM中新建用户，单独指定云解析权限
AliAccount:
  #你的阿里云AccessKeyID
  AccessKeyId: xxx-xxx-xxx
  #你的阿里云AccessKeySecret  -> AccessKeySecret在创建AccessKeyId时显示一次，无法找回。注意保存！
  AccessKeySecret: xxx-xxx-xxx
  #你的可用区ID/选填
  RegionId: xxx-xxx-xxx

#检查周期 /分钟
InspectionTime: 5
#需要动态解析的域名数组
Listens:
  - #DNS解析类型 AAAA->IPV6 A->IPV4
    Type: AAAA 
    #主机名
    RR: www
    #主域名
    DomainName: example.com
    #是否启用公网IP校验/关闭状态时默认解析获取到的第一个IP地址  (Yes/No)
    NetCheck: Yes
  - Type: A
    RR: "dev"
    DomainName: example.com
    NetCheck: Yes
```
