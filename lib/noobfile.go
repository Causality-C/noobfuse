package lib

import (
	"context"
	"fmt"
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
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat SSD file %s: %w", ssdPath, err)
	} else {
		a.Size = n.sizeBytes
	}

	log.Printf("Attr File: %s", n.relativePath)
	a.Inode = inodeForPath(n.relativePath)
	a.Mode = 0644
	return nil
}

func (n *NoobFile) ReadAll(ctx context.Context) ([]byte, error) {
	ssdPath := filepath.Join(ssdRoot, filepath.Base(n.relativePath))
	nfsPath := filepath.Join(nfsRoot, n.relativePath)

	// Try reading from SSD first
	if content, err := os.ReadFile(ssdPath); err == nil {
		log.Printf("ReadAll from ssd: %s", ssdPath)
		return content, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read from SSD %s: %w", ssdPath, err)
	}

	// If not in SSD, read from NFS
	time.Sleep(500 * time.Millisecond)
	content, err := os.ReadFile(nfsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fuse.ENOENT
		}
		if os.IsPermission(err) {
			return nil, fuse.EPERM
		}
		return nil, fmt.Errorf("failed to read from NFS %s: %w", nfsPath, err)
	}
	log.Printf("ReadAll from nfs: %s", nfsPath)

	// Write to SSD cache
	if err := os.MkdirAll(filepath.Dir(ssdPath), 0755); err != nil {
		log.Printf("Warning: failed to create SSD cache directory: %v", err)
		return content, nil // Return content anyway since we have it
	}

	if err := os.WriteFile(ssdPath, content, 0644); err != nil {
		log.Printf("Warning: failed to write to SSD cache: %v", err)
		return content, nil // Return content anyway since we have it
	}

	return content, nil
}
