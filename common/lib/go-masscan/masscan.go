package masscan

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os/exec"
)

type (
	Result struct {
		IP       string `json:"ip" bson:"ip"`
		Port     string `json:"port" bson:"port"`
		IsPublic bool   `json:"is_public" bson:"is_public"`
	}
	Masscan struct {
		SystemPath    string
		Args          []string
		Interface     string
		Ports         string
		Ranges        string
		Rate          string
		Exclude       string
		RouteIp       string
		WaitTime      string
		Result        []byte
		AutoInterface bool
	}
)

func (m *Masscan) SetSystemPath(systemPath string) {
	if systemPath != "" {
		m.SystemPath = systemPath
	}
}
func (m *Masscan) SetArgs(arg ...string) {
	m.Args = arg
}
func (m *Masscan) SetPorts(ports string) {
	m.Ports = ports
}
func (m *Masscan) SetRanges(ranges string) {
	m.Ranges = ranges
}
func (m *Masscan) SetInterface(interfaces string) {
	m.Interface = interfaces
}
func (m *Masscan) SetRate(rate string) {
	m.Rate = rate
}
func (m *Masscan) SetRouteIp(routeIp string) {
	m.RouteIp = routeIp
}
func (m *Masscan) SetExclude(exclude string) {
	m.Exclude = exclude
}
func (m *Masscan) SetWaitTime(waitTime string) {
	m.WaitTime = waitTime
}
func (m *Masscan) SetAutoInterface(enable bool) {
	m.AutoInterface = enable
}

// Start scanning
func (m *Masscan) Run() error {
	var (
		cmd        *exec.Cmd
		outb, errs bytes.Buffer
	)
	if m.Rate != "" {
		m.Args = append(m.Args, "--rate")
		m.Args = append(m.Args, m.Rate)
	}
	if m.Ranges != "" {
		m.Args = append(m.Args, "--range")
		m.Args = append(m.Args, m.Ranges)
	}
	if m.Ports != "" {
		m.Args = append(m.Args, "-p")
		m.Args = append(m.Args, m.Ports)
	}
	//if m.Exclude != "" {
	//	m.Args = append(m.Args, "--exclude")
	//	m.Args = append(m.Args, m.Exclude)
	//}
	// 需要手动配置网卡和路由地址
	if !m.AutoInterface {
		if m.RouteIp != "" {
			m.Args = append(m.Args, "--router-ip")
			m.Args = append(m.Args, m.RouteIp)
		}
		if m.Interface != "" {
			m.Args = append(m.Args, "-e")
			m.Args = append(m.Args, m.Interface)
		}
	}
	if m.WaitTime != "" {
		m.Args = append(m.Args, "--wait")
		m.Args = append(m.Args, m.WaitTime)
	}
	m.Args = append(m.Args, "-oX")
	m.Args = append(m.Args, "-")
	cmd = exec.Command(m.SystemPath, m.Args...)
	fmt.Println(cmd.Args)
	cmd.Stdout = &outb
	cmd.Stderr = &errs
	err := cmd.Run()
	if err != nil {
		if errs.Len() > 0 {
			return errors.New(errs.String())
		}
		return err
	}
	m.Result = outb.Bytes()
	return nil
}

// Parse scans result
func (m *Masscan) Parse() ([]Host, error) {
	var hosts []Host
	decoder := xml.NewDecoder(bytes.NewReader(m.Result))
	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if t == nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "host" {
				var host Host
				err := decoder.DecodeElement(&host, &se)
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, err
				}
				hosts = append(hosts, host)
			}
		default:
		}
	}
	return hosts, nil
}
func New() *Masscan {
	return &Masscan{
		SystemPath: "masscan",
	}
}
