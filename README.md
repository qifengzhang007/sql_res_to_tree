## SQL 查询结果快速树形化  

### 1.前言  
>   1.针对sql查询结果(本质上就是一个切片),进行树形化,支持有限级、无限级深度、单张树形表查询结果、多张树形表查询结果的联合(union)结果 数据进行树形化.      
>   2.本包搭配 gorm sql查询结果扫描函数 Scan 、 Find ,将获取的结果直接传递给本包，可以非常方便快捷地进行数据的树形化.  
>   3.关于性能，在我们测试的有限条sql查询结果树形化,耗费时间一直是0毫秒.  

###  2.核心树形化扫描函数    
>  1.核心函数，只有一个 ScanToTreeData(inSqlSlice, &dest), 其中 inSqlSlice 可以直接传递 gorm 的 sql 扫描函数 Find 、Scan获取的结果, dest 参数后续介绍.    
>  2.使用非常简单,就跟 gorm 的 Scan、Find函数类似，定义一个接受树形结果的结构体切片，传入地址,坐等扫描结果.  

### 3.集成到任何项目
> <font color=red> 1.注意: 请不要使用 v1.0.10 、v1.0.11 这两个过渡版本，永远推荐使用最新版本.</font>
```code  
# 安装前请自行在tag标签查询最新版本，本次我们以 v1.0.13为例

# 安装此包
go   get  github.com/qifengzhang007/sql_res_to_tree@v1.0.13

#调用sql结果树形化扫描函数， &dest  为接受树形结果的结构体切片，定义方式参考相关的单元测试示例代码  
sql_res_to_tree.CreateSqlResFormatFactory().ScanToTreeData(inSqlSlice, &dest);

```

