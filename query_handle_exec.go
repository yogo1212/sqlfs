package main

import (
	"context"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type queryHandleExecInfo struct{
	inode uint64

	s *queryHandleStateData
}

type QueryHandleExec struct{
	i *queryHandleExecInfo
}

func NewQueryHandleExec(parentInode uint64, s *queryHandleStateData) (h QueryHandleExec) {
	h.i = &queryHandleExecInfo{
		inode: fs.GenerateDynamicInode(parentInode, "exec"),
		s: s,
	}
	return
}

func (e QueryHandleExec) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = e.i.inode
	a.Uid = uid
	a.Gid = gid
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
