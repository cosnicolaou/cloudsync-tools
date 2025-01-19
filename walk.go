package main

import (
	"context"
	"fmt"

	"cloudeng.io/cmdutil/flags"
	"cloudeng.io/errors"
	"cloudeng.io/file/filewalk"
	"cloudeng.io/file/filewalk/find"
	"cloudeng.io/file/filewalk/localfs"
	"cloudeng.io/file/matcher"
)

type walkCmd struct {
}

type WalkFlags struct {
	Recurse bool `subcmd:"recurse,false,walk recursively"`
	Stats   bool `subcmd:"stats,false,print stats"`
}

type walkListFlags struct {
	WalkFlags
	Long bool `subcmd:"long,false,print long listing"`
	All  bool `subcmd:"all,false,'print all files, including hidden . files and directories'"`
}

type walkFindFlags struct {
	WalkFlags
	Prune    bool            `subcmd:"prune,false,stop search when a directory match is found"`
	Prefixes flags.Repeating `subcmd:"prefix,,prefix expression components"`
	Files    flags.Repeating `subcmd:"file,,file expression components"`
}

func (walkCmd) parse(args []string) (matcher.T, error) {
	items := []matcher.Item{}
	for _, a := range args {
		switch a {
		case "d", "dir", "directory":
			items = append(items, matcher.FileType("d"))
		case "f", "file":
			items = append(items, matcher.FileType("f"))
		case "l", "link":
			items = append(items, matcher.FileType("l"))
		case "or":
			items = append(items, matcher.OR())
		case "and":
			items = append(items, matcher.AND())
		default:
			items = append(items, matcher.Regexp(a))
		}
	}
	return matcher.New(items...)
}

func (wc walkCmd) find(ctx context.Context, values interface{}, args []string) error {
	ff := values.(*walkFindFlags)

	pm, err := wc.parse(ff.Prefixes.Values)
	if err != nil {
		return err
	}
	fm, err := wc.parse(ff.Files.Values)
	if err != nil {
		return err
	}

	fs := localfs.New()
	errs := &errors.M{}
	ch := make(chan find.Found, 1000)
	handler := find.New(fs, ch, pm, fm, ff.Prune)

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
		fmt.Println(fs.Join(f.Prefix, f.Name))

	}
	return errs.Err()
}

/*
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
*/
