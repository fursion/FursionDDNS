package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"gopkg.in/yaml.v3"
)

type ListensInfo struct {
	Type       string `yaml:"Type"`
	RR         string `yaml:"RR"`
	DomainName string `yaml:"DomainName"`
	NetCheck   bool   `yaml:"NetCheck"`
}
type ConfigInfo struct {
	AliAccount     AliAccount    `yaml:"AliAccount"`
	InspectionTime int           `yaml:"InspectionTime"`
	NetCheck       bool          `yaml:"NetCheck"`
	Listens        []ListensInfo `yaml:"Listens"`
}
type AliAccount struct {
	AccessKeyId     string `yaml:"AccessKeyId"`
	AccessKeySecret string `yaml:"AccessKeySecret"`
	RegionId        string `yaml:"RegionId"`
}

var AppConfig = ConfigInfo{}

func ReaderConfig() {

	println(os.Getenv("OS"))
	platform := runtime.GOOS
	var configpath string
	switch platform {
	case "linux":
		configpath = "/etc/fursion/ddns/config.yaml"
	case "windows":
		configpath = "config.yaml"
	default:
		configpath = "config.yaml"
	}
	bytes, err := os.ReadFile(configpath)
	if os.IsNotExist(err) {
		fmt.Printf("\033[0;31mPlatform %s配置文件不存在,请检查配置文件 %s\033[0m", platform, configpath)
		os.Exit(1)
	} else if err != nil {
		fmt.Println("配置文件访问异常")
		os.Exit(1)
	} else {
		y_err := yaml.Unmarshal(bytes, &AppConfig)
		if y_err != nil {
			fmt.Println("配置文件解析失败")
			os.Exit(1)
		}
		fmt.Printf("%.63s\n", "--------------------------------------------------------------------------------------------")
		fmt.Printf("|%.19s共%-d个监听|检测周期%-d分钟%.19s|\n", "--------------------------------------", len(AppConfig.Listens), AppConfig.InspectionTime, "---------------------------------------")
		fmt.Printf("%.63s\n", "--------------------------------------------------------------------------------------------")
		// fmt.Printf("|\tRR\tDomainName\tType\t公网校验\t|\n")
		fmt.Printf("|\t%-10s%-20s%-10s%-10s|\n", "RR", "DomainName", "Type", "公网校验")
		fmt.Printf("%.63s\n", "--------------------------------------------------------------------------------------------")
		for _, listen := range AppConfig.Listens {
			fmt.Printf("|\t%-10s%-20s%-10s%-14s|\n", listen.RR, listen.DomainName, listen.Type, strconv.FormatBool(listen.NetCheck))
			fmt.Printf("%.63s\n", "--------------------------------------------------------------------------------------------")
		}
	}
}
