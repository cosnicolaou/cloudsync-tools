package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"sync/atomic"

	"cloudeng.io/file"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/file/filewalk/localfs"
)

type walkCmd struct {
}

func (wc walkCmd) list(ctx context.Context, values interface{}, args []string) error {
	wlf := values.(*walkListFlags)
	fs := localfs.New()
	for _, dir := range args {
		w := filewalk.New(fs, &walker{fs: fs, flags: wlf})
		if err := w.Walk(ctx, dir); err != nil {
			return err
		}
	}
	return nil
}

type walkListFlags struct {
	Recurse bool `subcmd:"recurse,false,walk recursively"`
	Count   bool `subcmd:"count-only,false,print count of files and directories"`
	Long    bool `subcmd:"long,false,print long listing"`
	All     bool `subcmd:"all,false,'print all files, including hidden . files and directories'"`
}

type walker struct {
	fs          filewalk.FS
	flags       *walkListFlags
	numChildren int64
	outputLock  sync.Mutex
}

type walkerState struct {
	lf *formatter
}

func (w *walker) Prefix(_ context.Context, state *walkerState, prefix string, _ file.Info, err error) (bool, file.InfoList, error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error for directory %v: %v\n", prefix, err)
		return true, nil, err
	}
	state.lf = newFormatter()
	return false, nil, nil
}

func (w *walker) Contents(ctx context.Context, state *walkerState, prefix string, dirEntries []filewalk.Entry) (file.InfoList, error) {
	atomic.AddInt64(&w.numChildren, int64(len(dirEntries)))
	if !w.flags.Recurse && w.flags.Count {
		return nil, nil
	}
	children := make(file.InfoList, 0, len(dirEntries))
	for _, d := range dirEntries {
		filename := w.fs.Join(prefix, d.Name)
		info, err := w.fs.Lstat(ctx, filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lstat failed for file %v: %v\n", filename, err)
			continue
		}
		suffix := ""
		if info.IsDir() && w.flags.Recurse {
			children = append(children, info)
			suffix = "/"
		}
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			suffix = "@"
		}
		if w.flags.Count {
			continue
		}
		if !w.flags.All && d.Name[0] == '.' {
			continue
		}
		if !w.flags.Long {
			state.lf.append(d.Name + suffix)
			continue
		}
		fmt.Println(fs.FormatFileInfo(info))
	}
	return children, nil
}

func (w *walker) Done(ctx context.Context, state *walkerState, prefix string, err error) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	w.outputLock.Lock()
	defer w.outputLock.Unlock()
	fmt.Println(prefix)
	state.lf.flush()
	fmt.Println()
	return err
}
