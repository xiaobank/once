package docker

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	BackupDataDir   = "data"
	BackupRetention = 30 * 24 * time.Hour

	backupAppSettingsEntry = "once.application.json"
	backupVolSettingsEntry = "once.volume.json"
	backupTimeFormat       = "20060102-150405"
)

func (a *Application) Backup(ctx context.Context) error {
	if a.Settings.Backup.Path == "" {
		return fmt.Errorf("backup location is required")
	}

	return a.BackupToFile(ctx, a.Settings.Backup.Path, a.BackupName())
}

func (a *Application) BackupName() string {
	return fmt.Sprintf("%s-%s.tar.gz", a.Settings.Name, time.Now().Format(backupTimeFormat))
}

func (a *Application) BackupToFile(ctx context.Context, dir string, name string) error {
	uid, gid, err := prepareBackupDir(dir)
	if err != nil {
		slog.Error("Failed to create backup directory", "directory", dir, "error", err)
		return err
	}

	filePath := filepath.Join(dir, name)
	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("Failed to create backup file", "filename", filePath, "error", err)
		return fmt.Errorf("creating backup file: %w", err)
	}
	defer file.Close()

	if err := os.Chown(filePath, uid, gid); err != nil {
		slog.Error("Failed to set backup file ownership", "filename", filePath, "error", err)
		return fmt.Errorf("setting backup file ownership: %w", err)
	}

	err = a.backupToWriter(ctx, file)
	a.saveOperationResult(ctx, func(s *State) { s.RecordBackup(a.Settings.Name, err) })

	if err != nil {
		os.Remove(filePath)
		slog.Error("Failed to generate backup", "filename", filePath, "error", err)
		return err
	}

	slog.Info("Created backup file", "filename", filePath)

	return nil
}

func (a *Application) TrimBackups() error {
	if a.Settings.Backup.Path == "" {
		return nil
	}

	entries, err := os.ReadDir(a.Settings.Backup.Path)
	if err != nil {
		return fmt.Errorf("reading backup directory: %w", err)
	}

	var errs []error
	cutoff := time.Now().Add(-BackupRetention)

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		t, ok := parseBackupTime(a.Settings.Name, entry.Name())
		if !ok {
			continue
		}

		if t.Before(cutoff) {
			filename := filepath.Join(a.Settings.Backup.Path, entry.Name())
			if err := os.Remove(filename); err != nil {
				slog.Error("Failed to remove expired backup file", "filename", filename, "error", err)
				errs = append(errs, err)
			} else {
				slog.Info("Removed expired backup file", "filename", filename)
			}
		}
	}

	return errors.Join(errs...)
}

// Private

func (a *Application) backupToWriter(ctx context.Context, w io.Writer) error {
	containerName, err := a.ContainerName(ctx)
	if err != nil {
		return fmt.Errorf("getting container name: %w", err)
	}

	if err := a.runHookScript(ctx, containerName, "pre-backup"); err != nil {
		return fmt.Errorf("running pre-backup script: %w", err)
	}

	vol, err := a.Volume(ctx)
	if err != nil {
		return fmt.Errorf("getting volume: %w", err)
	}

	gw := gzip.NewWriter(w)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	if err := writeTarEntry(tw, backupAppSettingsEntry, []byte(a.Settings.Marshal())); err != nil {
		return fmt.Errorf("writing application settings: %w", err)
	}

	if err := writeTarEntry(tw, backupVolSettingsEntry, []byte(vol.Settings.Marshal())); err != nil {
		return fmt.Errorf("writing volume settings: %w", err)
	}

	reader, _, err := a.namespace.client.CopyFromContainer(ctx, containerName, AppVolumeMountTargets[0])
	if err != nil {
		return fmt.Errorf("copying from container: %w", err)
	}
	defer reader.Close()

	if err := copyTarEntriesWithPrefix(reader, tw, filepath.Base(AppVolumeMountTargets[0]), BackupDataDir); err != nil {
		return fmt.Errorf("copying volume contents: %w", err)
	}

	return nil
}

