package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thin-edge/tedge-oscar/internal/artifact"
	"github.com/thin-edge/tedge-oscar/internal/config"
	"github.com/thin-edge/tedge-oscar/internal/imagepull"
	"github.com/thin-edge/tedge-oscar/internal/registryauth"
)

var pullCmd = &cobra.Command{
	Use:     "pull [image]",
	Short:   "Pull a flow image from an OCI registry",
	Example: `tedge-oscar flows images pull ghcr.io/thin-edge/connectivity-counter:1.0`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Enable debug HTTP if logLevel is debug
		registryauth.SetDebugHTTP(logLevel)
		cfgPath := configPath
		if cfgPath == "" {
			cfgPath = config.DefaultConfigPath()
		}
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		saveAsTarball, _ := cmd.Flags().GetBool("tarball")
		tarballPath := ""

		imageRef := args[0]
		outputDir, _ := cmd.Flags().GetString("output-dir")
		name, err := artifact.ParseName(imageRef, false)
		if err != nil {
			return err
		}

		if outputDir == "" {
			outputDir = filepath.Join(cfg.ImageDir, name)
		}
		if saveAsTarball {
			tarballPath = filepath.Join(filepath.Dir(outputDir), fmt.Sprintf("%s.tar", name))
		}
		if err := imagepull.PullImage(cfg, imageRef, outputDir, tarballPath, false); err != nil {
			return err
		}
		if tarballPath == "" {
			fmt.Fprintf(cmd.ErrOrStderr(), "Image %s pulled to %s\n", imageRef, outputDir)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "Image %s pulled to %s\n", imageRef, tarballPath)
		}
		return nil
	},
}

func init() {
	pullCmd.Flags().String("output-dir", "", "Directory to download the artifact contents to (default: config image_dir)")
	pullCmd.Flags().Bool("tarball", false, "Save artifact as a tarball")
	imagesCmd.AddCommand(pullCmd)
}
