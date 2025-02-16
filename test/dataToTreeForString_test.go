package test

import (
	"encoding/json"
	"fmt"
	"github.com/qifengzhang007/sql_res_to_tree"
	"testing"
)

// 单元测试示例

// 模拟sql查询结果1
type SqlListString struct {
	SchoolId   int
	SchoolName string

	FkSchoolId int
	GradeId    string
	GradeName  string

	FkGradeId string
	ClassId   string
	ClassName string
}

// 查询数据示例（相当于gorm的Find、Scan 函数扫描结果）

//[
//{SchoolId:1 SchoolName:第一中学(高中)      FkSchoolId:1 GradeId:1 GradeName:高一           FkGradeId:1 ClassId:1 ClassName:文科班}
//{SchoolId:1 SchoolName:第一中学(高中)      FkSchoolId:1 GradeId:2 GradeName:高二           FkGradeId:2 ClassId:2 ClassName:理科班}
//{SchoolId:1 SchoolName:第一中学(高中)      FkSchoolId:1 GradeId:3 GradeName:高三           FkGradeId:3 ClassId:3 ClassName:实验班}

//{SchoolId:2 SchoolName:初级中学            FkSchoolId:2 GradeId:5 GradeName:初二           FkGradeId:5 ClassId:4 ClassName:普通班}
//{SchoolId:2 SchoolName:初级中学            FkSchoolId:2 GradeId:6 GradeName:初三           FkGradeId:6 ClassId:5 ClassName:实验班}
//{SchoolId:2 SchoolName:初级中学            FkSchoolId:2 GradeId:6 GradeName:初三           FkGradeId:6 ClassId:6 ClassName:中考冲刺班}
//]

// 指定目标接受数据的切片，程序自动从sql查询结果切片中扫描填充数据
func TestScanStringData(t *testing.T) {

	// 定义一个目标切片，用于接受最终的树形化数据
	type Stu struct {
		SchoolId   int    `primaryKey:"yes" json:"school_id"`
		SchoolName string `json:"school_name"`
		TestRemark string `json:"test_remark" default:"第一层结构体测试默认值"`
		Children   []struct {
			FkSchoolId int    `fid:"SchoolId"`
			GradeId    string `primaryKey:"yes"` //  本层的主键是string
			GradeName  string
			Children   []struct {
				FkGradeId string `fid:"GradeId"` //  本层外键需要和上级GradeId建立关联关系，那么数据类型必须相同
				ClassId   string `primaryKey:"yes"`
				ClassName string
				Remark    string `default:"为自定义字段使用default标签设置默认值"` //  允许目的变量中的字段可以在 sql 查询结果集中不存在，这样程序寻找default标签对应的值进行赋值，否则就是默认空值
				TestInt   int    `default:"100"`                    // default 标签支持 int  int16  int32  int64  string  bool
				TestBool  bool   `default:"true"`
			} `json:"children"`
		} `json:"children"`
	}
	var dest = make([]Stu, 0)

	// 模拟一份结构体切片格式的数据集(相当于gorm的sql函数 Scan Find的结果)
	// 测试有限层级的数据树形化
	in := mocStrData()

	if err := sql_res_to_tree.CreateSqlResFormatFactory().ScanToTreeData(in, &dest); err == nil {
		bytes, _ := json.Marshal(dest)
		fmt.Printf("树形化结果:\n%s\n", bytes)
	} else {
		t.Errorf("单元测试失败，错误：%s\n", err.Error())
	}
}

// 模拟一个具有多层次，但是每个结构体字段不同的结构体切片进行树形化
func mocStrData() []SqlListString {
	var demoList = make([]SqlListString, 0)

	var item = SqlListString{
		1,
		"第一中学(高中)",
		1,
		"1",
		"高一",
		"1",
		"1",
		"文科班",
	}
	demoList = append(demoList, item)
	//
	item = SqlListString{
		1,
		"第一中学(高中)",
		1,
		"2",
		"高二",
		"2",
		"2",
		"理科班",
	}
	demoList = append(demoList, item)

	item = SqlListString{
		1,
		"第一中学(高中)",
		1,
		"3",
		"高三",
		"3",
		"3",
		"实验班",
	}
	demoList = append(demoList, item)

	item = SqlListString{
		2,
		"初级中学",
		2,
		"5",
		"初二",
		"5",
		"4",
		"普通班",
	}
	demoList = append(demoList, item)

	item = SqlListString{
		2,
		"初级中学",
		2,
		"6",
		"初三",
		"6",
		"5",
		"实验班",
	}
	demoList = append(demoList, item)

	item = SqlListString{
		2,
		"初级中学",
		2,
		"6",
		"初三",
		"6",
		"6",
		"中考冲刺班",
	}
	demoList = append(demoList, item)

	item = SqlListString{
		2,
		"初级中学",
		2,
		"6",
		"初三",
		"6",
		"8",
		"普通班",
	}
	demoList = append(demoList, item)

	return demoList
}
