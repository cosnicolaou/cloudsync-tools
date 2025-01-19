package softlinks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloudeng.io/file"
	"cloudeng.io/file/filewalk"
)

type linkHandler interface {
	Handle(ctx context.Context, filename string, info file.Info) error
}

type common struct {
	fs filewalk.FS
	lh linkHandler
}

func (c *common) Prefix(_ context.Context, _ *struct{}, prefix string, _ file.Info, err error) (bool, file.InfoList, error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error for directory %v: %v\n", prefix, err)
		return true, nil, err
	}
	return false, nil, nil
}

func (c *common) Contents(ctx context.Context, _ *struct{}, prefix string, dirEntries []filewalk.Entry) (file.InfoList, error) {
	children := make(file.InfoList, 0, len(dirEntries))
	for _, d := range dirEntries {
		filename := c.fs.Join(prefix, d.Name)
		info, err := c.fs.Lstat(ctx, filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lstat failed for file %v: %v\n", filename, err)
			continue
		}
		if info.IsDir() {
			children = append(children, info)
			continue
		}
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			if err := c.lh.Handle(ctx, filename, info); err != nil {
				return nil, err
			}
		}
	}
	return children, nil
}

func (c *common) Done(_ context.Context, _ *struct{}, prefix string, err error) error {
	return err
}

type rewriter struct {
	oldRoot, newRoot, backupSuffix string
}

func (rw rewriter) Handle(_ context.Context, filename string, _ file.Info) error {
	target, err := os.Readlink(filename)
	if err != nil {
		return fmt.Errorf("readlink failed for file %v: %v\n", filename, err)
	}
	if !strings.HasPrefix(target, rw.oldRoot) {
		return nil
	}
	moved := strings.Replace(target, rw.oldRoot, rw.newRoot, 1)
	fmt.Printf("replacing %v with %v\n", target, moved)
	os.Rename(filename, filename+rw.backupSuffix)
	return os.Symlink(moved, filename)
}

func NewRewriter(fs filewalk.FS, oldRoot, newRoot, backupSuffix string) filewalk.Handler[struct{}] {
	rw := rewriter{
		oldRoot:      oldRoot,
		newRoot:      newRoot,
		backupSuffix: backupSuffix,
	}
	return &common{fs: fs, lh: rw}
}

type verifier struct {
	oldRoot, newRoot, backupSuffix string
	all                            bool
}

func NewVerifier(fs filewalk.FS, oldRoot, newRoot, backupSuffix string, showAll bool) filewalk.Handler[struct{}] {
	v := verifier{
		oldRoot:      oldRoot,
		newRoot:      newRoot,
		backupSuffix: backupSuffix,
		all:          showAll,
	}
	return &common{fs: fs, lh: v}
}

func (v verifier) Handle(_ context.Context, filename string, _ file.Info) error {
	if strings.HasSuffix(filename, v.backupSuffix) {
		return nil
	}
	target, err := os.Readlink(filename)
	if err != nil {
		return fmt.Errorf("readlink failed for file %v: %v\n", filename, err)
	}
	if filepath.IsAbs(target) && !strings.HasPrefix(target, v.newRoot) {
		fmt.Printf("\u274c  %v\n", filename)
		return nil
	}
	if v.all {
		fmt.Printf("\u2705  %v\n", filename)
	}
	return nil
}

func NewBackupRestore(fs filewalk.FS, backupSuffix string) filewalk.Handler[struct{}] {
	return &common{fs: fs, lh: &backupRestore{backupSuffix}}
}

type backupRestore struct {
	backupSuffix string
}

func (br *backupRestore) Handle(_ context.Context, filename string, _ file.Info) error {
	if !strings.HasSuffix(filename, br.backupSuffix) {
		return nil
	}
	restored := strings.TrimSuffix(filename, br.backupSuffix)
	fmt.Printf("restoring %v to %v\n", filename, restored)
	return os.Rename(filename, restored)
}

func NewBackupDelete(fs filewalk.FS, backupSuffix string) filewalk.Handler[struct{}] {
	return &common{fs: fs, lh: &backupRestore{backupSuffix}}
}

type backupDelete struct {
	backupSuffix string
}

func (bd *backupDelete) Handle(_ context.Context, filename string, _ file.Info) error {
	if !strings.HasSuffix(filename, bd.backupSuffix) {
		return nil
	}
	fmt.Printf("deleting %v\n", filename)
	return os.Remove(filename)
}
