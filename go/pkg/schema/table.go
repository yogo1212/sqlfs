package schema

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type tableInfo struct{
	inode uint64
}

type Table struct{
	i *tableInfo

	data *base.MountData
}

func NewTable(data *base.MountData, ts Tables, name string) (t Table, err error) {
	t.data = data
	t.i = &tableInfo{
		inode: fs.GenerateDynamicInode(ts.i.inode, name),
	}

	return
}

func (t Table) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = t.i.inode
	a.Uid = t.data.Uid
	a.Gid = t.data.Gid
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