func (n *Namespace) parseBackup(r io.Reader) (ApplicationSettings, ApplicationVolumeSettings, []byte, error) {
	var appSettings ApplicationSettings
	var volSettings ApplicationVolumeSettings
	var volumeData bytes.Buffer

	gr, err := gzip.NewReader(r)
	if err != nil {
		return appSettings, volSettings, nil, fmt.Errorf("%w: %v", ErrInvalidBackup, err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	tw := tar.NewWriter(&volumeData)
	defer tw.Close()

	foundApp := false
	foundVol := false

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return appSettings, volSettings, nil, fmt.Errorf("%w: %v", ErrInvalidBackup, err)
		}

		switch header.Name {
		case backupAppSettingsEntry:
			data, err := io.ReadAll(tr)
			if err != nil {
				return appSettings, volSettings, nil, fmt.Errorf("%w: reading application settings: %v", ErrInvalidBackup, err)
			}
			appSettings, err = UnmarshalApplicationSettings(string(data))
			if err != nil {
				return appSettings, volSettings, nil, fmt.Errorf("%w: parsing application settings: %v", ErrInvalidBackup, err)
			}
			foundApp = true

		case backupVolSettingsEntry:
			data, err := io.ReadAll(tr)
			if err != nil {
				return appSettings, volSettings, nil, fmt.Errorf("%w: reading volume settings: %v", ErrInvalidBackup, err)
			}
			volSettings, err = UnmarshalApplicationVolumeSettings(string(data))
			if err != nil {
				return appSettings, volSettings, nil, fmt.Errorf("%w: parsing volume settings: %v", ErrInvalidBackup, err)
			}
			foundVol = true

		default:
			if header.Name == BackupDataDir || strings.HasPrefix(header.Name, BackupDataDir+"/") {
				newHeader := *header
				if header.Name == BackupDataDir {
					newHeader.Name = "data"
				} else {
					newHeader.Name = "data" + strings.TrimPrefix(header.Name, BackupDataDir)
				}
				if err := tw.WriteHeader(&newHeader); err != nil {
					return appSettings, volSettings, nil, err
				}
				if header.Size > 0 {
					if _, err := io.Copy(tw, tr); err != nil {
						return appSettings, volSettings, nil, err
					}
				}
			}
		}
	}

	if !foundApp || !foundVol {
		return appSettings, volSettings, nil, fmt.Errorf("%w: missing required metadata files", ErrInvalidBackup)
	}

	return appSettings, volSettings, volumeData.Bytes(), nil
}

// Helpers

func prepareBackupDir(dir string) (int, int, error) {
	if dir == "" {
		return 0, 0, fmt.Errorf("backup location is required")
	}

	if !filepath.IsAbs(dir) {
		return 0, 0, ErrBackupPathRelative
	}

	uid, gid, err := findOwnership(dir)
	if err != nil {
		return 0, 0, fmt.Errorf("determining backup directory ownership: %w", err)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, 0, fmt.Errorf("creating backup directory: %w", err)
	}

	if err := chownNewDirs(dir, uid, gid); err != nil {
		return 0, 0, fmt.Errorf("setting backup directory ownership: %w", err)
	}

	return uid, gid, nil
}

func findOwnership(dir string) (int, int, error) {
	for path := dir; ; path = filepath.Dir(path) {
		info, err := os.Stat(path)
		if err == nil {
			stat := info.Sys().(*syscall.Stat_t)
			return int(stat.Uid), int(stat.Gid), nil
		}
		if !os.IsNotExist(err) {
			return 0, 0, err
		}
		if path == "/" {
			return 0, 0, fmt.Errorf("no existing parent directory found for %s", dir)
		}
	}
}

func chownNewDirs(dir string, uid, gid int) error {
	// Collect dirs from deepest to shallowest, stopping at the first
	// one that already has the correct ownership.
	var dirs []string
	for path := dir; ; path = filepath.Dir(path) {
		info, err := os.Stat(path)
		if err != nil {
			break
		}
		stat := info.Sys().(*syscall.Stat_t)
		if int(stat.Uid) == uid && int(stat.Gid) == gid {
			break
		}
		dirs = append(dirs, path)
		if path == "/" {
			break
		}
	}

	for _, d := range dirs {
		if err := os.Chown(d, uid, gid); err != nil {
			return err
		}
	}
	return nil
}

func parseBackupTime(appName, filename string) (time.Time, bool) {
	prefix := appName + "-"
	suffix := ".tar.gz"

	if !strings.HasPrefix(filename, prefix) || !strings.HasSuffix(filename, suffix) {
		return time.Time{}, false
	}

	middle := strings.TrimPrefix(filename, prefix)
	middle = strings.TrimSuffix(middle, suffix)

	t, err := time.Parse(backupTimeFormat, middle)
	if err != nil {
		return time.Time{}, false
	}

	return t, true
}

func writeTarEntry(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func copyTarEntriesWithPrefix(src io.Reader, dst *tar.Writer, oldPrefix, newPrefix string) error {
	tr := tar.NewReader(src)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if oldPrefix != "" && newPrefix != "" {
			if header.Name == oldPrefix {
				header.Name = newPrefix
			} else if strings.HasPrefix(header.Name, oldPrefix+"/") {
				header.Name = newPrefix + strings.TrimPrefix(header.Name, oldPrefix)
			}
		}

		if err := dst.WriteHeader(header); err != nil {
			return err
		}

		if header.Size > 0 {
			if _, err := io.Copy(dst, tr); err != nil {
				return err
			}
		}
	}
}
