package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// hard-coded link
const kyberDownloadURL = "https://github.com/LevelDreadnought/Kyber/raw/refs/heads/ver/beta10/Module/Kyber.dll"

// help prompt
func usage() {
	fmt.Println("Usage: kyber-updater [-v] [-c <container_name>] [-f <file_name>] [-d] [-h | --help]")
	fmt.Println("  -v         Enable verbose mode")
	fmt.Println("  -c         Specify a docker container name")
	fmt.Println("  -f         Specify input file (default: Kyber.dll)")
	fmt.Println("  -d         Download latest Kyber.dll instead of using a local file")
	fmt.Println("  -h, --help Show this help message and exit")
}

// runs specified command in bash shell in current directory
func runCommand(verbose bool, name string, args ...string) error {
	if verbose {
		fmt.Printf("Running: %s %s\n", name, strings.Join(args, " "))
	}

	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

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
		if line == container {
			return true, nil
		}
	}

	return false, nil
}

// downloads current Kyber.dll file from GitHub
func downloadFile(url, destination string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// main function
func main() {
	var (
		verbose       bool
		containerName string
		fileName      string
		download      bool
		showHelp      bool
	)

	// Support --help flag manually
	for _, arg := range os.Args[1:] {
		if arg == "--help" {
			usage()
			os.Exit(0)
		}
	}

	flag.BoolVar(&verbose, "v", false, "Enable verbose mode")
	flag.StringVar(&containerName, "c", "", "Docker container name")
	flag.StringVar(&fileName, "f", "Kyber.dll", "Input file name")
	flag.BoolVar(&download, "d", false, "Download Kyber.dll from hard-coded URL")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.Parse()

	// -h prints usage instructions
	if showHelp {
		usage()
		os.Exit(0)
	}

	// checks that a mandatory docker container name is passed
	if containerName == "" {
		fmt.Fprintln(os.Stderr, "Error: A Docker container name must be provided using -c")
		fmt.Fprintln(os.Stderr, "See --help for proper usage")
		os.Exit(1)
	}

	// checks if there are too many arguments when program is run
	if flag.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "Error: Too many arguments. See --help for proper usage")
		os.Exit(1)
	}

	// checks to see if docker is installed on the host system
	if !dockerExists() {
		fmt.Fprintln(os.Stderr, "Error: Docker is not installed or not in PATH")
		os.Exit(1)
	}

	// checks if the specified docker container exists
	exists, err := containerExists(containerName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error checking docker containers:", err)
		os.Exit(1)
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: Container '%s' does not exist\n", containerName)
		os.Exit(1)
	}

	// check if -f flag has been set
	fileFlagSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "f" {
			fileFlagSet = true
		}
	})

	// files allowed to be passed via -f
	allowedFiles := map[string]struct{}{
		"kyber.dll":                   {},
		"vanillabundleaggregation.kb": {},
		"ca_root.pem":                 {},
		"vivoxsdk.dll":                {},
	}

	// if filename doesn't match allowed files, exit with an error
	if fileFlagSet {
		// retrieves file name from path
		baseName := filepath.Base(fileName)

		// allows file names to be case-insensitive
		fileNameCase := strings.ToLower(baseName)

		if _, ok := allowedFiles[fileNameCase]; !ok {
			fmt.Fprintf(os.Stderr,
				"Error: invalid file '%s'\n",
				baseName,
			)
			os.Exit(1)
		}
	}

	// check if -d and -f are set at the same time
	if download && fileFlagSet {
		fmt.Fprintln(os.Stderr, "Error: -f and -d cannot be used together, see --help")
		os.Exit(1)
	}

	// Download file if requested by -d flag, don't download if -f flag is set
	// also check if Kyber.dll file exists
	if download && !fileFlagSet {
		if verbose {
			fmt.Printf("Downloading Kyber.dll from %s\n", kyberDownloadURL)
		}
		if err := downloadFile(kyberDownloadURL, "Kyber.dll"); err != nil {
			fmt.Fprintln(os.Stderr, "Error downloading file:", err)
			os.Exit(1)
		}
	} else {
		if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "Error: Host file '%s' does not exist\n", fileName)
			os.Exit(1)
		}
	}

	// get file name from path
	fileNameBase := filepath.Base(fileName)

	// Move old file inside container
	err = runCommand(
		verbose,
		"docker", "exec", containerName,
		"bash", "-c",
		fmt.Sprintf(
			"mv /root/.local/share/kyber/module/%s /root/.local/share/kyber/module/%s.old",
			fileNameBase,
			fileNameBase,
		),
	)
	if err != nil {
		os.Exit(1)
	}

	// Copy new file into container
	err = runCommand(
		verbose,
		"docker", "cp",
		fileName,
		fmt.Sprintf("%s:/root/.local/share/kyber/module/%s", containerName, fileNameBase),
	)
	if err != nil {
		os.Exit(1)
	}

	// Restart container
	err = runCommand(verbose, "docker", "restart", containerName)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("The new %s has been successfully added to the specified container\n", fileNameBase)
	fmt.Println("The docker container has been restarted and is ready for use")
}
