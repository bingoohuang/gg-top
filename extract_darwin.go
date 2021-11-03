package main

var extractConfig = darwinExtractConfig

func topCmd(pids []string) string {
	cmd := `top -l 1 -F`
	for _, pid := range pids {
		cmd += ` -pid ` + pid
	}

	return cmd
}
