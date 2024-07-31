package service

import (
	"fmt"
	"github.com/xiaohh-me/greateme_ddns/utils/alibaba"
	"log"
	"strings"
)

// SyncAllDomain 同步所有指定的域名到目前的公网IP上
func SyncAllDomain(domainName, dnsType string, availableIpv4 []string) error {

	// 获取公网IP地址
	/*wanIp, err := utils.GetWanIpAddress(dnsType)
	if err != nil {
		return err
	}
	log.Printf("成功获取当前的公网IP地址：%v\n", *wanIp)*/

	// 获取所有的域名列表
	domainList, err := alibaba.GetAllDomainList()
	if err != nil {
		return err
	}
	// 遍历所有需要同步DNS的域名，查询到它的二级域名
	fmt.Printf("开始尝试同步域名：%v\n", domainName)

	// 遍历所有域名，找到需要同步的二级域名和rr记录值
	level2Domain, rr := resolveDomainAndRR(domainName, domainList)
	if level2Domain == "" || rr == "" {
		log.Printf("非常抱歉域名: %v, 可能不属于您，请您确认你的阿里云账户的域名信息！\n", domainName)
	}
	fmt.Printf("成功查询到: %v, 域名信息信息，二级域名：%v，rr值：%v\n", domainName, level2Domain, rr)

	dnsList, err := alibaba.GetAllDNSListByDomainNameAndRR(&level2Domain, &rr)
	if err != nil {
		log.Printf("查询%v域名解析记录时候发生错误，错误信息：%v，将继续同步下一个域名\n", domainName, err)
	}

	// var targetRecord *alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord

	vs := make([]string, 0, len(*dnsList))
	// 判断记录类型是否存在
	for _, record := range *dnsList {
		if strings.Compare(*record.Type, "A") == 0 || // IPv4记录类型
			strings.Compare(*record.Type, "AAAA") == 0 || // IPv6记录类型
			strings.Compare(*record.Type, "CNAME") == 0 || // CNAME记录类型
			strings.Compare(*record.Type, "TXT") == 0 { // TXT记录类型
			vs = append(vs, *record.Value)
			if !containsString(*record.Value, availableIpv4) {
				// 减少
				err = alibaba.DeleteDNSRecord(record.RecordId)
				if err != nil {
					log.Printf("删除%v解析的时候发生错误，错误信息：%v\n", domainName, err)
				} else {
					fmt.Printf("删除%v解析记录成功，删除的IP地址为%v\n", domainName, *record.Value)
				}
			}
		}
	}

	for _, s := range difference(availableIpv4, vs) {
		// 增加
		err = alibaba.AddDNSRecord(&level2Domain, &rr, &s, &dnsType)
		if err != nil {
			log.Printf("新增%v解析的时候发生错误，错误信息：%v\n", domainName, err)
		} else {
			fmt.Printf("新增%v解析记录成功，解析到IP地址为%v\n", domainName, s)
		}
	}

	// alibaba.OpenDNSSLB(&domainName, &level2Domain, &rr, &dnsType)

	/*if targetRecord == nil {
		// 需要新增
		err = alibaba.AddDNSRecord(&level2Domain, &rr, &wanIp, &dnsType)
		if err != nil {
			log.Printf("新增%v解析的时候发生错误，错误信息：%v\n", domainName, err)
		} else {
			fmt.Printf("新增%v解析记录成功，解析到IP地址为%v\n", domainName, wanIp)
		}
	} else if strings.Compare(*targetRecord.Type, *alibaba.GetDNSType(&dnsType)) != 0 ||
		strings.Compare(*targetRecord.Value, wanIp) != 0 ||
		strings.Compare(*targetRecord.Line, "default") != 0 ||
		*targetRecord.TTL != 600 {
		// 需要修改
		err = alibaba.UpdateDNSRecord(targetRecord.RecordId, &rr, &wanIp, &dnsType)
		if err != nil {
			log.Printf("修改%v解析的时候发生错误，错误信息：%v\n", domainName, err)
		} else {
			fmt.Printf("修改%v解析记录成功，解析到IP地址：%v，原类型：%v，原记录值：%v\n", domainName, wanIp, *targetRecord.Type, *targetRecord.Value)
		}
	} else {
		fmt.Printf("无需修改%v的解析记录，记录值为：%v\n", domainName, wanIp)
	}*/

	return nil
}

func containsString(item string, items []string) bool {
	for i := range items {
		if items[i] == item {
			return true
		}
	}
	return false
}

func difference(slice1, slice2 []string) []string {
	m := make(map[string]struct{})
	for _, item := range slice2 {
		m[item] = struct{}{}
	}
	var diff []string
	for _, item := range slice1 {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return diff
}

func resolveDomainAndRR(domainName string, domainList *[]string) (string, string) {
	var level2Domain string
	for _, domain := range *domainList {
		// 如果二级域名比三级域名还要长，说明不是这个域名
		if len(domain) > len(domainName) {
			continue
		}
		// 判断后缀是否相等，如果相等那么就找到了这个二级域名
		if strings.HasSuffix(domainName, fmt.Sprintf(".%v", domain)) || domainName == domain {
			level2Domain = domain
			// 判断RR值
			if domainName == domain {
				return level2Domain, "@"
			}
			return level2Domain, strings.TrimSuffix(domainName, fmt.Sprintf(".%v", domain))
		}
	}
	return "", ""
}
