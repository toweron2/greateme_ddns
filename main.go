package main

import (
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/xiaohh-me/greateme_ddns/conf"
	"github.com/xiaohh-me/greateme_ddns/service"
	"github.com/xiaohh-me/greateme_ddns/utils/alibaba"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
)

// 域名服务文档地址：https://next.api.aliyun.com/api/Domain/2018-01-29/SaveSingleTaskForCreatingOrderActivate?lang=GO
// 云解析文档地址：https://next.api.aliyun.com/api/Alidns/2015-01-09/AddCustomLine?lang=GO

func main() {
	// 配置文件的路径
	var configFilePath *string
	if len(os.Args) >= 2 {
		// 自定义配置文件路径，读取执行参数的第二个值，也就是下标为1的值
		configFilePath = tea.String(os.Args[1])
	} else {
		// 读取默认配置文件路径
		configFilePath = tea.String("./config.yaml")
	}

	var c conf.Config
	conf.MustLoad(*configFilePath, &c)

	// 初始化阿里云域名客户端
	err := alibaba.InitClient(c.Aliyun.AccessKeyId, c.Aliyun.AccessKeySecret, c.Aliyun.DomainEndpoint, c.Aliyun.DnsEndpoint)
	if err != nil {
		log.Fatalf("初始化阿里云域名客户端的时候发生了错误：%v\n", err)
	}
	fmt.Println("域名和DNS解析客户端初始化成功")
	fmt.Printf("执行任务：每\t%s\t执行一次的任务\n", c.Time.DurationMinute.String())

	// 创建一个定时器，每10分钟触发一次
	ticker := time.NewTicker(c.Time.DurationMinute)
	defer ticker.Stop() // 确保在函数退出时停止定时器

	// 循环执行任务
	for {
		select {
		case <-ticker.C:
			// 这里放置你要执行的任务代码
			for _, dns := range c.Dns {
				// 开始同步
				err = service.SyncAllDomain(dns.Domain, dns.DnsType, checkAvailability(dns.Ipv4))
				if err != nil {
					log.Printf("同步域名信息的时候发生了异常：%v\n", err)
				}
			}
			runtime.GC()
			// 如果需要在程序中止时清理资源或其他操作，可以监听信号
			// case <-stopSignal:
			//   return
		}
	}
}

func checkAvailability(targets []string) []string {
	fmt.Printf("-----------------------%s-------------------------\n", time.Now().Format(time.DateTime))
	var available []string
	for _, t := range targets {
		conn, err := net.DialTimeout("tcp", t, 5*time.Second)
		if err != nil {
			fmt.Printf("✘\t%s: %v\n", t, err.Error())
		} else {
			fmt.Printf("✔\t%s\n", t)
			available = append(available, strings.Split(t, ":")[0])
			conn.Close()
		}
	}
	if len(available) == 0 {
		log.Printf("!!! 没有可用的地址")
	}
	fmt.Println("-------------------------------------------------------------------")
	return available
}
