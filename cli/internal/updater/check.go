package updater

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// updateCache represents the cached update check result.
type updateCache struct {
	LastCheck     int64  `json:"last_check"`
	LatestVersion string `json:"latest_version"`
}

// cacheFilePath returns the path to the update check cache file.
func cacheFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".blueprint", ".update-check")
}

// CheckForUpdate checks GitHub Releases API for newer versions.
// Returns the latest version string if an update is available, or "" if current/dev/error.
// Results are cached for 24 hours. Silent on all errors.
func CheckForUpdate(currentVersion string) string {
	// Skip for dev builds
	if currentVersion == "dev" || currentVersion == "" {
		return ""
	}

	cachePath := cacheFilePath()
	if cachePath == "" {
		return ""
	}

	// Check cache first
	if cached := readCache(cachePath); cached != nil {
		elapsed := time.Since(time.Unix(cached.LastCheck, 0))
		if elapsed < 24*time.Hour {
			if IsNewer(cached.LatestVersion, currentVersion) {
				return cached.LatestVersion
			}
			return ""
		}
	}

	// Fetch latest release from GitHub
	latest := fetchLatestVersion()
	if latest == "" {
		return ""
	}

	// Write cache
	writeCache(cachePath, latest)

	if IsNewer(latest, currentVersion) {
		return latest
	}
	return ""
}

// IsNewer returns true if latest is a newer version than current.
// Both should be semver-like strings (with or without "v" prefix).
// Returns false for empty strings or "dev".
func IsNewer(latest, current string) bool {
	if latest == "" || current == "" || current == "dev" || latest == "dev" {
		return false
	}

	latestClean := strings.TrimPrefix(latest, "v")
	currentClean := strings.TrimPrefix(current, "v")

	if latestClean == currentClean {
		return false
	}

	latestParts := strings.Split(latestClean, ".")
	currentParts := strings.Split(currentClean, ".")

	// Pad to 3 parts
	for len(latestParts) < 3 {
		latestParts = append(latestParts, "0")
	}
	for len(currentParts) < 3 {
		currentParts = append(currentParts, "0")
	}

	for i := 0; i < 3; i++ {
		l := parseNum(latestParts[i])
		c := parseNum(currentParts[i])
		if l > c {
			return true
		}
		if l < c {
			return false
		}
	}

	return false
}

func parseNum(s string) int {
	// Strip any pre-release suffix (e.g., "0-beta")
	if idx := strings.IndexByte(s, '-'); idx >= 0 {
		s = s[:idx]
	}
	n := 0
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}
	return n
}

func fetchLatestVersion() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/skaisser/blueprint/releases/latest")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	return release.TagName
}

func readCache(path string) *updateCache {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var cache updateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}
	return &cache
}

func writeCache(path, version string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	cache := updateCache{
		LastCheck:     time.Now().Unix(),
		LatestVersion: version,
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return
	}
	os.WriteFile(path, data, 0644)
}
