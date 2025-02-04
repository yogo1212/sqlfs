package schema

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type schemaInfo struct {
	inode uint64
	name  string

	tables *Tables
}

type Schema struct{
	data *base.MountData
	i *schemaInfo
}

func NewSchema(data *base.MountData, parentInode uint64, name string) (s Schema, err error) {
	s.data = data
	s.i = &schemaInfo{
		inode: fs.GenerateDynamicInode(parentInode, name),
		name: name,
	}

	return
}

func (s Schema) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = s.i.inode
	a.Uid = s.data.Uid
	a.Gid = s.data.Gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (s Schema) Lookup(ctx context.Context, name string) (fs.Node, error) {
	switch name {
	case "tables":
		if s.i.tables != nil {
			return s.i.tables, nil
		}

		t, err := NewTables(s.data, s)
		if err == nil {
			s.data.PrintErr(err)
			s.i.tables = &t
		}

		return t, err
	// TODO data types, functions, operators
	default:
	}
	return nil, syscall.ENOENT
}

func (s Schema) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	dirs := []fuse.Dirent{
		{Inode: fs.GenerateDynamicInode(s.i.inode, "tables"), Name: "tables", Type: fuse.DT_Dir},
	}
	return dirs, nil
}
