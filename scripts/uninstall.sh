#!/usr/bin/env bash
set -u

APP_NAME="lanshare"
DISPLAY_NAME="LanShare"
INSTALL_PATH="$HOME/.local/bin/$APP_NAME"

echo "Uninstalling $DISPLAY_NAME..."

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

case "$OS" in
    linux)
        SERVICE_FILE="$HOME/.config/systemd/user/lanshare.service"

        if command -v systemctl >/dev/null 2>&1; then
            echo "Stopping and disabling systemd user service..."
            systemctl --user disable --now lanshare.service 2>/dev/null || true
        fi

        if [ -f "$SERVICE_FILE" ]; then
            echo "Removing systemd user service at $SERVICE_FILE..."
            rm -f "$SERVICE_FILE"

            if command -v systemctl >/dev/null 2>&1; then
                systemctl --user daemon-reload
            fi
        fi

        # Clean up the autostart entry used by older installers.
        LEGACY_AUTOSTART_FILE="$HOME/.config/autostart/lanshare.desktop"
        if [ -f "$LEGACY_AUTOSTART_FILE" ]; then
            echo "Removing legacy autostart entry at $LEGACY_AUTOSTART_FILE..."
            rm -f "$LEGACY_AUTOSTART_FILE"
        fi

        LOG_DIR="$HOME/.local/state/lanshare/logs"
        ;;
    darwin)
        LABEL="com.lanshare.app"
        PLIST_FILE="$HOME/Library/LaunchAgents/$LABEL.plist"
        USER_ID="$(id -u)"

        if command -v launchctl >/dev/null 2>&1; then
            echo "Unregistering LaunchAgent..."
            if ! launchctl bootout "gui/$USER_ID/$LABEL" 2>/dev/null; then
                # Compatibility with the LaunchAgent registration used by
                # older installers.
                launchctl unload "$PLIST_FILE" 2>/dev/null || true
            fi
        fi

        if [ -f "$PLIST_FILE" ]; then
            echo "Removing LaunchAgent at $PLIST_FILE..."
            rm -f "$PLIST_FILE"
        fi

        LOG_DIR="$HOME/Library/Logs/LanShare"
        ;;
    *)
        echo "Unsupported operating system: $OS"
        exit 1
        ;;
esac

# Stop a process left over from a legacy installation or failed service teardown.
if pgrep -x "$APP_NAME" > /dev/null 2>&1; then
    echo "Stopping remaining $DISPLAY_NAME process..."
    pkill -x "$APP_NAME" || true
fi

if [ -f "$INSTALL_PATH" ]; then
    echo "Removing binary at $INSTALL_PATH..."
    rm -f "$INSTALL_PATH"
fi

if [ -d "$LOG_DIR" ]; then
    echo "Removing logs at $LOG_DIR..."
    rm -rf "$LOG_DIR"
fi

echo "$DISPLAY_NAME has been completely uninstalled!"
