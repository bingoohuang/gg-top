package main

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// ExtractType defines how to extract info.
type ExtractType int

const (
	// ExtractWhole extracts value as a whole
	ExtractWhole ExtractType = iota
	// ExtractValueKey extracts [value key] patterns.
	ExtractValueKey
	// ExtractTable extracts a text table info.
	ExtractTable
)

// ExtractConfig defines configuration for config.
type ExtractConfig struct {
	Start    string
	End      string
	Type     ExtractType
	Names    []string
	Includes []string
	Excludes []string
	SortBy   string
}

func (c ExtractConfig) capture(s string) (string, bool) {
	p := strings.Index(s, c.Start)
	if p < 0 {
		return "", false
	}

	var t string
	// 表格，退回到换行处开始，表格为了计算左对齐右对齐，需要保留空格
	if c.Type == ExtractTable {
		if lastNewline := strings.LastIndex(s[:p], "\n"); lastNewline >= 0 {
			t = s[lastNewline+1:]
		} else {
			t = s
		}
	} else {
		t = s[p+len(c.Start):]
	}

	if c.End != "" {
		q := strings.Index(t, c.End)
		if q < 0 {
			return "", false
		}
		t = t[:q]
	}

	if c.Type == ExtractTable {
		return t, true
	}

	return strings.TrimFunc(t, func(r rune) bool {
		return unicode.IsSpace(r) || r == ':' || r == ','
	}), true
}

var reNum = regexp.MustCompile(`[\d.]+`)

var reValueKey = regexp.MustCompile(`(\w+)\s+(\w+)`)

// ExtractTop extracts top output.
func ExtractTop(timestamp, s string) (fields []string, result string) {
	return ExtractTopWithConfig(timestamp, s, extractConfig)
}

var linuxExtractConfig = []ExtractConfig{
	{Start: "Load Avg:", End: "\n", Type: ExtractWhole, Names: []string{"load1", "load5", "load15"}},
	{Start: "MemRegions", End: "\n", Type: ExtractValueKey},
	{Start: "PID ", Type: ExtractTable, Includes: []string{"COMMAND", "MEM", "%CPU"}, SortBy: "PID"},
}

var darwinExtractConfig = []ExtractConfig{
	{Start: "Load Avg:", End: "\n", Type: ExtractWhole, Names: []string{"load1", "load5", "load15"}},
	{Start: "MemRegions", End: "\n", Type: ExtractValueKey},
	{Start: "PID ", Type: ExtractTable, Includes: []string{"COMMAND", "MEM", "%CPU"}, SortBy: "PID"},
}

// ExtractTopWithConfig extracts top output.
func ExtractTopWithConfig(timestamp, s string, configs []ExtractConfig) (fields []string, result string) {
	fields = []string{"timestamp"}
	result = "['" + timestamp + "'"
	for _, c := range configs {
		switch c.Type {
		case ExtractWhole:
			t, ok := c.capture(s)
			if !ok {
				continue
			}
			fields = append(fields, c.Names...)
			result += "," + t
		case ExtractValueKey:
			t, ok := c.capture(s)
			if !ok {
				continue
			}

			for _, item := range reValueKey.FindAllStringSubmatch(t, -1) {
				result += "," + wrap(item[1])
				if len(c.Names) == 0 {
					fields = append(fields, item[2])
				}
			}

			fields = append(fields, c.Names...)
		case ExtractTable:
			t, ok := c.capture(s)
			if !ok {
				continue
			}

			result, fields = c.extractTable(t, result, fields)
		}
	}

	return fields, result + "]"
}

func (c ExtractConfig) extractTable(t, result string, fields []string) (string, []string) {
	p := strings.Index(t, "\n")
	header := t[:p]
	t = t[p+1:]

	sortBy := c.SortBy
	headerColumns := strings.Fields(header)
	fieldIndices := make([]int, len(headerColumns))

	var left int
	for i, col := range headerColumns {
		j := strings.Index(header, col)
		header = header[j+len(col):]
		if i > 0 {
			fieldIndices[i] = left + j
		}
		left += j + len(col)
	}

	headerMap := map[string]int{}
	for i, col := range headerColumns {
		headerMap[col] = i
		if sortBy == "" {
			sortBy = col
		}
	}

	includes := make(map[string]bool)
	for _, include := range c.Includes {
		includes[include] = true
	}
	excludes := make(map[string]bool)
	for _, exclude := range c.Excludes {
		excludes[exclude] = true
	}

	var includeFunc func(col string) bool
	if len(includes) > 0 {
		includeFunc = func(col string) bool {
			return includes[col]
		}
	} else if len(excludes) > 0 {
		includeFunc = func(col string) bool {
			return !excludes[col]
		}
	} else {
		includeFunc = func(col string) bool {
			return true
		}
	}

	var sortValues []string
	sortLines := map[string][]string{}

	for _, line := range strings.Split(t, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fs := strings.Fields(line)
		if len(fs) > len(headerColumns) {
			fs = nil
			for i := 0; i < len(headerColumns); i++ {
				var v string
				if i+1 < len(headerColumns) {
					v = line[fieldIndices[i]:fieldIndices[i+1]]
				} else {
					v = line[fieldIndices[i]:]
				}

				fs = append(fs, strings.TrimSpace(v))
			}
		}

		sortField := fs[headerMap[sortBy]]
		sortValues = append(sortValues, sortField)
		sortLines[sortField] = fs
	}

	sort.Strings(sortValues)

	for _, sortValue := range sortValues {
		fs := sortLines[sortValue]
		fp := fs[headerMap[sortBy]] + "-"
		for i, f := range fs {
			if includeFunc(headerColumns[i]) {
				result += "," + wrap(f)
			}
		}

		for i, f := range headerColumns {
			if includeFunc(f) {
				fields = append(fields, fp+headerColumns[i])
			}
		}
	}
	return result, fields
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
