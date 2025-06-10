package lib

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type NoobDir struct{ relativePath string }

// Attr implements fs.Node.
func (n *NoobDir) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Printf("Attr Dir: %s", n.relativePath)
	a.Inode = inodeForPath(n.relativePath)
	a.Mode = os.ModeDir | 0755 // superuser can do everything, everyone else can only read and execute
	return nil
}

func (n *NoobDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Printf("ReadDirAll: %s", n.relativePath)
	entries := []fuse.Dirent{}

	dirPath := filepath.Join(nfsRoot, n.relativePath)
	res, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fuse.ENOENT
		}
		if os.IsPermission(err) {
			return nil, fuse.EPERM
		}
		return nil, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, entry := range res {
		if entry.IsDir() {
			entries = append(entries, fuse.Dirent{
				Name:  entry.Name(),
				Type:  fuse.DT_Dir,
				Inode: inodeForPath(entry.Name()),
			})
		} else {
			entries = append(entries, fuse.Dirent{
				Name:  entry.Name(),
				Type:  fuse.DT_File,
				Inode: inodeForPath(entry.Name()),
			})
		}
	}
	return entries, nil
}

func (n *NoobDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	fullPath := filepath.Join(nfsRoot, n.relativePath, name)
	log.Printf("Lookup: %s", fullPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fuse.ENOENT
		}
		if os.IsPermission(err) {
			return nil, fuse.EPERM
		}
		return nil, fmt.Errorf("failed to stat path %s: %w", fullPath, err)
	}

	if info.IsDir() {
		return &NoobDir{relativePath: filepath.Join(n.relativePath, name)}, nil
	} else {
		return &NoobFile{relativePath: filepath.Join(n.relativePath, name), sizeBytes: uint64(info.Size())}, nil
	}
}
