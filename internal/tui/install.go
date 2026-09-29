package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/openspec"
	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

type installState int

const (
	installIdle installState = iota
	installPreparing
	installConfirming
	installConfirmingOverwrite
	installWriting
	installDone
	installFailed
)

type InstallPrepared struct {
	Plan openspec.Plan
	Err  error
}

type InstallFinished struct {
	Name       string
	Validation openspec.Validation
	Unverified string
	Err        error
}

type installFlow struct {
	state installState
	row   registry.Row
	plan  openspec.Plan

	err        error
	validation openspec.Validation
	unverified string

	installer *Installer
}

func newInstallFlow(installer *Installer) *installFlow {
	return &installFlow{installer: installer}
}

func (f *installFlow) active() bool { return f.state != installIdle }

func (f *installFlow) start(row registry.Row) (tea.Cmd, string) {
	if f.installer == nil || !f.installer.InProject {
		return nil, openspec.ErrNotAProject.Error()
	}

	if !row.Installable {
		switch row.Origin {
		case registry.OriginBuiltIn:
			return nil, row.Name + " ships with OpenSpec and needs no installing"
		default:
			return nil, row.Name + " already resolves for this project, so there is nothing to install"
		}
	}

	f.state = installPreparing
	f.row = row
	f.err = nil

	return f.installer.PrepareCmd(row), ""
}

func (f *installFlow) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case InstallPrepared:
		if msg.Err != nil {
			f.state = installFailed
			f.err = msg.Err
			return nil
		}

		f.plan = msg.Plan
		if f.plan.Occupied {
			f.state = installConfirmingOverwrite
		} else {
			f.state = installConfirming
		}
		return nil

	case InstallFinished:
		if msg.Err != nil {
			f.state = installFailed
			f.err = msg.Err
			return nil
		}

		f.state = installDone
		f.validation = msg.Validation
		f.unverified = msg.Unverified
		return nil

	case tea.KeyPressMsg:
		return f.handleKey(msg)
	}

	return nil
}

func (f *installFlow) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.String()

	switch f.state {
	case installConfirming:
		switch key {
		case "y", "enter":
			f.state = installWriting
			return f.installer.WriteCmd(f.plan, false)
		case "n", "esc", "q":
			f.state = installIdle
		}

	case installConfirmingOverwrite:
		switch key {
		case "o":
			f.state = installWriting
			return f.installer.WriteCmd(f.plan, true)
		case "n", "esc", "q", "y", "enter":
			f.state = installIdle
		}

	case installDone, installFailed:
		switch key {
		case "esc", "q", "enter":
			f.state = installIdle
		}
	}

	return nil
}

func (f *installFlow) View(width int) string {
	var b strings.Builder

	switch f.state {
	case installPreparing:
		b.WriteString(headingStyle.Render("Install " + f.row.Name))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("Fetching the source…"))

	case installConfirming:
		b.WriteString(headingStyle.Render("Install " + f.plan.Name))
		b.WriteString("\n\n")
		b.WriteString(f.writeList(width))
		b.WriteString("\n\n")
		b.WriteString(keyStyle.Render("y") + " write these files    " + keyStyle.Render("n") + " cancel")

	case installConfirmingOverwrite:
		b.WriteString(headingStyle.Render("Overwrite " + f.plan.Name + "?"))
		b.WriteString("\n\n")
		b.WriteString(wrap(f.plan.Relative+" already holds a schema. ossm cannot tell whether it came from this entry, because OpenSpec records no provenance. Overwriting replaces it entirely.", width))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("%d file(s) there now, %d would be written", len(f.plan.Existing), len(f.plan.Files))))
		b.WriteString("\n\n")
		b.WriteString(f.writeList(width))
		b.WriteString("\n\n")
		b.WriteString(keyStyle.Render("o") + " overwrite it    " + keyStyle.Render("n") + " leave it alone")

	case installWriting:
		b.WriteString(headingStyle.Render("Installing " + f.plan.Name))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("Writing…"))

	case installDone:
		b.WriteString(headingStyle.Render(f.plan.Name + " installed"))
		b.WriteString("\n\n")
		b.WriteString(truncate(f.plan.Destination, width))
		b.WriteString("\n\n")
		b.WriteString(f.verdict(width))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("The project default is unchanged. Set it from the Project tab with s."))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("enter · esc · q  close"))

	case installFailed:
		b.WriteString(headingStyle.Render("Install failed"))
		b.WriteString("\n\n")
		b.WriteString(wrap(f.err.Error(), width))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("Nothing was written. enter · esc · q  close"))
	}

	return b.String()
}

func (f *installFlow) writeList(width int) string {
	var b strings.Builder

	b.WriteString("into " + truncate(f.plan.Relative, width-5) + "  in " + truncate(f.plan.Root, width-10))
	b.WriteString("\n")

	const shown = 12
	for i, file := range f.plan.Files {
		if i == shown {
			b.WriteString("\n  " + dimStyle.Render(fmt.Sprintf("and %d more", len(f.plan.Files)-shown)))
			break
		}
		b.WriteString("\n  " + truncate(file, width-2))
	}

	return b.String()
}

func (f *installFlow) verdict(width int) string {
	if f.unverified != "" {
		return dimStyle.Render(wrap("Installed but unverified: "+f.unverified, width))
	}

	if f.validation.Valid {
		return "OpenSpec validated it."
	}

	var b strings.Builder
	b.WriteString(headingStyle.Render("OpenSpec rejected it:"))
	for _, issue := range f.validation.Issues {
		b.WriteString("\n  ")
		b.WriteString(truncate(issue.Message, width-2))
	}

	return b.String()
}

type Installer struct {
	Fetcher   source.Fetcher
	CLI       openspec.CLI
	Root      string
	InProject bool
}

func (i *Installer) PrepareCmd(row registry.Row) tea.Cmd {
	return func() tea.Msg {
		if row.Entry == nil {
			return InstallPrepared{Err: errors.New("this schema has no source to install from")}
		}

		dir, err := i.Fetcher.Fetch(context.Background(), source.Source{
			Repo: row.Entry.Source.Repo,
			Path: row.Entry.Source.Path,
			Ref:  row.Entry.Source.Ref,
		}, false)
		if err != nil {
			return InstallPrepared{Err: err}
		}

		plan, err := openspec.PlanInstall(i.Root, dir, row.Name)

		return InstallPrepared{Plan: plan, Err: err}
	}
}

func (i *Installer) WriteCmd(plan openspec.Plan, overwrite bool) tea.Cmd {
	return func() tea.Msg {
		if err := openspec.Install(plan, overwrite); err != nil {
			return InstallFinished{Name: plan.Name, Err: err}
		}

		validation, err := openspec.ValidateInstalled(context.Background(), i.CLI, i.Root, plan.Name)
		if err != nil {
			return InstallFinished{Name: plan.Name, Unverified: err.Error()}
		}

		return InstallFinished{Name: plan.Name, Validation: validation}
	}
}
