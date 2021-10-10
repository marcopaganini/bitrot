[![Go Report Card](https://goreportcard.com/badge/github.com/marcopaganini/bitrot)](https://goreportcard.com/report/github.com/marcopaganini/bitrot)

# bitrot

# Description

Bitrot Scrubber - Scrubs your disks (or array) looking for "bit rot" (silent
disk data corruption.)

The size of hard drives has grown steadily over the years, but their bit error
rate has remained relatively unchanged. As an example, the Seagate "Barracuda"
line has a "Non Recoverable Bit Error per bit Read" rate of 1 bit per 1e14 bits
read ([Barracuda datasheet](http://www.seagate.com/staticfiles/docs/pdf/datasheet/disc/barracuda-ds1737-1-1111us.pdf)).
This means that 1 bit may be incorrectly reported every 1e14 bits read from the
drive, or 11 TB (bytes) of data. This may seem a lot, but with 4-6+ TB drives
becoming commonplace, the problem becomes very obvious.

Another problem is the ever growing sizes of solid state media. Many flash
cards fail silently and don't report any error conditions the host operating
system.  By the time the problem is detected, usually even backups already have
a copy of corrupt data.

Some forms of read errors are detectable by the drive firmware. These errors
are usually very visible to the system operator and generate ominous error
messages in the system log. If a disk array is used, such conditions will
normally cause the offending drive to be marked as defective and removed from
the array. Our main concern are the "invisible" bit read errors, popularly
known as bitrot, which are not detected by the drive firmware and could cause
silent data corruption.

The "proper" solution for bitrot protection is to use an advanced file system
such as ZFS, which was specifically created to address data integrity issues.
Unfortunately, not everybody is willing to move to ZFS and deal with the
resulting technical complications or licensing structure.

This program is a simple alternative to detect silent data corruption on disk
drives. The idea is to "scrub" your collection of files periodically, so that
an eventual corruption might be detected and the file restored from backups.
This program will not protect your data from write errors or cache/buffer
corruptions in memory (use ECC RAM!). Also, it was *not* meant to be used as a
security audit tool (it is, in fact, useless in that capacity as it purposely
ignores files with metadata changes.)

# Installation

The easiest way to install bitrot is to download it from the releases page in
github: https://github.com/marcopaganini/bitrot/releases/tag/v1.0.0. Just
download the desired version and give it a go.

If you prefer to install from source:

* Clone the repo locally: `git clone https://github.com/marcopaganini/bitrot`
* Change  to the repo directory: `cd bitrot`
* Build the version: `make`
* Install in your local binaries directory: `sudo make install`

# Usage

Run `bitrot <directory>` periodically and bitrot will read all regular files
under that directory, recursively, calculating their MD5 hashes. The program
then compares the MD5 hash of each file read from disk with a saved version
from a previous run. If a file exists with the same filesystem metadata
(modification date, size) and the MD5 hash doesn't match, the scrubber assumes
data corruption happened and reports the file. The program assumes that
modified files (metadata changes) contain valid modifications and it won't try
to compare the MD5 hash on those. Instead, the current MD5 file is saved for
future runs.
