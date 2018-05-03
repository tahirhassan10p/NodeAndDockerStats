package main

import (
	"github.com/docker/docker/api/types"
)

type Configuration struct {
	URL      string
	WaitTime int
}

type InfoBase struct {
	NodeInfo      NodeInfo
	ContainerInfo []types.ContainerJSON
}

type NodeInfo struct {
	RuntimeOS                string
	TotalMemory              string
	FreeMemory               string
	PercentUsedMemory        string
	DiskSerialNumber         string
	TotalDiskSpace           string
	UsedDiskSpace            string
	FreeDiskSpace            string
	PercentDiskSpaceUsed     string
	CpuInfo                  []CPUInfo
	HostName                 string
	Uptime                   string
	NumberOfProcessesRunning string
	OperatingSystem          string
	Plateform                string
	HostId                   string
	NetworkInfo              []NetworkInfo
}

type CPUInfo struct {
	IndexNumber   string
	VendorId      string
	Family        string
	NumberOfCores string
	ModelName     string
	Speed         string
	PercentUsed   string
}

type NetworkInfo struct {
	InterfaceName     string
	MacAddress        string
	InterfaceBehavior string
	IpAddress         string
}
