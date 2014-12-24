// This file is part of bitrot, a bitrot scrubber in Go
// http://github.com/marcopaganini/bitrot
// (C) Dec/2014 by Marco Paganini <paganini@paganini.net>

package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/marcopaganini/logger"
	"io"
	"os"
	"os/user"
	"path/filepath"
)

const (
	stateDirPrefix = ".bitrot"
	stateDirMode   = 0770
)

type cmdLineOpts struct {
	root    string
	verbose bool
}

var (
	opt cmdLineOpts
)

// Return a unique state db file based on root path
func stateFile(root string) string {
	hd := md5.New()
	io.WriteString(hd, root)
	return fmt.Sprintf("bitrot_%x.db", hd.Sum(nil))
}

// Return a directory for the state database. The directory
// is created if it doesn't yet exist.
func stateDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("Unable to get information for current user: %q", err)
	}
	cdir := filepath.Join(usr.HomeDir, stateDirPrefix)

	// Create if it doesn't exist
	fi, err := os.Stat(cdir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("Unable to stat state DB directory: %q", err)
	}
	if os.IsNotExist(err) {
		err := os.Mkdir(cdir, stateDirMode)
		if err != nil {
			return "", fmt.Errorf("Unable to create state DB directory: %q", err)
		}
	} else {
		if !fi.Mode().IsDir() {
			return "", fmt.Errorf("A non-directory named %s already exists.", cdir)
		}
	}
	return cdir, nil
}

// Parse command-line flags
func parseFlags() error {
	flag.BoolVar(&opt.verbose, "verbose", false, "Verbose mode")
	flag.BoolVar(&opt.verbose, "v", false, "Verbose mode (shorthand)")
	flag.Parse()

	if flag.NArg() != 1 {
		return fmt.Errorf("Use: bitrot [-v|--verbose] directory")
	}
	root, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		return fmt.Errorf("Unable to convert root directory \"%s\" to absolute path: %q", flag.Arg(0), err)
	}
	opt.root = root
	return nil
}

func main() {
	log := logger.New("")

	err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}

	dt := NewDirTree(opt.root, []*string{})

	// Load DirTree from disk if the right state DB file exists
	cdir, err := stateDir()
	if err != nil {
		log.Fatal(err)
	}
	cfile := filepath.Join(cdir, stateFile(opt.root))

	fi, err := os.Stat(cfile)
	if err == nil && fi.Mode().IsRegular() {
		r, err := os.Open(cfile)
		if err != nil {
			log.Fatal("Error opening state DB file:", err)
		}
		defer r.Close()
		err = dt.Load(r)
		if err != nil {
			log.Fatal(err)
		}
	}

	dt.Compare()

	// Save State DB
	w, err := os.Create(cfile)
	if err != nil {
		log.Fatal("Error creating state DB file:", err)
	}
	err = dt.Save(w)
	if err != nil {
		log.Fatal(err)
	}

	defer w.Close()
}
