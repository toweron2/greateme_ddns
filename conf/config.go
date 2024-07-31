package conf

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"time"
)

type Config struct {
	Aliyun struct {
		AccessKeyId     string `yaml:"accessKeyId"`
		AccessKeySecret string `yaml:"accessKeySecret"`
		DomainEndpoint  string `yaml:"domainEndpoint"`
		DnsEndpoint     string `yaml:"dnsEndpoint"`
	} `yaml:"aliyun"`

	Dns []struct {
		// DomainList 需要被解析的域名列表
		Domain string   `yaml:"domain"`
		Ipv4   []string `yaml:"ipv4"`
		// DnsType 解析类型，只能是 ipv4 和 ipv6 （注意全部小写且不能为大写）
		DnsType string `yaml:"dnsType"`
	} `yaml:"dns"`

	Time struct {
		// ExecType 执行类型，可选值：single 和 repetition ，single：只执行一次，需要配合系统的定时任务执行。repetition重复执行，需要配合durationMinute配置项执行
		Type string `yaml:"type"`
		// DurationMinute 时隔多久同步一次域名解析，单位为分钟
		DurationMinute time.Duration `yaml:"durationMinute"`
	} `yaml:"time"`
}

func MustLoad(path string, v any) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取配置文件时候发生错误：%v\n", err)
	}
	err = yaml.Unmarshal(content, v)
	if err != nil {
		log.Fatalf("配置文件解析错误：%v\n", err)
	}

}
