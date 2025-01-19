package main

import (
	"context"

	"cloudeng.io/file/filewalk"
	"cloudeng.io/file/filewalk/localfs"
	"github.com/cosnicolaou/cloudsync-tools/internal/softlinks"
)

type softlinkCmds struct {
}

type softlinkVerifyFlags struct {
	All bool `subcmd:"all,false,'show all softlinks, including those that are valid'"`
}

const softlinkBackupSuffix = ".softlink-bak"

func (sc softlinkCmds) rewrite(ctx context.Context, values interface{}, args []string) error {
	fs := localfs.New()
	w := filewalk.New(fs,
		softlinks.NewRewriter(fs, args[0], args[1], softlinkBackupSuffix))
	return w.Walk(ctx, args[1])
}

func (sc softlinkCmds) verify(ctx context.Context, values interface{}, args []string) error {
	vf := values.(*softlinkVerifyFlags)
	fs := localfs.New()
	w := filewalk.New(fs,
		softlinks.NewVerifier(fs, args[0], args[1], softlinkBackupSuffix, vf.All))
	return w.Walk(ctx, args[1])
}

func (sc softlinkCmds) delete(ctx context.Context, values interface{}, args []string) error {
	fs := localfs.New()
	w := filewalk.New(fs, softlinks.NewBackupDelete(fs, softlinkBackupSuffix))
	return w.Walk(ctx, args[0])
}

func (sc softlinkCmds) restore(ctx context.Context, values interface{}, args []string) error {
	fs := localfs.New()
	w := filewalk.New(fs, softlinks.NewBackupRestore(fs, softlinkBackupSuffix))
	return w.Walk(ctx, args[0])
}
