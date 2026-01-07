package mod

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"red-cloud/mod/gologger"
	"red-cloud/utils"
	"time"
)

type CaseState string

const (
	StateRunning CaseState = "running"
	StateStopped CaseState = "stopped"
	StateError   CaseState = "error"
	StateCreated CaseState = "created"
	StatePending CaseState = "pending"
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

// GenerateCaseID 生成 ID (64字符 hex string)
// 本质是 32 字节 (256 bit) 的随机数
func GenerateCaseID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// 极端情况下随机数生成失败，回退到时间戳+简单的随机，或者直接 panic
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// CaseScene 场景参数判断
func CaseScene(t string) ([]string, error) {
	var par []string
	switch t {
	case "cs-49", "c2-new", "snowc2":
		par = RVar(
			fmt.Sprintf("node_count=%d", Node),
			fmt.Sprintf("domain=%s", Domain),
		)
	case "aws-proxy", "aliyun-proxy", "asm":
		par = RVar(fmt.Sprintf("node_count=%d", Node))
	case "dnslog", "xraydnslog", "interactsh":
		if Domain == "360.com" {
			return par, fmt.Errorf("创建 dnslog 时,域名不可为默认值")
		}
		par = RVar(fmt.Sprintf("domain=%s", Domain))
	case "pss5", "frp", "frp-loki", "nps":
		par = []string{fmt.Sprintf("base64_command=%s", Base64Command)}
	case "asm-node":
		par = RVar(
			fmt.Sprintf("node_count=%d", Node),
			fmt.Sprintf("domain2=%s", Domain2),
			fmt.Sprintf("doamin=%s", Domain),
		)
	}
	return par, nil
}

func (p *RedcProject) CaseCreate(CaseName string, User string, Name string) (*Case, error) {
	// 创建新的 case 目录,这里不需要检测是否存在,因为名称是采用nanoID
	gologger.Info().Msgf("正在创建场景 「%s」", CaseName)
	uid := GenerateCaseID()

	// 从模版文件夹复制模版
	tpPath := filepath.Join("redc-templates", CaseName)
	casePath := filepath.Join(p.ProjectPath, uid)

	// 复制 tf文件
	gologger.Debug().Msgf("复制模版中 %s", uid)
	if err := utils.Dir(tpPath, casePath); err != nil {
		return nil, fmt.Errorf("复制模版出错！\n%v", err)
	}

	// 在次 init,防止万一
	if err := TfInit2(casePath); err != nil {
		gologger.Error().Msgf("二次初始化失败！%s", err.Error())
		return nil, err
	}

	// 初始化结构参数
	par, err := CaseScene(CaseName)
	if err != nil {
		gologger.Error().Msgf("场景参数校验失败！%s", err.Error())
		return nil, err
	}

	// 初始化实例名称
	if Name == "" {
		Name = RandomName()
	}

	// 初始化实例
	c := &Case{
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Id:         uid,
		Name:       Name,
		Operator:   User,
		Path:       casePath,
		Type:       CaseName,
		Parameter:  par,
		State:      StateCreated,
	}

	// 构建场景
	if err := c.TfPlan(); err != nil {
		gologger.Error().Msgf("场景创建失败！%s", err.Error())
		return nil, err
	}
	gologger.Info().Msgf("场景创建成功！%s", uid)
	// 确认场景创建无误后,才会写入到配置文件中
	err = p.AddCase(c)
	err = p.SaveProject()
	if err != nil {
		gologger.Error().Msgf("项目配置保存失败！")
		return nil, err
	}
	RedcLog("创建成功 " + p.ProjectPath + uid + " " + CaseName)
	return c, nil
}

func (c *Case) TfApply() error {
	var err error
	err = TfApply(c.Path, c.Parameter...)
	if err != nil {
		return err
	}
	c.StateTime = time.Now().Format("2006-01-02 15:04:05")
	c.State = StateRunning
	return nil
}

func (c *Case) TfPlan() error {
	var err error
	err = TfPlan(c.Path, c.Parameter...)
	if err != nil {
		return err
	}
	c.StateTime = time.Now().Format("2006-01-02 15:04:05")
	c.State = StateCreated
	return nil
}

func (c *Case) TfDestroy() error {
	err := TfDestroy(c.Path, c.Parameter)
	if err != nil {
		gologger.Error().Msgf("场景销毁失败！%s", err.Error())
		return err
	}
	c.StateTime = time.Now().Format("2006-01-02 15:04:05")
	c.State = StateStopped
	return nil
}
func (c *Case) Remove() error {
	if c.State == StateRunning {
		return fmt.Errorf("场景正在运行中，请先停止场景后删除！")
	}
	c.StateTime = time.Now().Format("2006-01-02 15:04:05")
	err := os.RemoveAll(c.Path)
	if err != nil {
		return fmt.Errorf("删除场景文件失败！%s", err.Error())
	}
	return nil
}

// Stop 停止场景
func (c *Case) Stop() error {

	err := c.TfDestroy()
	if err != nil {
		return err
	}
	return nil
}

// Kill 强制销毁场景
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

// Change 重建场景
func (c *Case) Change() error {
	// 销毁场景，不删除项目
	if err := c.TfDestroy(); err != nil {
		return err
	}
	// 重建场景
	if err := c.TfApply(); err != nil {
		return err
	}
	return nil
}

func (c *Case) Status() error {
	TfStatus(c.Path)
	return nil
}

// humanDuration 计算时间差并返回 Docker 风格的字符串
// 例如: "Up 2 hours", "Up 5 minutes"
func humanDuration(t time.Time) string {
	duration := time.Since(t)
	seconds := int(duration.Seconds())

	switch {
	case seconds < 60:
		return fmt.Sprintf("%d seconds", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%d minutes", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%d hours", seconds/3600)
	default:
		return fmt.Sprintf("%d days", seconds/86400)
	}
}

// parseTime 将字符串时间转为 time.Time
func parseTime(timeStr string) time.Time {
	// 对应你代码中的 time.Now().Format("2006-01-02 15:04:05")
	layout := "2006-01-02 15:04:05"
	t, err := time.ParseInLocation(layout, timeStr, time.Local)
	if err != nil {
		return time.Now() // 解析失败则返回当前时间，避免 panic
	}
	return t
}
