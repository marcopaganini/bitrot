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
	configDirPrefix = ".bitrot"
	configDirMode   = 640
)

type cmdLineOpts struct {
	root    string
	verbose bool
}

var (
	opt cmdLineOpts
)

// Return a unique configuration file based on root path
func configFile(root string) string {
	hd := md5.New()
	io.WriteString(hd, root)
	return fmt.Sprintf("bitrot_%x.db", hd.Sum(nil))
}

// Return a directory for the configuration file. The directory
// is created if it doesn't yet exist.
func configDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("Unable to get information for current user: %q", err)
	}
	cdir := filepath.Join(usr.HomeDir, configDirPrefix)

	// Create if it doesn't exist
	fi, err := os.Stat(cdir)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("Unable to stat config directory: %q", err)
	}
	if os.IsNotExist(err) {
		err := os.Mkdir(cdir, configDirMode)
		if err != nil {
			return "", fmt.Errorf("Unable to create config directory: %q", err)
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

	fmt.Println(flag.NArg())
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

	// Load DirTree from disk if it already exists
	cdir, err := configDir()
	if err != nil {
		log.Fatal(err)
	}
	cfile := filepath.Join(cdir, configFile(opt.root))
	fmt.Println("Using cfile", cfile)

	dt := NewDirTree(opt.root, []*string{})
	dt.Compare()
	fmt.Println(configFile(opt.root))
}
