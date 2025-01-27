package main

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type ResultColumn struct{
	inode uint64
	data  []byte
}

func NewResultColumn(parentInode uint64, name string, value []byte) (c ResultColumn) {
	c.inode = fs.GenerateDynamicInode(parentInode, name)
	c.data = value
	return
}

func (c ResultColumn) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = c.inode
	a.Uid = uid
	a.Gid = gid
	a.Mode = 0o444
	a.Size = uint64(len(c.data))
	return nil
}

func (c ResultColumn) ReadAll(ctx context.Context) ([]byte, error) {
	return c.data, nil
}
