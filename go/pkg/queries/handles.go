package queries

import (
	"context"
	"os"
	"sync"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
	"github.com/yogo1212/sqlfs.git/go/pkg/queries/handle"
)

type queryHandlesInfo struct {
	inode uint64

	m sync.Mutex
	handles map[string]*handle.QueryHandle
}

type QueryHandles struct{
	i *queryHandlesInfo

	data *base.MountData
}

func NewQueryHandles(data *base.MountData, parentInode uint64) (s QueryHandles) {
	s.data = data
	s.i = &queryHandlesInfo{
		inode: fs.GenerateDynamicInode(parentInode, "handles"),
		handles: make(map[string]*handle.QueryHandle),
	}
	return
}

func (q QueryHandles) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = q.i.inode
	a.Uid = q.data.Uid
	a.Gid = q.data.Gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (q QueryHandles) Lookup(ctx context.Context, name string) (fs.Node, error) {
	q.i.m.Lock()
	defer q.i.m.Unlock()

	h, ok := q.i.handles[name]
	if !ok {
		return nil, syscall.ENOENT
	}

	return *h, nil
}

func (q QueryHandles) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	q.i.m.Lock()
	defer q.i.m.Unlock()

	if _, ok := q.i.handles[req.Name]; ok {
		return nil, syscall.EEXIST
	}

	h := handle.NewQueryHandle(q.data, q.i.inode, req.Name)
	q.i.handles[req.Name] = &h
	return h, nil
}

func (q QueryHandles) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	q.i.m.Lock()
	defer q.i.m.Unlock()

	dirs := make([]fuse.Dirent, 0, len(q.i.handles))
	for n, _ := range q.i.handles {
		dirs = append(dirs, fuse.Dirent{Inode: fs.GenerateDynamicInode(q.i.inode, n), Name: n, Type: fuse.DT_Dir})
	}
	return dirs, nil
}

func (q QueryHandles) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	if !req.Dir {
		return syscall.ENOSYS
	}

	q.i.m.Lock()
	defer q.i.m.Unlock()

	h, ok := q.i.handles[req.Name]
	if !ok {
		return syscall.ENOENT
	}

	delete(q.i.handles, req.Name)
	go h.Cleanup()
	return nil
}

