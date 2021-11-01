package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

	// elastic-java-28418:a kafka-23134-java:b
	fields, result := Extract("2021-11-01T12:48", s)
	assert.Equal(t, []string{"timestamp", "load1", "load5", "load15",
		"memTotal", "memFree", "memUsed", "memBuff",
		"elastic+-28418-java-PID", "elastic+-28418-java-USER", "elastic+-28418-java-PR", "elastic+-28418-java-NI",
		"elastic+-28418-java-VIRT", "elastic+-28418-java-RES", "elastic+-28418-java-SHR", "elastic+-28418-java-S",
		"elastic+-28418-java-%CPU", "elastic+-28418-java-%MEM", "elastic+-28418-java-TIME+", "elastic+-28418-java-COMMAND",
		"kafka-23134-java-PID", "kafka-23134-java-USER", "kafka-23134-java-PR", "kafka-23134-java-NI",
		"kafka-23134-java-VIRT", "kafka-23134-java-RES", "kafka-23134-java-SHR", "kafka-23134-java-S",
		"kafka-23134-java-%CPU", "kafka-23134-java-%MEM", "kafka-23134-java-TIME+", "kafka-23134-java-COMMAND"}, fields)
	assert.Equal(t, "['2021-11-01T12:48'," +
		"0.40, 0.22, 0.22," +
		"8009040,599628,3094756,4314656," +
		"28418,'elastic+',20,0,4304336,1.0,128396,'S',0.0,13.1,'21:51.44','java'," +
		"23134,'kafka',20,0,5276184,1.1,17672,'S',0.0,14.0,'101:24.10','java']", result)
}
