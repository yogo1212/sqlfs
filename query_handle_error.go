package main

import (
	"context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type queryHandleErrorInfo struct{
	inode uint64

	s *queryHandleStateData
}

type QueryHandleError struct{
	i *queryHandleErrorInfo
}

func NewQueryHandleError(parentInode uint64, s *queryHandleStateData) (e QueryHandleError) {
	e.i = &queryHandleErrorInfo{
		inode: fs.GenerateDynamicInode(parentInode, "error"),
		s: s,
	}
	return
}

func (e QueryHandleError) Attr(ctx context.Context, a *fuse.Attr) error {
	s := e.i.s

	a.Inode = e.i.inode
	a.Uid = uid
	a.Gid = gid
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

	return s.err.Bytes(), nil
}
