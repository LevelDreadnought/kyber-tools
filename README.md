# kyber-tools


# kyber-updater

`kyber-updater` is a small linux CLI tool written in Go designed to update **Kyber module files** inside an existing dedicated server Docker container.
It can either:

* Copy a **local file** from the host into the container, or
* **Download the latest Kyber.dll** and install it automatically.

After updating the file, the tool restarts the Docker container to apply changes.

## Requirements

* **Docker** installed and available in `PATH`
* A **running or stopped container** that already exists
* Go **1.20+** (only required if building from source)

---

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

---

## Examples

### Update using a local `Kyber.dll`

```bash
./kyber-updater -c kyber-container -f ~/Kyber.dll
```

### Download and install the latest `Kyber.dll`

```bash
./kyber-updater -c kyber-container -d
```

### Verbose mode

```bash
./kyber-updater -c kyber-container -d -v
```

---


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

---

## Error Handling

The program will exit with a clear error message if:

* Docker is not installed
* The container does not exist
* `-f` and `-d` are used together
* An invalid file is passed with `-f`
* The local file does not exist
* Docker commands fail

---

