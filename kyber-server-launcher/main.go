package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const imageName = "ghcr.io/armchairdevelopers/kyber-server:latest"

type Config struct {
	ContainerName        string
	MaximaEmail          string
	MaximaPassword       string
	KyberToken           string
	ServerName           string
	ServerDescription    string
	ServerPassword       string
	MaxPlayers           string
	MapRotation          string
	ModuleChannel        string
	GameDataPath         string
	ModFolderPath        string
	PluginFolderPath     string
	RestartUnlessStopped bool
}

func main() {

	// checks to see if docker is installed on the host system
	if !dockerExists() {
		fmt.Fprintln(os.Stderr, "Error: Docker is not installed or not in PATH")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	cfg := Config{
		ModuleChannel: "main",
	}

	fmt.Println("Kyber Dedicated Server Docker Setup")
	fmt.Println("----------------------------------")

	cfg.MaximaEmail = promptRequired(reader, "EA account email")
	cfg.MaximaPassword = promptRequired(reader, "EA account password")

	cfg.KyberToken = promptRequired(reader, "Kyber token")
	cfg.ServerName = promptRequired(reader, "Server name")
	cfg.ServerDescription = promptOptional(reader, "Server description (optional, leave blank for none)")
	cfg.ServerPassword = promptOptional(reader, "Server password (optional, leave blank for none)")
	cfg.MaxPlayers = promptRequired(reader, "Max players")
	cfg.MapRotation = promptRequired(reader, "Map rotation BASE64 string")

	module := promptOptional(reader, "Kyber module channel (default: main)")
	if module != "" {
		cfg.ModuleChannel = module
	}

	cfg.GameDataPath = promptRequired(reader, "Path to game data on host")

	cfg.ModFolderPath = promptOptional(reader, "Path to mod folder on host (leave blank if not using mods)")
	cfg.PluginFolderPath = promptOptional(reader, "Path to plugin folder on host (leave blank if not using plugins)")

	cfg.ContainerName = promptContainerName(reader, "Docker container name (no spaces, use - or _)")

	cfg.RestartUnlessStopped = promptYesNo(reader, "Automatically restart container unless stopped? (y/n)")

	command := buildDockerCommand(cfg)

	fmt.Println("\nWhat would you like to do?")
	fmt.Println("1) Run the command")
	fmt.Println("2) Save the command to a file")
	fmt.Println("3) Run the command and save it to a file")
	fmt.Println("4) Print the command only")

	choice := promptRequired(reader, "Select an option (1-4)")

	switch choice {
	case "1":
		runCommand(command)
	case "2":
		saveCommand(reader, command)
	case "3":
		saveCommand(reader, command)
		runCommand(command)
	case "4":
		fmt.Println("\nDocker command:\n")
		fmt.Println(command)
	default:
		fmt.Println("Invalid option.")
	}
}

// function to check if Docker is installed on the host system
func dockerExists() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func buildDockerCommand(cfg Config) string {
	var parts []string

	parts = append(parts, "docker run -it")

	parts = append(parts, "--name", cfg.ContainerName)

	if cfg.RestartUnlessStopped {
		parts = append(parts, "--restart=unless-stopped")
	}

	maximaCreds := cfg.MaximaEmail + ":" + cfg.MaximaPassword

	parts = append(parts,
		"-e MAXIMA_CREDENTIALS="+quote(maximaCreds),
		"-e KYBER_TOKEN="+cfg.KyberToken,
		"-e KYBER_SERVER_NAME="+quote(cfg.ServerName),
		"-e KYBER_SERVER_MAX_PLAYERS="+cfg.MaxPlayers,
		"-e KYBER_MAP_ROTATION="+cfg.MapRotation,
	)

	if cfg.ModuleChannel != "main" {
		parts = append(parts, "-e KYBER_MODULE_CHANNEL="+cfg.ModuleChannel)
	}

	if cfg.ServerDescription != "" {
		parts = append(parts, "-e KYBER_SERVER_DESCRIPTION="+quote(cfg.ServerDescription))
	}

	if cfg.ServerPassword != "" {
		parts = append(parts, "-e KYBER_SERVER_PASSWORD="+quote(cfg.ServerPassword))
	}

	parts = append(parts,
		"-v "+quote(cfg.GameDataPath+":/mnt/battlefront"),
	)

	if cfg.ModFolderPath != "" {
		parts = append(parts,
			"-v "+quote(cfg.ModFolderPath)+":/mnt/battlefront/mods",
			"-e KYBER_MOD_FOLDER=/mnt/battlefront/mods",
		)
	}

	if cfg.PluginFolderPath != "" {
		parts = append(parts,
			"-v "+quote(cfg.PluginFolderPath)+":/mnt/battlefront/plugins",
			"-e KYBER_SERVER_PLUGINS_PATH=/mnt/battlefront/plugins",
		)
	}

	parts = append(parts, imageName)

	return strings.Join(parts, " ")
}

func runCommand(cmd string) {
	fmt.Println("\nRunning command...\n")
	command := exec.Command("/bin/sh", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if err := command.Run(); err != nil {
		fmt.Println("Error running docker command:", err)
	}
}

func saveCommand(reader *bufio.Reader, cmd string) {
	path := promptRequired(reader, "Enter file path to save command")
	err := os.WriteFile(path, []byte(cmd+"\n"), 0644)
	if err != nil {
		fmt.Println("Failed to save file:", err)
		return
	}
	// sets the new file as executable
	os.Chmod(path, 0755)
	fmt.Println("Command saved to", path)
}

func promptRequired(reader *bufio.Reader, label string) string {
	for {
		fmt.Print(label + ": ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text != "" {
			return text
		}
		fmt.Println("This value is required.")
	}
}

func promptOptional(reader *bufio.Reader, label string) string {
	fmt.Print(label + ": ")
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func promptYesNo(reader *bufio.Reader, label string) bool {
	for {
		fmt.Print(label + ": ")
		text, _ := reader.ReadString('\n')
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter y or n.")
		}
	}
}

func promptContainerName(reader *bufio.Reader, label string) string {
	for {
		fmt.Print(label + ": ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "" {
			fmt.Println("Container name is required.")
			continue
		}

		if strings.Contains(text, " ") {
			fmt.Println("Container name cannot contain spaces. Use '-', '_', or '.' instead.")
			continue
		}

		for _, r := range text {
			if !(r >= 'a' && r <= 'z' ||
				r >= '0' && r <= '9' ||
				r == '-' || r == '_' || r == '.') {
				fmt.Println("Container name must be lowercase and may only contain a-z, 0-9, '-', '_' and '.'")
				goto retry
			}
		}

		return text
	retry:
	}
}

func quote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
