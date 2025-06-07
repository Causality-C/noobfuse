package main

import (
	"log"
	"noobfuse/lib"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func main() {
	// TODO: create the nfs and ssd directories
	mountpoint := lib.Mountpoint

	// Note: we need to create the mountpoint directory here, otherwise the kernel will not be able to mount the filesystem
	if err := os.MkdirAll(mountpoint, 0755); err != nil {
		log.Fatal(err)
	}

	// Reset SSD directory
	lib.ResetSsd()

	// Note: ReadOnly() only restricts the client from writing to the filesystem
	c, err := fuse.Mount(mountpoint, fuse.FSName("noobfuse"), fuse.Subtype("noobfuse"), fuse.ReadOnly())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigc
		log.Printf("Received signal %s, unmounting...", sig)
		fuse.Unmount(mountpoint)
		os.Exit(0)
	}()

	err = fs.Serve(c, &lib.NoobFS{})
	if err != nil {
		log.Fatal(err)
	}
}
