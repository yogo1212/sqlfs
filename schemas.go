package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type schemasInfo struct {
	schemas map[string]*Schema
}

type Schemas struct{
	i *schemasInfo
}

func NewSchemas() (s Schemas, err error) {
	s.i = &schemasInfo{
		schemas: make(map[string]*Schema),
	}

	rows, err := db.Query(`
select
	nspname
from pg_namespace`)
	if err != nil {
		err = fmt.Errorf("error listing schemas: %w", err)
		return
	}

	for rows.Next() {
		var n string
		err = rows.Scan(&n)
		if err != nil {
			err = fmt.Errorf("error listing schemas: %w", err)
			return
		}

		s.i.schemas[n] = nil
	}

	return
}

func (s Schemas) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = InodeSchemas
	a.Uid = uid
	a.Gid = gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (s Schemas) Lookup(ctx context.Context, name string) (fs.Node, error) {
	schema, ok := s.i.schemas[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	if schema != nil {
		return *schema, nil
	}

	n, err := prTE(NewSchema(name))
	if err == nil {
		s.i.schemas[name] = &n
	}

	return n, err
}

func (s Schemas) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	schemas := make([]fuse.Dirent, 0, len(s.i.schemas))
	for n, _ := range s.i.schemas {
		schemas = append(schemas, fuse.Dirent{Inode: fs.GenerateDynamicInode(InodeSchemas, n), Name: n, Type: fuse.DT_Dir})
	}
	return schemas, nil
}
