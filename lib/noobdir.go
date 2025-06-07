package lib

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"syscall"

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

	res, err := os.ReadDir(filepath.Join(nfsRoot, n.relativePath))
	if err != nil {
		return nil, err
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
		return nil, syscall.ENOENT
	}

	if info.IsDir() {
		return &NoobDir{relativePath: filepath.Join(n.relativePath, name)}, nil
	} else {
		return &NoobFile{relativePath: filepath.Join(n.relativePath, name), sizeBytes: uint64(info.Size())}, nil
	}
}
