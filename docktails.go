package main

// Docktails outputs docker container logs without exiting when the container is restarted.

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"golang.org/x/term"
)

const (
	defaultLogCount = 5
	logBufSize      = 4096
)

func selectContainer(cli *client.Client) (string, error) {
	fmt.Println("Select a container:")
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return "", err
	}

	for idx, container := range containers {
		fmt.Printf("[%d] %s\n", idx+1, container.Names[0][1:])
	}

	fmt.Print("Enter the number of the container to tail logs: ")
	var selection int
	_, err = fmt.Scanln(&selection)
	if err != nil {
		return "", err
	}

	if selection < 1 || selection > len(containers) {
		return "", fmt.Errorf("invalid selection")
	}

	return containers[selection-1].Names[0][1:], nil
}

func printColorizedLog(log string) {
	lines := strings.Split(strings.TrimSuffix(log, "\n"), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ERROR") || strings.Contains(line, "error") {
			color.Red(line)
		} else if strings.Contains(line, "WARNING") || strings.Contains(line, "warn") {
			color.Yellow(line)
		} else {
			color.White(line)
		}
	}
}

func printTitleBar(title string, barWidth int) {
	bar := strings.Repeat("-", barWidth)
	fmt.Println(bar)
	fmt.Printf("| %s |\n", title)
	fmt.Println(bar)
}

func main() {
	// Initialize Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		fmt.Println("Error creating Docker client:", err)
		os.Exit(1)
	}

	var containerName string

	// Select or input container name
	if len(os.Args) == 2 {
		containerName = os.Args[1]
	} else {
		containerName, err = selectContainer(cli)
		if err != nil {
			fmt.Println("Error listing containers:", err)
			os.Exit(1)
		}
	}

	// Print the selected container name
	fmt.Printf("Tailing logs for container: %s\n", containerName)

	// Create context and cancel function for graceful exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set options for fetching container logs
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "all",
	}

	// Get logs stream
	logs, err := cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		fmt.Println("Error getting container logs:", err)
		os.Exit(1)
	}
	defer logs.Close()

	// Set up interrupt signal handling
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Create a goroutine to handle the interrupt signal
	go func() {
		<-signalCh
		fmt.Println("Received interrupt signal. Stopping docktails...")
		cancel()
	}()

	// Set up variables for log buffering and title bar width
	buf := make([]byte, logBufSize)
	firstIteration := true
	lines := []string{}

	termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Error getting terminal size:", err)
		os.Exit(1)
	}

	barWidth := termWidth
	if len(containerName)+4 < termWidth {
		barWidth = len(containerName) + 4
	}

	// Main loop for reading and displaying logs
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := logs.Read(buf)
			if err != nil {
				return
			}
			data := string(buf[:n])

			if firstIteration {
				lines = append(lines, strings.Split(data, "\n")...)
				if len(lines) > defaultLogCount {
					lines = lines[len(lines)-defaultLogCount:]
				}
				firstIteration = false
			} else {
				lines = append(lines, strings.Split(data, "\n")...)
			}

			if len(lines) > 0 {
				if len(lines) > defaultLogCount {
					lines = lines[len(lines)-defaultLogCount:]
				}
				printTitleBar(containerName, barWidth)
				printColorizedLog(strings.Join(lines, "\n"))
			}
		}
	}
}
