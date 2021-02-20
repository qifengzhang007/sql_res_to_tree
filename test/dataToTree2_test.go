package test

import (
	"encoding/json"
	"fmt"
	"sql_res_to_tree"
	"testing"
)

// 单元测试示例2

// 模拟sql查询结果2,以 省份、城市无限级数据为例
type SqlCityList struct {
	Id       int64
	CityName string
	Fid      int64
	Status   int
	Remark   string
}

// 查询数据示例(本质上就是 GormDbMysql.Raw("select * from  ...").Find(&receive))
// 以下结果本质上就是gorm返回的结果

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

//  指定目标接受数据的切片，程序自动从sql查询结果切片中扫描填充数据
func TestScanWay2(t *testing.T) {

	// 定义一个目标切片，用于接受最终的树形化数据
	//  模拟sql查询结果2
	type ProvinceCity struct {
		Id       int64 `primaryKey:"yes"`
		CityName string
		Fid      int64 `fid:"Id"`
		Status   int
		Children []ProvinceCity
	}
	var dest = make([]ProvinceCity, 0)
	in := mocData2()

	// 如果被处理的sql结果需要后续使用，那么必须复制一份被处理的数据
	// 因为经过本次处理以后，原始sql结果集相关的上下关联关系键值会被改变
	//var tmp = make([]SqlList, len(in))
	//copy(tmp, in) // 复制一份原始数据到 tmp 变量

	fmt.Printf("%+v\n", in)
	if err := sql_res_to_tree.CreateSqlResFormatFactory().ScanToTreeData(in, &dest); err == nil {
		bytes, _ := json.Marshal(dest)
		fmt.Printf("最终树形结果:%s\n", bytes)
	} else {
		t.Errorf("单元测试失败，错误：%s\n", err.Error())
	}

}

//  模拟一个具有多层次，但是每个结构体字段不同的结构体切片进行树形化

// 模拟一个多层次，无限嵌套的，拥有相同字段的结构体切片
func mocData2() []SqlCityList {
	res := make([]SqlCityList, 0)

	var tmp = SqlCityList{
		1,
		"上海",
		0,
		1,
		"一级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		2,
		"上海市",
		1,
		1,
		"上海市(二级节点)",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		3,
		"徐汇区",
		2,
		1,
		"三级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		4,
		"松江区",
		2,
		1,
		"三级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		5,
		"田林路",
		3,
		1,
		"徐汇区，街道之一",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		6,
		"宜山路",
		3,
		1,
		"徐汇区，街道之一",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		7,
		"佘山镇",
		4,
		1,
		"松江区，镇之一",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		8,
		"泗泾镇",
		4,
		1,
		"松江区，镇之一",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		9,
		"河北省",
		0,
		1,
		"一级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		10,
		"邯郸市",
		9,
		1,
		"二级节点",
	}
	res = append(res, tmp)

	tmp = SqlCityList{
		11,
		"邯郸区",
		10,
		1,
		"市区划分",
	}
	res = append(res, tmp)
	tmp = SqlCityList{
		12,
		"复兴区",
		10,
		1,
		"市区划分",
	}
	res = append(res, tmp)

	return res
}
