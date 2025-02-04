package queries

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type queriesInfo struct {
	queryHandles *QueryHandles
}

type Queries struct{
	i *queriesInfo
	data *base.MountData

	inode uint64
}

func NewQueries(data *base.MountData, parentInode uint64) (q Queries) {
	q.data = data
	q.i = &queriesInfo{}
	q.inode = fs.GenerateDynamicInode(parentInode, "queries")
	return
}

func (q Queries) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = q.inode
	a.Uid = q.data.Uid
	a.Gid = q.data.Gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (q Queries) Lookup(ctx context.Context, name string) (fs.Node, error) {
	switch name {
	case "handles":
		if q.i.queryHandles == nil {
			s := NewQueryHandles(q.data, q.inode)
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
		{Inode: fs.GenerateDynamicInode(q.inode, "handles"), Name: "handles", Type: fuse.DT_Dir},
		// {Inode: fs.GenerateDynamicInode(q.i.inode, "single"), Name: "single", Type: fuse.DT_File},
		// {Inode: fs.GenerateDynamicInode(q.i.inode, "handles"), Name: "handles", Type: fuse.DT_Dir},
		// {Inode: fs.GenerateDynamicInode(q.i.inode, "results"), Name: "results", Type: fuse.DT_File},
	}
	return entries, nil
}
