# LanShare

LanShare makes it simple to send files between computers on the same local network.

Instead of opening a browser, uploading to cloud storage, or manually sharing network folders, both people run LanShare and copy files into a peer's Send folder. LanShare discovers connected computers on the LAN and transfers the files automatically.

It is designed for trusted home, office, or local-network use. It is not an internet file-sharing service and current transfers are not encrypted.

## Install LanShare

Install LanShare on every computer that should send or receive files.

### Windows

Open PowerShell and run:

```powershell
irm https://raw.githubusercontent.com/dhruv-sharma007/lan_sharing/master/scripts/install.ps1 | iex
```

The installer downloads the latest Windows release, starts LanShare, and creates a `LanShare` Scheduled Task that starts it when you sign in and restarts it after failures.

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/dhruv-sharma007/lan_sharing/master/scripts/install.sh | bash
```

Linux requires a systemd-based distribution. The installer creates and enables the `lanshare.service` systemd user service.

### macOS

```bash
curl -fsSL https://raw.githubusercontent.com/dhruv-sharma007/lan_sharing/master/scripts/install.sh | bash
```

The installer downloads the matching release binary and registers the `com.lanshare.app` LaunchAgent, which starts LanShare when you sign in, keeps it running, and writes logs to `~/Library/Logs/LanShare/`.

## Uninstall LanShare

Run the matching command for your operating system:

### Windows

```powershell
irm https://raw.githubusercontent.com/dhruv-sharma007/lan_sharing/master/scripts/uninstall.ps1 | iex
```

### Linux and macOS

```bash
curl -fsSL https://raw.githubusercontent.com/dhruv-sharma007/lan_sharing/master/scripts/uninstall.sh | bash
```

The uninstaller stops and unregisters the background task or service, then removes the installed binary. It also cleans up automatic-start entries created by earlier installer versions.

### Manual run for development or testing

Build and run LanShare from a dedicated working directory:

```bash
go build -o lanshare ./cmd/lanshare
./lanshare
```

On Windows, build `lanshare.exe` and run it from PowerShell. LanShare creates its configuration and shared folders in its working directory.

## Set up the computers

1. Connect all computers to the same local network.
2. Use a private/trusted network, not guest Wi-Fi, client-isolated Wi-Fi, VPN-only networking, or a public hotspot.
3. On Windows, set the Wi-Fi/Ethernet network profile to **Private**.
4. Allow LanShare through the firewall. It needs TCP port `3598` for peer connections and file transfers, plus UDP port `5353` for local-network discovery.
5. Start LanShare on each computer and wait for them to connect.
6. Create a shortcut to the `LanShare` folder in a convenient place, such as the Desktop, Quick Access, or Finder sidebar. Use this shortcut whenever you want to send or view received files.

### Find the LanShare folder and make a shortcut

The application is installed in the following locations:

| System | Application location | Default LanShare folder |
| --- | --- | --- |
| Windows | `%LOCALAPPDATA%\\LanShare\\lanshare.exe` | `%LOCALAPPDATA%\\LanShare\\LanShare` |
| Linux | `~/.local/bin/lanshare` | `~/LanShare` when started from your home folder |
| macOS | `~/.local/bin/lanshare` | `~/LanShare` when started from your home folder |

On Windows, press `Win + R`, enter `%LOCALAPPDATA%\\LanShare`, and open the `LanShare` folder inside it. Right-click the folder and choose **Send to → Desktop (create shortcut)**.

On Linux, open `~/LanShare` in your file manager and choose **Create Link** (or drag it to the desktop, depending on your desktop environment). On macOS, open `~/LanShare` in Finder and drag the folder to the Desktop or Finder sidebar while holding `Option` and `Command` to create an alias.

The application binary and the file-sharing folder are different: only the `LanShare` folder needs a shortcut. It contains the `Send` and `Received` folders used every day.

After two computers connect, each computer creates a peer-specific Send folder:

```text
LanShare/
├── Send/
│   └── <peer-hostname>-<peer-node-id>/
└── Received/
```

For example:

```text
LanShare/Send/acer-ACERdhruv-29316217849356388/
```

The unique node ID prevents conflicts when two computers have the same hostname.

## Send and receive files

To send a file to another computer:

1. Open `LanShare/Send/`.
2. Open the folder named for the computer that should receive the file.
3. Copy a file or folder into that peer folder.
4. Leave the file unchanged for a few seconds. LanShare waits for the copy to finish, then transfers it automatically.

The other computer receives it under:

```text
LanShare/Received/
```

Folder structure is preserved. If a file with the same name already exists, LanShare keeps the original and creates a non-conflicting name such as `photo (1).jpg`.

Do not place files directly inside `LanShare/Send/`; they must be inside a specific peer folder.
Tip: create shortuct of LanShare app to your desired path 

## Troubleshooting

- **No peer folder appears:** Confirm both computers are running LanShare, connected to the same private LAN, and allowed through the firewall. Restart both apps after correcting the network settings.
- **A file does not send:** Confirm it is inside the correct peer folder and does not end with `.part`.
- **Files are slow:** Transfers are limited by Wi-Fi/LAN and disk speed. Files for the same recipient are sent one at a time, so a small file waits behind an earlier large file.
- **Wrong or old peer folder name:** Restart after updating LanShare. Existing folders created with the old malformed name are migrated when that peer connects.

## Current limitations

- Use only on a trusted local network; transport encryption is not implemented yet.
- There is no graphical file picker or transfer-progress window. The Send and Received folders are the user interface.
- Docker Desktop on Windows has file-watcher limitations; use the native desktop app for normal file sharing.
