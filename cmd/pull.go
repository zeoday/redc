package cmd

import (
	"red-cloud/mod"
	"red-cloud/mod/gologger"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// 定义命令行变量
var opts struct {
	Registry string
	// Dir 字段已移除，改为直接绑定 mod.TemplateDir
	Force   bool
	Timeout time.Duration
}

var pullCmd = &cobra.Command{
	Use:   "pull <image>[:tag]",
	Short: "Pull a template from registry",
	Args:  cobra.ExactArgs(1), // 必须传入 1 个参数
	RunE: func(cmd *cobra.Command, args []string) error {

		pullOpts := mod.PullOptions{
			RegistryURL: opts.Registry,
			Force:       opts.Force,
			Timeout:     opts.Timeout,
		}

		err := mod.Pull(cmd.Context(), args[0], pullOpts)

		if err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				gologger.Warning().Msg("❌ Operation canceled by user.")
				return nil
			}
			return err
		}

		return nil
	},
}

func init() {
	// 绑定 Registry 参数
	pullCmd.Flags().StringVarP(&opts.Registry, "registry", "r", "https://redc.wgpsec.org", "Registry URL")
	pullCmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force pull (overwrite)")
	pullCmd.Flags().DurationVar(&opts.Timeout, "timeout", 60*time.Second, "Download timeout")

	rootCmd.AddCommand(pullCmd)
}
