package lib

import "bazil.org/fuse/fs"

type NoobFS struct{}

func (fs *NoobFS) Root() (fs.Node, error) {
	return &NoobDir{}, nil
}
