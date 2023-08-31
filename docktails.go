package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/fatih/color"
	"golang.org/x/term"
)

const (
	defaultLogCount = 5
	logBufSize      = 4096
)

var (
	logLevels = map[string]color.Attribute{
		"error":   color.FgRed,
		"warning": color.FgYellow,
	}
)


// printFormattedLog prints log lines with customizable formatting.
func printFormattedLog(log string, logFormat LogFormat) {
	lines := strings.Split(strings.TrimSuffix(log, "\n"), "\n")
	for _, line := range lines {
		formattedLine := formatLogLine(line, logFormat)
		fmt.Println(formattedLine)
	}
}

// formatLogLine formats a log line based on the specified format.
func formatLogLine(line string, logFormat LogFormat) string {
	formattedLine := line
	if logFormat.Timestamp {
		formattedLine = "[" + getTimeStamp() + "] " + formattedLine
	}
	if logFormat.LogLevel {
		formattedLine = "[" + getLogLevel(line) + "] " + formattedLine
	}
	if logFormat.Truncate > 0 && len(formattedLine) > logFormat.Truncate {
		formattedLine = formattedLine[:logFormat.Truncate] + "..."
	}
	return formattedLine
}

// getTimeStamp returns the current timestamp in a formatted string.
func getTimeStamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// getLogLevel retrieves the log level from a log line.
func getLogLevel(line string) string {
	for level, colorCode := range logLevels {
		if strings.Contains(strings.ToLower(line), level) {
			return color.New(colorCode).Sprint(level)
		}
	}
	return "info"
}

// LogFormat contains settings for custom log formatting.
type LogFormat struct {
	Timestamp bool
	LogLevel  bool
	Truncate  int
}

// selectContainers prompts the user to select containers to tail logs from.
func selectContainers(cli *client.Client) ([]string, error) {
	fmt.Println("Select containers (comma-separated, e.g., 1,2,3):")
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}

	for idx, container := range containers {
		fmt.Printf("[%d] %s\n", idx+1, container.Names[0][1:])
	}

	fmt.Print("Enter the numbers of the containers to tail logs (comma-separated): ")
	var selectionStr string
	_, err = fmt.Scanln(&selectionStr)
	if err != nil {
		return nil, err
	}

	selections := strings.Split(selectionStr, ",")
	selectedContainers := []string{}
	for _, selection := range selections {
		idx := parseIndex(selection)
		if idx >= 1 && idx <= len(containers) {
			selectedContainers = append(selectedContainers, containers[idx-1].Names[0][1:])
		}
	}

	if len(selectedContainers) == 0 {
		return nil, fmt.Errorf("invalid selection")
	}

	return selectedContainers, nil
}

// parseIndex converts the input string to an integer index.
func parseIndex(input string) int {
	var idx int
	fmt.Sscanf(input, "%d", &idx)
	return idx
}


// readUserInput reads user input from the console.
func readUserInput() string {
	var input string
	fmt.Scanln(&input)
	return input
}


// printColorizedLog prints log lines with color highlighting based on log level.
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

// printTitleBar prints a title bar for the selected container.
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

	var containerNames []string

	// Select or input container names
	if len(os.Args) > 1 {
		containerNames = os.Args[1:]
	} else {
		containerNames, err = selectContainers(cli)
		if err != nil {
			fmt.Println("Error listing containers:", err)
			os.Exit(1)
		}
	}


	// Print the selected container names
	fmt.Printf("Tailing logs for containers: %s\n", strings.Join(containerNames, ", "))

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

	// Define log formatting options
	logFormat := LogFormat{
		Timestamp: true,
		LogLevel:  true,
		Truncate:  0, // Customize truncation length if needed
	}


	// Main loop for reading and displaying logs for selected containers
	for _, containerName := range containerNames {

		// Create a goroutine to handle each container
		go func(containerName string) {
			// Get logs stream for the current container
			logs, err := cli.ContainerLogs(ctx, containerName, options)
			if err != nil {
				fmt.Printf("Error getting logs for container %s: %v\n", containerName, err)
				return
			}
			defer logs.Close()

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

					// Use the printFormattedLog function here with custom formatting
					printFormattedLog(strings.Join(lines, "\n"), logFormat)

					// Use the printColorizedLog function here
					printColorizedLog(strings.Join(lines, "\n"))
				}
				}
			}
		}(containerName)
	}

	// Set up interrupt signal handling
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal to gracefully exit
	<-signalCh
	fmt.Println("Received interrupt signal. Stopping docktails...")
	cancel()
}
