package base

import (
	"database/sql"

	"bazil.org/fuse/fs"
)

type MountData struct {
	Uid, Gid uint32

	DB *sql.DB

	FuseServer *fs.Server

	PrintErr func(error)
}
