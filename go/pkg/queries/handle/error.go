package handle

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type queryHandleErrorInfo struct{
	inode uint64

	s *queryHandleStateData
}

type QueryHandleError struct{
	data *base.MountData

	i *queryHandleErrorInfo
}

func NewQueryHandleError(data *base.MountData, parentInode uint64, s *queryHandleStateData) (e QueryHandleError) {
	e.data = data
	e.i = &queryHandleErrorInfo{
		inode: fs.GenerateDynamicInode(parentInode, "error"),
		s: s,
	}
	return
}

func (e QueryHandleError) Attr(ctx context.Context, a *fuse.Attr) error {
	s := e.i.s

	a.Inode = e.i.inode
	a.Uid = e.data.Uid
	a.Gid = e.data.Gid
	a.Mode = 0o444
	a.Size = uint64(s.err.Len())
	return nil
}

func (e QueryHandleError) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags = fuse.OpenDirectIO | fuse.OpenNonSeekable
	return e, nil
}

//func (h QueryHandleReadAllAsAscii) Poll(ctx context.Context, req *fuse.PollRequest, resp *fuse.PollResponse) error {
//func (r QueryHandleReadAllAsAscii) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {

func (e QueryHandleError) ReadAll(ctx context.Context) ([]byte, error) {
	s := e.i.s
	s.m.Lock()
	defer s.m.Unlock()

	buf := make([]byte, s.err.Len() + 1)
	buf = append(buf, s.err.Bytes()...)
	buf = append(buf, '\n')

	return buf, nil
}
