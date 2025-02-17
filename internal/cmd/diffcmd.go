package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type diffCmdConfig struct {
	Command        string                `mapstructure:"command"`
	Args           []string              `mapstructure:"args"`
	Exclude        *chezmoi.EntryTypeSet `mapstructure:"exclude"`
	Pager          string                `mapstructure:"pager"`
	include        *chezmoi.EntryTypeSet
	recursive      bool
	useBuiltinDiff bool
}

func (c *Config) newDiffCmd() *cobra.Command {
	diffCmd := &cobra.Command{
		Use:     "diff [target]...",
		Short:   "Print the diff between the target state and the destination state",
		Long:    mustLongHelp("diff"),
		Example: example("diff"),
		RunE:    c.runDiffCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadMockWrite,
		},
	}

	flags := diffCmd.Flags()
	flags.VarP(c.Diff.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Diff.include, "include", "i", "Include entry types")
	flags.BoolVarP(&c.Diff.recursive, "recursive", "r", c.Diff.recursive, "Recurse into subdirectories")
	flags.StringVar(&c.Diff.Pager, "pager", c.Diff.Pager, "Set pager")
	flags.BoolVarP(&c.Diff.useBuiltinDiff, "use-builtin-diff", "", c.Diff.useBuiltinDiff, "Use the builtin diff")

	return diffCmd
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	sb := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	if c.Diff.useBuiltinDiff || c.Diff.Command == "" {
		color := c.Color.Value(c.colorAutoFunc)
		gitDiffSystem := chezmoi.NewGitDiffSystem(dryRunSystem, &sb, c.DestDirAbsPath, color)
		if err := c.applyArgs(cmd.Context(), gitDiffSystem, c.DestDirAbsPath, args, applyArgsOptions{
			include:   c.Diff.include.Sub(c.Diff.Exclude),
			recursive: c.Diff.recursive,
			umask:     c.Umask,
		}); err != nil {
			return err
		}
		return c.pageOutputString(sb.String(), c.Diff.Pager)
	}
	diffSystem := chezmoi.NewExternalDiffSystem(dryRunSystem, c.Diff.Command, c.Diff.Args, c.DestDirAbsPath)
	defer diffSystem.Close()
	return c.applyArgs(cmd.Context(), diffSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:   c.Diff.include.Sub(c.Diff.Exclude),
		recursive: c.Diff.recursive,
		umask:     c.Umask,
	})
}
