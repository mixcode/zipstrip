package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultBackupExt = "bak"
)

var (
	// Actions
	showinfo = false // show ZIP truncate info. this is the default
	copyfile = false // copy arg[0] to arg[1] removing garbages
	truncate = false // truncate the ZIP file in-place

	// flags
	force = false // for show, display zip size even if no truncation is needed
	// for copy, copy even if no truncation is needed

	backupext = defaultBackupExt // default extension of the backup file
	nobackup  = false            // do not make backup when truncate

	maxTruncate int64 = 4 * 1024 * 1024 // Max size to truncate
)

// get file size and zip size
func getZipSz(filename string) (filesize, zipsize int64, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	// get file size
	filesize, err = f.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}

	// get zip size
	zipsize, err = zipLength(f, maxTruncate)
	return
}

// Copy a file to another file.
// if sz>0, then exact sz bytes are copied. otherwise all data is copied
func copyFileN(src, dst string, sz int64) (written int64, err error) {
	// create output file
	fileflag := os.O_CREATE | os.O_WRONLY
	if force {
		// truncate existing file
		fileflag |= os.O_TRUNC
	} else {
		// file must not exist
		fileflag |= os.O_EXCL
	}
	fo, err := os.OpenFile(dst, fileflag, 0644)
	if err != nil {
		return
	}
	defer fo.Close()

	fi, err := os.Open(src)
	if err != nil {
		return
	}
	defer fi.Close()

	if sz > 0 {
		return io.CopyN(fo, fi, sz)
	}
	return io.Copy(fo, fi)
}

// show zip size
func doShowInfo(filename string) (err error) {
	fileSz, zipSz, err := getZipSz(filename)
	if err != nil {
		return
	}
	if zipSz != fileSz || force {
		diff := fileSz - zipSz
		fmt.Printf("%d %d %d %s\n", zipSz, diff, fileSz, filename)
	}
	return
}

// Copy zip bytes to another file or directory
func doCopyZip(src, dst string) (err error) {
	fileSz, zipSz, err := getZipSz(src)
	if err != nil {
		return
	}
	if fileSz == zipSz && !force {
		// no need to copy
		return
	}

	// if dst is a directory, then make a same named file to the dst directory.
	finfo, err := os.Stat(dst)
	if err == nil && finfo.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}
	err = nil

	// copy
	written, err := copyFileN(src, dst, zipSz)
	if err != nil {
		return
	}
	if written != zipSz {
		err = fmt.Errorf("Copy size not match; expected %d, actual %d", zipSz, written)
	}

	return
}

// truncate the zip file in-place
func doTruncateZip(filename string) (err error) {
	fileSz, zipSz, err := getZipSz(filename)
	if err != nil {
		return
	}
	if fileSz == zipSz && !force {
		// no need to truncate
		return
	}
	if !nobackup { // make a backup file
		// determine backup file name
		ext := backupext
		if ext == "" {
			ext = defaultBackupExt
		}
		ext = "." + strings.TrimLeft(ext, ".")
		backupFile := filepath.Join(filepath.Dir(filename), filepath.Base(filename)+ext)

		_, err = copyFileN(filename, backupFile, -1)
		if err != nil {
			return
		}
	}

	if fileSz == zipSz {
		return
	}

	// truncate the file
	err = os.Truncate(filename, zipSz)
	return
}

// action coordinator
func run() (err error) {

	// check filename
	args := flag.Args()
	if len(args) < 1 {
		err = fmt.Errorf("Filename not given. (use --help for guide)")
		return
	}

	// check actions
	n := 0
	if showinfo {
		n++
	}
	if copyfile {
		n++
	}
	if truncate {
		n++
	}
	if n == 0 {
		err = fmt.Errorf("No command specified")
		return
	}
	if n > 1 {
		err = fmt.Errorf("-s, -c and -t are mutually exclusive; must specify one")
		return
	}

	// call actions
	if showinfo {
		return doShowInfo(args[0])
	}

	if copyfile {
		if len(args) < 2 {
			err = fmt.Errorf("copy destination not specified")
			return
		}
		return doCopyZip(args[0], args[1])
	}

	if truncate {
		return doTruncateZip(args[0])
	}

	return
}

func main() {
	var err error

	// command line flags
	flag.Usage = func() {
		o := flag.CommandLine.Output()
		fmt.Fprintf(o, "\n%s: Detect and cleanup garbages at the end of a ZIP archive.\n", os.Args[0])
		fmt.Fprintf(o, "\nUsage: %s [OPTIONS] zip_filename [copy_dest_filename]\n", os.Args[0])
		fmt.Fprintf(o, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.BoolVar(&showinfo, "s", showinfo, "Show actual ZIP size, garbage size, and original file size in bytes. By default, no size is shown if there are no garbages at the end of the file. (set -f to override.)")
	flag.BoolVar(&copyfile, "c", copyfile, "Copy ZIP portion to a file. By default, file copy occurs only if there are garbages at the end of the file. (set -f to override.)")
	flag.BoolVar(&truncate, "t", truncate, "Make a proper ZIP by truncating the file in-place. By default, a backup file is created only if there are garbages at the end of the file. (set -f to override.)")
	flag.StringVar(&backupext, "k", backupext, "Set the file extension of backup file for truncation.")
	flag.BoolVar(&nobackup, "nobackup", nobackup, "Do not make backup file when truncation.")
	flag.BoolVar(&force, "f", force, "Force actions.")

	flag.Int64Var(&maxTruncate, "maxtruncate", maxTruncate, "Maximum allowed size of garbage search and truncation. If no ZIP end-of-file header is found in this range from the end of file, then the file is not regarded as a ZIP.")

	flag.Parse()

	if !copyfile && !truncate {
		showinfo = true
	}

	err = run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
