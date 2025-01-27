package main

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type tableInfo struct{
	inode uint64
}

type Table struct{
	i *tableInfo
}

func NewTable(ts Tables, name string) (t Table, err error) {
	t.i = &tableInfo{
		inode: fs.GenerateDynamicInode(ts.i.inode, name),
	}

	return
}

func (t Table) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = t.i.inode
	a.Uid = uid
	a.Gid = gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (t Table) Lookup(ctx context.Context, name string) (fs.Node, error) {
	switch name {
	// columns, sequences, indexes
	default:
	}
	return nil, syscall.ENOENT
}

func (t Table) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	tables := []fuse.Dirent{
		{Inode: fs.GenerateDynamicInode(t.i.inode, "columns"), Name: "columns", Type: fuse.DT_Dir},
		{Inode: fs.GenerateDynamicInode(t.i.inode, "schemata"), Name: "schemata", Type: fuse.DT_Dir},
		{Inode: fs.GenerateDynamicInode(t.i.inode, "indexes"), Name: "indexes", Type: fuse.DT_Dir},
	}
	return tables, nil
}
