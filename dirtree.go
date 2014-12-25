// This file is part of bitrot, a bitrot scrubber in Go
// http://github.com/marcopaganini/bitrot
// (C) 2014 by Marco Paganini <paganini@paganini.net>

package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type FileInfo struct {
	Size   int64
	Mode   os.FileMode
	Mtime  time.Time
	Md5sum []byte
}

type DirTree struct {
	Root      string
	ExcludeRe []*string
	Files     map[string]*FileInfo
}

// Return a new DirTree object.
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
		Log.Verbosef(1, "Created new memory entry for %s, md5=%x\n", fname, md5sum)
	} else {
		if osinfo.ModTime() != memfi.Mtime || osinfo.Mode() != memfi.Mode || osinfo.Size() != memfi.Size {
			// Exists: If medatada changed, replace entry silently
			Log.Verbosef(1, "Metadata changes detected for %s\n", fname)
			memfi.Size = osinfo.Size()
			memfi.Mode = osinfo.Mode()
			memfi.Mtime = osinfo.ModTime()
			memfi.Md5sum = md5sum
			d.Files[fname] = memfi
		} else {
			Log.Verbosef(1, "No metadata changes for %s. Comparing md5sum.", fname)
			// Exists: No metadata changes. Report md5 differences
			for k, _ := range md5sum {
				if memfi.Md5sum[k] != md5sum[k] {
					return fmt.Sprintf("MD5sum mismatch for path: %s", fname), nil
				}
			}
		}
	}
	return "", nil
}

// Compare files in the DirTree to the files on disk. Adds new files found on
// disk to DirTree and replace Files with changed attributes. Report in case a
// file exists with the exact same attributes but different checksum.
func (d *DirTree) Compare() error {
	filepath.Walk(d.Root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			//fmt.Printf("DEBUG ROOT: Got ERROR %q on %s\n", err, path)
			return nil
		}
		if fi.Mode().IsRegular() {
			//fmt.Printf("DEBUG ROOT: Sending %s to channel\n", path)
			d.compareFile(path)
		}
		return nil
	})
	return nil
}

// Write a JSON representation of the current DirTree struct
// to the specified io.Writer.
func (d *DirTree) Save(writer io.Writer) error {
	enc := json.NewEncoder(writer)
	err := enc.Encode(d)
	if err != nil {
		return err
	}
	return nil
}

// Load a JSON representation of DirTree from the specified io.Reader.
func (d *DirTree) Load(reader io.Reader) error {
	enc := json.NewDecoder(reader)
	err := enc.Decode(d)
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
