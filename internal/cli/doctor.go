package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/weiyong1024/clawsandbox/internal/config"
	"github.com/weiyong1024/clawsandbox/internal/container"
)

var doctorCmd = &cobra.Command{
	Use:           "doctor",
	Short:         "Run a local preflight check and show the next step",
	Example:       "  clawsandbox doctor",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE:          runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Println("ClawSandbox Doctor")
	fmt.Println()

	cli, err := container.NewClient()
	if err != nil {
		fmt.Println("Docker: NOT READY")
		fmt.Println("  ClawSandbox could not reach your local Docker engine.")
		fmt.Println("  See the detail below, then retry once Docker is reachable.")
		fmt.Println()
		fmt.Println("Current path: blocked before startup")
		fmt.Println("Estimated wait: 1-2 minutes after Docker starts")
		fmt.Println("Next step:")
		fmt.Println("  1. Start Docker Desktop (or your Docker engine).")
		fmt.Println("  2. Wait until `docker version` succeeds.")
		fmt.Println("  3. Run `clawsandbox doctor` again.")
		fmt.Println()
		fmt.Println("Details:")
		for _, line := range strings.Split(err.Error(), "\n") {
			fmt.Printf("  %s\n", line)
		}
		return silentExitError{code: 1}
	}

	fmt.Println("Docker: READY")
	fmt.Println("  Your local Docker engine is reachable.")
	fmt.Println()

	imageRef := cfg.ImageRef()
	imageExists, err := container.ImageExists(cli, imageRef)
	if err != nil {
		return err
	}

	if imageExists {
		fmt.Println("Image: READY")
		fmt.Printf("  Local image `%s` is already available.\n", imageRef)
		fmt.Println()
		fmt.Println("Current path: ready to create claws now")
		fmt.Println("Estimated wait: Dashboard starts in seconds")
		fmt.Println("Next step:")
		fmt.Println("  1. Run `clawsandbox dashboard serve`.")
		fmt.Println("  2. Open `http://localhost:8080`.")
		fmt.Println("  3. Create your first claw.")
		fmt.Println("  4. Click `Configure` to finish onboarding.")
		return nil
	}

	fmt.Println("Image: NOT PRESENT LOCALLY")
	fmt.Printf("  Local image `%s` is not available on this machine yet.\n", imageRef)
	fmt.Println()
	fmt.Println("Current path: create now, then let ClawSandbox prepare the image")
	fmt.Println("Estimated wait: the first pull or local build can take several minutes")
	fmt.Println("Next step:")
	fmt.Println("  1. Run `clawsandbox dashboard serve`.")
	fmt.Println("  2. Open `http://localhost:8080`, then create your first claw.")
	fmt.Println("  3. Click `Configure` to finish onboarding.")
	fmt.Println("Optional:")
	fmt.Println("  Run `clawsandbox build` first if you prefer to prebuild locally for offline use or customization.")
	return nil
}
