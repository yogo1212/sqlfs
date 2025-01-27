package main

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

const (
	InodeRoot uint64 = iota + 1
	InodeSchemas
	InodeQueries
)

type FS struct{}

func (FS) Root() (fs.Node, error) {
	var i rootInfo
	return Root{&i}, nil
}

type rootInfo struct {
	queries *Queries
	schemas *Schemas
}

type Root struct {
	i *rootInfo
}

func (Root) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = InodeRoot
	a.Uid = uid
	a.Gid = gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (r Root) Lookup(ctx context.Context, name string) (fs.Node, error) {
	// TODO why not a map?
	switch name {
	case "queries":
		if r.i.queries != nil {
			return *r.i.queries, nil
		}

		q := NewQueries()
		r.i.queries = &q

		return q, nil
	case "schemas":
		if r.i.schemas != nil {
			return *r.i.schemas, nil
		}

		s, err := prTE(NewSchemas())
		if err == nil {
			r.i.schemas = &s
		}

		return s, err
	default:
	}

	return nil, syscall.ENOENT
}

func (Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var entries = []fuse.Dirent{
		{Inode: InodeQueries, Name: "queries", Type: fuse.DT_Dir},
		{Inode: InodeSchemas, Name: "schemas", Type: fuse.DT_Dir},
	}

	return entries, nil
}
