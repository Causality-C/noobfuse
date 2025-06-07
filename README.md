# noobfuse

You like FUSE? You like Go?  
This is a toy **read-only FUSE filesystem** written in Go.

It mounts a virtual filesystem at `/tmp/mnt/all-projects`, pulls files from a simulated `nfs/` backend, and caches them on first read to a local `ssd/` directory.

## ✨ Features

- Read-only FUSE mount (no writes, no deletes)
- Directory listing directly from `nfs/`
- File reads with transparent SSD cache
- 500ms delay on first NFS read to simulate slowness
- Caches files on disk (not memory)
- Treats files with the same name as interchangeable (no content hash)

## 🗂 Directory Structure

```bash
nfs
├── empty.py
├── project-1
│   ├── common-lib.py
│   └── main.py
├── project-2
│   └── common-lib.py
└── project-3
    ├── embed
    │   └── embedding.py
    └── project-3.py
ssd/                # starts empty, populated on first read
```

## 🚀 How It Works

1. Mount the FUSE FS at `/tmp/mnt/all-projects`
2. `ls` lists files directly from `nfs/`
3. `cat` or `python` reads a file:
   - If cached in `ssd/`: read immediately
   - If not cached: sleep 500ms, read from `nfs/`, write to `ssd/`, return

## 🧪 Run & Test

### 1. Install FUSE

You need FUSE installed to run any FUSE-based filesystem.

🐧 Linux

```bash
# Debian / Ubuntu
sudo apt update
sudo apt install fuse3

# Arch
sudo pacman -S fuse3

# Red Hat / CentOS
sudo yum install fuse3
```
---

### 2. Run the Filesystem

```bash
go mod tidy
go run main.go
```

You'll see log output for mounting, lookups, file reads, and caching.

Control + C to exit

---

### 3. Try It Out

```bash
ls /tmp/mnt/all-projects/
stat /tmp/mnt/all-projects/project-1/main.py
cat /tmp/mnt/all-projects/project-2/entrypoint.py
head /tmp/mnt/all-projects/project-2/entrypoint.py
```

✔️ The first `cat` adds a 500ms delay and copies to `ssd/`  
✔️ The second read is instant (cache hit)

### 4. (Optional) Run the Tests

```
go test -v
```
Look into `noobfs_test.go` for more info on what we test


## ⚠️ Limitations / Improvements
1. One of the requirements of this project is that files are cached on one level meaning that two files that have the same name but different directories will lead to the same cache hit. This means that by the chance both files are different, the cache will return whatever file came first
2. Another thing is that this file system is read-only, which mean only a subset of methods for the FS, Dir, and File are implemented; a more complete implementation will allow modifying files and changing their permissions
3. The caching strategy is also naive, and a more substantial example can include a cache invalidation strategy, which can perhaps address limitation 1 (ie. evicting the old file of the same name)
4. The way the current implementation is, we cannot control where the projects directory is pointing to (its static); we can see further versions implement that improvement