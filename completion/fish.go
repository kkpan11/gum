package completion

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"
)

// Fish is a fish shell completion generator.
type Fish struct{}

// Run generates fish completion script.
func (f Fish) Run(ctx *kong.Context) error {
	var buf strings.Builder
	buf.WriteString(`# Fish shell completion for gum
# Generated by gum completion

# disable file completion unless explicitly enabled
complete -c gum -f

`)
	node := ctx.Model.Node
	f.gen(&buf, node)
	_, err := fmt.Fprint(ctx.Stdout, buf.String())
	if err != nil {
		return fmt.Errorf("unable to generate fish completion: %w", err)
	}
	return nil
}

func (f Fish) gen(buf io.StringWriter, cmd *kong.Node) {
	root := cmd
	for root.Parent != nil {
		root = root.Parent
	}
	rootName := root.Name
	if cmd.Parent == nil {
		_, _ = buf.WriteString(fmt.Sprintf("# %s\n", rootName))
	} else {
		_, _ = buf.WriteString(fmt.Sprintf("# %s\n", cmd.Path()))
		_, _ = buf.WriteString(
			fmt.Sprintf("complete -c %s -f -n '__fish_use_subcommand' -a %s -d '%s'\n",
				rootName,
				cmd.Name,
				cmd.Help,
			),
		)
	}

	for _, f := range cmd.Flags {
		if f.Hidden {
			continue
		}
		if cmd.Parent == nil {
			_, _ = buf.WriteString(
				fmt.Sprintf("complete -c %s -f",
					rootName,
				),
			)
		} else {
			_, _ = buf.WriteString(
				fmt.Sprintf("complete -c %s -f -n '__fish_seen_subcommand_from %s'",
					rootName,
					cmd.Name,
				),
			)
		}
		if !f.IsBool() {
			enums := flagPossibleValues(f)
			if len(enums) > 0 {
				_, _ = buf.WriteString(fmt.Sprintf(" -xa '%s'", strings.Join(enums, " ")))
			} else {
				_, _ = buf.WriteString(" -x")
			}
		}
		if f.Short != 0 {
			_, _ = buf.WriteString(fmt.Sprintf(" -s %c", f.Short))
		}
		_, _ = buf.WriteString(fmt.Sprintf(" -l %s", f.Name))
		_, _ = buf.WriteString(fmt.Sprintf(" -d \"%s\"", f.Help))
		_, _ = buf.WriteString("\n")
	}
	_, _ = buf.WriteString("\n")

	for _, c := range cmd.Children {
		if c == nil || c.Hidden {
			continue
		}
		f.gen(buf, c)
	}
}
