package sql_res_to_tree

// 可能的错误常量
const (
	inSliceErrMustValidSlice       = "参数一(inSlice) 必须是一个不为空值的结构体切片"
	destSlicePtrErrMustPtr         = "参数二(destSlicePtr) 必须是一个指针"
	destSlicePtrErrMustSlice       = "参数二(destSlicePtr) 必须是一个结构体切片的指针"
	structErrMustPrimaryKey        = "每级结构体必须设置 primaryKey 标签，指定每一层结构体的主键"
	structErrMustFid               = "子结构体必须设置 fid 标签，指定的父级键名也必须存在,错误的键名："
	structPrimaryKeyMustUpper      = "结构体主键字段（primaryKey 标签所在键）必须以大写字母开头（扫描原始数据时本程序需要修改主键字段值的权限）"
	destStructFieldNotExists       = "参数二(destSlicePtr) 结构体定义的字段（包括类型）必须在 inSlice 参数传递的结构体中存在，否则无法获取值,请检查字段名称书写是否有误，错误字段名："
	destStructFidFieldNotExists    = "子结构体设置的 fid 标签对应的字段不存在于 inSlice 参数传递的结构体中存在,fid设置的错误字段名："
	allowMaxRows                   = 100000
	overAllowMaxRows               = "程序遍历次数已经超过了 100000 次,可能已经选入了死循环,请检查传入的数据是否符合要求,引起该错误的原因 primaryKey 标签字段和 fid 标签字段值不要出现互相嵌套，此外保证 primaryKey 标签键值唯一"
	primaryKeyDataTypeError        = "主键数据类型必须是 int系列(int、int16、int32、int64) 和 string 类型，其他类型不支持,错误发生的主键为： "
	subKeyDataTypeIsNotIntError    = "原始切片中, 主键、子健数据类型不一致，请确保子键数据类型为 int 系列, 相关子键(外键)："
	subKeyDataTypeIsNotStringError = "原始切片中, 主键、子健数据类型不一致，请确保子键数据类型为 string 系列, 相关子键(外键)："
)
