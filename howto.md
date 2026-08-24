# Build Instructions for ShareApp

Since ShareApp is written in Go, you can easily cross-compile it for almost any operating system and architecture directly from your current machine.

Here are the commands to build the application for the major platforms. 

## General Syntax
The general pattern for cross-compiling in Go is setting the `GOOS` (target Operating System) and `GOARCH` (target Architecture) environment variables before running the build command.

We also use `-ldflags="-s -w"` to strip debug information, which makes the final binary significantly smaller.

---

### Windows
To build for Windows (64-bit):
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o shareapp.exe ./cmd/shareapp
```

### Linux
To build for Linux (64-bit PC/Server):
```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o shareapp-linux-amd64 ./cmd/shareapp
```

To build for Linux (ARM64 - e.g., Raspberry Pi 4/5, AWS Graviton):
```bash
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o shareapp-linux-arm64 ./cmd/shareapp
```

### macOS (Darwin)
To build for newer Macs with Apple Silicon (M1/M2/M3):
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o shareapp-macos-arm64 ./cmd/shareapp
```

To build for older Macs with Intel processors:
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o shareapp-macos-intel ./cmd/shareapp
```

---

## Testing Locally (Single Machine)

If you want to test peer discovery with two instances running on the same machine, you can run them directly without building by specifying different ports and names:

**Terminal 1:**
```bash
go run ./cmd/shareapp -port 3498 -name NodeA
```

**Terminal 2:**
```bash
go run ./cmd/shareapp -port 3499 -name NodeB
```
