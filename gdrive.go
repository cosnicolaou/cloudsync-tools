package main

import (
	"context"

	"cloudeng.io/cmdutil/subcmd"
	"github.com/cosnicolaou/cloudsync-tools/internal/gdrive"
)

const gdriveCommand = `- name: gdrive
  summary: work with google drive
  commands:
    - name: list-drives
      summary: list the available drives
`

func newGDriveCLI() subcmd.Extension {
	return gdriveCLI{}
}

type gdriveCLI struct{}

func (gcli gdriveCLI) YAML() string {
	return gdriveCommand
}

func (gcli gdriveCLI) Name() string {
	return "gdrive"
}

func (gcli gdriveCLI) Set(cmdSet *subcmd.CommandSetYAML) error {
	cmds := &gcliCommands{}
	cmdSet.Set("gdrive", "list-drives").MustRunner(cmds.listDrives, &struct{}{})
	return nil
}

type gcliCommands struct {
}

func (gcli *gcliCommands) listDrives(ctx context.Context, params interface{}, args []string) error {
	return gdrive.ListFiles(ctx)
}
