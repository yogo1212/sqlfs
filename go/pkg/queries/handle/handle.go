package handle

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"sync"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/yogo1212/sqlfs.git/go/pkg/base"
)

type queryHandleState uint

const (
	queryHandleStateStart queryHandleState = iota
	queryHandleStateExec
	queryHandleStateReady
)

type queryHandleStateData struct {
	m sync.Mutex
	state queryHandleState
	notifyStateChange func()

	query bytes.Buffer
	paramStarted bool
	params []bytes.Buffer

	err bytes.Buffer
	notifyError func(error)

	rows *sql.Rows
	notifyRowsChange func()
}

type queryHandleInfo struct {
	inode uint64

	s queryHandleStateData

	exec *QueryHandleExec
	params *QueryHandleParams
	readAllAsAscii *QueryHandleReadAllAsAscii

	err *QueryHandleError
}

type QueryHandle struct{
	data *base.MountData

	i *queryHandleInfo
}

func NewQueryHandle(data *base.MountData, parentInode uint64, name string) (q QueryHandle) {
	q.data = data
	q.i = &queryHandleInfo{
		inode: fs.GenerateDynamicInode(parentInode, name),
		s: queryHandleStateData{
			notifyStateChange: func () {
				err := data.FuseServer.InvalidateNodeData(q)
				if err != nil && err != fuse.ErrNotCached {
					data.PrintErr(err)
				}
			},
			notifyError: func (e error) {
				q.i.s.err.Reset()
				q.i.s.err.WriteString(e.Error())

				go func() {
					err := data.FuseServer.InvalidateEntry(q, "error")
					if err != nil && err != fuse.ErrNotCached {
						data.PrintErr(err)
					}
				}()

				go q.i.s.notifyStateChange()
			},
			notifyRowsChange: func () {
				go func() {
					err := data.FuseServer.InvalidateEntry(q, "read_all_as_ascii")
					if err != nil && err != fuse.ErrNotCached {
						data.PrintErr(err)
					}
				}()

				go q.i.s.notifyStateChange()
			},
		},
	}
	return
}

func (q QueryHandle) Cleanup() {
	s := q.i.s
	s.m.Lock()
	defer s.m.Unlock()

	if s.rows != nil {
		s.rows.Close()
		s.rows = nil
	}
}

func (h QueryHandle) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = h.i.inode
	a.Uid = h.data.Uid
	a.Gid = h.data.Gid
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (q QueryHandle) Lookup(ctx context.Context, name string) (fs.Node, error) {
	s := q.i.s
	s.m.Lock()
	defer s.m.Unlock()

	switch name {
		case "error":
			if s.err.Len() == 0 {
				return nil, syscall.ENOENT
			}

			if q.i.err != nil {
				return *q.i.err, nil
			}

			n := NewQueryHandleError(q.data, q.i.inode, &q.i.s)
			q.i.err = &n

			return n, nil
	default:
	}

	switch q.i.s.state {
	case queryHandleStateStart:
		switch name {
		case "exec":
			if q.i.exec != nil {
				return *q.i.exec, nil
			}

			n := NewQueryHandleExec(q.data, q.i.inode, &q.i.s)
			q.i.exec = &n

			return n, nil
		default:
			return nil, syscall.ENOENT
		}
	case queryHandleStateExec:
		switch name {
		case "params":
			if q.i.params != nil {
				return *q.i.params, nil
			}

			n := NewQueryHandleParams(q.data, q.i.inode, &q.i.s)
			q.i.params = &n

			return n, nil
		default:
			return nil, syscall.ENOENT
		}
	case queryHandleStateReady:
		switch name {
		case "read_all_as_ascii":
			if q.i.readAllAsAscii != nil {
				return *q.i.readAllAsAscii, nil
			}

			n := NewQueryHandleReadAllAsAscii(q.data, q.i.inode, &q.i.s)
			q.i.readAllAsAscii = &n

			return n, nil
		default:
			return nil, syscall.ENOENT
		}
	default:
		return nil, syscall.ENOENT
	}

	return nil, syscall.ENOENT
}

func (q QueryHandle) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var dirs []fuse.Dirent

	s := q.i.s
	s.m.Lock()
	defer s.m.Unlock()

	switch q.i.s.state {
	case queryHandleStateStart:
		dirs = []fuse.Dirent{
			{Inode: fs.GenerateDynamicInode(q.i.inode, "exec"), Name: "exec", Type: fuse.DT_Char},
			// {Inode: fs.GenerateDynamicInode(q.i.inode, "direct"), Name: "direct", Type: fuse.DT_Char},
			// {Inode: fs.GenerateDynamicInode(q.i.inode, "prepare"), Name: "prepare", Type: fuse.DT_Char},
		}
	case queryHandleStateExec:
		dirs = []fuse.Dirent{
			{Inode: fs.GenerateDynamicInode(q.i.inode, "params"), Name: "params", Type: fuse.DT_Char},
		}
	case queryHandleStateReady:
		dirs = []fuse.Dirent{}
		if s.rows != nil {
			dirs = append(dirs, fuse.Dirent{Inode: fs.GenerateDynamicInode(q.i.inode, "read_all_as_ascii"), Name: "read_all_as_ascii", Type: fuse.DT_Char},)
		}
	}

	if s.err.Len() > 0 {
		dirs = append(dirs, fuse.Dirent{Inode: fs.GenerateDynamicInode(q.i.inode, "error"), Name: "error", Type: fuse.DT_Char})
	}
	return dirs, nil
}

func (q QueryHandle) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	s := &q.i.s
	s.m.Lock()
	defer s.m.Unlock()

	if req.Name == "params" && !req.Dir {
		if s.state == queryHandleStateExec {
			s.state = queryHandleStateReady
			s.notifyStateChange()

			params := make([]any, 0, len(s.params))
			for _, p := range s.params {
				params = append(params, p.String())
			}

			var err error
			s.rows, err = q.data.DB.Query(s.query.String(), params...)
			if err != nil {
				q.data.PrintErr(err)
				s.notifyError(err)
				return syscall.EIO
			}

			return nil
		}
		return syscall.ENOSYS
	}
	return syscall.ENOENT
}