### 4.核心转换原理
> 1.数据需要严谨的排序，从左到右，左侧的数据范围必须包括右边的子级数据，如果数据是乱序的，本包查询子级是基于左侧一条数据，依次向下提取子级的，最终会导致漏掉某些数据.    
> 2.以上核心思想总结就是一句话：左侧到右侧，数据范围依次缩小，相同性质的数据必须放在一起.    
- 4.1 数据库sql查询结果  
  ![转换逻辑1](https://www.ginskeleton.com/images/sql0.png)
- 4.2 gorm函数Scan、Find等获取的go切片原始数据  
  ![转换逻辑2](https://www.ginskeleton.com/images/sql1.png)
- 4.3 原始数据转换为树形数据的过程逻辑    
  ![转换逻辑2](https://www.ginskeleton.com/images/process2.png)

###  5.sql结果转换为树形化结果的主要场景  
1. [sql结果有限级且支持个性化设置子结构体字段树形化](./test/dataToTree_test.go)  
2. [单张表sql结果无限级树形化](./test/dataToTree2_test.go) ,单张树形表，例如：省份城市树形表,特点默认情况下是id是唯一的.
3. [多张表sql结果无限级树形化](./test/dataToTree3_test.go) ,多张表树形表，例如：A表是树形表、B表是树形表，B表是A表的子表（B挂接在A表底下），这种关系下 A、B 两张表ID是存在重复值的,我们依旧可以通过变换实现树形化.   
4. 主要思路如下(多张树形关系表转换为一张树形表关系,处理语法就和单张树形表一致了)：

![转换1](https://www.ginskeleton.com/images/tree_conv1.png)
![转换2](https://www.ginskeleton.com/images/tree_conv2.png)

###  6.场景1：平行数据，每一块数据通过外键关联  
- 更多的示例代码您可以直接参考、复制单元测试目录内的示例代码，[查看详更多示例代码](./test/)  
#### 6.1.有限层级的数据,支持每一层拥有不同字段的结构体树形化     
[详细实现过程](./test/dataToTree_test.go)  
- 注意细节：SchoolId、GradeId等每一级的主键首字母必须是大写的（允许本包进行修改字段值），一般来说gorm查询的结果都是符合此条件的，
- 如果是手动模拟输入以下数据，则必须要注意此项。  
```code   
//原始数据如下(gorm的 Find、 Scan 函数扫描结果都符合以下结构)：
[
	{SchoolId:1 SchoolName:第一中学(高中) FkSchoolId:1 GradeId:1 GradeName:高一 FkGradeId:1 ClassId:1 ClassName:文科班} 
	{SchoolId:1 SchoolName:第一中学(高中) FkSchoolId:1 GradeId:2 GradeName:高二 FkGradeId:2 ClassId:2 ClassName:理科班} 
	{SchoolId:1 SchoolName:第一中学(高中) FkSchoolId:1 GradeId:3 GradeName:高三 FkGradeId:3 ClassId:3 ClassName:实验班} 

	{SchoolId:2 SchoolName:初级中学 FkSchoolId:2 GradeId:5 GradeName:初二 FkGradeId:5 ClassId:4 ClassName:普通班}

	{SchoolId:2 SchoolName:初级中学 FkSchoolId:2 GradeId:6 GradeName:初三 FkGradeId:6 ClassId:5 ClassName:实验班} 
	{SchoolId:2 SchoolName:初级中学 FkSchoolId:2 GradeId:6 GradeName:初三 FkGradeId:6 ClassId:6 ClassName:中考冲刺班}
]
```
####  6.2 使用本包函数 ScanToTreeData(inSqlSlice, &dest),直接将 dest 变量转换为树形化结果：
- 使用非常简单，本包只有三个语法关键词
- 1 `primaryKey:"yes"` 定义每一级的主键
- 2 `fid:"父级键名"` 在子级特定的键上指定此标签，这样就把该键和父级结构体中的键建立了绑定(关联)关系
- 3 `default:"默认值"` 定义的dest结构体中的字段如果在被扫描的sql结果集中不存在，那么使用该默认值填充. 
- 4 `Children` 结构体挂载下一级的字段(成员)，必须使用 `Children` 单词,您可以通过 `json` 修改为其他单词。 

```code
	
	// 接受树形结果的结构体要求如下：
	// 1.主键必须使用 primaryKey:"yes" 标签定义，类型必须是  int  int64 in32 等int系列，以及 string 类型，不能是其他类型
	// 2.子结构体关联父级结构体的键必须定义 `fid:"父级键名"`  标签，父子关联键数据类型必须相同，要么都是 int 系列，要么都是 string
	
	// 定义一个目标切片，用于接受最终的树形化数据
	type Stu struct {
		SchoolId   int    `primaryKey:"yes" json:"school_id"`
		SchoolName string `json:"school_name"`
		Children   []struct {
			FkSchoolId int `fid:"SchoolId"`  // 这里FkSchoolId的值就会和父级SchoolId的值建立关联关系，他们的数据类型必须一致
			GradeId    int `primaryKey:"yes"`
			GradeName  string
			Children   []struct {
				FkGradeId int `fid:"GradeId"`  // 这里FkGradeId的值就会和父级GradeId的值建立关联关系，他们的数据类型必须一致
				ClassId   int `primaryKey:"yes"`
				ClassName string
				Remark    string `default:"为自定义字段使用default标签设置默认值"` //  允许目的变量中的字段可以在 sql 查询结果集中不存在，这样程序寻找default标签对应的值进行赋值，否则就是默认空值
				TestInt   int    `default:"100"`   // default 标签只支持为 int  int16  int32  int64  string  bool设置默认值
			} `json:"children"`
		} `json:"children"`
	}
	var dest = make([]Stu, 0)

    // inSqlSlice 表示以上sql查询的结构体切片结果
    sql_res_to_tree.CreateSqlResFormatFactory().ScanToTreeData(inSqlSlice, &dest);

```
- 6.3  最终将 dest 变量使用 json.Marshal函数 json化  
#### 效果图1  
![效果图1](https://www.ginskeleton.com/images/tree1.jpg)  


###  7.场景2：无限层级的数据树形
[详细实现过程](./test/dataToTree2_test.go)    
- 注意细节：Id 作为每一级的主键首字母必须是大写的（允许本包进行修改字段值），一般来说gorm查询的结果都是符合此条件的，
- 如果是手动模拟输入以下数据，则必须要注意此项。
```code   
//原始数据如下(gorm的 Find、 Scan 函数扫描结果都符合以下结构)：  
[
{Id:1 CityName:上海      Fid:0   Satatus:1 Remark:上海(一级节点)}
    {Id:2 CityName:上海市      Fid:1   Satatus:1 Remark:上海市(二级节点)}
	  {Id:3 CityName:徐汇区      Fid:2   Satatus:1 Remark:""}
	    {Id:5 CityName:田林路      Fid:3   Satatus:1 Remark:"街道"}
	    {Id:6 CityName:宜山路      Fid:3   Satatus:1 Remark:"街道"}

	{Id:4 CityName:松江区      Fid:2   Satatus:1 Remark:""}
	    {Id:7 CityName:佘山      Fid:4   Satatus:1 Remark:""}
	    {Id:8 CityName:泗泾镇      Fid:4   Satatus:1 Remark:""}

    {Id:9 CityName:河北省      Fid:0   Satatus:1 Remark:""}
	    {Id:10 CityName:邯郸市      Fid:9   Satatus:1 Remark:"二级城市节点"}
	      {Id:11 CityName:邯山区      Fid:10   Satatus:1 Remark:"市区划分"}
	      {Id:12 CityName:复兴区      Fid:10   Satatus:1 Remark:"市区划分"}
]
```

- 7.1 使用本包函数 ScanToTreeData(inSqlSlice, &dest),直接将 dest 变量json化结果： 
```code
	// 接受树形结果的结构体要求如下：
	// 1.主键必须使用 primaryKey:"yes" 标签定义，类型必须是  int  int64 in32 等int系列，以及 string 类型，不能是其他类型
	// 2.子结构体关联父级结构体的键必须定义 `fid:"父级主键"`  标签，父子关联键数据类型必须都是 int  int32  int64 等int系列和 string 类型
	
	// 定义一个目标切片，格式为树形结构，结构体自己嵌套自己，用于接受最终的树形化数据
	type ProvinceCity struct {
		Id       int64 `primaryKey:"yes"`
		CityName string
		Fid      int64 `fid:"Id"`
		Status   int
		Children []ProvinceCity
	}
	var dest = make([]ProvinceCity, 0)

    // inSqlSlice 表示以上sql查询的结构体切片结果
    sql_res_to_tree.CreateSqlResFormatFactory().ScanToTreeData(inSqlSlice, &dest);

```
- 7.2  最终将 dest 变量使用 json.Marshal函数 json化  
#### 效果图2  
![效果图2](https://www.ginskeleton.com/images/tree2.jpg)  

