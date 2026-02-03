package cmd

import (
	"fmt"
	"os"
	"red-cloud/mod"
	"red-cloud/mod/gologger"
	"strings"
	"text/tabwriter"
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
var searchCmd = &cobra.Command{
	Use:   "search xxx",
	Short: "Search registry ",
	Args:  cobra.ExactArgs(1), // 必须传入 1 个参数
	RunE: func(cmd *cobra.Command, args []string) error {

		pullOpts := mod.PullOptions{
			RegistryURL: opts.Registry,
			Force:       opts.Force,
			Timeout:     opts.Timeout,
		}
		res, err := mod.Search(cmd.Context(), args[0], pullOpts)

		if err != nil {
			if strings.Contains(err.Error(), "context canceled") {
				gologger.Warning().Msg("❌ Operation canceled by user.")
				return nil
			}
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', 0)
		// 打印表头
		fmt.Fprintln(w, "NAME\tVERSION\tAUTHOR\tDESCRIPTION")

		for _, item := range res {
			// 处理描述：移除换行符 + 截断过长文本
			desc := cleanDescription(item.Description)

			// 格式化输出
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.Key,     // NAME (例如 aliyun/ecs)
				item.Version, // VERSION
				item.Author,  // AUTHOR
				desc,         // DESCRIPTION
			)
		}

		// 刷新缓冲区，将内容输出到终端
		w.Flush()

		return nil
	},
}

// cleanDescription 辅助函数：清理和截断描述文本
func cleanDescription(desc string) string {
	// 1. 替换所有换行符为空格，防止表格错乱
	desc = strings.ReplaceAll(desc, "\n", " ")
	desc = strings.ReplaceAll(desc, "\r", "")

	// 2. 截断长度 (例如 60 个字符)，并在末尾加 "..."
	// 注意：如果有中文，直接用切片 desc[:60] 可能会乱码，建议使用 rune
	const maxLen = 60
	runes := []rune(desc)

	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return desc
}

func init() {
	// 绑定 Registry 参数
	pullCmd.Flags().StringVarP(&opts.Registry, "registry", "r", "https://redc.wgpsec.org", "Registry URL")
	pullCmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force pull (overwrite)")
	pullCmd.Flags().DurationVar(&opts.Timeout, "timeout", 60*time.Second, "Download timeout")
	searchCmd.Flags().StringVarP(&opts.Registry, "registry", "r", "https://redc.wgpsec.org", "Registry URL")
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(searchCmd)
}
