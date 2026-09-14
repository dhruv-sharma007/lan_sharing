#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="dhruv-sharma007"
REPO_NAME="lan_sharing"

APP_NAME="lanshare"
DISPLAY_NAME="LanShare"

INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/$APP_NAME"

echo "Installing $DISPLAY_NAME..."

# --------------------------------------------------
# Detect OS
# --------------------------------------------------

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

case "$OS" in
    linux)
        OS="linux"
        ;;
    darwin)
        OS="darwin"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# --------------------------------------------------
# Detect architecture
# --------------------------------------------------

ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Detected: $OS/$ARCH"

# --------------------------------------------------
# Build GitHub download URL
# --------------------------------------------------

BINARY_NAME="lanshare-${OS}-${ARCH}"

DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${BINARY_NAME}"

echo "Downloading latest release..."
echo "Asset: $BINARY_NAME"

mkdir -p "$INSTALL_DIR"

TMP_FILE="$(mktemp)"

cleanup() {
    rm -f "$TMP_FILE"
}

trap cleanup EXIT

curl \
    --fail \
    --location \
    --silent \
    --show-error \
    --retry 3 \
    "$DOWNLOAD_URL" \
    --output "$TMP_FILE"

chmod +x "$TMP_FILE"

# Atomic-ish replacement of existing binary
mv "$TMP_FILE" "$INSTALL_PATH"

# TMP_FILE no longer exists after mv
trap - EXIT

echo "Installed binary:"
echo "  $INSTALL_PATH"

# --------------------------------------------------
# Linux - systemd user service
# --------------------------------------------------

if [ "$OS" = "linux" ]; then

    if ! command -v systemctl >/dev/null 2>&1; then
        echo "systemd was not found."
        echo "This installer currently supports systemd-based Linux distributions."
        exit 1
    fi

    SERVICE_DIR="$HOME/.config/systemd/user"
    SERVICE_FILE="$SERVICE_DIR/lanshare.service"

    mkdir -p "$SERVICE_DIR"

    echo "Creating systemd user service..."

    cat > "$SERVICE_FILE" <<'EOF'
[Unit]
Description=LanShare LAN File Sharing Service
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%h/.local/bin/lanshare

Restart=always
RestartSec=3

TimeoutStopSec=10

[Install]
WantedBy=default.target
EOF

    echo "Reloading systemd..."

    systemctl --user daemon-reload

    echo "Enabling LanShare..."

    systemctl --user enable lanshare.service

    echo "Restarting LanShare..."

    systemctl --user restart lanshare.service

    echo
    echo "LanShare systemd service installed."
    echo
    echo "Useful commands:"
    echo
    echo "  systemctl --user status lanshare"
    echo "  systemctl --user restart lanshare"
    echo "  systemctl --user stop lanshare"
    echo "  journalctl --user -u lanshare -f"

# --------------------------------------------------
# macOS - launchd LaunchAgent
# --------------------------------------------------

elif [ "$OS" = "darwin" ]; then

    LABEL="com.lanshare.app"

    LAUNCH_AGENT_DIR="$HOME/Library/LaunchAgents"
    PLIST_FILE="$LAUNCH_AGENT_DIR/$LABEL.plist"

    LOG_DIR="$HOME/Library/Logs/LanShare"

    mkdir -p "$LAUNCH_AGENT_DIR"
    mkdir -p "$LOG_DIR"

    echo "Creating launchd LaunchAgent..."

    cat > "$PLIST_FILE" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC
    "-//Apple//DTD PLIST 1.0//EN"
    "http://www.apple.com/DTDs/PropertyList-1.0.dtd">

<plist version="1.0">
<dict>

    <key>Label</key>
    <string>$LABEL</string>

    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_PATH</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <key>ProcessType</key>
    <string>Background</string>

    <key>ThrottleInterval</key>
    <integer>3</integer>

    <key>StandardOutPath</key>
    <string>$LOG_DIR/lanshare.log</string>

    <key>StandardErrorPath</key>
    <string>$LOG_DIR/lanshare-error.log</string>

</dict>
</plist>
EOF

    USER_ID="$(id -u)"

    echo "Registering LaunchAgent..."

    # Remove old loaded version if present
    launchctl bootout \
        "gui/$USER_ID/$LABEL" \
        2>/dev/null || true

    launchctl bootstrap \
        "gui/$USER_ID" \
        "$PLIST_FILE"

    launchctl kickstart \
        -k \
        "gui/$USER_ID/$LABEL"

    echo
    echo "LanShare LaunchAgent installed."
    echo
    echo "Useful commands:"
    echo
    echo "  launchctl print gui/$USER_ID/$LABEL"
    echo "  launchctl kickstart -k gui/$USER_ID/$LABEL"
    echo "  launchctl bootout gui/$USER_ID/$LABEL"
    echo
    echo "Logs:"
    echo "  $LOG_DIR/lanshare.log"
    echo "  $LOG_DIR/lanshare-error.log"

fi

echo
echo "--------------------------------------"
echo "$DISPLAY_NAME installed successfully"
echo "--------------------------------------"
echo
echo "Binary:"
echo "  $INSTALL_PATH"