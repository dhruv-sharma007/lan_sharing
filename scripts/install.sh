#!/usr/bin/env bash
set -e

REPO_OWNER="dhruv-sharma007"
REPO_NAME="lan_sharing"

echo "Installing LanShare..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

if [ "$OS" = "linux" ] || [ "$OS" = "darwin" ]; then
    BINARY_NAME="lanshare-$OS-$ARCH"
else
    echo "Unsupported OS: $OS"
    exit 1
fi

RELEASES_URL="https://api.github.com/repos/$REPO_OWNER/$REPO_NAME/releases/latest"

echo "Fetching latest release..."
DOWNLOAD_URL=$(curl -fsSL "$RELEASES_URL" | grep "browser_download_url" | grep "$BINARY_NAME" | head -n 1 | cut -d '"' -f 4)

if [ -z "$DOWNLOAD_URL" ]; then
    echo "Could not find asset $BINARY_NAME in the latest release."
    exit 1
fi

INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/lanshare"

mkdir -p "$INSTALL_DIR"

echo "Downloading $BINARY_NAME..."
curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"

echo "Configuring startup..."
if [ "$OS" = "linux" ]; then
    AUTOSTART_DIR="$HOME/.config/autostart"
    mkdir -p "$AUTOSTART_DIR"
    DESKTOP_FILE="$AUTOSTART_DIR/lanshare.desktop"
    cat > "$DESKTOP_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=LanShare
Exec=$INSTALL_PATH
X-GNOME-Autostart-enabled=true
EOF
elif [ "$OS" = "darwin" ]; then
    LAUNCHAGENT_DIR="$HOME/Library/LaunchAgents"
    mkdir -p "$LAUNCHAGENT_DIR"
    PLIST_FILE="$LAUNCHAGENT_DIR/com.lanshare.app.plist"
    cat > "$PLIST_FILE" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.lanshare.app</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_PATH</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
EOF
    # Attempt to load the agent
    launchctl load "$PLIST_FILE" || true
fi

echo "Starting LanShare..."
"$INSTALL_PATH" &

echo "LanShare installed successfully!"
