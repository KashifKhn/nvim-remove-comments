package walker

import (
	"os"
	"path/filepath"

	"github.com/KashifKhn/nvim-remove-comments/cli/internal/languages"
	"github.com/boyter/gocodewalker"
)

type FileEntry struct {
	Path string
	Ext  string
	Lang languages.LangConfig
}

func Walk(root string, langFilter string, maxFileSize int64) ([]FileEntry, []error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, []error{err}
	}
	if !info.IsDir() {
		return walkSingleFile(root, langFilter, maxFileSize)
	}

	queue := make(chan *gocodewalker.File, 512)

	exts := buildAllowList(langFilter)

	fw := gocodewalker.NewFileWalker(root, queue)
	fw.AllowListExtensions = exts
	fw.ExcludeDirectory = []string{".git", "node_modules", "vendor", ".idea", ".vscode"}

	var walkErr error
	fw.SetErrorHandler(func(e error) bool {
		walkErr = e
		return true
	})

	go func() {
		_ = fw.Start()
	}()

	var entries []FileEntry
	var errs []error

	for f := range queue {
		ext := filepath.Ext(f.Filename)
		cfg, ok := languages.Get(ext)
		if !ok {
			continue
		}
		if langFilter != "" && cfg.Name != langFilter {
			continue
		}

		fi, statErr := os.Stat(f.Location)
		if statErr != nil {
			errs = append(errs, statErr)
			continue
		}
		if maxFileSize > 0 && fi.Size() > maxFileSize {
			continue
		}

		entries = append(entries, FileEntry{
			Path: f.Location,
			Ext:  ext,
			Lang: cfg,
		})
	}

	if walkErr != nil {
		errs = append(errs, walkErr)
	}

	return entries, errs
}

func walkSingleFile(path string, langFilter string, maxFileSize int64) ([]FileEntry, []error) {
	ext := filepath.Ext(path)
	cfg, ok := languages.Get(ext)
	if !ok {
		return nil, nil
	}
	if langFilter != "" && cfg.Name != langFilter {
		return nil, nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return nil, []error{err}
	}
	if maxFileSize > 0 && fi.Size() > maxFileSize {
		return nil, nil
	}
	return []FileEntry{{Path: path, Ext: ext, Lang: cfg}}, nil
}

func buildAllowList(langFilter string) []string {
	var source []string
	if langFilter == "" {
		source = languages.Supported()
	} else {
		for _, ext := range languages.Supported() {
			cfg, ok := languages.Get(ext)
			if ok && cfg.Name == langFilter {
				source = append(source, ext)
			}
		}
	}
	exts := make([]string, 0, len(source))
	for _, ext := range source {
		if len(ext) > 1 && ext[0] == '.' {
			exts = append(exts, ext[1:])
		} else {
			exts = append(exts, ext)
		}
	}
	return exts
}
