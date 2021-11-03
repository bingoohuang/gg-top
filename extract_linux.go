package main

import "strings"

var extractConfig = linuxExtractConfig

func topCmd(pids []string) string {
	return "top -bn1 -p " + strings.Join(pids, ",")
}
