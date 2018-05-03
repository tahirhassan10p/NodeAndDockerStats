package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	//  "github.com/shirou/gopsutil/docker"
	"encoding/json"
	"runtime"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func dealwithErr(err error) {
	if err != nil {
		fmt.Println(err)
		//os.Exit(-1)
	}
}

func GetHardwareData(node *NodeInfo) {
	runtimeOS := runtime.GOOS
	// memory
	vmStat, err := mem.VirtualMemory()
	dealwithErr(err)

	// disk - start from "/" mount point for Linux
	// might have to change for Windows!!
	// don't have a Window to test this out, if detect OS == windows
	// then use "\" instead of "/"

	diskStat, err := disk.Usage("/")
	dealwithErr(err)

	// cpu - get CPU number of cores and speed
	cpuStat, err := cpu.Info()
	dealwithErr(err)

	percentage, err := cpu.Percent(0, true)
	dealwithErr(err)

	// host or machine kernel, uptime, platform Info
	hostStat, err := host.Info()
	dealwithErr(err)

	// get interfaces MAC/hardware address
	interfStat, err := net.Interfaces()
	dealwithErr(err)

	node.RuntimeOS = runtimeOS
	node.TotalMemory = strconv.FormatUint(vmStat.Total, 10)
	node.FreeMemory = strconv.FormatUint(vmStat.Free, 10)
	node.PercentUsedMemory = strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64)

	// get disk serial number.... strange... not available from disk package at compile time
	// undefined: disk.GetDiskSerialNumber
	serial := disk.GetDiskSerialNumber("/dev/sda")

	node.DiskSerialNumber = serial

	node.TotalDiskSpace = strconv.FormatUint(diskStat.Total, 10)
	node.UsedDiskSpace = strconv.FormatUint(diskStat.Used, 10)
	node.FreeDiskSpace = strconv.FormatUint(diskStat.Free, 10)
	node.PercentDiskSpaceUsed = strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64)

	// since my machine has one CPU, I'll use the 0 index
	// if your machine has more than 1 CPU, use the correct index
	// to get the proper data
	for _, cpu := range cpuStat {
		var cpuinfo CPUInfo

		cpuinfo.IndexNumber = strconv.FormatInt(int64(cpu.CPU), 10)
		cpuinfo.VendorId = cpu.VendorID
		cpuinfo.Family = cpu.Family
		cpuinfo.NumberOfCores = strconv.FormatInt(int64(cpu.Cores), 10)
		cpuinfo.ModelName = cpu.ModelName
		cpuinfo.Speed = strconv.FormatFloat(cpu.Mhz, 'f', 2, 64)

		node.CpuInfo = append(node.CpuInfo, cpuinfo)
	}

	for idx, cpupercent := range percentage {
		node.CpuInfo[idx].PercentUsed = strconv.FormatFloat(cpupercent, 'f', 2, 64)
	}

	node.HostName = hostStat.Hostname
	node.Uptime = strconv.FormatUint(hostStat.Uptime, 10)
	node.NumberOfProcessesRunning = strconv.FormatUint(hostStat.Procs, 10)

	// another way to get the operating system name
	// both darwin for Mac OSX, For Linux, can be ubuntu as platform
	// and linux for OS

	node.OperatingSystem = hostStat.OS
	node.Plateform = hostStat.Platform

	// the unique hardware id for this machine
	node.HostId = hostStat.HostID

	for _, interf := range interfStat {
		var networkInfo NetworkInfo

		networkInfo.InterfaceName = interf.Name

		if interf.HardwareAddr != "" {
			networkInfo.MacAddress = interf.HardwareAddr
		}

		for _, flag := range interf.Flags {
			networkInfo.InterfaceBehavior = flag
		}

		for _, addr := range interf.Addrs {
			networkInfo.IpAddress = addr.String()
		}

		node.NetworkInfo = append(node.NetworkInfo, networkInfo)
	}
}

func getConfiguration() Configuration {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	return configuration
}

func getAPIURL() string {
	return getConfiguration().URL
}

func main() {
	config := getConfiguration()
	fmt.Println(config)
	if config.WaitTime > 0 {
		for {
			fmt.Println("Start Processing...")
			processInfo()
			time.Sleep(time.Second * time.Duration(config.WaitTime))
		}
	} else {
		processInfo()
	}
}

func processInfo() {
	var info InfoBase
	//var nodeInfo NodeInfo
	GetHardwareData(&info.NodeInfo)

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		containerInfo, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			panic(err)
		}

		info.ContainerInfo = append(info.ContainerInfo, containerInfo)
	}

	jsonValue, _ := json.Marshal(info)
	fmt.Println(string(jsonValue))

	_, err = http.Post(getAPIURL(), "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		// data, _ := ioutil.ReadAll(response.Body)
		// fmt.Println(string(data))
	}
}
