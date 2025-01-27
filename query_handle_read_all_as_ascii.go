package main

import (
	"bytes"
	"context"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type queryHandleReadAllAsAsciiInfo struct{
	inode uint64

	s *queryHandleStateData
}

type QueryHandleReadAllAsAscii struct{
	i *queryHandleReadAllAsAsciiInfo
}

func NewQueryHandleReadAllAsAscii(parentInode uint64, s *queryHandleStateData) (r QueryHandleReadAllAsAscii) {
	r.i = &queryHandleReadAllAsAsciiInfo{
		inode: fs.GenerateDynamicInode(parentInode, "read_all_as_ascii"),
		s: s,
	}
	return
}

func (r QueryHandleReadAllAsAscii) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = r.i.inode
	a.Uid = uid
	a.Gid = gid
	a.Mode = 0o444
	return nil
}

// TODO somehow, the char devices turn out to be regular files (even with AllowDev())
// files are only read up until their reported size.
// the design could be changed to collect the data in a buffer, using poll, and such.
// right now, just use directio

func (r QueryHandleReadAllAsAscii) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags = fuse.OpenDirectIO | fuse.OpenNonSeekable
	return r, nil
}


//func (h QueryHandleReadAllAsAscii) Poll(ctx context.Context, req *fuse.PollRequest, resp *fuse.PollResponse) error {
//func (r QueryHandleReadAllAsAscii) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {

func (r QueryHandleReadAllAsAscii) ReadAll(ctx context.Context) ([]byte, error) {
	s := r.i.s
	s.m.Lock()
	defer s.m.Unlock()

	if s.rows == nil {
		return nil, syscall.ENOENT
	}

	var res bytes.Buffer

	appendRows := func () error {
		c, err := prTE(s.rows.Columns())
		if err != nil {
			return syscall.EIO
		}

		cCount := len(c)

		row := make([]any, cCount)
		for i := 0; i < cCount; i++ {
			var s string
			row[i] = &s
		}

		first := true

		for s.rows.Next() {
			if first {
				first = false
			} else {
				res.WriteByte(0x1e)
			}

			err = prE(s.rows.Scan(row...))
			if err != nil {
				return syscall.EIO
			}

			for i := range row {
				if row[i] != nil {
					res.WriteString(*(row[i].(*string)))
				}
				if i < cCount - 1 {
					res.WriteByte(0x1f)
				}
			}
		}
		return nil
	}

	err := appendRows()
	if err != nil {
		s.notifyError(err)
		return nil, err
	}

	for s.rows.NextResultSet() {
		res.WriteByte(0x1d)

		err = appendRows()
		if err != nil {
			s.notifyError(err)
			return nil, err
		}
	}

	return res.Bytes(), nil
}
