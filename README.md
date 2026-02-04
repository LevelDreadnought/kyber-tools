# Kyber Dedicated Server Tools

A collection of linux command line tools written in Go to assist in the creation and administration of Kyber dedicated servers via the official Kyber Docker container.

## CLI Tools

| Plugin                                                | Description                                                                                      |
|-------------------------------------------------------|--------------------------------------------------------------------------------------------------|
| [Kyber Server Launcher](#Kyber-Server-Launcher)       | An interactive Linux CLI tool for creating and running a **Kyber dedicated server** using Docker |
| [Kyber Server Log Manager](#Kyber-Server-Log-Manager) | A simple, interactive **CLI tool** for extracting Kyber dedicated server `.log` files            |
| [Kyber Module Updater](#Kyber-Module-Updater)         | A small CLI tool designed to update **Kyber module files** inside an existing dedicated server   |

---

# Kyber Server Launcher

An interactive Linux CLI tool for creating and running a **Kyber dedicated server** using Docker.

This tool walks you through all required configuration options, then generates a valid `docker run` command for the official Kyber server image. You can choose to run the command immediately, save it to a file, or simply print it to the command line.

## Features

* Interactive, guided setup (no Docker command memorization required)
* Automatically builds a valid `docker run` command
* Secure handling of credentials (entered at runtime)
* Optional automatic restart (`--restart=unless-stopped`)
* Optional mod and plugin volume variable support
* Save generated commands as executable scripts
* Print-only mode for inspection or manual use

## Requirements

* **Linux**
* **Docker** installed and available in `$PATH`
* Go **1.20+** (only required to build)

> ⚠️ If running or saving the generated Docker command, you will need or run this tool with `sudo`.

## What this Tool Does

The program will prompt you for:

### Required Inputs

* EA account email
* EA account password
* Kyber token
* Server name
* Max players
* Map rotation (BASE64 from the Kyber client)
* Path to Battlefront II game data on the host
* Docker container name (lowercase, no spaces)

### Optional Inputs

* Mod folder path on the host
* Plugin folder path on the host
* Server description
* Server password
* Module channel (**leave blank unless you know what you are doing**)
* Docker restart policy

Based on your answers, it generates a valid `docker run` command for the Kyber dedicated server image.

## Docker Container Naming Rules

The Docker container name **must**:

* Be **lowercase**
* Contain **no spaces**
* Only use:

  * `a–z`
  * `0–9`
  * `-` `_` `.`

## Output Options

After configuration, you can choose one of the following:

1. **Run the Docker command immediately**
2. **Save the command to a file**
3. **Run the command and save it to a file**
4. **Print the command only**

Saved Docker command files are automatically marked executable.

## Example Generated Command

```bash
docker run -it \
  --name kyber-server-01 \
  --restart=unless-stopped \
  -e MAXIMA_CREDENTIALS="email:password" \
  -e KYBER_TOKEN=123456 \
  -e KYBER_SERVER_NAME="My Server" \
  -e KYBER_SERVER_MAX_PLAYERS=40 \
  -e KYBER_MAP_ROTATION=BASE64STRING \
  -v "/path/to/game/files:/mnt/battlefront" \
  -v "/path/to/mods:/mnt/battlefront/mods" \
  ghcr.io/armchairdevelopers/kyber-server:latest
```

## Notes & Tips

* Mod and plugin environment variables are **only added if their volume paths are set**
* The module channel is omitted unless changed from `main` (leave blank for most cases)
* Generated scripts can be reused, versioned, or shared
* The tool does **not** manage existing containers — name collisions will fail normally via Docker

---

# Kyber Server Log Manager

A simple, interactive **Go CLI tool** for extracting Kyber dedicated server `.log` files from a running Docker container.

This tool is designed to make log retrieval easier: it detects Docker, validates containers, lists available log files, and lets you selectively (or fully) copy them to your local machine.


## Features

* Verifies Docker is installed
* Confirms the container exists *and* is running
* Lists all `.log` files inside the Kyber server container
* Supports flexible selection:

  * Single file (`1`)
  * Multiple files (`1,3,5`)
  * Ranges (`2-6`)
  * Everything (`all`)
* Copies logs to:

  * Current working directory, or
  * Any custom destination directory
* Robust error handling and clear prompts


## Requirements

* **Docker** installed and available in your `PATH`
* A **running Kyber dedicated server container**

> ⚠️ You will need to be in the `docker` group or run this tool with `sudo`.


## How It Works

The tool connects to a running Docker container and looks for log files in the **Kyber log directory**:

```
/root/.local/share/maxima/wine/prefix/drive_c/users/root/AppData/Roaming/ArmchairDevelopers/Kyber/Logs
```

Once found, you’re prompted to choose which logs to extract, and where to save them on your host system.


## Usage

Run the program from your terminal:

```bash
./kyber-log-manager
```

### Example Session

1. **Enter the container name**

   ```
   Enter container name: kyber-server
   ```

2. **Select log files**

   ```
   Log files found:
   1) kyber-server_2026-01-30.log
   2) kyber-server_2026-01-31.log
   3) kyber-server_2026-02-01.log
   all) All log files
   
   Select log files to extract (e.g. 1, 1-3, 1-3,5): 1-2
   ```

3. **Choose destination directory**

   ```
   Destination directory (leave empty for current directory): ./logs
   ```

4. Logs are copied using `docker cp` and saved locally.

   ```
   Copied kyber-server_2026-01-30.log
   Copied kyber-server_2026-01-31.log

   Done.
   ```


## Error Handling

The program will exit if:

* Docker is not installed
* The container name is empty
* The container does not exist
* The container exists but is not running
* No `.log` files are found

Clear error messages are printed with actionable guidance.


## Why This Exists

This tool was built to simplify Kyber server administration—especially when logs live deep inside Wine paths in the Kyber Docker container. No manual Docker commands are required which simplifies log retrieval.


## Notes

* Logs are copied, not moved or deleted
* Duplicate selections are automatically ignored
* Destination directories are created if missing

---

# Kyber Module Updater

`kyber-updater` is a small linux CLI tool written in Go designed to update **Kyber module files** inside an existing dedicated server Docker container.
It can either:

* Copy a **local file** from the host into the container, or
* **Download the latest Kyber.dll** and install it automatically.

After updating the file, the tool restarts the Docker container to apply changes.

## Requirements

* **Docker** installed and available in `PATH`
* A **running or stopped container** that already exists
* Go **1.20+** (only required if building from source)
> ⚠️ You will need to be in the `docker` group or run this tool with `sudo`.


## Usage

```bash
kyber-updater [-v] [-c <container_name>] [-f <file_name>] [-d] [-h | --help]
```

### Options

|     Flag | Description                                                   |
| -------: | ------------------------------------------------------------- |
|     `-c` | **(Required)** Docker container name                          |
|     `-f` | Input file to copy into the container (default: `Kyber.dll`)  |
|     `-d` | Download the latest `Kyber.dll` instead of using a local file |
|     `-v` | Enable verbose output                                         |
|     `-h` | Show help message                                             |
| `--help` | Show help message                                             |


## Examples

### Update using a local `Kyber.dll`

```bash
./kyber-updater -c <container_name> -f /path/to/Kyber.dll
```

### Download and install the latest `Kyber.dll` automatically

```bash
./kyber-updater -c <container_name> -d
```

### Update using a `Kyber.dll` in the same directory

```bash
./kyber-updater -c <container_name>
```

### Verbose mode

```bash
./kyber-updater -c <container_name> -d -v
```


## Behavior Details

### File Handling

* The existing file inside the container is renamed to:

  ```
  <filename>.old
  ```
* The new file is copied into:

  ```
  /root/.local/share/kyber/module/
  ```

### Container Restart

After the file is updated, the container is automatically restarted:

```bash
docker restart <container_name>
```


## Error Handling

The program will exit with a clear error message if:

* Docker is not installed
* The container does not exist
* `-f` and `-d` are used together
* An invalid file is passed with `-f`
* The local file does not exist
* Docker commands fail

---





