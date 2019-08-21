package server

import (
	"net"
	"strings"

	"github.com/darkMoon1973/PortScanner/common/util"
)

type (
	ipInfo struct {
		Ip         string `bson:"ip"`
		IsPublic   bool   `bson:"is_public"`
		InsertTime string `bson:"insert_time"`
		Info       string `bson:"info"`
	}
)

// 从 mongodb 获取黑白名单IP并去重
func getScanList() []ipInfo {
	var (
		scanList  []ipInfo
		whiteList []ipInfo
		result    []ipInfo
	)
	// 从 Mongodb 中获取扫描 IP 和白名单 IP，否则从文件中获取任务
	scanList = parseIp("scan_list")
	logic.Log.Warn("Get scan ip success, len ip is: ", len(scanList))
	whiteList = parseIp("white_list")
	logic.Log.Warn("Get white list ip success, len ip is: ", len(whiteList))
	// remove white list
	result = removeList2(scanList, whiteList)
	logic.Log.Warn("After remove white list ip, len ip is: ", len(result))
	return result
}

// 从 Mongodb 读取 IP 列表并去重拆分
func parseIp(table string) []ipInfo {
	n, err := get(table)
	if err != nil {
		logic.Log.Errorf("Get ip from mongodb fail ", err)
	}
	ipList := removeDuplicate(n)
	for _, v := range ipList {
		ip, err := getIpList(v)
		if err != nil {
			logic.Log.Errorf("Ipnet %s to ip fail, ", v, err)
		}
		ipList = append(ipList, ip...)
	}
	return ipList
}

// 从文件读取黑白名单，导入 Mongodb 数据库
func insertIp(scanList, whiteList string) {
	var (
		sList ipInfo
		wList ipInfo
	)
	scan := util.ReadFile(scanList)
	white := util.ReadFile(whiteList)

	sl := strings.Split(scan, "\n")
	wl := strings.Split(white, "\n")
	l := util.RemoveDuplicate(sl)
	w := util.RemoveDuplicate(wl)
	for _, v := range l {
		ipAddress, _, err := net.ParseCIDR(v)
		if err != nil {
			logic.Log.Errorf("Parse %s to net.Ip fail, ", v, err)
		}
		sList.IsPublic = util.IsPublicIP(ipAddress)
		sList.Ip = v
		err = upsert("scan_list", sList)
		if err != nil {
			logic.Log.Error("Insert scan ip list to mongodb failed", err)
		}
	}
	for _, v := range w {
		ipAddress, _, err := net.ParseCIDR(v)
		if err != nil {
			logic.Log.Errorf("Parse %s to net.Ip fail, ", v, err)
		}
		wList.IsPublic = util.IsPublicIP(ipAddress)
		wList.Ip = v
		err = upsert("white_list", wList)
		if err != nil {
			logic.Log.Error("Insert white ip list to mongodb failed", err)
		}
	}
}

// 将网段拆分成IP
func getIpList(ipCidr ipInfo) ([]ipInfo, error) {
	var ipList []ipInfo
	if ipCidr.Ip == "" {
		return nil, nil
	}
	if !strings.Contains(ipCidr.Ip, "/") {
		ipList = append(ipList, ipCidr)
		return ipList, nil
	}
	i := strings.Split(ipCidr.Ip, "/")
	if i[1] == "32" {
		ipCidr.Ip = i[0]
		ipList = append(ipList, ipCidr)
		return ipList, nil
	}
	ip, ipNet, err := net.ParseCIDR(ipCidr.Ip)
	if err != nil {
		return nil, err
	}
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); getNextIp(ip) {
		ipCidr.Ip = ip.String()
		ipList = append(ipList, ipCidr)
	}
	return ipList[1:], nil
}

func getNextIp(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}

func removeList2(list1, list2 []ipInfo) (result []ipInfo) {
	tmp := make(map[ipInfo]bool)
	for _, v := range list1 {
		tmp[v] = true
	}
	for _, v := range list2 {
		tmp[v] = false
	}
	for k, v := range tmp {
		if v == true {
			result = append(result, k)
		}
	}
	return result
}

func removeDuplicate(list []ipInfo) []ipInfo {
	result := make([]ipInfo, 0, len(list))
	temp := map[string]struct{}{}
	for _, item := range list {
		if _, ok := temp[item.Ip]; !ok {
			temp[item.Ip] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
