package walker

import (
	"os"
	"path/filepath"

	"github.com/KashifKhn/remove-comments/cli/internal/languages"
	"github.com/boyter/gocodewalker"
)

type FileEntry struct {
	Path string
	Ext  string
	Lang languages.LangConfig
}

func Walk(root string, langFilter string, maxFileSize int64, excludePatterns []string) ([]FileEntry, []error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, []error{err}
	}
	if !info.IsDir() {
		return walkSingleFile(root, langFilter, maxFileSize, excludePatterns)
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
		if matchesAny(f.Location, excludePatterns) {
			continue
		}

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

func walkSingleFile(path string, langFilter string, maxFileSize int64, excludePatterns []string) ([]FileEntry, []error) {
	if matchesAny(path, excludePatterns) {
		return nil, nil
	}
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

func matchesAny(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	base := filepath.Base(path)
	normalized := filepath.ToSlash(path)
	for _, pattern := range patterns {
		p := filepath.ToSlash(pattern)
		if matched, _ := filepath.Match(p, base); matched {
			return true
		}
		if matched, _ := filepath.Match(p, normalized); matched {
			return true
		}
		if matchDoubleGlob(normalized, p) {
			return true
		}
	}
	return false
}

func matchDoubleGlob(path, pattern string) bool {
	if len(pattern) < 3 {
		return false
	}
	idx := 0
	for idx <= len(pattern)-2 {
		if pattern[idx] == '*' && pattern[idx+1] == '*' {
			prefix := pattern[:idx]
			suffix := pattern[idx+2:]
			if len(suffix) > 0 && suffix[0] == '/' {
				suffix = suffix[1:]
			}
			if prefix != "" {
				if matched, _ := filepath.Match(prefix+"*", path[:min(len(path), len(prefix)+1)]); !matched {
					if !startsWith(filepath.ToSlash(path), filepath.ToSlash(prefix)) {
						return false
					}
				}
			}
			if suffix == "" {
				return true
			}
			matched, _ := filepath.Match(suffix, filepath.Base(path))
			return matched
		}
		idx++
	}
	return false
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
