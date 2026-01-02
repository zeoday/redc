package mod

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"red-cloud/mod/gologger"
	"red-cloud/utils"
	"text/tabwriter"
	"time"

	uuid "github.com/satori/go.uuid"
	"gopkg.in/ini.v1"
)

func RandomName() string {
	var lastName = []string{
		"red", "blue", "yellow", "brown", "purple", "anger", "lazy", "shy", "huge", "rare",
		"fast", "stupid", "sluggish", "boring", "rigid", "rigorous", "clever", "dexterity",
		"white", "black", "dark", "idiot", "shiny", "friendly", "integrity", "happy", "sad",
		"lively", "lonely", "ugly", "leisurely", "calm", "young", "tenacious"}
	var firstName = []string{
		"pig", "cow", "sheep", "mouse", "dragon", "serpent", "tiger", "fox", "frog", "chicken",
		"fish", "shrimp", "hippocampus", "helicopter", "crab", "dolphin", "whale", "chinchilla",
		"bunny", "mole", "rabbit", "horse", "monkey", "dog", "shark", "panda", "bear", "lion",
		"rhino", "leopard", "giraffe", "deer", "wolf", "parrot", "camel", "antelope", "turtle", "zebra"}
	var lastNameLen = len(lastName)
	var firstNameLen = len(firstName)
	rand.Seed(time.Now().UnixNano())     //设置随机数种子
	var first string                     //名
	for i := 0; i <= rand.Intn(1); i++ { //随机产生2位或者3位的名
		first = fmt.Sprint(firstName[rand.Intn(firstNameLen-1)])
	}
	return fmt.Sprint(lastName[rand.Intn(lastNameLen-1)]) + first
}

func CaseCreate(path string, CaseName string, User string, Name string) error {
	// 创建新的 case 目录,这里不需要检测是否存在,因为名称是采用nanoID
	uid := uuid.NewV4()
	tpPath := filepath.Join("redc-templates", CaseName)
	// workdir
	casePath := filepath.Join(ProjectPath, path, uid.String())

	// 复制 tf文件
	err := utils.Dir(tpPath, casePath)
	if err != nil {
		return fmt.Errorf("复制模版出错！\n%v", err)
	}
	gologger.Info().Msgf("case id: %s \n关闭命令: ./redc -stop %s", uid.String(), uid.String())

	// 在次 init,防止万一
	err = TfInit2(casePath)
	if err != nil {
		gologger.Error().Msgf("二次初始化失败！%s", err.Error())
		return err
	}

	gologger.Info().Msgf("正在初始化")
	if Name == "" {
		Name = RandomName()
	}
	c := &Case{
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Domain:     Domain,
		Domain2:    Domain2,
		Id:         uid.String(),
		Name:       Name,
		Node:       Node,
		Operator:   User,
		Path:       casePath,
		Type:       CaseName,
	}

	err = c.Apply()
	if err != nil {
		gologger.Error().Msgf("场景创建失败！%s", err.Error())
		return err
	}
	gologger.Info().Msgf("场景创建成功！%s\n关闭命令: ./redc -stop  %s", uid.String(), uid.String())
	// 确认场景创建无误后,才会写入到配置文件中
	pro, err := ProjectByName(path)
	if err != nil {
		gologger.Error().Msgf("未找到项目配置")
		return err
	}
	err = pro.AddCase(c)
	if err != nil {
		gologger.Error().Msgf("项目配置保存失败！")
		return err
	}

	RedcLog("创建成功 " + ProjectPath + uid.String() + " " + CaseName)
	return nil

}

