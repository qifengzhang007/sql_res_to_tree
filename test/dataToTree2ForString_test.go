package test

import (
	"encoding/json"
	"fmt"
	"github.com/qifengzhang007/sql_res_to_tree"
	"testing"
)

// 单元测试示例2

// 模拟sql查询结果2,以 省份、城市无限级数据为例
type SqlCityListString struct {
	Id       string
	CityName string
	Fid      string
	Status   int
	Remark   string
}

// 查询数据示例(相当于gorm的Find、Scan 函数扫描结果)

//[
//{id:1 CityName:上海      Fid:0   Status:1 Remark:上海(一级节点)}
//	{id:2 CityName:上海市      Fid:1   Status:1 Remark:上海市(二级节点)}
//	  {id:3 CityName:徐汇区      Fid:2   Status:1 Remark:""}
//    {id:5 CityName:田林路      Fid:3   Status:1 Remark:"街道"}
//    {id:6 CityName:宜山路      Fid:3   Status:1 Remark:"街道"}

//	{id:4 CityName:松江区      Fid:2   Status:1 Remark:""}
//    {id:7 CityName:佘山      Fid:4   Status:1 Remark:""}
//    {id:8 CityName:泗泾镇      Fid:4   Status:1 Remark:""}

//    {id:9 CityName:河北省      Fid:0   Status:1 Remark:""}
//    {id:10 CityName:邯郸市      Fid:9   Status:1 Remark:"二级城市节点"}
//    {id:11 CityName:邯山区      Fid:10   Status:1 Remark:"市区划分"}
//    {id:12 CityName:复兴区      Fid:10   Status:1 Remark:"市区划分"}
//]

// 指定目标接受数据的切片，程序自动从sql查询结果切片中扫描填充数据
func TestScanString2(t *testing.T) {

	// 定义一个目标切片，用于接受最终的树形化数据
	//  模拟sql查询结果2
	type ProvinceCity struct {
		Id       string `primaryKey:"yes"`
		CityName string
		Fid      string `fid:"Id"`
		Status   int
		Children *[]ProvinceCity
	}
	var dest = make([]ProvinceCity, 0)
	// 模拟一份结构体切片格式的数据集(相当于gorm的sql函数 Scan Find的结果)
	// 测试无限层级的数据树形化（自己嵌套自己）
	in := mocStringData2()
	if err := sql_res_to_tree.CreateSqlResFormatFactory().ScanToTreeData(in, &dest); err == nil {
		bytes, _ := json.Marshal(dest)
		fmt.Printf("最终树形结果:\n%s\n", bytes)
	} else {
		t.Errorf("单元测试失败，错误：%s\n", err.Error())
	}
}

// 模拟一个多层次，无限嵌套的，拥有相同字段的结构体切片
func mocStringData2() []SqlCityListString {
	res := make([]SqlCityListString, 0)

	var tmp = SqlCityListString{
		"1",
		"上海",
		"0",
		1,
		"一级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"2",
		"上海市",
		"1",
		1,
		"上海市(二级节点)",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"3",
		"徐汇区",
		"2",
		1,
		"三级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"4",
		"松江区",
		"2",
		1,
		"三级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"5",
		"田林路",
		"3",
		1,
		"徐汇区，街道之一",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"6",
		"宜山路",
		"3",
		1,
		"徐汇区，街道之一",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"7",
		"佘山镇",
		"4",
		1,
		"松江区，镇之一",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"8",
		"泗泾镇",
		"4",
		1,
		"松江区，镇之一",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"9",
		"河北省",
		"0",
		1,
		"一级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"10",
		"邯郸市",
		"9",
		1,
		"二级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityListString{
		"11",
		"邯山区",
		"10",
		1,
		"市区划分",
	}
	res = append(res, tmp)
	tmp = SqlCityListString{
		"12",
		"复兴区",
		"10",
		1,
		"市区划分",
	}
	res = append(res, tmp)

	return res
}
