package main

import (
	//"fmt"
	//"maunium.net/go/mautrix"
	//event "maunium.net/go/mautrix/event"
	id "maunium.net/go/mautrix/id"
	"strconv"
	"time"
	//"strings"
	ps_mem "github.com/shirou/gopsutil/v3/mem"
	ps_host "github.com/shirou/gopsutil/v3/host"
	ps_cpu "github.com/shirou/gopsutil/v3/cpu"
	ps_load "github.com/shirou/gopsutil/v3/load"
	ps_net "github.com/shirou/gopsutil/v3/net"
	ps_disk "github.com/shirou/gopsutil/v3/disk"
	human "github.com/dustin/go-humanize"
)


type ServerStatus struct {
	StartTime time.Time
}

func NewServerStatus() *ServerStatus {
    return &ServerStatus{
    	StartTime: 	time.Now(),
    }
}

const Year uint64 = 365.2425 * 24 * 60 * 60
const Month uint64 = 30.436875 * 24 * 60 * 60
const Day uint64 = 24 * 60 * 60
const Hour uint64 = 60 * 60
const Minute uint64 = 60
func humaniseUptime(uptime uint64) string {
	timeleft := uptime

	Years := timeleft / Year
	timeleft -= Years * Year

	Months := timeleft / Month
	timeleft -= Months * Month

	Days := timeleft / Day
	timeleft -= Days * Day

	Hours := timeleft / Hour
	timeleft -= Hours * Hour

	Minutes := timeleft / Minute
	timeleft -= Minutes * Minute

	Seconds := timeleft

	uptime_str := ""
	if Years > 0 {
		uptime_str += strconv.FormatUint(Years, 10) + " Years "
	}
	if Months > 0 {
		uptime_str += strconv.FormatUint(Months, 10) + " Months "
	}
	if Days > 0 {
		uptime_str += strconv.FormatUint(Days, 10) + " Days "
	}
	if Hours > 0 {
		uptime_str += strconv.FormatUint(Hours, 10) + " Hours "
	}
	if Minutes > 0 {
		uptime_str += strconv.FormatUint(Minutes, 10) + " Minutes "
	}

	return uptime_str + strconv.FormatUint(Seconds, 10) + " Seconds"
}

func (ss *ServerStatus) generateStatusMessage() string {
	uptime_str := ""
	uptime, err := ps_host.Uptime()
	if err != nil {
		uptime_str = "Error getting Uptime\n"
	} else {
		uptime_str = humaniseUptime(uptime) + "\n"
	}

	// Bot Uptime
	uptimebot_str := humaniseUptime(uint64(time.Now().Unix() - ss.StartTime.Unix())) + "\n"

	memory_str := ""
	memory, err := ps_mem.VirtualMemory()
	if err != nil {
		memory_str = "Error getting Memory\n"
	} else {
		memory_str = "Total -> " + human.IBytes(memory.Total) + " - Used -> " + human.IBytes(memory.Used) +
					" - Used -> " + strconv.FormatFloat(memory.UsedPercent, 'f', 2, 64) + "% - Free -> " + human.IBytes(memory.Free) + "\n"
	}

	swap_str := ""
	swap, err := ps_mem.SwapMemory()
	if err != nil {
		swap_str = "Error getting SWAP\n"
	} else {
		swap_str = "Total -> " + human.IBytes(swap.Total) + " - Used -> " + human.IBytes(swap.Used) +
					" - Used -> " + strconv.FormatFloat(swap.UsedPercent, 'f', 2, 64) + "% - Free -> " + human.IBytes(swap.Free) + "\n"
	}

	cpu_str := ""
	cpu, err1 := ps_cpu.Counts(true)
	loadavg, err2 := ps_load.Avg()
	if err1 != nil || err2 != nil {
		cpu_str = "Error getting CPU Info\n"
	} else {
		cpu_str = 	"Count: " 	+ strconv.Itoa(cpu) + 
					" - Loadavg 1m: " 	+ strconv.FormatFloat(loadavg.Load1, 'f', 2, 64) +
					" - Loadavg 5m: " + strconv.FormatFloat(loadavg.Load5, 'f', 2, 64) + 
					" - Loadavg 15m: " 	+ strconv.FormatFloat(loadavg.Load15, 'f', 2, 64) + 
		"\n"
	}

	network_str := ""
	network, err := ps_net.IOCounters(false)
	if err != nil {
		network_str = "Error getting network interfaces\n"
	} else {
		network_str +=  
				" - RX: " + human.IBytes(network[0].BytesRecv) +
				" - TX: " + human.IBytes(network[0].BytesSent) + 
		"\n"
	}

	disk_str := ""
	disks, err := ps_disk.Partitions(false)
	if err != nil {
		disk_str = "Error getting disk interfaces\n"
	} else {
		disk_str += "Device | Filesystem | Total | Free | Used | Used% | Mount"
		for _, partition := range disks {
			disk_str += "\n\t" + partition.Device + " | " + partition.Fstype
			partinfo, err2 := ps_disk.Usage(partition.Mountpoint)
			if err2 != nil {
				disk_str += " | Could not query details!"
			} else {
				disk_str += " | " + human.IBytes(partinfo.Total) + " | " +
									human.IBytes(partinfo.Free) + " | " + 
									human.IBytes(partinfo.Used) + " | " + 
									strconv.FormatFloat(partinfo.UsedPercent, 'f', 2, 64)
			}
			disk_str += " | " + partition.Mountpoint
		}
	}

	finalstring := "Status as of " + time.Now().Format(time.RFC1123) + ":\n"
	finalstring += "Host Uptime: " + uptime_str
	finalstring += "Bot Uptime: " + uptimebot_str
	finalstring += "Mem: " + memory_str
	finalstring += "Swap: " + swap_str
	finalstring += "CPU: " + cpu_str
	finalstring += "Network: " + network_str
	finalstring += "Disks: " + disk_str

	return finalstring
}


func HandleServerStatusHandler(boottime time.Time) CommandCallback {
	serverstatus := NewServerStatus()
	if !boottime.IsZero() {
		serverstatus.StartTime = boottime
	}

	return func(ca CommandArgs) BotReply {
		return BotPrintSimple(ca.room, serverstatus.generateStatusMessage())
	}
}

func CronServerStatusHandler(cmdhdlr *CommandHandler, boottime time.Time, room id.RoomID) func() {
	serverstatus := NewServerStatus()
	if !boottime.IsZero() {
		serverstatus.StartTime = boottime
	}
	targetroom := room

	return func() {
		cmdhdlr.BotPrint(BotPrintSimple(targetroom, serverstatus.generateStatusMessage()))
	}
}