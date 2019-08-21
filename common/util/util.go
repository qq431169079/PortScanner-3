package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 从 list1 中移除 list2
func RemoveList2(list1, list2 []interface{}) (result []interface{}) {
	tmp := make(map[interface{}]bool)
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

// 判断是否是公网IP
func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}

// 读取文件
func ReadFile(filename string) string {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	return string(buf)
}

// 时间戳转字符串, intTime = time.Now().Unix()
func TimeToStr(intTime int64) string {
	timeLayout := "2006-01-02 15:04:05"
	dateTime := time.Unix(intTime, 0).Format(timeLayout)
	return dateTime
}

// 时间戳转日期
func TimeToDate(intTime int64) string {
	timeLayout := "2006-01-02"
	dataTime := time.Unix(intTime, 0).Format(timeLayout)
	return dataTime
}

// 切片去重，利用空结构体
func RemoveDuplicate(list []string) []string {
	result := make([]string, 0, len(list))
	temp := map[string]struct{}{}
	for _, item := range list {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// 解析 ip 路由地址
func parseIpRoute(output []byte) (localIP, interfaceName, routeIp string, err error) {
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == "via" && fields[3] == "dev" {
			localIP = fields[6]
			interfaceName = fields[4]
			routeIp = fields[2]
			return
		}
	}
	return
}

// 获取绝对路径
func GetAbsPath() (string, error) {
	path, err := filepath.Abs(filepath.Dir(os.Args[0]))
	return path, err
}

// 获取本机 IP
func GetLocalIp() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Are you connected to the network?")
}
