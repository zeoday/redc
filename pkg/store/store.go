package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"red-cloud/mod"
	"red-cloud/proto"

	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

const (
	DBPath            = "redc.db"
	BucketProjectMeta = "Project_Meta"
)

// getCaseBucketName 获取存放特定项目 Case 的桶名
func getCaseBucketName(projectName string) []byte {
	return []byte(fmt.Sprintf("Cases_%s", projectName))
}

// execute 统一的带超时事务执行器
func execute(fn func(tx *bolt.Tx) error) error {
	// 设置 5 秒超时，解决多进程并发冲突
	opts := &bolt.Options{Timeout: 5 * time.Second}
	db, err := bolt.Open(DBPath, 0600, opts)
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return fmt.Errorf("保存配置失败！")
		}
		return err
	}
	defer db.Close()
	return db.Update(fn)
}

// ==========================================
// 转换逻辑 (Mapper)
// ==========================================

// ToProto  Domain(mod.Case) -> Storage(pb.Case)
func ToProto(c *mod.Case) *pb.Case {
	// 处理 output map，序列化存入
	outputBytes, _ := json.Marshal(c.Output)

	return &pb.Case{
		Id:         c.Id,
		Name:       c.Name,
		Type:       c.Type,
		Module:     c.Module,
		Operator:   c.Operator,
		Path:       c.Path,
		Node:       int32(c.Node),
		CreateTime: c.CreateTime,
		StateTime:  c.StateTime,
		Parameter:  c.Parameter,
		// 假设 mod.CaseState 底层是 string，这里需要一个映射转换
		// 如果 mod.CaseState 是 string，这里需要手动 switch case 转 enum
		// 这里简化演示假设你改造成了 int 或者做好了 map 映射
		State:      pb.CaseState(convertStateStringToInt(c.State)),
		OutputJson: string(outputBytes),
	}
}

// FromProto Storage(pb.Case) -> Domain(mod.Case)
func FromProto(p *pb.Case) *mod.Case {
	c := &mod.Case{
		Id:         p.Id,
		Name:       p.Name,
		Type:       p.Type,
		Module:     p.Module,
		Operator:   p.Operator,
		Path:       p.Path,
		Node:       int(p.Node),
		CreateTime: p.CreateTime,
		StateTime:  p.StateTime,
		Parameter:  p.Parameter,
		State:      convertStateEnumToString(p.State),
	}

	// 还原 Output
	if len(p.OutputJson) > 0 {
		// 注意：这里需要 mod 包暴露 Output 字段，或者你通过 SetOutput 方法赋值
		json.Unmarshal([]byte(p.OutputJson), &c.Output)
	}

	// 绑定运行时函数 (saveHandler 等)
	// 这一步通常在业务层调用更合适，或者这里不处理，由业务层 Load 完数据后统一 Bind
	return c
}

// 辅助转换函数 (你需要根据你的实际 State 定义来实现)
func convertStateStringToInt(s mod.CaseState) pb.CaseState {
	switch s {
	case mod.StateRunning:
		return pb.CaseState_RUNNING
	case mod.StateStopped:
		return pb.CaseState_STOPPED
	default:
		return pb.CaseState_UNKNOWN
	}
}

func convertStateEnumToString(s pb.CaseState) mod.CaseState {
	switch s {
	case pb.CaseState_RUNNING:
		return mod.StateRunning
	case pb.CaseState_STOPPED:
		return mod.StateStopped
	default:
		return mod.StateUnknown
	}
}

// ==========================================
// 数据库操作 (对外接口)
// ==========================================

// SaveCase 保存单个 Case
func SaveCase(projectName string, c *mod.Case) error {
	return execute(func(tx *bolt.Tx) error {
		// 1. 确保桶存在
		b, err := tx.CreateBucketIfNotExists(getCaseBucketName(projectName))
		if err != nil {
			return err
		}

		// 2. 转换
		pbData := ToProto(c)

		// 3. 序列化
		data, err := proto.Marshal(pbData)
		if err != nil {
			return err
		}

		// 4. 存入
		return b.Put([]byte(c.Id), data)
	})
}

// GetCase 读取单个 Case
func GetCase(projectName string, caseId string) (*mod.Case, error) {
	var p pb.Case
	err := execute(func(tx *bolt.Tx) error {
		b := tx.Bucket(getCaseBucketName(projectName))
		if b == nil {
			return fmt.Errorf("项目不存在")
		}

		data := b.Get([]byte(caseId))
		if data == nil {
			return fmt.Errorf("未找到 Case")
		}

		return proto.Unmarshal(data, &p)
	})
	if err != nil {
		return nil, err
	}

	return FromProto(&p), nil
}

// DeleteCase 删除 Case
func DeleteCase(projectName string, caseId string) error {
	return execute(func(tx *bolt.Tx) error {
		b := tx.Bucket(getCaseBucketName(projectName))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(caseId))
	})
}

// ListCases 获取项目下所有 Case
func ListCases(projectName string) ([]*mod.Case, error) {
	var cases []*mod.Case
	err := execute(func(tx *bolt.Tx) error {
		b := tx.Bucket(getCaseBucketName(projectName))
		if b == nil {
			return nil
		}

		return b.ForEach(func(k, v []byte) error {
			var p pb.Case
			if err := proto.Unmarshal(v, &p); err == nil {
				cases = append(cases, FromProto(&p))
			}
			return nil
		})
	})
	return cases, err
}

// SaveProjectMeta 保存项目元数据
func SaveProjectMeta(p *mod.RedcProject) error {
	return execute(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(BucketProjectMeta))

		pbProj := &pb.Project{
			ProjectName: p.ProjectName,
			ProjectPath: p.ProjectPath,
			CreateTime:  p.CreateTime,
			User:        p.User,
		}
		data, _ := proto.Marshal(pbProj)

		// 同时创建该项目的 Case 桶
		tx.CreateBucketIfNotExists(getCaseBucketName(p.ProjectName))

		return b.Put([]byte(p.ProjectName), data)
	})
}

// GetProjectMeta 读取项目元数据 (不含 Cases)
func GetProjectMeta(name string) (*mod.RedcProject, error) {
	var p pb.Project
	err := execute(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketProjectMeta))
		if b == nil {
			return fmt.Errorf("无项目数据")
		}
		data := b.Get([]byte(name))
		if data == nil {
			return fmt.Errorf("项目不存在")
		}
		return proto.Unmarshal(data, &p)
	})

	if err != nil {
		return nil, err
	}

	return &mod.RedcProject{
		ProjectName: p.ProjectName,
		ProjectPath: p.ProjectPath,
		CreateTime:  p.CreateTime,
		User:        p.User,
		// Case 列表需要在业务层调用 ListCases 单独填充
	}, nil
}
