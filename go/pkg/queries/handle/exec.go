package handle

import (
	"context"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type queryHandleExecInfo struct{
	inode uint64

	s *queryHandleStateData
}

type QueryHandleExec struct{
	data *base.MountData
	i *queryHandleExecInfo
}

func NewQueryHandleExec(data *base.MountData, parentInode uint64, s *queryHandleStateData) (h QueryHandleExec) {
	h.data = data
	h.i = &queryHandleExecInfo{
		inode: fs.GenerateDynamicInode(parentInode, "exec"),
		s: s,
	}
	return
}

func (e QueryHandleExec) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = e.i.inode
	a.Uid = e.data.Uid
	a.Gid = e.data.Gid
	a.Mode = 0o222
	return nil
}

func (e QueryHandleExec) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) (err error) {
	s := e.i.s
	s.m.Lock()
	defer s.m.Unlock()

	resp.Size, err = s.query.Write(req.Data)
	return
}

func (e QueryHandleExec) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	s := e.i.s
	s.m.Lock()
	defer s.m.Unlock()

	if s.state != queryHandleStateStart {
		return syscall.ENOTSUP
	}

	s.state = queryHandleStateExec
	s.notifyStateChange()
	return nil
}
