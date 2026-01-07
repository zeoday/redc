package main

import (
	_ "embed"
	"flag"
	"os"
	redc "red-cloud/mod"
	"red-cloud/mod/gologger"
	"red-cloud/utils"
)

const banner = `
██████╗  ███████╗ ██████╗   ██████╗ 
 ██╔══██╗ ██╔════╝ ██╔══██╗ ██╔════╝ 
 ██████╔╝ █████╗   ██║  ██║ ██║      
 ██╔══██╗ ██╔══╝   ██║  ██║ ██║      
 ██║  ██║ ███████╗ ██████╔╝ ╚██████╗ 
 ╚═╝  ╚═╝ ╚══════╝ ╚═════╝   ╚═════╝
`

func Banner() {
	gologger.Print().Msgf("%s\nVersion: %s\n\n", banner, redc.Version)
}
func main() {
	flag.Parse()
	// -version 显示版本号
	if redc.V {
		Banner()
		os.Exit(0)
	}

	// 解析配置文件
	if err := redc.LoadConfig("./config.yaml"); err != nil {
		gologger.Fatal().Msgf("配置文件加载失败! %s", err.Error())
	}

	// -init 初始化
	if redc.Init {
		redc.RedcLog("进行初始化")
		gologger.Info().Msgf("初始化中")
		// 先删除文件夹
		err := os.RemoveAll("redc-templates")
		if err != nil {
			gologger.Error().Msgf("初始化过程中删除模板文件夹失败: %s", err)
		}
		// 释放 templates 资源
		utils.ReleaseDir("redc-templates")

		// 遍历 redc-templates 文件夹,不包括子目录
		_, dirs := utils.GetFilesAndDirs("./redc-templates")
		for _, v := range dirs {
			err = redc.TfInit(v)
			if err != nil {
				gologger.Error().Msgf("「%s」场景初始化失败\n %s", v, err)
			} else {
				gologger.Info().Msgf("✅「%s」场景初始化任务完成！", v)
			}
		}
		return
	}

	// 解析项目名称
	pro, err := redc.ProjectParse(redc.Project, redc.U)
	if err != nil {
		gologger.Fatal().Msgf("项目解析失败: %s", err)
	}

	// list 操作查看项目里所有 case
	if redc.List {
		pro.CaseList()
	}

	// start 操作,去调用 case 创建方法
	if redc.Start != "" {
		redc.RedcLog("start " + redc.Start)
		if redc.Start == "pte" {
			redc.Start = "pte_arm"
		}
		err = pro.CaseCreate(redc.Start, redc.U, redc.Name)
		if err != nil {
			gologger.Error().Msgf("「%s」场景创建失败\n %s", redc.Start, err)
			return
		}
		return
	}

	var targetID string // 用来存用户输入的那个 ID

	// 先看用户用了哪个 flag，把 ID 拿出来
	switch {
	case redc.Stop != "":
		targetID = redc.Stop
	case redc.Kill != "":
		targetID = redc.Kill
	case redc.Change != "":
		targetID = redc.Change
	case redc.Status != "":
		targetID = redc.Status
	default:
		// 如果都不是，说明没输命令，直接结束
		return
	}

	// 根据 ID 查找 Case 对象 (只查一次)
	c, err := pro.GetCaseByUid(targetID)
	if err != nil {
		gologger.Error().Msgf("操作失败: 找不到 ID 为「%s」的场景\n错误: %s", targetID, err)
		return
	}
	redc.RedcLog("Action on " + targetID)
	var actionErr error // 接收执行错误

	switch {
	case redc.Stop != "":
		// Stop 有特殊逻辑，需要额外处理
		actionErr = c.Stop()
		if actionErr == nil {
			actionErr = pro.HandleCase(c)
		}
	case redc.Kill != "":
		actionErr = c.Kill()
	case redc.Change != "":
		actionErr = c.Change()
	case redc.Status != "":
		actionErr = c.Status()
	}
	// 统一报错打印
	if actionErr != nil {
		gologger.Error().Msgf("执行失败: %v", actionErr)
	}
}
