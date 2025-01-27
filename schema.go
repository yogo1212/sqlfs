package main

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type schemaInfo struct {
	inode uint64
	name  string

	tables *Tables
}

type Schema struct{
	i *schemaInfo
}

func NewSchema(name string) (s Schema, err error) {
	s.i = &schemaInfo{
		inode: fs.GenerateDynamicInode(InodeSchemas, name),
		name: name,
	}

	return
}

func (s Schema) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = s.i.inode
	a.Uid = uid
	a.Gid = gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (s Schema) Lookup(ctx context.Context, name string) (fs.Node, error) {
	switch name {
	case "tables":
		if s.i.tables != nil {
			return s.i.tables, nil
		}

		t, err := prTE(NewTables(s))
		if err == nil {
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