func CaseStop(name string, uid string) error {
	project, err := ProjectByName(name)
	if err != nil {
		return fmt.Errorf("项目不存在 %s", err.Error())
	}
	c, err := project.GetCaseByUid(uid)
	if err != nil {
		return err
	}
	casePath := filepath.Join(project.ProjectPath, uid)

	err = TfDestroy(casePath, c.Parameter)
	if err != nil {
		gologger.Error().Msgf("场景销毁失败！%s", err.Error())
		return err
	}

	// 成功销毁场景后,删除 case 文件夹
	err = os.RemoveAll(casePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	err = project.HandleCase(uid)
	if err != nil {
		gologger.Error().Msgf("项目保存失败")
	}

	return nil

}

func CaseKill(name string, uid string) {
	// 在次 init,防止万一
	dirs := utils.ChechDirMain(filepath.Join(ProjectPath, uid))
	for _, v := range dirs {
		err := utils.CheckFileName(v, "tf")
		if err {
			TfInit2(v)
		}
	}
	// 销毁场景
	err := CaseStop(name, uid)
	if err != nil {
		gologger.Error().Msgf("场景销毁失败！%s", err.Error())
		return
	}
}

func CaseChange(ProjectPath string, UUID string) {

	filePath := ProjectPath + "/project.ini"
	cfg, err := ini.Load(filePath)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(3)
	}

	if cfg.Section(UUID).Key("Type").String() == "cs-49" || cfg.Section(UUID).Key("Type").String() == "c2-new" || cfg.Section(UUID).Key("Type").String() == "snowc2" {
		C2Change(ProjectPath + "/" + UUID)
	} else if cfg.Section(UUID).Key("Type").String() == "aliyun-proxy" {
		AliyunProxyChange(ProjectPath + "/" + UUID)
	} else if cfg.Section(UUID).Key("Type").String() == "asm" {
		AsmChange(ProjectPath + "/" + UUID)
	} else {
		fmt.Printf("不适用与当前场景")
		os.Exit(3)
	}

}

func CaseStatus(ProjectPath string, UUID string) {
	filePath := ProjectPath + "/project.ini"
	cfg, err := ini.Load(filePath)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(3)
	}

	fmt.Println("操作人员:", cfg.Section(UUID).Key("Operator").String())
	fmt.Println("项目名称:", cfg.Section(UUID).Key("Name").String())
	fmt.Println("场景类型:", cfg.Section(UUID).Key("Type").String())
	fmt.Println("创建时间:", cfg.Section(UUID).Key("CreateTime").String())

	TfStatus(ProjectPath + "/" + UUID)

}

func CaseList(name string) {
	// 读取项目 JSON 文件
	project, err := ProjectByName(name)
	if err != nil {
		gologger.Fatal().Msgf("项目读取失败:/%s", err.Error())
		return
	}

	// 使用 tabwriter 创建表格输出
	w := tabwriter.NewWriter(os.Stdout, 15, 0, 1, ' ', tabwriter.AlignRight)
	fmt.Fprintln(w, "UUID\tType\tName\tOperator\tCreateTime\t")

	// 遍历项目中的所有 Case
	for id, c := range project.Case {
		// 鉴权：只显示当前用户或 system 用户的 Case
		if c.Operator == U || U == "system" {
			// 从 Case 结构中获取 Type（从 Name 字段获取类型信息，或者需要额外的 Type 字段）
			caseType := c.Name // 假设 Name 包含类型信息，或者需要在 Case 结构中添加 Type 字段
			fmt.Fprintln(w, id, "\t", caseType, "\t", c.Name, "\t", c.Operator, "\t", c.CreateTime)
		}
	}

	err = w.Flush()
	if err != nil {
		gologger.Fatal().Msgf("表格输出失败:/%s", err.Error())
	}

}

func CheckUser(name string, uid string) error {

	pro, err := ProjectByName(name)
	if err != nil {
		gologger.Debug().Msgf("项目读取失败%s", err.Error())
		return fmt.Errorf("项目读取失败")
	}
	cs, err := pro.GetCaseByUid(uid)
	if err != nil {
		gologger.Debug().Msgf("case读取失败%s", err.Error())
		return fmt.Errorf("未找到uid为 %s 的case", uid)
	}
	if cs.Operator != U && U != "system" {
		return fmt.Errorf("当前用户无权限访问")
	}

	return nil
}
