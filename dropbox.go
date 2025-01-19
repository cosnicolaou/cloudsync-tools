package main

import (
	"bytes"
	"context"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"regexp"

	"cloudeng.io/cmdutil/subcmd"
	"cloudeng.io/errors"
	"cloudeng.io/file"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/file/filewalk/find"
	"cloudeng.io/file/filewalk/localfs"
	"cloudeng.io/file/matcher"
)

const dropboxCommand = `- name: dropbox
  summary: work with dropbox
  commands:
	- name: conflicts
	  summary: find drop box conflicts
	  arguments:
		- <directory>
		- ...
`

func newDropboxCLI() subcmd.Extension {
	return dropboxCLI{}
}

type dropboxCLI struct{}

func (gcli dropboxCLI) YAML() string {
	return dropboxCommand
}

func (gcli dropboxCLI) Name() string {
	return "dropbox"
}

func (gcli dropboxCLI) Set(cmdSet *subcmd.CommandSetYAML) error {
	cmds := &dropboxCommands{}
	re, err := regexp.Compile(`(.*) \(.* conflicted copy \d\d\d\d-\d\d-\d\d\)(.*)`)
	if err != nil {
		return err
	}
	cmds.conflictRE = re
	cmdSet.Set("dropbox", "conflicts").MustRunner(cmds.findConflicts, &dropboxConflictsFlags{})
	return nil
}

type dropboxCommands struct {
	conflictRE *regexp.Regexp
}

type dropboxConflictsFlags struct {
	Detail bool `subcmd:"details,false,display details of the original and conflict files"`
}

func (dbox *dropboxCommands) findConflicts(ctx context.Context, values interface{}, args []string) error {
	dbf := values.(*dropboxConflictsFlags)
	m, err := matcher.New(
		matcher.Regexp(`\(.* conflicted copy \d\d\d\d-\d\d-\d\d\)`),
	)
	if err != nil {
		return err
	}

	errs := &errors.M{}
	fs := localfs.New()
	ch := make(chan find.Found, 1000)
	handler := find.New(fs, ch, matcher.T{}, m, false)

	go func() {
		defer close(ch)
		wk := filewalk.New(fs, handler)
		for _, dir := range args {
			if err := wk.Walk(ctx, dir); err != nil {
				errs.Append(err)
				return
			}
		}
	}()

	for f := range ch {
		p := fs.Join(f.Prefix, f.Name)
		if len(f.Name) == 0 {
			p += "/"
		}
		if !dbf.Detail {
			fmt.Println(p)
			continue
		}
		if err := dbox.detail(ctx, fs, f); err != nil {
			errs.Append(err)
		}
	}
	return errs.Err()
}

func (dbx *dropboxCommands) detail(ctx context.Context, fs filewalk.FS, f find.Found) error {
	cfile := fs.Join(f.Prefix, f.Name)

	orig := dbx.conflictRE.ReplaceAllString(f.Name, `$1$2`)
	ofile := fs.Join(f.Prefix, orig)

	fmt.Println(f.Prefix)

	ci, oi, err := dbx.stat(ctx, fs, cfile, ofile)
	if err != nil {
		return err
	}

	var timeCmp string
	if ci.ModTime().After(oi.ModTime()) {
		timeCmp = fmt.Sprintf("%q > %q\n", orig, f.Name)
	} else {
		timeCmp = fmt.Sprintf("%q <= %q\n", orig, f.Name)
	}
	differ := ci.Size() != oi.Size()
	var csum, osum []byte
	if !differ {
		csum, osum, err = dbx.sha512s(cfile, ofile)
		if err != nil {
			return err
		}
		differ = bytes.Equal(csum, osum)

	}
	if differ {
		fmt.Printf("files difer, mod times: %v\n", timeCmp)
	}
	fmt.Println()
	return nil
}

func (dbx *dropboxCommands) stat(ctx context.Context, fs filewalk.FS, a, b string) (ai, bi file.Info, err error) {
	ai, err = fs.Lstat(ctx, a)
	if err != nil {
		return
	}
	bi, err = fs.Lstat(ctx, b)
	if err != nil {
		return
	}
	return
}

func (dbx *dropboxCommands) sha512(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (dbx *dropboxCommands) sha512s(a, b string) (asum, bsum []byte, err error) {
	asum, err = dbx.sha512(a)
	if err != nil {
		return
	}
	bsum, err = dbx.sha512(b)
	if err != nil {
		return
	}
	return
}
