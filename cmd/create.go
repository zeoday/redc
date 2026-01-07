package cmd

import (
	redc "red-cloud/mod"
	"red-cloud/mod/gologger"

	"github.com/spf13/cobra"
)

var (
	userName    string
	projectName string
)

var runCmd = &cobra.Command{
	Use:     "run [template_name]",
	Short:   "创建并立即启动一个场景",
	Example: "redc run ecs",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		c := createLogic(templateName)
		if err := c.TfApply(); err != nil {
			gologger.Error().Msgf("场景启动失败！%s", err.Error())
		}
	},
}

var createCmd = &cobra.Command{
	Use:     "create [template_name]",
	Short:   "创建一个新的基础设施场景",
	Example: "redc create ecs -u team1 -n operation_alpha",
	Args:    cobra.ExactArgs(1), // 强制要求输入一个模板名，例如 pte
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]
		createLogic(templateName)
	},
}

func createLogic(templateName string) *redc.Case {

	// 别名处理
	if templateName == "pte" {
		templateName = "pte_arm"
	}

	// 解析 Project (这里需要确保 Config 已经在 root.go 加载了)
	pro, err := redc.ProjectParse(redc.Project, userName) // 注意：这里使用了 flag 传入的 userName
	if err != nil {
		gologger.Fatal().Msgf("项目解析失败: %s", err)
	}

	// 创建 Case
	c, err := pro.CaseCreate(templateName, userName, projectName)
	if err != nil {
		gologger.Error().Msgf("❌「%s」场景创建失败: %v", templateName, err)
		return nil
	}
	gologger.Info().Msgf("✅「%s」场景创建完成！", templateName)
	return c
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(runCmd)
	createCmd.Flags().StringVarP(&userName, "user", "u", "system", "指定用户/操作员")
	createCmd.Flags().StringVarP(&projectName, "name", "n", "", "指定项目/任务名称")
}
