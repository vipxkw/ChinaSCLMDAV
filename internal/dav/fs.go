// Package dav implements the WebDAV protocol handler backed by the store's
// per-user filesystem, with delete-to-trash and overwrite version snapshots.
package dav

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	webdav "golang.org/x/net/webdav"

	"chinasclmdav/internal/fsys"
	"chinasclmdav/internal/store"
)

// file is a wrapper adding filtered directory listing.
type file struct {
	*os.File
	owner    *FileSystem
	infos    []os.FileInfo // lazily loaded directory entries
	pos      int
	listing  bool
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	if !f.listing {
		// Could happen if a client reads a file's directory; fall back to os.File.
		return f.File.Readdir(count)
	}
	if f.infos == nil {
		infos, err := f.File.Readdir(-1)
		if err != nil && err != io.EOF {
			return nil, err
		}
		filtered := infos[:0]
		for _, fi := range infos {
			if fsys.IsHidden(fi.Name(), f.owner.ignoreSecrets()) {
				continue
			}
			filtered = append(filtered, fi)
		}
		f.infos = filtered
	}
	if f.pos >= len(f.infos) {
		f.pos = len(f.infos)
		if count >= 0 {
			return []os.FileInfo{}, io.EOF
		}
		return []os.FileInfo{}, io.EOF
	}
	end := len(f.infos)
	if count > 0 && f.pos+count < end {
		end = f.pos + count
	}
	r := f.infos[f.pos:end]
	f.pos = end
	return r, nil
}

// FileSystem implements x/net/webdav.FileSystem over a user's workspace.
type FileSystem struct {
	store *store.Store
	user  *store.User
	root  string // user workspace root
}

// NewFileSystem builds a new FileSystem for a user.
func NewFileSystem(st *store.Store, u *store.User) *FileSystem {
	return &FileSystem{store: st, user: u, root: st.UserRoot(u.ID)}
}

func (f *FileSystem) ignoreSecrets() []string {
	ig, _ := f.store.GetSetting("ignore", ".git,.DS_Store,*.log")
	return strings.Split(ig, ",")
}

func (f *FileSystem) path(name string) (string, error) {
	return f.store.Resolve(f.user.ID, name)
}

// Mkdir implements webdav.FileSystem.
func (f *FileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	abs, err := f.path(name)
	if err != nil {
		return err
	}
	return os.Mkdir(abs, perm)
}

// OpenFile implements webdav.FileSystem.
func (f *FileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	abs, err := f.path(name)
	if err != nil {
		return nil, err
	}
	// Snapshots the old content before an overwrite.
	if flag&(os.O_WRONLY|os.O_RDWR) != 0 {
		if err := f.snapshotVersion(name, abs); err != nil {
			return nil, err
		}
		if flag&os.O_CREATE != 0 {
			if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
				return nil, err
			}
		}
	}
	osf, err := os.OpenFile(abs, flag, perm)
	if err != nil {
		return nil, err
	}
	// Determine whether this is an existing directory for listing.
	info, err := osf.Stat()
	if err == nil && info.IsDir() {
		return &file{File: osf, owner: f, listing: true}, nil
	}
	return &file{File: osf, owner: f}, nil
}

// snapshotVersion copies the old file content to the version store.
func (f *FileSystem) snapshotVersion(name, abs string) error {
	info, err := os.Stat(abs)
	if err != nil || !info.Mode().IsRegular() {
		return nil // nothing to snapshot
	}
	clean := strings.TrimPrefix(filepath.Clean("/"+strings.TrimPrefix(name, "/")), "/")
	key := sanitize(clean)
	dir := filepath.Join(f.store.UserVersionDir(f.user.ID), key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	ts := time.Now().UTC().Format("20060102_150405.000000000")
	snap := filepath.Join(dir, ts+"_"+info.Name())
	if err := copyFile(abs, snap); err != nil {
		return err
	}
	_, err = f.store.AddVersion(f.user.ID, clean, info.Name(), snap, info.Size())
	return err
}

// RemoveAll implements webdav.FileSystem (moves to trash).
func (f *FileSystem) RemoveAll(ctx context.Context, name string) error {
	abs, err := f.path(name)
	if err != nil {
		return err
	}
	return f.trashItem(name, abs)
}

// trashItem moves a file/folder to the user's trash and records it.
func (f *FileSystem) trashItem(name, abs string) error {
	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	trash := f.store.UserTrashDir(f.user.ID)
	dst := filepath.Join(trash, fmt.Sprintf("%d_%s", time.Now().UnixNano(), info.Name()))
	if err := os.Rename(abs, dst); err != nil {
		return err
	}
	size, _ := fsys.DirSize(dst, info.IsDir())
	clean := strings.TrimPrefix(filepath.Clean("/"+strings.TrimPrefix(name, "/")), "/")
	_, err = f.store.AddTrash(f.user.ID, "/"+clean, info.Name(), dst, info.IsDir(), size)
	if err != nil {
		// Best-effort: keep the file moved even if record fails.
		_ = os.Rename(dst, abs)
		return err
	}
	return nil
}

// Rename implements webdav.FileSystem.
func (f *FileSystem) Rename(ctx context.Context, oldName, newName string) error {
	oldAbs, err := f.path(oldName)
	if err != nil {
		return err
	}
	newAbs, err := f.path(newName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(newAbs), 0o755); err != nil {
		return err
	}
	return os.Rename(oldAbs, newAbs)
}

// Stat implements webdav.FileSystem.
func (f *FileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	abs, err := f.path(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(abs)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func sanitize(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteString("_")
		}
	}
	return b.String()
}