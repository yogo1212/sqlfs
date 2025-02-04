package pkg

import (
	"context"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
	"github.com/yogo1212/sqlfs.git/go/pkg/queries"
	"github.com/yogo1212/sqlfs.git/go/pkg/schema"
)

const inodeRoot = uint64(1)

type FS struct{
	data *base.MountData
}

func NewFS(data *base.MountData) FS {
	return FS{
		data,
	}
}

func (f FS) Root() (fs.Node, error) {
	var i rootInfo
	return Root{&i, f.data, &f}, nil
}

type rootInfo struct {
	queries *queries.Queries
	schemas *schema.Schemas
}

type Root struct {
	i *rootInfo

	data *base.MountData

	fs *FS
}

func (r Root) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = inodeRoot
	a.Uid = r.data.Uid
	a.Gid = r.data.Gid
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

		q := queries.NewQueries(r.data, inodeRoot)
		r.i.queries = &q

		return q, nil
	case "schemas":
		if r.i.schemas != nil {
			return *r.i.schemas, nil
		}

		s, err := schema.NewSchemas(r.data, inodeRoot)
		if err == nil {
			r.data.PrintErr(err)
			r.i.schemas = &s
		}

		return s, err
	default:
	}

	return nil, syscall.ENOENT
}

func (Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var entries = []fuse.Dirent{
		{Inode: fs.GenerateDynamicInode(inodeRoot, "queries"), Name: "queries", Type: fuse.DT_Dir},
		{Inode: fs.GenerateDynamicInode(inodeRoot, "schemas"), Name: "schemas", Type: fuse.DT_Dir},
	}

	return entries, nil
}
