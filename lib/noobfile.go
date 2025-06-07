package lib

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"bazil.org/fuse"
)

type NoobFile struct {
	relativePath string
	sizeBytes    uint64
}

func (n *NoobFile) Attr(ctx context.Context, a *fuse.Attr) error {
	ssdPath := filepath.Join(ssdRoot, filepath.Base(n.relativePath))

	// Custom cache logic: if the file exists in ssd, use the size from ssd
	if info, err := os.Stat(ssdPath); err == nil {
		a.Size = uint64(info.Size())
	} else {
		a.Size = n.sizeBytes
	}

	log.Printf("Attr File: %s", n.relativePath)
	a.Inode = inodeForPath(n.relativePath)
	a.Mode = 0644
	return nil
}

func (n *NoobFile) ReadAll(ctx context.Context) ([]byte, error) {
	// Implement nfs and ssd read
	ssdPath := filepath.Join(ssdRoot, filepath.Base(n.relativePath))
	// ssdPath := filepath.Join(ssdRoot, n.relativePath)
	// Custom cache logic
	if content, err := os.ReadFile(ssdPath); err == nil {
		log.Printf("ReadAll from ssd: %s", ssdPath)
		return content, nil
	}
	time.Sleep(500 * time.Millisecond)
	nfsPath := filepath.Join(nfsRoot, n.relativePath)
	content, err := os.ReadFile(nfsPath)
	if err != nil {
		return nil, err
	}
	log.Printf("ReadAll from nfs: %s", nfsPath)

	// Write to ssd
	os.MkdirAll(filepath.Dir(ssdPath), 0755)
	os.WriteFile(ssdPath, content, 0644)
	return content, nil
}
