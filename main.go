package main

import (
	"context"

	"cloudeng.io/cmdutil/subcmd"
)

const commands = `name: cloudsync-tools
summary: tool for working with cloud sync services such as google drive etc.

commands:
	- name: softlinks
	  summary: work with soft links
	  commands:
		- name: rewrite
		  arguments:
			- <old-root-directory>
			- <new-root-directory>
		- name: verify
			arguments:
			- <old-root-directory>
			- <new-root-directory>
		- name: backups
		  summary: restore/delete backups of softlinks
		  commands:
			- name: restore
			  summary: restore backups of softlinks created by rewrite
			  arguments:
				- <directory>
			- name: delete
			  summary: delete backups of softlinks created by rewrite
			  arguments:
				- <directory>

	- name: walk
	  summary: efficiently walk large directories and directory trees
	  commands:
		- name: list
		  arguments:
			- <directory>
			- ...

	{{range subcmdExtension "gdrive"}}{{.}}
	{{end}}
`

func cli() *subcmd.CommandSetYAML {
	cmdSet := subcmd.MustFromYAMLTemplate(commands, newGDriveCLI())
	cmdSet.MustAddExtensions()

	sl := softlinkCmds{}
	cmdSet.Set("softlinks", "rewrite").MustRunner(sl.rewrite, &struct{}{})
	cmdSet.Set("softlinks", "verify").MustRunner(sl.verify, &softlinkVerifyFlags{})
	cmdSet.Set("softlinks", "backups", "restore").MustRunner(sl.restore, &struct{}{})
	cmdSet.Set("softlinks", "backups", "delete").MustRunner(sl.delete, &struct{}{})

	wk := walkCmd{}
	cmdSet.Set("walk", "list").MustRunner(wk.list, &walkListFlags{})

	return cmdSet
}

func main() {
	cli().MustDispatch(context.Background())
}
