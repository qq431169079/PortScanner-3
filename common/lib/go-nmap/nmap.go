package nmap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
)

type (
	Nmap struct {
		SystemPath    string
		Args          []string
		Ports         string
		Hosts         string
		MaxRttTimeOut string
		HostTimeOut   string
		Exclude       string
		MaxRetries    string
		Result        []byte
	}
	Result struct {
		IP         string      `json:"ip" bson:"ip"`
		Port       string      `json:"port" bson:"port"`
		Protocol   string      `json:"protocol" bson:"protocol"`
		ScanDate   string      `json:"scan_time" bson:"scan_time"`
		TimeStamp  int64       `json:"time_stamp" bson:"time_stamp"`
		MachineIp  string      `json:"machine_ip" bson:"machine_ip"`
		IsPublic   bool        `json:"is_public" bson:"is_public"`
		ScanFailed bool        `json:"scan_failed" bson:"scan_failed"`
		Service    PortService `json:"service" bson:"service"`
	}
	PortService struct {
		Name    string `json:"name" bson:"name"`
		Product string `json:"product" bson:"product"`
		Version string `json:"version" bson:"version"`
		Info    string `json:"info" bson:"info"`
	}
)

func (n *Nmap) SetSystemPath(systemPath string) {
	if systemPath != "" {
		n.SystemPath = systemPath
	}
}
func (n *Nmap) SetArgs(arg ...string) {
	n.Args = arg
}
func (n *Nmap) SetPorts(ports string) {
	n.Ports = ports
}
func (n *Nmap) SetHosts(hosts string) {
	n.Hosts = hosts
}
func (n *Nmap) SetMaxRttTimeOut(rttTimeOut string) {
	n.MaxRttTimeOut = rttTimeOut
}
func (n *Nmap) SetHostTimeOut(hostTimeOut string) {
	n.HostTimeOut = hostTimeOut
}
func (n *Nmap) SetMaxRetries(maxRetries string) {
	n.MaxRetries = maxRetries
}

// 排除扫描IP/IP段
func (n *Nmap) SetExclude(exclude string) {
	n.Exclude = exclude
}

func (n *Nmap) Run() error {
	var (
		cmd        *exec.Cmd
		outb, errs bytes.Buffer
	)
	n.Args = append(n.Args, "-sV")
	n.Args = append(n.Args, "-n")
	n.Args = append(n.Args, "-T4")
	n.Args = append(n.Args, "-Pn")
	n.Args = append(n.Args, "-open")

	if n.Hosts != "" {
		n.Args = append(n.Args, n.Hosts)
	}

	if n.Ports != "" {
		n.Args = append(n.Args, "-p")
		n.Args = append(n.Args, n.Ports)
	}

	if n.Exclude != "" {
		n.Args = append(n.Args, "--exclude")
		n.Args = append(n.Args, n.Exclude)
	}
	if n.MaxRetries != "" {
		n.Args = append(n.Args, "--max-retries")
		n.Args = append(n.Args, n.MaxRetries)
	}
	if n.HostTimeOut != "" {
		n.Args = append(n.Args, "--host-timeout")
		n.Args = append(n.Args, n.HostTimeOut)
	}
	if n.MaxRttTimeOut != "" {
		n.Args = append(n.Args, "--max-rtt-timeout")
		n.Args = append(n.Args, n.MaxRttTimeOut)
	}

	n.Args = append(n.Args, "-oX")
	n.Args = append(n.Args, "-")

	cmd = exec.Command(n.SystemPath, n.Args...)
	fmt.Println(cmd.Args)
	cmd.Stdout = &outb
	cmd.Stderr = &errs
	err := cmd.Run()

	if errs.Len() > 0 {
		return errors.New(errs.String())
	}
	if err != nil {
		return err
	}
	n.Result = outb.Bytes()
	return nil
}

// Parse takes a byte array of nmap xml data and unmarshals it into an
// NmapRun struct. All elements are returned as strings, it is up to the caller
// to check and cast them to the proper type.
func (n *Nmap) Parse() (*NmapRun, error) {
	results := &NmapRun{}
	err := xml.Unmarshal(n.Result, results)
	return results, err
}

func New() *Nmap {
	return &Nmap{
		SystemPath: "nmap",
	}
}
