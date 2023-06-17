package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	console "github.com/alibabacloud-go/tea-console/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

var interrupt = make(chan os.Signal, 1)

func main() {
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	ReaderConfig()
	go ProtectionTask()
	<-interrupt

}

/**
 * 守护程序，按照设定的周期运行
 */
func ProtectionTask() {
	ticker := time.NewTicker(time.Duration(AppConfig.InspectionTime*60) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case sig := <-interrupt:
			fmt.Println("Task interrupted: ", sig)
			return
		case <-ticker.C:
			fmt.Println("Task executed at", time.Now())
			ipv4, ipv6, err := GetHostIP()
			if err == nil {
				for _, listen := range AppConfig.Listens {
					switch listen.Type {
					case "A":
						V4listenHander(ipv4, listen)
					case "AAAA":
						V6listenHander(ipv6, listen)
					default:
						return
					}
				}
			}

		}
	}

}

/**
* Initialization  初始化公共请求参数
 */
func Initialization() (_result *alidns.Client, _err error) {
	config := &openapi.Config{}
	// 您的AccessKey ID
	config.AccessKeyId = tea.String(AppConfig.AliAccount.AccessKeyId)
	// 您的AccessKey Secret
	config.AccessKeySecret = tea.String(AppConfig.AliAccount.AccessKeySecret)
	// 您的可用区ID
	// config.RegionId = &AppConfig.AliAccount.RegionId
	config.Endpoint = tea.String("alidns.cn-hangzhou.aliyuncs.com")
	_result = &alidns.Client{}
	_result, _err = alidns.NewClient(config)
	return _result, _err
}

/**
 * 修改解析记录
 */
func UpdateDomainRecord(client *alidns.Client, req *alidns.UpdateDomainRecordRequest) (_err error) {
	runtime := &util.RuntimeOptions{}
	tryErr := func() (_e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		rep, _err := client.UpdateDomainRecordWithOptions(req, runtime)
		if _err != nil {
			console.Error(tea.String(_err.Error()))
			return _err
		}
		console.Log(util.ToJSONString(tea.ToMap(rep)))

		return nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 如有需要，请打印 error
		_, _err = util.AssertAsString(error.Message)
		if _err != nil {
			console.Error(tea.String(_err.Error()))
			return _err
		}
	}
	return _err
}

func V4listenHander(addrs []net.IP, listen ListensInfo) {
	if listen.NetCheck {
		netip, err := GetNetIPV4()
		if err != nil {
			console.Error(tea.String("网络错误，联网校验失败"))
			return
		}
		if !contains(addrs, netip) {
			console.Error(tea.String("没有有效的公网IPV4地址"))
			return
		}
		record, recordId, _err := GetDescribeDomainRecords(listen)
		if _err != nil {
			return
		}
		if !record.Equal(netip) {
			fmt.Printf("云解析记录:%s 与当前地址:%s 不一致", record.String(), netip.String())
			updateDomainRecordRequest := alidns.UpdateDomainRecordRequest{
				RR:       tea.String(listen.RR),
				Type:     tea.String(listen.Type),
				Value:    tea.String(netip.String()),
				RecordId: tea.String(recordId),
			}
			println(*updateDomainRecordRequest.RecordId, recordId)
			client, err := Initialization()
			if err != nil {

				return
			}
			UpdateDomainRecord(client, &updateDomainRecordRequest)
			return
		}
		console.Info(tea.String("当前主机地址与云解析一致"))
		//
	} else {
		CheckAndUpdate(addrs[0], listen)
	}
}
func V6listenHander(addrs []net.IP, listen ListensInfo) {
	if listen.NetCheck {
		netip, err := GetNetIPV6()
		if err != nil {
			console.Error(tea.String("IPV6网络环境错误,联网校验失败"))
			return
		}
		console.Info(tea.String(fmt.Sprintf("获取到公网IPV6地址:%s", netip.String())))
		if !contains(addrs, netip) {
			console.Info(tea.String("没有有效的公网地址"))
			return
		}
		CheckAndUpdate(netip, listen)
	} else {
		CheckAndUpdate(addrs[0], listen)
	}
}
func CheckAndUpdate(netip net.IP, listen ListensInfo) {
	record, recordId, _err := GetDescribeDomainRecords(listen)
	if _err != nil {
		return
	}
	if !record.Equal(netip) {
		fmt.Printf("云解析记录:%s 与当前地址:%s 不一致", record.String(), netip.String())
		updateDomainRecordRequest := alidns.UpdateDomainRecordRequest{
			RR:       tea.String(listen.RR),
			Type:     tea.String(listen.Type),
			Value:    tea.String(netip.String()),
			RecordId: tea.String(recordId),
		}
		println(*updateDomainRecordRequest.RecordId, recordId)
		client, err := Initialization()
		if err != nil {

			return
		}
		UpdateDomainRecord(client, &updateDomainRecordRequest)
		return
	}
	console.Info(tea.String("当前主机地址与云解析一致"))
}

