package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtract(t *testing.T) {
	s := `[root@localhost ~]# top -bn1 -p 28418,23134
top - 12:39:14 up 3 days,  2:51, 10 users,  load average: 0.40, 0.22, 0.22
Tasks:   2 total,   0 running,   2 sleeping,   0 stopped,   0 zombie
%Cpu(s):  1.7 us,  5.0 sy,  0.0 ni, 93.3 id,  0.0 wa,  0.0 hi,  0.0 si,  0.0 st
KiB Mem :  8009040 total,   599628 free,  3094756 used,  4314656 buff/cache
KiB Swap:  8257532 total,  8245244 free,    12288 used.  4357044 avail Mem

  PID USER      PR  NI    VIRT    RES    SHR S  %CPU %MEM     TIME+ COMMAND
28418 elastic+  20   0 4304336   1.0g 128396 S   0.0 13.1  21:51.44 java
23134 kafka     20   0 5276184   1.1g  17672 S   0.0 14.0 101:24.10 java`

	configs := []ExtractConfig{
		{Start: "load average", End: "\n", Type: ExtractWhole, Names: []string{"load1", "load5", "load15"}},
		{Start: "KiB Mem", End: "\n", Type: ExtractValueKey, Names: []string{"memTotal", "memFree", "memUsed", "memBuff"}},
		{Start: "PID ", Type: ExtractTable, Excludes: []string{"PR", "NI", "S", "USER"}, SortBy: "PID"},
	}

	fields, result := ExtractTopWithConfig("2021-11-01T12:48", s, configs, true)
	assert.Equal(t, []string{
		"timestamp", "load1", "load5", "load15",
		"memTotal", "memFree", "memUsed", "memBuff",
		"23134-PID",
		"23134-VIRT", "23134-RES", "23134-SHR",
		"23134-%CPU", "23134-%MEM", "23134-TIME+", "23134-COMMAND",
		"28418-PID",
		"28418-VIRT", "28418-RES", "28418-SHR",
		"28418-%CPU", "28418-%MEM", "28418-TIME+", "28418-COMMAND",
	}, fields)
	assert.Equal(t, "[\"2021-11-01T12:48\","+
		"0.40, 0.22, 0.22,"+
		"8009040,599628,3094756,4314656,"+
		"23134,5276184,1153433.6,17672,0.0,14.0,\"101:24.10\",\"java\","+
		"28418,4304336,1048576,128396,0.0,13.1,\"21:51.44\",\"java\"]",
		result)

	mac := `# top -l 1 -F -pid 99921 -pid 69330
Processes: 600 total, 2 running, 598 sleeping, 3162 threads
2021/11/02 14:34:29
Load Avg: 2.61, 3.03, 3.25
CPU usage: 4.64% user, 10.28% sys, 85.7% idle
MemRegions: 249570 total, 4457M resident, 0B private, 2568M shared.
PhysMem: 16G used (2669M wired), 136M unused.
VM: 12T vsize, 0B framework vsize, 22281275(0) swapins, 23914209(0) swapouts.
Networks: packets: 26981829/27G in, 22854927/25G out.
Disks: 17344532/467G read, 11675489/375G written.

PID    COMMAND          %CPU TIME     #TH #WQ #PORTS MEM  PURG  CMPRS PGRP  PPID  STATE    BOOSTS %CPU_ME %CPU_OTHRS UID FAULTS  COW  MSGSENT  MSGRECV  SYSBSD  SYSMACH   CSW     PAGEINS IDLEW  POWER INSTRS CYCLES USER       #MREGS RPRVT VPRVT VSIZE KPRVT KSHRD
99921  Google Chrome He 0.0  00:54.07 15  1   144    110M 0B    99M   57371 57371 sleeping *0[5]  0.00000 0.00000    501 171850  2308 129683   61074    221189  429923    254106  860     3821   0.0   0      0      bingoobjca N/A    N/A   N/A   N/A   N/A   N/A
69330  WeChat           0.0  15:51.88 37  9   49211  288M 7444K 137M  69330 69294 sleeping *5[15] 0.00000 0.00000    501 7275450 3595 72299264 32956646 8521736 117179313 9048963 7472    517256 0.0   0      0      bingoobjca N/A    N/A   N/A   N/A   N/A   N/A
`

	fields, result = ExtractTopWithConfig("2021-11-01T12:48", mac, []ExtractConfig{
		{Start: "Load Avg:", End: "\n", Type: ExtractWhole, Names: []string{"load1", "load5", "load15"}},
		{Start: "MemRegions", End: "\n", Type: ExtractValueKey},
		{Start: "PID ", Type: ExtractTable, Includes: []string{"COMMAND", "MEM", "%CPU"}, SortBy: "PID"},
	}, true)
	assert.Equal(t, []string{"timestamp", "load1", "load5", "load15", "total", "resident", "private", "shared", "69330-COMMAND", "69330-%CPU", "69330-MEM", "99921-COMMAND", "99921-%CPU", "99921-MEM"}, fields)
	assert.Equal(t, "[\"2021-11-01T12:48\",2.61, 3.03, 3.25,249570,4563968,0,2629632,\"WeChat\",0.0,294912,\"Google Chrome He\",0.0,112640]", result)
}

func TestWrap(t *testing.T) {
	assert.Equal(t, "1048576", wrap("1.0g", true))
	assert.Equal(t, "1153433.6", wrap("1.1g", true))
	assert.Equal(t, "1.0", wrap("1.0g", false))
	assert.Equal(t, "1.1", wrap("1.1g", false))
}
