package main

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type ResultColumn struct{
	mountData *base.MountData

	inode uint64
	data  []byte
}

func NewResultColumn(mountData *base.MountData, parentInode uint64, name string, value []byte) (c ResultColumn) {
	c.inode = fs.GenerateDynamicInode(parentInode, name)
	c.data = value
	c.mountData = mountData
	return
}

func (c ResultColumn) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = c.inode
	a.Uid = c.mountData.Uid
	a.Gid = c.mountData.Gid
	a.Mode = 0o444
	a.Size = uint64(len(c.data))
	return nil
}

func (c ResultColumn) ReadAll(ctx context.Context) ([]byte, error) {
	return c.data, nil
}
