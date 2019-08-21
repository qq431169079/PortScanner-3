package agent

import (
	"strconv"
	"strings"

	"github.com/darkMoon1973/PortScanner/common/lib/go-masscan"
	"github.com/darkMoon1973/PortScanner/common/lib/go-nmap"
)

// masscan 扫描核心逻辑
// 自动获取网卡时 masscan --rate 2500 --range 114.114.114.114 -p 1-65535 --wait 1 -oX -
func masScanner(ipRange, portRange, rate string, isPublic bool) ([]masscan.Result, error) {
	var masscanResults []masscan.Result
	m := masscan.New()
	m.SetRanges(ipRange)
	m.SetPorts(portRange)
	m.SetRate(rate)
	m.SetWaitTime("1")
	m.AutoInterface = logic.Conf.Masscan.autoInterface
	if isPublic {
		m.SetInterface(logic.Conf.Masscan.publicInterface)
		m.SetRouteIp(logic.Conf.Masscan.publicRouteAddr)
	} else {
		m.SetInterface(logic.Conf.Masscan.privateInterface)
		m.SetRouteIp(logic.Conf.Masscan.privateRouteAddr)
	}
	err := m.Run()
	if err != nil {
		logic.Log.Error("Masscan scan fail, ", err)
		if !strings.Contains(err.Error(), "ranges overlapped something in an exclude file range") {
			logic.Log.Error("Masscan exclude file range error, ", err)
			return nil, err
		}
		return nil, err
	}
	results, err := m.Parse()
	if err != nil {
		logic.Log.Error("Parse masscan result fail, ", err)
		return nil, err
	}
	for _, result := range results {
		for _, v := range result.Ports {
			masscanResults = append(masscanResults, masscan.Result{
				IP:       result.Address.Addr,
				Port:     v.Portid,
				IsPublic: isPublic,
			})
		}
	}
	// 移除开放端口超过 5000 个的IP, 如 F5 或其他 VIP
	removed := removeDuplicate(masscanResults, 5000)
	logic.Log.Info(ipRange, " Scan result is: ", removeDuplicate)
	return removed, err
}

// 扫描并解析 Nmap 结果
func nmapScanner(host, port string, isPublic bool) ([]nmap.Result, error) {
	n := nmap.New()
	n.SetHosts(host)
	n.SetPorts(port)
	n.SetMaxRetries("3")
	n.SetHostTimeOut("100s")
	n.SetMaxRttTimeOut("10s")
	n.SetArgs("--version-all")
	err := n.Run()
	if err != nil {
		logic.Log.Error("Start nmap scan fail, ", err)
		return nil, err
	}
	results, err := n.Parse()
	var nmapResults []nmap.Result
	// 如果 nmap 规定时间内未扫描出结果, 返回 masscan 结果
	if results == nil {
		logic.Log.Error("Nmap scan result is nil, ", host, port)
	} else {
		for i := 0; i < len(results.Hosts); i++ {
			if results.Hosts[i].Status.State == "up" {
				var (
					portDesc    nmap.Result
					portList    []nmap.Result
					service     nmap.PortService
					serviceList []nmap.PortService
				)
				Host := results.Hosts[i].Addresses[0].Addr
				for j := 0; j < len(results.Hosts[i].Ports); j++ {
					portDesc.Port = strconv.Itoa(results.Hosts[i].Ports[j].PortId)
					portDesc.Protocol = results.Hosts[i].Ports[j].Protocol
					portDesc.ScanFailed = false
					service.Name = results.Hosts[i].Ports[j].Service.Name
					service.Product = results.Hosts[i].Ports[j].Service.Product
					service.Version = results.Hosts[i].Ports[j].Service.Version
					service.Info = results.Hosts[i].Ports[j].Service.ExtraInfo
					portList = append(portList, portDesc)
					serviceList = append(serviceList, service)
				}
				if len(portList) != 0 {
					for k, v := range portList {
						nmapResults = append(nmapResults, nmap.Result{
							IP:        Host,
							Port:      v.Port,
							Protocol:  v.Protocol,
							MachineIp: logic.localIp,
							IsPublic:  isPublic,
							Service:   serviceList[k],
						})
					}
				} else {
					nmapResults = append(nmapResults, nmap.Result{
						IP:         host,
						Port:       port,
						IsPublic:   isPublic,
						MachineIp:  logic.localIp,
						ScanFailed: true,
					})
				}
			}
		}
		return nmapResults, err
	}
	return nil, err
}

func removeDuplicate(m []masscan.Result, portCount int) (n []masscan.Result) {
	if len(m) == 0 {
		return
	}
	var duplicateCount = make(map[string]int)
	var ipList []string
	for _, v := range m {
		_, exist := duplicateCount[v.IP]
		if exist {
			duplicateCount[v.IP] += 1
		} else {
			duplicateCount[v.IP] = 1
		}
	}
	for v, k := range duplicateCount {
		if k < portCount {
			ipList = append(ipList, v)
		}
	}
	for _, v := range m {
		for _, k := range ipList {
			if k == v.IP {
				n = append(n, masscan.Result{IP: v.IP, Port: v.Port})
			}
		}
	}
	return n
}
