#!/usr/bin/env bash

echo "Uninstalling LanShare..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

# Stop the running process if it exists
if pgrep -x "lanshare" > /dev/null; then
    echo "Stopping LanShare process..."
    pkill -x "lanshare"
fi

# Remove binary
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="$INSTALL_DIR/lanshare"
if [ -f "$INSTALL_PATH" ]; then
    echo "Removing binary at $INSTALL_PATH..."
    rm "$INSTALL_PATH"
fi

# Remove autostart / launch agents
if [ "$OS" = "linux" ]; then
    AUTOSTART_FILE="$HOME/.config/autostart/lanshare.desktop"
    if [ -f "$AUTOSTART_FILE" ]; then
        echo "Removing autostart entry at $AUTOSTART_FILE..."
        rm "$AUTOSTART_FILE"
    fi
elif [ "$OS" = "darwin" ]; then
    PLIST_FILE="$HOME/Library/LaunchAgents/com.lanshare.app.plist"
    if [ -f "$PLIST_FILE" ]; then
        echo "Unloading LaunchAgent..."
        launchctl unload "$PLIST_FILE" 2>/dev/null || true
        echo "Removing LaunchAgent at $PLIST_FILE..."
        rm "$PLIST_FILE"
    fi
fi

# Remove logs
if [ "$OS" = "linux" ]; then
    LOG_DIR="$HOME/.local/state/lanshare/logs"
elif [ "$OS" = "darwin" ]; then
    LOG_DIR="$HOME/Library/Logs/LanShare"
fi

if [ -d "$LOG_DIR" ]; then
    echo "Removing logs at $LOG_DIR..."
    rm -rf "$LOG_DIR"
fi

echo "LanShare has been completely uninstalled!"
