package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	dmesgCmd := exec.Command("dmesg")
	output, err := dmesgCmd.Output()
	if err != nil {
		fmt.Println("Error running dmesg command:", err)
		return
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.Contains(line, " drop: ") {
			srcIP, dstIP, srcPort, dstPort, container := extractInfo(line)
			fmt.Printf("Source: %s:%d (Container: %s) -> Destination: %s:%d\n", srcIP, srcPort, container, dstIP, dstPort)
		}
	}
}

func extractInfo(line string) (srcIP string, dstIP string, srcPort int, dstPort int, container string) {
	re := regexp.MustCompile(`SRC=([\d.]+).*DST=([\d.]+).*SPT=(\d+) DPT=(\d+).*drop:.*\s+(whalewall-\w+-\w+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 6 {
		srcIP = matches[1]
		dstIP = matches[2]
		srcPort = atoi(matches[3])
		dstPort = atoi(matches[4])
		container = matches[5]
	}
	return
}

func atoi(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
