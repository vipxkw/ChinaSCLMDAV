package fsys

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Entry describes one directory or file entry for listing.
type Entry struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	ModTime  int64  `json:"mod_time"`
}

// Breadcrumb is one segment of the navigation trail.
type Breadcrumb struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Category is a storage classification bucket.
type Category struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Size  int64  `json:"size"`
	Count int64  `json:"count"`
}

// IsHidden reports whether a name should be hidden from listings
// (system dirs and files in the ignore list).
func IsHidden(name string, ignore []string) bool {
	if name == "." || name == ".." {
		return true
	}
	for _, p := range ignore {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if matched, err := filepath.Match(p, name); err == nil && matched {
			return true
		}
		if strings.Contains(p, "*") {
			if re, err := regexp.Compile("^" + globToRegex(p) + "$"); err == nil && re.MatchString(name) {
				return true
			}
		}
	}
	return false
}

func globToRegex(p string) string {
	var b strings.Builder
	for _, r := range p {
		switch r {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteString(".")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// List reads a directory. pattern, when non-empty, is a regex matched against
// file names (recursively up to a sane depth).
func List(abs string, pattern string) ([]Entry, error) {
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	var out []Entry
	aRel := abs
	for _, de := range entries {
		info, err := de.Info()
		if err != nil {
			continue
		}
		e := Entry{
			Name:    de.Name(),
			Path:    filepath.Join(aRel, de.Name()),
			IsDir:   de.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Unix(),
		}
		if pattern != "" {
			re, err := regexp.Compile(pattern)
			if err == nil && !re.MatchString(de.Name()) {
				continue
			}
		}
		out = append(out, e)
	}
	return out, nil
}

// WalkSearch recursively searches for names matching a regex pattern.
func WalkSearch(abs, pattern string, maxDepth int) ([]Entry, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	var out []Entry
	var walk func(dir string, depth int) error
	walk = func(dir string, depth int) error {
		if depth > maxDepth {
			return nil
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}
		for _, de := range entries {
			info, err := de.Info()
			if err != nil {
				continue
			}
			if re.MatchString(de.Name()) {
				out = append(out, Entry{
					Name:    de.Name(),
					Path:    filepath.Join(dir, de.Name()),
					IsDir:   de.IsDir(),
					Size:    info.Size(),
					ModTime: info.ModTime().Unix(),
				})
			}
			if de.IsDir() {
				if err := walk(filepath.Join(dir, de.Name()), depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(abs, 0); err != nil {
		return nil, err
	}
	return out, nil
}

// Breadcrumbs builds the breadcrumb trail from an absolute path.
func Breadcrumbs(abs, rootAbs string) []Breadcrumb {
	rel, err := filepath.Rel(rootAbs, abs)
	if err != nil {
		return nil
	}
	if rel == "." {
		return []Breadcrumb{}
	}
	parts := strings.Split(filepath.Clean(rel), string(filepath.Separator))
	var crumbs []Breadcrumb
	cur := rootAbs
	for _, p := range parts {
		if p == "" {
			continue
		}
		cur = filepath.Join(cur, p)
		crumbs = append(crumbs, Breadcrumb{Name: p, Path: cur})
	}
	return crumbs
}

// DirSize computes the total size (and optionally count) under a path.
func DirSize(abs string, isDir bool) (int64, int64) {
	info, err := os.Stat(abs)
	if err != nil {
		return 0, 0
	}
	if !isDir {
		return info.Size(), 1
	}
	var size, count int64
	filepath.Walk(abs, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			count++
			return nil
		}
		size += fi.Size()
		count++
		return nil
	})
	return size, count
}

// categorize maps a filename to a storage category.
func categorize(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	low := strings.ToLower(name)
	switch {
	case ext == ".pdf":
		return "documents"
	case ext == ".doc" || ext == ".docx" || ext == ".txt" || ext == ".md":
		return "documents"
	case ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp" ||
		ext == ".svg" || ext == ".ico" || ext == ".bmp":
		return "images"
	case ext == ".mp4" || ext == ".mkv" || ext == ".mov" || ext == ".avi" || ext == ".webm":
		return "video"
	case ext == ".mp3" || ext == ".wav" || ext == ".flac" || ext == ".ogg" || ext == ".aac":
		return "audio"
	case ext == ".zip" || ext == ".tar" || ext == ".gz" || ext == ".7z" || ext == ".rar" ||
		ext == ".bz2" || ext == ".xz":
		return "archives"
	case strings.HasPrefix(low, "README") || low == "LICENSE":
		return "documents"
	default:
		return "other"
	}
}

// Categories computes storage usage by file type under a directory tree.
func Categories(abs string) ([]Category, int64, int64, error) {
	agg := map[string]int64{} // key -> size
	cnt := map[string]int64{} // key -> count
	var totalSize, fileCount, dirCount int64
	err := filepath.Walk(abs, func(_ string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		name := fi.Name()
		if name == ".trash" || name == ".versions" || strings.HasPrefix(name, ".") {
			if fi.IsDir() && (name == ".trash" || name == ".versions") {
				return filepath.SkipDir
			}
		}
		if fi.IsDir() {
			dirCount++
			return nil
		}
		key := categorize(name)
		agg[key] += fi.Size()
		cnt[key]++
		totalSize += fi.Size()
		fileCount++
		return nil
	})
	if err != nil {
		return nil, 0, 0, err
	}
	labels := map[string]string{
		"documents": "Documents",
		"images":    "Images",
		"video":     "Videos",
		"audio":     "Audio",
		"archives":  "Archives",
		"other":     "Other",
	}
	order := []string{"documents", "images", "video", "audio", "archives", "other"}
	var out []Category
	for _, k := range order {
		if cnt[k] == 0 {
			continue
		}
		out = append(out, Category{Key: k, Label: labels[k], Size: agg[k], Count: cnt[k]})
	}
	return out, totalSize, fileCount, nil
}

// Zip writes the contents of a directory or file to w as a zip archive.
// Rel paths inside the archive are relative to baseAbs (the item's parent
// for directories, or the file name for files).
func Zip(w io.Writer, abs, baseName string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	return addToZip(zw, abs, baseName)
}

func addToZip(zw *zip.Writer, abs, zipPath string) error {
	info, err := os.Stat(abs)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return addFile(zw, abs, zipPath)
	}
	// walk directory
	return filepath.Walk(abs, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(filepath.Join(zipPath, rel))
		return addFile(zw, path, rel)
	})
}

func addFile(zw *zip.Writer, abs, name string) error {
	name = strings.TrimPrefix(name, "/")
	h := &zip.FileHeader{Name: name, Method: zip.Deflate}
	h.SetModTime(fileInfoModTime(abs))
	w, err := zw.CreateHeader(h)
	if err != nil {
		return err
	}
	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

func fileInfoModTime(path string) (t time.Time) {
	if fi, err := os.Stat(path); err == nil {
		return fi.ModTime()
	}
	return
}