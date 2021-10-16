// This file is part of bitrot, a bitrot scrubber in Go
// http://github.com/marcopaganini/bitrot
// (C) Dec/2014 by Marco Paganini <paganini@paganini.net>

package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/marcopaganini/logger"
)

// BuildVersion is filled by go build -ldflags during build.
var BuildVersion string

const (
	stateDirPrefix = ".bitrot"
	stateFileMask  = "bitrot_%x.db.gz"
	stateDirMode   = 0770
)

type cmdLineOpts struct {
	root    string
	verbose bool
	version bool
}

// Globals -- Use with care
var (
	Opt cmdLineOpts
	Log *logger.Logger
)

// Return a unique state db file based on root path
func stateFile(root string) string {
	hd := md5.New()
	io.WriteString(hd, root)
	return fmt.Sprintf(stateFileMask, hd.Sum(nil))
}

// Return a directory for the state database. The directory
// is created if it doesn't yet exist.
func stateDir() (string, error) {
	home, err := homeDir()
	if err != nil {
		return "", fmt.Errorf("unable to find home directory for the current user: %v", err)
	}
	cdir := filepath.Join(home, stateDirPrefix)

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
			return "", fmt.Errorf("a non-directory named %s already exists", cdir)
		}
	}
	return cdir, nil
}

// Parse command-line flags
func parseFlags() error {
	flag.BoolVar(&Opt.version, "version", false, "Display version")
	flag.BoolVar(&Opt.verbose, "verbose", false, "Verbose mode")
	flag.BoolVar(&Opt.verbose, "v", false, "Verbose mode (shorthand)")
	flag.Parse()

	if Opt.version {
		fmt.Printf("bitrot build: %s\n", BuildVersion)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		usage("Error: Starting directory not specified.")
	}
	root, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		return fmt.Errorf("Unable to convert root directory \"%s\" to absolute path: %q", flag.Arg(0), err)
	}
	Opt.root = root
	return nil
}

// Load the state database from disk, if it exists
func loadStateFromFile(d *DirTree) error {
	cdir, err := stateDir()
	if err != nil {
		return err
	}
	cfile := filepath.Join(cdir, stateFile(Opt.root))

	fi, err := os.Stat(cfile)
	if err == nil && fi.Mode().IsRegular() {
		Log.Verbosef(1, "Reading state DB from %s\n", cfile)
		r, err := os.Open(cfile)
		if err != nil {
			return err
		}
		defer r.Close()
		err = d.Load(r)
		if err != nil {
			return err
		}
	}
	return nil
}

// Save state to file
func saveStateToFile(d *DirTree) error {
	cdir, err := stateDir()
	if err != nil {
		return err
	}
	cfile := filepath.Join(cdir, stateFile(Opt.root))

	w, err := os.Create(cfile)
	if err != nil {
		return err
	}
	defer w.Close()

	err = d.Save(w)
	if err != nil {
		return err
	}
	return nil
}

// homeDir returns the user's home directory or an error if the variable HOME
// is not set, or os.user fails, or the directory cannot be found.
func homeDir() (string, error) {
	// Get home directory from the HOME environment variable first. If it's
	// not set, fallback to os.user. This workaround is needed due to:
	// https://github.com/golang/go/issues/6376
	home := os.Getenv("HOME")
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("Unable to get information for current user: %q", err)
		}
		home = usr.HomeDir
	}
	_, err := os.Stat(home)
	if os.IsNotExist(err) || !os.ModeType.IsDir() {
		return "", fmt.Errorf("Home directory %q must exist and be a directory", home)
	}
	// Other errors than file not found.
	if err != nil {
		return "", err
	}
	return home, nil
}

// Prints error message and program usage to stderr, exit the program.
func usage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", msg)
	}
	fmt.Fprintf(os.Stderr, "Use: %s [flags] start_dir\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

// Good ole main()
func main() {
	Log = logger.New("")

	err := parseFlags()
	if err != nil {
		Log.Fatalln(err)
	}
	if Opt.verbose {
		Log.SetVerboseLevel(1)
	}

	dt := NewDirTree(Opt.root, []*string{})
	err = loadStateFromFile(dt)
	if err != nil {
		Log.Fatalln(fmt.Sprintf("Error loading state DB: %q", err))
	}
	dt.Compare()
	err = saveStateToFile(dt)
	if err != nil {
		Log.Fatalln(fmt.Sprintf("Error saving state DB: %q", err))
	}
}
