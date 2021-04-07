package test

import (
	"encoding/json"
	"fmt"
	"github.com/qifengzhang007/sql_res_to_tree"
	"testing"
)

// 单元测试示例3

// 模拟sql查询结果3,假设该结果是多张表联合查询结果（本次以3张表为例）
// 那么我们的 sql 查询语法按照如下格式编写
//  select  ***   form  **
//   union
//  select   *** from  ***
//   union
//  select  ***from  ***

// 每一个 select 段请使用类似  NodeType:dept   等字段标记该sql查询的类型
// 因为三段 sql 他们是联合查询(拼接),他们的主键可能会相同，用一个类似  NodeType:dept  标记这样在处理业务时，可以先锁定不同的业务类型，然后获取id键值

type SqlDeptMenuButton struct {
	Id       int64  `primaryKey:"yes"`
	OrgFid   int64  `fid:"Id"`
	OrgTitle string `json:"org_title"`
	NodeType string `unionPrimaryKey:"yes" json:"node_type"`
	Expand   int    `json:"expand"`
}

// 查询数据示例(相当于gorm的Find、Scan 函数扫描结果)
// NodeType  字段是为了数据隶属于不同的业务类型、字段名称可以是任何类型，处理业务时根据不同的业务类型获取id值
// 本包树形化时并不需要该字段

//[
//{id:1 OrgTitle:上海仪电数字技术股份有限公司      OrgFid:0   NodeType:dept  Expand:1 }
//{id:35 OrgTitle:信息化中心      OrgFid:1   NodeType:dept  Expand:1  }
//	  {id:36 CityName:超级管理员      OrgFid:35   NodeType:dept  Expand:1 }

//	  {id:238 CityName:系统配置      OrgFid:36   NodeType:menu  Expand:1 }
//	  {id:330 CityName:组织机构      OrgFid:328   NodeType:menu  Expand:0 }
//	  {id:333 CityName:用户管理      OrgFid:328   NodeType:menu  Expand:0 }

//	  {id:343 CityName:系统配置      OrgFid:1   NodeType:menu  Expand:1 }
//	  {id:344 CityName:公共权限      OrgFid:343   NodeType:menu  Expand:0 }
//	  {id:27 CityName:组织机构      OrgFid:344   NodeType:button  Expand:0 }
//	  {id:37 CityName:文件上传      OrgFid:344   NodeType:button  Expand:0 }

//	  {id:1 CityName:新增      OrgFid:330   NodeType:button  Expand:0 }
//	  {id:35 CityName:删除      OrgFid:330   NodeType:button  Expand:0 }
//	  {id:36 CityName:修改      OrgFid:330   NodeType:button  Expand:0 }
//	  {id:238 CityName:查询      OrgFid:330   NodeType:button  Expand:0 }
//]

//  指定目标接受数据的切片，程序自动从sql查询结果切片中扫描填充数据
func TestScanWay3(t *testing.T) {

	// 定义一个目标切片，用于接受最终的树形化数据
	// 结构体的定义类似 示例2 ，和示例2不同点在于sql查询语法上有差异
	type SqlDeptMenuButton struct {
		Id       int64               `primaryKey:"yes"`
		OrgTitle string              `json:"org_title"`
		OrgFid   int64               `fid:"Id" json:"org_fid"`
		NodeType string              `json:"node_type"`
		Expand   int                 `json:"expand"`
		Children []SqlDeptMenuButton `json:"children"`
	}
	var dest = make([]SqlDeptMenuButton, 0)
	// 模拟一份结构体切片格式的数据集(相当于gorm的sql函数 Scan Find的结果)
	// 测试无限层级的数据树形化（自己嵌套自己）
	in := mocData3()
	if err := sql_res_to_tree.CreateSqlResFormatFactory().ScanToTreeData(in, &dest); err == nil {
		bytes, _ := json.Marshal(dest)
		fmt.Printf("最终树形结果:\n%s\n", bytes)
	} else {
		t.Errorf("单元测试失败，错误：%s\n", err.Error())
	}
}

// 模拟一个多层次，无限嵌套的，拥有相同字段的结构体切片
func mocData3() []SqlDeptMenuButton {
	res := make([]SqlDeptMenuButton, 0)

	var tmp = SqlDeptMenuButton{
		1,
		0,
		"上海仪电数字技术股份有限公司",
		"dept",
		1,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		35,
		1,
		"信息化中心",
		"dept",
		1,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		36,
		35,
		"超级管理员",
		"dept",
		1,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		328,
		36,
		"系统配置",
		"menu",
		1,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		330,
		328,
		"组织机构",
		"menu",
		0,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		333,
		328,
		"用户管理",
		"menu",
		0,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		343,
		1,
		"系统配置",
		"menu",
		0,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		344,
		343,
		"公共权限",
		"menu",
		0,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		27,
		344,
		"组织机构",
		"button",
		0,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		37,
		344,
		"文件上传",
		"button",
		0,
	}
	res = append(res, tmp)

	tmp = SqlDeptMenuButton{
		1,
		330,
		"新增",
		"button",
		0,
	}
	res = append(res, tmp)
	tmp = SqlDeptMenuButton{
		35,
		330,
		"删除",
		"button",
		0,
	}
	res = append(res, tmp)

	return res
}
