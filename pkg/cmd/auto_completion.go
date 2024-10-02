package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewAutoCompleteCommand() *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(kubectl-kruise completion bash)

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ kubectl-kruise completion zsh > "${fpath[1]}/_kubectl-kruise"

Fish:
  $ kubectl-kruise completion fish | source

PowerShell:
  PS> kubectl-kruise completion powershell | Out-String | Invoke-Expression
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell type %q", args[0])
			}
		},
	}
	return completionCmd
}
