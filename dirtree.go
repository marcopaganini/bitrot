// This file is part of bitrot, a bitrot scrubber in Go
// http://github.com/marcopaganini/bitrot
// (C) 2014 by Marco Paganini <paganini@paganini.net>

package main

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileInfo holds file metadata about each file in the filesystem.
type FileInfo struct {
	Size   int64
	Mode   os.FileMode
	Mtime  time.Time
	Md5sum []byte
}

// DirTree contains the metadata information about all files in a given
// mountpoint.
type DirTree struct {
	Root      string
	ExcludeRe []*string
	Files     map[string]*FileInfo
}

// NewDirTree creates a new DirTree struct.
func NewDirTree(root string, excludeRe []*string) *DirTree {
	dt := &DirTree{
		Root:      root,
		ExcludeRe: excludeRe,
	}
	dt.Files = make(map[string]*FileInfo)
	return dt
}

func (d *DirTree) compareFile(fname string) (string, error) {
	var (
		memfi  *FileInfo
		exists bool
	)

	osinfo, err := os.Stat(fname)
	if err != nil {
		return "", err
	}

	md5sum, err := md5sum(fname)
	if err != nil {
		return "", err
	}

	if memfi, exists = d.Files[fname]; !exists {
		// If medatada changed, replace entry silently
		memfi = &FileInfo{
			Size:   osinfo.Size(),
			Mode:   osinfo.Mode(),
			Mtime:  osinfo.ModTime(),
			Md5sum: md5sum,
		}
		d.Files[fname] = memfi
		Log.Verbosef(1, "[New] %s (%x)\n", fname, md5sum)
	} else {
		if osinfo.ModTime() != memfi.Mtime || osinfo.Mode() != memfi.Mode || osinfo.Size() != memfi.Size {
			// Exists: If medatada changed, replace entry silently
			Log.Verbosef(1, "[Metadata changes] %s (%x)\n", fname, md5sum)
			memfi.Size = osinfo.Size()
			memfi.Mode = osinfo.Mode()
			memfi.Mtime = osinfo.ModTime()
			memfi.Md5sum = md5sum
			d.Files[fname] = memfi
		} else {
			Log.Verbosef(1, "[No metadata changes] %s (%x)\n", fname, md5sum)
			// Exists: No metadata changes. Report md5 differences
			for k := range md5sum {
				if memfi.Md5sum[k] != md5sum[k] {
					return fmt.Sprintf("[MD5 Mismatch] %s (%x -> %x)", fname, memfi.Md5sum, md5sum), nil
				}
			}
		}
	}
	return "", nil
}

// Compare compares all files in the DirTree to the files on disk. Adds new
// files found on disk to DirTree and replace Files with changed attributes.
// Report in case a file exists with the exact same attributes but different
// checksum.
func (d *DirTree) Compare() error {
	filepath.Walk(d.Root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.Mode().IsRegular() {
			msg, err := d.compareFile(path)
			if err != nil {
				return (fmt.Errorf("Error comparing file: %s: %q", path, err))
			}
			if msg != "" {
				fmt.Println(msg)
			}
		}
		return nil
	})
	return nil
}

// Save writes a JSON representation of the current DirTree struct to the
// specified io.Writer.
func (d *DirTree) Save(writer io.Writer) error {
	zwriter := gzip.NewWriter(writer)
	defer zwriter.Close()

	enc := json.NewEncoder(zwriter)
	err := enc.Encode(d)
	if err != nil {
		return err
	}
	return nil
}

// Load a JSON representation of DirTree from the specified io.Reader.
func (d *DirTree) Load(reader io.Reader) error {
	zreader, err := gzip.NewReader(reader)
	if err != nil {
		return (fmt.Errorf("Error uncompressing state DB: %q", err))
	}
	defer zreader.Close()

	dec := json.NewDecoder(zreader)
	err = dec.Decode(d)
	if err != nil {
		return err
	}
	return nil
}

// Calculate the md5sum of a file
func md5sum(fname string) ([]byte, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return []byte{}, err
	}
	defer fd.Close()

	hd := md5.New()
	_, err = io.Copy(hd, fd)
	if err != nil {
		return []byte{}, err
	}
	return hd.Sum(nil), nil
}
