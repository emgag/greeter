package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/hako/durafmt"
	"github.com/logrusorgru/aurora/v3"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

const titleWidth = 8

func main() {
	title := func(t string) string {
		return fmt.Sprintf(
			"* %s%s",
			aurora.Cyan(aurora.Bold(t+":")),
			strings.Repeat(" ", titleWidth-len(t)),
		)
	}

	if info, err := host.Info(); err == nil {
		fmt.Printf("%s%s\n", title("Host"), info.Hostname)

		kernel := strings.Trim(
			fmt.Sprintf("%s %s", strings.Title(info.OS), info.KernelVersion),
			" ",
		)

		platform := strings.Trim(
			fmt.Sprintf("%s %s", strings.Title(info.Platform), info.PlatformVersion),
			" ",
		)

		fmt.Printf("%s%s (%s)\n", title("OS"), kernel, platform)

		bootime := time.Unix(int64(info.BootTime), 0)
		uptime := time.Duration(info.Uptime) * time.Second
		fmt.Printf("%s%v, %s\n", title("Uptime"), bootime, durafmt.Parse(uptime))
	} else {
		fmt.Println(aurora.Red("Error getting host info: "), err.Error())
	}

	if loadavg, err := load.Avg(); err == nil {
		cpuinfo := ""

		if info, err := cpu.Info(); err == nil {
			// count cores
			cores := make(map[string]bool)

			for _, i := range info {
				cores[i.CoreID] = true
			}

			if len(cores) > 1 {
				cpuinfo = fmt.Sprintf("%d Cores", len(cores))
			} else {
				cpuinfo = "1 Core"
			}

			cpuinfo += ": " + info[0].ModelName
		} else {
			cpuinfo = "Error getting CPU info"
		}

		fmt.Printf(
			"%s%.2f, %.2f, %.2f [%s]\n",
			title("Load"),
			loadavg.Load1,
			loadavg.Load5,
			loadavg.Load15,
			cpuinfo,
		)
	} else {
		fmt.Println(aurora.Red("Error getting loadavg averages: "), err.Error())
	}

	if parts, err := disk.Partitions(false); err == nil {
		fmt.Print(title("Storage"))

		devices := make(map[string]bool)
		var out []string

		for _, p := range parts {
			if usage, err := disk.Usage(p.Mountpoint); err == nil {
				if _, ok := devices[p.Device]; ok {
					continue
				}

				devices[p.Device] = true

				c := aurora.Green

				if usage.UsedPercent >= 98 || usage.InodesUsedPercent >= 75 {
					c = func(arg interface{}) aurora.Value {
						return aurora.Red(aurora.Bold(arg))
					}
				} else if usage.UsedPercent >= 90 || usage.InodesUsedPercent >= 50 {
					c = aurora.Red
				} else if usage.UsedPercent >= 80 || usage.InodesUsedPercent >= 25 {
					c = aurora.Brown
				}

				o := c(fmt.Sprintf(
					"%s => (%s, %.0f%%/%.0f%%, %s free)",
					p.Mountpoint,
					p.Fstype,
					usage.UsedPercent,
					usage.InodesUsedPercent,
					humanize.IBytes(usage.Free),
				)).String()

				out = append(out, o)
			} else {
				out = append(
					out,
					aurora.Red(
						fmt.Sprintf("%s => (%s)", p.Mountpoint, err),
					).String(),
				)
			}
		}

		fmt.Printf("%s\n", strings.Join(out, " "))
	} else {
		fmt.Println(aurora.Red("Error getting disk stats: "), err.Error())
	}

	if memstat, err := mem.VirtualMemory(); err == nil {
		c := aurora.Green

		if memstat.UsedPercent >= 98 {
			c = func(arg interface{}) aurora.Value {
				return aurora.Red(aurora.Bold(arg))
			}
		} else if memstat.UsedPercent >= 90 {
			c = aurora.Red
		} else if memstat.UsedPercent >= 80 {
			c = aurora.Brown
		}

		fmt.Print(
			title("Memory"),
			"RAM: ",
			c(fmt.Sprintf("%s (%.0f%%, u: %s, f: %s, s: %s, b: %s, a: %s) ",
				humanize.IBytes(memstat.Total),
				memstat.UsedPercent,
				humanize.IBytes(memstat.Used),
				humanize.IBytes(memstat.Free),
				humanize.IBytes(memstat.Shared),
				humanize.IBytes(memstat.Buffers),
				humanize.IBytes(memstat.Available),
			)),
		)

		if memstat.SwapTotal > 0 {
			swapPercent := float64(memstat.SwapTotal-memstat.SwapFree) / float64(memstat.SwapTotal)

			if swapPercent >= 98 {
				c = func(arg interface{}) aurora.Value {
					return aurora.Red(aurora.Bold(arg))
				}
			} else if swapPercent >= 90 {
				c = aurora.Red
			} else if swapPercent >= 80 {
				c = aurora.Brown
			}

			fmt.Print("Swap: ",
				c(fmt.Sprintf("%s (%.0f%%, f: %s) ",
					humanize.IBytes(memstat.SwapTotal),
					swapPercent,
					humanize.IBytes(memstat.SwapFree),
				)),
			)

		} else {
			fmt.Print("Swap: none")
		}

		fmt.Println()

	} else {
		fmt.Println(aurora.Red("Error getting memory stats: "), err.Error())
	}
}
