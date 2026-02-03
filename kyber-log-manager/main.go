package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const containerLogPath = "/root/.local/share/maxima/wine/prefix/drive_c/users/root/AppData/Roaming/ArmchairDevelopers/Kyber/Logs"

// function to check if Docker is installed on the host system
func dockerExists() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// function to check if specified Docker container exists
func containerExists(container string) (bool, error) {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == container {
			return true, nil
		}
	}

	return false, nil
}

// function to check if user specified Docker container is running
func containerRunning(container string) (bool, error) {
	cmd := exec.Command(
		"docker", "inspect",
		"-f", "{{.State.Running}}",
		container,
	)

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) == "true", nil
}

func listLogFiles(container string) ([]string, error) {
	cmd := exec.Command(
		"docker", "exec", container,
		"sh", "-c",
		fmt.Sprintf("ls -1 %s/*.log 2>/dev/null || true", containerLogPath),
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	lines := strings.Split(outputStr, "\n")
	var files []string

	for _, line := range lines {
		if line == "" {
			continue
		}
		files = append(files, filepath.Base(line))
	}

	return files, nil
}

// parses user input on which log files to output
func parseSelection(input string, files []string) ([]string, error) {
	parts := strings.Split(input, ",")
	seen := make(map[int]bool)
	var selected []string

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// handle range (e.g. 2-5)
		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}

			start, err1 := strconv.Atoi(strings.TrimSpace(bounds[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(bounds[1]))

			if err1 != nil || err2 != nil || start < 1 || end > len(files) || start > end {
				return nil, fmt.Errorf("invalid range: %s", part)
			}

			for i := start; i <= end; i++ {
				if !seen[i] {
					seen[i] = true
					selected = append(selected, files[i-1])
				}
			}
			continue
		}

		// handle single number
		index, err := strconv.Atoi(part)
		if err != nil || index < 1 || index > len(files) {
			return nil, fmt.Errorf("invalid number: %s", part)
		}

		if !seen[index] {
			seen[index] = true
			selected = append(selected, files[index-1])
		}
	}

	return selected, nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	// checks to see if docker is installed on the host system
	if !dockerExists() {
		fmt.Fprintln(os.Stderr, "Error: Docker is not installed or not in PATH")
		os.Exit(1)
	}

	// Get container name
	fmt.Print("Enter container name: ")
	containerName, _ := reader.ReadString('\n')
	containerName = strings.TrimSpace(containerName)

	// Check if container name passed is empty
	if containerName == "" {
		fmt.Println("Error: Container name cannot be empty")
		os.Exit(1)
	}

	// Check if the user specified docker container exists
	exists, err := containerExists(containerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error checking docker containers:", err)
		os.Exit(1)
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Container '%s' does not exist\n", containerName)
		os.Exit(1)
	}

	// Check if user specified docker container is running
	running, err := containerRunning(containerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error checking container state:", err)
		os.Exit(1)
	}
	if !running {
		fmt.Fprintf(
			os.Stderr,
			"Error: Container '%s' exists but is not running.\nStart it with: docker start %s\n",
			containerName,
			containerName,
		)
		os.Exit(1)
	}

	// Get list of .log files
	logFiles, err := listLogFiles(containerName)
	if err != nil {
		fmt.Println("Error listing log files:", err)
		os.Exit(1)
	}

	// Exits if no log files are found
	if len(logFiles) == 0 {
		fmt.Println("No .log files found in container.")
		return
	}

	// Prints the list of log files found
	fmt.Println("\nLog files found:")
	for i, file := range logFiles {
		fmt.Printf("%d) %s\n", i+1, file)
	}

	// Get selected files from the user
	fmt.Print("\nSelect log files to extract (e.g. 1, 1-3, 1-3,5): ")
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	// Gets parsed log files by user selection via a function
	selectedFiles, err := parseSelection(selection, logFiles)
	if err != nil {
		fmt.Println("Invalid selection:", err)
		os.Exit(1)
	}

	// User selects destination directory
	fmt.Print("\nDestination directory (leave empty for current directory): ")
	destDir, _ := reader.ReadString('\n')
	destDir = strings.TrimSpace(destDir)

	if destDir == "" {
		destDir, _ = os.Getwd()
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Println("Error: Failed to create destination directory:", err)
		os.Exit(1)
	}

	// Copy files from Docker container to host system
	for _, file := range selectedFiles {
		src := fmt.Sprintf("%s:%s/%s", containerName, containerLogPath, file)
		dst := filepath.Join(destDir, file)

		cmd := exec.Command("docker", "cp", src, dst)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to copy %s: %v\n", file, err)
			continue
		}

		fmt.Printf("Copied %s\n", file)
	}

	fmt.Println("\nDone.")
}
