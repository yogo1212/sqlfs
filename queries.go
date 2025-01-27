package main

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type queriesInfo struct {
	queryHandles *QueryHandles
}

type Queries struct{
	i *queriesInfo
}

func NewQueries() (q Queries) {
	q.i = &queriesInfo{}
	return
}

func (Queries) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = InodeQueries
	a.Uid = uid
	a.Gid = gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (q Queries) Lookup(ctx context.Context, name string) (fs.Node, error) {
	switch name {
	case "handles":
		if q.i.queryHandles == nil {
			s := NewQueryHandles(InodeQueries)
			q.i.queryHandles = &s
			return s, nil
		}

		return *q.i.queryHandles, nil
	default:
	}

	return nil, syscall.ENOENT
}

func (q Queries) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	entries := []fuse.Dirent{
		{Inode: fs.GenerateDynamicInode(InodeQueries, "handles"), Name: "handles", Type: fuse.DT_Dir},
		// {Inode: fs.GenerateDynamicInode(q.i.inode, "single"), Name: "single", Type: fuse.DT_File},
		// {Inode: fs.GenerateDynamicInode(q.i.inode, "handles"), Name: "handles", Type: fuse.DT_Dir},
		// {Inode: fs.GenerateDynamicInode(q.i.inode, "results"), Name: "results", Type: fuse.DT_File},
	}
	return entries, nil
}
