package metrics

import (
	"encoding/json"
	"fmt"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Statistics struct {
	HostInfo          *host.InfoStat         `json:"host_info"`
	LoadAvg           *load.AvgStat          `json:"load_avg"`
	VirtualMemoryStat *mem.VirtualMemoryStat `json:"virtual_memory"`
	SwapMemoryStat    *mem.SwapMemoryStat    `json:"swap_memory"`
	DiskUsage         *disk.UsageStat        `json:"disk_usage"`
	DiskPartition     []disk.PartitionStat   `json:"disk_partition"`
	CpuInfo           []cpu.InfoStat         `json:"cpu_info"`
	CpuTimes          []cpu.TimesStat        `json:"cpu_times"`
	NetIOCounters     []net.IOCountersStat   `json:"net_io_counters"`
}

func main() {
	statis := Statistics{}
	statis.VirtualMemoryStat, _ = mem.VirtualMemory()
	statis.SwapMemoryStat, _ = mem.SwapMemory()
	statis.CpuInfo, _ = cpu.Info()
	statis.CpuTimes, _ = cpu.Times(true)
	statis.DiskUsage, _ = disk.Usage("/")
	statis.DiskPartition, _ = disk.Partitions(true)
	statis.HostInfo, _ = host.Info()
	statis.LoadAvg, _ = load.Avg()
	statis.NetIOCounters, _ = net.IOCounters(true)

	statisJson, _ := json.Marshal(statis)

	fmt.Println(string(statisJson))
}
