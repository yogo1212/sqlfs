package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "github.com/lib/pq"
)

var db *sql.DB

func prE(err error) (error) {
	if err != nil {
		log.Println(err)
	}
	return err
}

func prTE[A any](a A, err error) (A, error) {
	if err != nil {
		log.Println(err)
	}
	return a, err
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintf(os.Stderr, "  %s CONNECT_STRING MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintln(os.Stderr, `
Examples for CONNECT_STRING:
  postgres://user:password@localhost/db_name?sslmode=verify-full
  user=user dbname=pqgotest host=localhost port=5432 sslmode=verify-full

Hints regarding the chosen postgres library (pq):
The default sslmode is 'prefer' but for postgres, it's set to 'require'.
The .pgpass mechanism is supported but PGPASSFILE needs to set on Windows.
`)
}

var uid, gid uint32

func setup() (db *sql.DB, f *fuse.Conn, err error) {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "verbose debug")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(1)
	}

	uid = uint32(syscall.Getuid())
	gid = uint32(syscall.Getgid())

	if verbose {
		fuse.Debug = func (msg interface{}) {
			log.Println(msg)
		}
	}

	driver := "postgres" // currently, nothing else is supported

	connectString := flag.Arg(0)
  db, err = sql.Open(driver, connectString)
	if err != nil {
		err = fmt.Errorf("error opening database connection: %w", err)
		return
	}

	mountpoint := flag.Arg(1)

	f, err = fuse.Mount(
		mountpoint,
		fuse.FSName("sqlfs"),
		fuse.Subtype(driver),
	)
	if err != nil {
		db.Close()
		err = fmt.Errorf("error setting up fuse: %w", err)
		return
	}

	return
}

var fuseServer *fs.Server

func main() {
	var (
		f   *fuse.Conn
		err error
	)
	db, f, err = setup()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	defer func () {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	fuseServer = fs.New(f, nil)

	err = fuseServer.Serve(FS{})
	if err != nil {
		log.Fatal(err)
	}
}
