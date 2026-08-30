# zeopoxa-exporter

`zeopoxa-exporter` is a command-line tool designed to convert cycling tracks from the proprietary database format used by the [Zeopoxa Cycling](https://www.zeopoxa.com/cycling.html) Android app into standard [GPX (GPS Exchange Format)](https://en.wikipedia.org/wiki/GPS_Exchange_Format) files.

This allows you to migrate your workout history to other platforms like Strava, Garmin Connect, or Komoot.

## Features

- Converts SQLite database backups to individual GPX files.
- Preserves track points, elevation, speed, and heart rate data (if available).
- Supports cross-platform binaries (Linux, macOS, Windows).

## How to Use

### 1. Get your database
To use this tool, you need the SQLite database file from your Zeopoxa app. Usually, this is obtained by creating a backup within the app and locating the resulting file (often named with a `.db` extension) on your phone's storage.

### 2. Run the exporter
Download the binary for your platform from the **Releases** and run it via the terminal:

```bash
./zeopoxa_exporter -d path/to/your/zeopoxa_backup.db -o ./my_gpx_tracks
```

### Command Line Options

| Flag | Description |
|------|-------------|
| `-d` | **(Required)** Path to the Zeopoxa SQLite backup file |
| `-o` | Directory where GPX files will be saved | 
| `-z` | Timezone for parsing start times (e.g., `Europe/London`) |
| `-t` | Timeout for database processing |
| `-l` | Log level (`error`, `info`, `warn`, `debug`) |
| `-v` | Show version information |

## Installation

### Download Binaries
Pre-compiled binaries for Windows, Linux, and macOS (Intel/Apple Silicon) are available in the **Releases** tab of this repository. Select the latest successful workflow run and look for the `zeopoxa_exporter-binaries` artifact.

### Build from Source
If you have [Go](https://go.dev/) installed (version 1.27 or higher), you can build the project manually:

1. Clone the repository:
   ```bash
   git clone https://github.com/securjty/zeopoxa-exporter.git
   cd zeopoxa-exporter
   ```

2. Build for your current architecture:
   ```bash
   make build
   ```
   The binary will be located in the `build/` directory.

3. (Optional) Build for all supported platforms:
   ```bash
   make build-all
   ```

