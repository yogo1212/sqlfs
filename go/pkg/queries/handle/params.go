package handle

import (
	"bytes"
	"context"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type queryHandleParamsInfo struct{
	inode uint64

	s *queryHandleStateData
}

type QueryHandleParams struct{
	data *base.MountData

	i *queryHandleParamsInfo
}

func NewQueryHandleParams(data *base.MountData, parentInode uint64, s *queryHandleStateData) (p QueryHandleParams) {
	p.data = data
	p.i = &queryHandleParamsInfo{
		inode: fs.GenerateDynamicInode(parentInode, "params"),
		s: s,
	}
	return
}

func (p QueryHandleParams) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = p.i.inode
	a.Uid = p.data.Uid
	a.Gid = p.data.Gid
	a.Mode = 0o222
	return nil
}

func (p QueryHandleParams) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) (err error) {
	s := p.i.s
	s.m.Lock()
	defer s.m.Unlock()

	if !s.paramStarted {
		if len(s.params) == 0 {
			s.params = make([]bytes.Buffer, 1)
		} else {
			s.params = append(s.params, bytes.Buffer{})
		}
	}

	resp.Size, err = s.params[len(s.params) - 1].Write(req.Data)
	return
}

func (p QueryHandleParams) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	s := p.i.s
	s.m.Lock()
	defer s.m.Unlock()

	if s.state != queryHandleStateExec {
		return syscall.ENOTSUP
	}

	s.paramStarted = false
	return nil
}
