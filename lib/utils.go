package lib

import (
	"hash/fnv"
	"log"
	"os"
)

func inodeForPath(path string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(path))
	return h.Sum64()
}

// We simulate a filesystem with two directories: nfs and ssd
var (
	nfsRoot = "./nfs"
	ssdRoot = "./ssd"
)

const Mountpoint = "/tmp/mnt/all-projects"

func ResetSsd() {
	log.Printf("Resetting ssd directory: %s", ssdRoot)
	err := os.RemoveAll(ssdRoot)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(ssdRoot, 0755)
	if err != nil {
		log.Fatal(err)
	}
}
