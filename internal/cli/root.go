package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "clawsandbox",
	Short: "Deploy and manage a fleet of OpenClaw instances",
	Long: `ClawSandbox lets you spin up multiple isolated OpenClaw instances
on a single machine. Each instance runs in its own Docker container
with a full Linux desktop, accessible via your browser.`,
}

type silentExitError struct {
	code int
}

func (e silentExitError) Error() string {
	return ""
}

func Execute() {
	rootCmd.AddCommand(
		buildCmd,
		doctorCmd,
		createCmd,
		listCmd,
		startCmd,
		stopCmd,
		restartCmd,
		destroyCmd,
		desktopCmd,
		logsCmd,
		dashboardCmd,
		configCmd,
		configureCmd,
		versionCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		var silentErr silentExitError
		if errors.As(err, &silentErr) {
			os.Exit(silentErr.code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
