package mod

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"red-cloud/mod/gologger"
	"red-cloud/utils"
	"time"

	uuid "github.com/satori/go.uuid"
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

func (p *RedcProject) CaseCreate(CaseName string, User string, Name string) error {
	// 创建新的 case 目录,这里不需要检测是否存在,因为名称是采用nanoID
	gologger.Info().Msgf("正在创建场景 「%s」", CaseName)
	uid := uuid.NewV4()

	// 从模版文件夹复制模版
	tpPath := filepath.Join("redc-templates", CaseName)
	casePath := filepath.Join(p.ProjectPath, uid.String())

	// 复制 tf文件
	gologger.Debug().Msgf("复制模版中 %s", uid.String())
	if err := utils.Dir(tpPath, casePath); err != nil {
		return fmt.Errorf("复制模版出错！\n%v", err)
	}

	// 在次 init,防止万一
	if err := TfInit2(casePath); err != nil {
		gologger.Error().Msgf("二次初始化失败！%s", err.Error())
		return err
	}

	// 初始化结构参数
	par, err := CaseScene(CaseName)
	if err != nil {
		gologger.Error().Msgf("场景参数校验失败！%s", err.Error())
		return err
	}

	// 初始化实例名称
	if Name == "" {
		Name = RandomName()
	}

	// 初始化实例
	c := &Case{
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Id:         uid.String(),
		Name:       Name,
		Operator:   User,
		Path:       casePath,
		Type:       CaseName,
		Parameter:  par,
	}

	// 构建场景
	if err := c.TfApply(); err != nil {
		gologger.Error().Msgf("场景创建失败！%s", err.Error())
		return err
	}
	gologger.Info().Msgf("场景创建成功！%s\n关闭命令: ./redc -stop  %s", uid.String(), uid.String())
	// 确认场景创建无误后,才会写入到配置文件中
	err = p.AddCase(c)
	err = p.SaveProject()
	if err != nil {
		gologger.Error().Msgf("项目配置保存失败！")
		return err
	}
	RedcLog("创建成功 " + p.ProjectPath + uid.String() + " " + CaseName)
	return nil
}

func (c *Case) TfApply() error {
	var err error
	err = TfApply(c.Path, c.Parameter...)
	if err != nil {
		return err
	}
	return nil
}
func (c *Case) TfDestroy() error {
	err := TfDestroy(c.Path, c.Parameter)
	if err != nil {
		gologger.Error().Msgf("场景销毁失败！%s", err.Error())
		return err
	}
	return nil
}
func (c *Case) Stop() error {

	err := c.TfDestroy()
	if err != nil {
		return err
	}

	// 成功销毁场景后,删除 case 文件夹
	err = os.RemoveAll(c.Path)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	return nil

}

func (c *Case) Kill() error {
	// 在次 init,防止万一
	dirs := utils.ChechDirMain(c.Path)
	for _, v := range dirs {
		err := utils.CheckFileName(v, "tf")
		if err {
			TfInit2(v)
		}
	}
	// 销毁场景
	if err := c.Stop(); err != nil {
		gologger.Error().Msgf("场景销毁失败！%s", err.Error())
		return err
	}
	return nil
}

func (c *Case) Change() error {
	// 销毁场景，不删除项目
	if err := c.TfDestroy(); err != nil {
		return err
	}
	// 重建场景
	if err := c.TfApply(); err != nil {
		return err
	}

	//if cfg.Section(UUID).Key("Type").String() == "cs-49" || cfg.Section(UUID).Key("Type").String() == "c2-new" || cfg.Section(UUID).Key("Type").String() == "snowc2" {
	//	C2Change(ProjectPath + "/" + UUID)
	//} else if cfg.Section(UUID).Key("Type").String() == "aliyun-proxy" {
	//	AliyunProxyChange(ProjectPath + "/" + UUID)
	//} else if cfg.Section(UUID).Key("Type").String() == "asm" {
	//	AsmChange(ProjectPath + "/" + UUID)
	//} else {
	//	fmt.Printf("不适用与当前场景")
	//	os.Exit(3)
	//}
	return nil
}

func (c *Case) Status() error {
	TfStatus(c.Path)
	return nil
}
