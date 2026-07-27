// Package youtube provides YouTube playlist polling and Taskwarrior task creation.
// It invokes yt-dlp to fetch a public playlist and creates a task for each new
// video URL. A state file tracks which video IDs have already been processed so
// re-runs are idempotent. Taskwarrior on-add hooks are expected to enrich the
// task from the URL (title, tags, etc.).
package youtube

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/clobrano/wui/internal/config"
	"github.com/clobrano/wui/internal/core"
)

// Syncer fetches a YouTube playlist and creates Taskwarrior tasks for new videos.
type Syncer struct {
	playlist  *config.YoutubePlaylist
	ytDlpBin  string
	stateFile string
	svc       core.TaskService
}

// NewSyncer creates a Syncer for the given playlist entry.
// stateFile is auto-derived from the playlist ID when the playlist's StateFile is empty.
func NewSyncer(playlist *config.YoutubePlaylist, ytDlpBin string, svc core.TaskService) *Syncer {
	return &Syncer{
		playlist:  playlist,
		ytDlpBin:  ytDlpBin,
		stateFile: resolveStateFile(playlist),
		svc:       svc,
	}
}

// SyncResult summarises a single sync run.
type SyncResult struct {
	Added   int
	Skipped int
}

type ytVideo struct {
	ID  string `json:"id"`
	URL string `json:"webpage_url"`
}

// Sync fetches the playlist, compares against the state file, and creates
// Taskwarrior tasks for any videos not yet seen.
func (s *Syncer) Sync() (*SyncResult, error) {
	seen, err := loadStateFile(s.stateFile)
	if err != nil {
		return nil, fmt.Errorf("load state file: %w", err)
	}

	videos, err := s.fetchPlaylist()
	if err != nil {
		return nil, fmt.Errorf("fetch playlist: %w", err)
	}

	result := &SyncResult{}
	var newIDs []string

	for _, v := range videos {
		if seen[v.ID] {
			result.Skipped++
			continue
		}

		desc := buildDescription(v.URL, s.playlist.TaskProject, s.playlist.TaskTags)

		if _, err := s.svc.Add(desc); err != nil {
			slog.Warn("YouTube sync: failed to add task", "url", v.URL, "error", err)
			continue
		}

		newIDs = append(newIDs, v.ID)
		result.Added++
		slog.Info("YouTube sync: task created", "url", v.URL)
	}

	if len(newIDs) > 0 {
		if err := appendStateFile(s.stateFile, newIDs); err != nil {
			return result, fmt.Errorf("save state file: %w", err)
		}
	}

	return result, nil
}

func buildDescription(url, project string, tags []string) string {
	parts := []string{url}
	if project != "" {
		parts = append(parts, "project:"+project)
	}
	for _, tag := range tags {
		parts = append(parts, "+"+tag)
	}
	return strings.Join(parts, " ")
}

func (s *Syncer) fetchPlaylist() ([]ytVideo, error) {
	cmd := exec.Command(s.ytDlpBin, "--flat-playlist", "-j", "--no-warnings", s.playlist.PlaylistURL)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("yt-dlp: %w\n%s", err, msg)
		}
		return nil, fmt.Errorf("yt-dlp: %w", err)
	}

	var videos []ytVideo
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var v ytVideo
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			slog.Warn("YouTube sync: failed to parse yt-dlp line", "error", err)
			continue
		}
		if v.ID != "" && v.URL != "" {
			videos = append(videos, v)
		}
	}
	return videos, scanner.Err()
}

func resolveStateFile(pl *config.YoutubePlaylist) string {
	if pl.StateFile != "" {
		return pl.StateFile
	}
	// Extract the list= parameter from the URL for a stable, readable filename.
	id := "unknown"
	if u, err := url.Parse(pl.PlaylistURL); err == nil {
		if list := u.Query().Get("list"); list != "" {
			id = list
		}
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "wui", "youtube_"+id+".txt")
}

func loadStateFile(path string) (map[string]bool, error) {
	seen := make(map[string]bool)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return seen, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		id := strings.TrimSpace(scanner.Text())
		if id != "" {
			seen[id] = true
		}
	}
	return seen, scanner.Err()
}

func appendStateFile(path string, ids []string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, id := range ids {
		fmt.Fprintln(w, id)
	}
	return w.Flush()
}
