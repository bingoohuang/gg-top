package main

import (
	"regexp"
	"sort"
	"strings"

	"github.com/bingoohuang/gg/pkg/ss"
)

var reNum = regexp.MustCompile(`[\d.]+`)

// ExtractTop extracts top output.
func ExtractTop(timestamp, s string) (fields []string, result string) {
	key := "load average:"
	p := strings.Index(s, key)
	s = s[p+len(key):]

	p = strings.Index(s, "\n")
	loadAverage := strings.TrimSpace(s[:p])
	s = s[p+1:]

	result = "['" + timestamp + "'," + loadAverage
	fields = []string{"timestamp", "load1", "load5", "load15"}
	key = "KiB Mem"
	p = strings.Index(s, key)
	s = s[p+len(key):]
	p = strings.Index(s, "\n")
	mem := s[:p]
	s = s[p+1:]
	for _, rs := range strings.Fields(mem) {
		if reNum.MatchString(rs) {
			result += "," + rs
		}
	}

	fields = append(fields, []string{"memTotal", "memFree", "memUsed", "memBuff"}...)
	key = "PID"
	p = strings.Index(s, key)
	s = s[p:]
	p = strings.Index(s, "\n")
	header := strings.TrimSpace(s[:p])
	s = s[p+1:]
	headerColumns := strings.Fields(header)
	headerMap := map[string]int{}
	for i, c := range headerColumns {
		headerMap[c] = i
	}

	var pids []int
	pidLines := map[int][]string{}

	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fs := strings.Fields(line)
		pid := ss.ParseInt(fs[headerMap["PID"]])
		pids = append(pids, pid)
		pidLines[pid] = fs
	}

	sort.Ints(pids)

	for _, pid := range pids {
		fs := pidLines[pid]
		fp := fs[headerMap["USER"]] + "-" + fs[headerMap["PID"]] + "-" + fs[headerMap["COMMAND"]] + "-"
		for _, f := range fs {
			result += "," + wrap(f)
		}

		for i := range headerColumns {
			fields = append(fields, fp+headerColumns[i])
		}
	}

	return fields, result + "]"
}

func wrap(s string) string {
	if p := strings.Index(s, ":"); p >= 0 {
		return `'` + s + `'` // ignore time like 21:51.44
	}

	if v := reNum.FindString(s); v != "" {
		return v
	}

	return `'` + s + `'`
}
