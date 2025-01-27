package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type tablesInfo struct{
	schema Schema

	inode uint64
	tables map[string]*Table
}

type Tables struct {
	i *tablesInfo
}

func NewTables(s Schema) (t Tables, err error) {
	t.i = &tablesInfo{
		inode: fs.GenerateDynamicInode(s.i.inode, "tables"),
		schema: s,
		tables: make(map[string]*Table),
	}

	rows, err := db.Query(`
select
	table_name
from information_schema.tables
where table_schema = $1
  and table_type = 'BASE TABLE'`, s.i.name)
	if err != nil {
		err = fmt.Errorf("error listing tables for schema \"%s\": %w", s.i.name, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var n string
		err = rows.Scan(&n)
		if err != nil {
			err = fmt.Errorf("error fetching table record in schema \"%s\": %w", s.i.name, err)
			return
		}

		t.i.tables[n] = nil
	}

	return
}

func (t Tables) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = t.i.inode
	a.Uid = uid
	a.Gid = gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (t Tables) Lookup(ctx context.Context, name string) (fs.Node, error) {
	table, ok := t.i.tables[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	if table != nil {
		return *table, nil
	}

	n, err := prTE(NewTable(t, name))
	if err == nil {
		t.i.tables[name] = &n
	}

	return n, err
}

func (t Tables) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	tables := make([]fuse.Dirent, 0, len(t.i.tables))
	for name, _ := range t.i.tables {
		tables = append(tables, fuse.Dirent{Inode: fs.GenerateDynamicInode(t.i.inode, name), Name: name, Type: fuse.DT_Dir})
	}
	return tables, nil
}
