package process

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	tea "charm.land/bubbletea/v2"
	"github.com/shirou/gopsutil/v3/process"

	"samouly.fun/bubbleMonitor/internal/data"
	"samouly.fun/bubbleMonitor/internal/msg"
)

// CachedProcessInfo stores the actual process object and its static data
type CachedProcessInfo struct {
	Proc          *process.Process
	Name          string
	Username      string
	Cmdline       string
	CmdlineLower  string
	NameLower     string
	UsernameLower string
	CreateTime    int64
	Ppid          int32
	Nice          int32
}

var (
	processCache      = make(map[int32]*CachedProcessInfo)
	cacheMutex        sync.RWMutex
	isFetching        atomic.Bool
	initialPidsLoaded atomic.Bool
)

// pruneProcessCache drops every entry whose pid is not in alivePids. Pass
// an empty map to drop everything (used on error paths so dead PIDs don't
// linger indefinitely).
func pruneProcessCache(alivePids map[int32]bool) {
	cacheMutex.Lock()
	for pid := range processCache {
		if !alivePids[pid] {
			delete(processCache, pid)
		}
	}
	cacheMutex.Unlock()
}

// PidsOnlyCmd fetches only PIDs for cache warming (lightweight)
func PidsOnlyCmd() tea.Cmd {
	return func() tea.Msg {
		pids, err := process.Pids()
		if err != nil {
			return msg.ProcessCountMsg(0)
		}
		initialPidsLoaded.Store(true)
		return msg.ProcessCountMsg(len(pids))
	}
}

// ProcessesCmd fetches running processes and sorts them
func ProcessesCmd(sortBy string, sortDirection string) tea.Cmd {
	return func() tea.Msg {
		// Prevent concurrent execution to avoid race conditions on *process.Process
		// and unbounded goroutine growth.
		if !isFetching.CompareAndSwap(false, true) {
			return nil
		}
		defer isFetching.Store(false)

		pids, err := process.Pids()
		if err != nil {
			// Prune whatever we can so a transient error doesn't keep
			// dead PIDs in the cache forever. Pass an empty alive set;
			// the next successful fetch will repopulate.
			pruneProcessCache(map[int32]bool{})
			return msg.ProcessesMsg{}
		}

		// Use slice pool to reduce allocation
		// Note: We don't Put back here. The receiver (Update loop) is responsible for recycling
		// the previous slice once it's replaced.
		procListPtr := GetProcSlice()
		procList := *procListPtr
		// Ensure capacity
		if cap(procList) < len(pids) {
			procList = make([]data.ProcessInfo, 0, len(pids))
		}

		currentPids := make(map[int32]bool, len(pids))

		for _, pid := range pids {
			currentPids[pid] = true

			cacheMutex.RLock()
			cached, exists := processCache[pid]
			cacheMutex.RUnlock()

			if !exists {
				newProc, err := process.NewProcess(pid)
				if err != nil {
					continue
				}

				name, _ := newProc.Name()
				createTime, _ := newProc.CreateTime()
				nice, _ := newProc.Nice()
				ppid, _ := newProc.Ppid()

				// Intern strings to save memory
				nameInterned := Intern(name)

				cached = &CachedProcessInfo{
					Proc:          newProc,
					Name:          nameInterned,
					Username:      "", // Lazy loaded when needed
					Cmdline:       "", // Lazy loaded when needed
					CmdlineLower:  "",
					CreateTime:    createTime,
					Nice:          nice,
					Ppid:          ppid,
					NameLower:     strings.ToLower(name),
					UsernameLower: "",
				}

				cacheMutex.Lock()
				processCache[pid] = cached
				cacheMutex.Unlock()
			}

			// Fetch dynamic data
			// CPUPercent acts on the persistent object (cached.Proc)
			// This is safe now because isFetching guarantees only one goroutine touches it at a time.
			cpuPercent, _ := cached.Proc.CPUPercent()
			memPercent, _ := cached.Proc.MemoryPercent()
			status, _ := cached.Proc.Status()
			memInfo, _ := cached.Proc.MemoryInfo()

			var memBytes uint64
			if memInfo != nil {
				memBytes = memInfo.RSS
			}

			statusStr := strings.Join(status, ",")
			if statusStr == "" {
				statusStr = "running"
			}
			statusStr = Intern(statusStr)

			procList = append(procList, data.ProcessInfo{
				Name:          cached.Name,
				Pid:           pid,
				Cpu:           cpuPercent,
				Memory:        float64(memPercent),
				Status:        statusStr,
				Username:      cached.Username,
				CreateTime:    cached.CreateTime,
				Cmdline:       cached.Cmdline,
				MemoryBytes:   memBytes,
				Nice:          cached.Nice,
				Ppid:          cached.Ppid,
				NameLower:     cached.NameLower,
				UsernameLower: cached.UsernameLower,
				CmdlineLower:  cached.CmdlineLower,
			})
		}

		// Prune dead processes
		pruneProcessCache(currentPids)

		sort.Slice(procList, func(i, j int) bool {
			var less bool
			switch sortBy {
			case "cpu":
				less = procList[i].Cpu > procList[j].Cpu
			case "memory", "mem":
				less = procList[i].Memory > procList[j].Memory
			case "pid":
				less = procList[i].Pid > procList[j].Pid
			case "name":
				// Handle dotted and parenthesized names (put them at the end)
				nameI := procList[i].NameLower
				nameJ := procList[j].NameLower

				isSpecialI := strings.HasPrefix(nameI, ".") || strings.HasPrefix(nameI, "(")
				isSpecialJ := strings.HasPrefix(nameJ, ".") || strings.HasPrefix(nameJ, "(")

				if isSpecialI != isSpecialJ {
					less = !isSpecialI // Special chars go to the end (so they are "greater")
				} else {
					less = nameI < nameJ
				}
			default:
				return false
			}

			// Invert result if descending
			if sortDirection == "desc" {
				return !less
			}
			return less
		})

		return msg.ProcessesMsg(procList)
	}
}

// ProcessCountCmd only fetches the number of running processes (Lightweight)
func ProcessCountCmd() tea.Cmd {
	return func() tea.Msg {
		pids, err := process.Pids()
		if err != nil {
			return msg.ProcessCountMsg(0)
		}
		return msg.ProcessCountMsg(len(pids))
	}
}