/**
*获取解析记录
 */
func GetDescribeDomainRecords(listen ListensInfo) (_record net.IP, _recordId string, _err error) {
	describeDomainRecordsRequest := &alidns.DescribeDomainRecordsRequest{
		DomainName:  tea.String(listen.DomainName),
		RRKeyWord:   tea.String(listen.RR),
		TypeKeyWord: tea.String(listen.Type),
	}
	client, err := Initialization()
	if err != nil {
		return nil, "", err
	}
	runtime := &util.RuntimeOptions{}
	record, _recordId, tryErr := func() (ip net.IP, id string, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		resp, _err := client.DescribeDomainRecordsWithOptions(describeDomainRecordsRequest, runtime)
		if _err != nil {
			return nil, "", _err
		}
		// 获取云解析中记录的值
		var net_value = *resp.Body.DomainRecords.Record[0].Value
		var recordId = *resp.Body.DomainRecords.Record[0].RecordId
		return net.ParseIP(net_value), recordId, _err
	}()
	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 如有需要，请打印 error
		_, _err := util.AssertAsString(error.Message)
		if _err != nil {
			println(_err.Error())
			return nil, "", _err
		}
	}
	return record, _recordId, tryErr
}
func contains(addrs []net.IP, addr net.IP) bool {
	for _, v := range addrs {
		if v.Equal(addr) {
			return true
		}
	}
	return false
}
func CheckDns() {

}

/**
* 获取本机IP地址
 */
func GetHostIP() ([]net.IP, []net.IP, error) {
	var IPV6List []net.IP
	var IPV4List []net.IP
	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return IPV4List, IPV6List, err
	}
	for _, addr := range interfaces {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.IsGlobalUnicast() {
				if ipnet.IP.To4() != nil {
					IPV4List = append(IPV4List, ipnet.IP)
				} else {
					IPV6List = append(IPV6List, ipnet.IP)
				}
			}
		}
	}
	return IPV4List, IPV6List, err
}

/**
* 获取公网IPV6出口地址
 */
func GetNetIPV6() (net.IP, error) {
	httpclient := http.Client{Timeout: 2 * time.Second}
	resp, err := httpclient.Get("https://6.ipw.cn")

	if err != nil {
		console.Error(tea.String("IPV6无法访问公网"))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		println("获取响应出错")
	}
	return net.ParseIP(string(body)).To16(), err
}

/**
*获取公网IPV4出口地址
 */
func GetNetIPV4() (net.IP, error) {
	httpclient := http.Client{Timeout: 2 * time.Second}
	resp, err := httpclient.Get("https://4.ipw.cn")

	if err != nil {
		console.Error(tea.String("无法访问公网"))
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		console.Error(tea.String("获取响应出错"))
	}
	return net.ParseIP(string(body)).To4(), err
}
