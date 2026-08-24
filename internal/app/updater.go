package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/minio/selfupdate"
)

const repoOwner = "dhruv-sharma007"
const repoName = "lan_sharing"

// StartUpdater runs a background goroutine to check for updates every 6 hours.
func StartUpdater(ctx context.Context, currentVersion string) {
	time.AfterFunc(5*time.Second, func() {
		checkAndUpdate(currentVersion)
	})

	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				checkAndUpdate(currentVersion)
			}
		}
	}()
}

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

func checkAndUpdate(currentVersion string) {
	if currentVersion == "dev" {
		slog.Info("Skipping auto-update check for dev version")
		return
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		slog.Error("Failed to check for updates", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("GitHub API returned non-200 status", "status", resp.Status)
		return
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		slog.Error("Failed to decode release info", "error", err)
		return
	}

	if release.TagName == currentVersion || release.TagName == "v"+currentVersion {
		slog.Info("Already on the latest version", "version", currentVersion)
		return
	}

	slog.Info("New version available", "current", currentVersion, "new", release.TagName)

	var targetAssetUrl string
	expectedAssetName := fmt.Sprintf("lanshare-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedAssetName += ".exe"
	}

	for _, asset := range release.Assets {
		if asset.Name == expectedAssetName {
			targetAssetUrl = asset.BrowserDownloadUrl
			break
		}
	}

	if targetAssetUrl == "" {
		slog.Error("No matching binary found for current platform in latest release")
		return
	}

	// Download the asset
	slog.Info("Downloading update", "url", targetAssetUrl)
	dlResp, err := http.Get(targetAssetUrl)
	if err != nil {
		slog.Error("Failed to download update", "error", err)
		return
	}
	defer dlResp.Body.Close()

	if err := selfupdate.Apply(dlResp.Body, selfupdate.Options{}); err != nil {
		slog.Error("Update application failed", "error", err)
		return
	}

	slog.Info("Update successful, restarting...")
	Restart()
}

// Restart replaces the current process with the new binary.
func Restart() {
	execPath, err := os.Executable()
	if err != nil {
		slog.Error("Failed to get executable path for restart", "error", err)
		return
	}

	cmd := exec.Command(execPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		slog.Error("Failed to restart", "error", err)
		return
	}

	os.Exit(0)
}