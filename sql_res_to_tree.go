package sql_res_to_tree

import (
	"errors"
	"reflect"
)

// 可能的错误常量
const (
	inSliceErrMustValidSlice  = "参数一(inSlice) 必须是一个不为空值的结构体切片"
	destSlicePtrErrMustPtr    = "参数二(destSlicePtr) 必须是一个指针"
	destSlicePtrErrMustSlice  = "参数二(destSlicePtr) 必须是一个结构体切片的指针"
	structErrMustPrimaryKey   = "结构体必须设置 primaryKey 标签，指定每一层结构体的主键"
	structErrMustFid          = "子结构体必须设置 fid 标签，指定父级键名"
	structPrimaryKeyMustUpper = "结构体主键字段（primaryKey 标签所在键）必须以大写字母开头（扫描原始数据时本程序需要修改主键字段值的权限）"
)

func CreateSqlResFormatFactory() *sqlResFormatTree {
	return &sqlResFormatTree{make(map[string]string)}
}

type sqlResFormatTree struct {
	primaryKey map[string]string
}

func (s *sqlResFormatTree) ScanToTreeData(inSlice interface{}, destSlicePtr interface{}) (err error) {

	inTypeOf := reflect.TypeOf(inSlice)
	if inTypeOf.Kind() != reflect.Slice {
		return errors.New(inSliceErrMustValidSlice)
	}

	inValueOf := reflect.ValueOf(inSlice)

	inLen := inValueOf.Len()
	if inLen == 0 {
		return errors.New(inSliceErrMustValidSlice)
	}

	destTypeOf := reflect.TypeOf(destSlicePtr)
	if destTypeOf.Kind() != reflect.Ptr {
		return errors.New(destSlicePtrErrMustPtr)
	}

	valueOf := reflect.ValueOf(destSlicePtr).Elem()

	if valueOf.Type().Kind() != reflect.Slice {
		return errors.New(destSlicePtrErrMustSlice)
	}

	originSlice := reflect.MakeSlice(valueOf.Type(), 0, 0)

	// type 获取的 slice，继续获取内部的结构体单元
	structElem := valueOf.Type().Elem()
	primaryKeyName := s.getCurStructPrimaryKeyName(structElem)
	if primaryKeyName == "" {
		return errors.New(structErrMustPrimaryKey)
	} else {
		s.storeprimaryKey(primaryKeyName)
	}
	// 返回结构体指针
	TmpOriginStruct := reflect.New(structElem)

	structElemTypeOf := TmpOriginStruct.Type().Elem() //  相当于  typeOf
	structElemValueOf := TmpOriginStruct.Elem()       //  相当于  valueOf
	//分析第一层结构体字段
	fieldNum := structElemTypeOf.NumField()

	//遍历sql查询结果集的行
	for rowIndex := 0; rowIndex < inLen; rowIndex++ {
		row := inValueOf.Index(rowIndex) //  struct
		mainKeyField := row.FieldByName(primaryKeyName)
		mainId := mainKeyField.Int()
		if mainId > 0 {
			for i := 0; i < fieldNum; i++ {
				if structElem.Field(i).Name == "Children" && structElem.Field(i).Type.Kind() == reflect.Slice {
					if val, err := s.analysisChildren(row, inValueOf, structElem.Field(i).Type); err == nil {
						structElemValueOf.Field(i).Set(val)
					} else {
						return err
					}
				} else {
					structElemValueOf.Field(i).Set(row.FieldByName(structElem.Field(i).Name))
				}
			}
			originSlice = reflect.Append(originSlice, structElemValueOf)
		}
	}
	valueOf.Set(originSlice)
	return nil
}

func (s *sqlResFormatTree) storeprimaryKey(keyName string) {
	if s.primaryKey[keyName] != keyName {
		s.primaryKey[keyName] = keyName
	}
}

// 设置已经遍历过的所有主键失效
func (s *sqlResFormatTree) setUsedKeyInvalid(rValue reflect.Value) error {
	for key := range s.primaryKey {
		if rValue.FieldByName(key).CanSet() {
			rValue.FieldByName(key).SetInt(-1)

		} else {
			return errors.New(structPrimaryKeyMustUpper)
		}
	}
	return nil
}

//  获取正在分析(处理)的结构体主键键名（unique）
func (s *sqlResFormatTree) getCurStructPrimaryKeyName(value reflect.Type) string {
	numField := value.NumField()
	for i := 0; i < numField; i++ {
		if value.Field(i).Tag.Get("primaryKey") == "yes" {
			return value.Field(i).Name
		}
	}
	return ""
}

// 获取当前结构体对应的父级键名（父子关系键中的父级键,也就是 fid 标签的值）
func (s *sqlResFormatTree) getCurStructParentKeyName(value reflect.Type) string {
	numField := value.NumField()
	for i := 0; i < numField; i++ {
		parentKey := value.Field(i).Tag.Get("fid")
		if parentKey != "" {
			return parentKey
		}
	}
	return ""
}

// 获取当前结构体的 子外键名 （父子关系键中的子外键键,也就是 fid 标签所在的字段名称）
func (s *sqlResFormatTree) getCurStructSubFKeyName(value reflect.Type) string {
	numField := value.NumField()
	for i := 0; i < numField; i++ {
		if _, ok := value.Field(i).Tag.Lookup("fid"); ok {
			return value.Field(i).Name
		}
	}
	return ""
}

// 继续分析 children 结构体
func (s *sqlResFormatTree) analysisChildren(parentField reflect.Value, inSlice reflect.Value, children reflect.Type) (reflect.Value, error) {
	resChildren := reflect.MakeSlice(children, 0, 0)
	vType := children.Elem()
	newStruct := reflect.New(vType)

	// 分析新结构体的每个字段
	newTypeOf := newStruct.Type().Elem()
	newValueOf := newStruct.Elem()
	fieldNum := newTypeOf.NumField()

	inLen := inSlice.Len()

	parentKeyName := s.getCurStructParentKeyName(newTypeOf)
	if parentKeyName == "" {
		return reflect.Value{}, errors.New(structErrMustFid)
	}

	parentPrimaryKeyName := s.getCurStructPrimaryKeyName(newTypeOf)
	if parentPrimaryKeyName == "" {
		return reflect.Value{}, errors.New(structErrMustPrimaryKey)
	}
	mainId := parentField.FieldByName(parentKeyName).Int()

	s.storeprimaryKey(parentPrimaryKeyName)

	if mainId > 0 {
		for subRowIndex := 0; subRowIndex < inLen; subRowIndex++ {
			subRow := inSlice.Index(subRowIndex)
			subKeyName := s.getCurStructSubFKeyName(newTypeOf)
			subKeyField := subRow.FieldByName(subKeyName)

			subPrimaryKeyName := s.getCurStructPrimaryKeyName(newTypeOf)
			subprimaryKeyField := subRow.FieldByName(subPrimaryKeyName)
			subKeyId := subKeyField.Int()

			s.storeprimaryKey(subPrimaryKeyName)

			if subKeyId > 0 && subKeyId == mainId && subprimaryKeyField.Int() > 0 {
				for j := 0; j < fieldNum; j++ {
					if newTypeOf.Field(j).Type.Kind() == reflect.Slice && newTypeOf.Field(j).Name == "Children" {
						if val, err := s.analysisChildren(subRow, inSlice, newTypeOf.Field(j).Type); err == nil {
							newValueOf.Field(j).Set(val)
						} else {
							newValueOf.Field(j).Set(reflect.Value{})
							return reflect.Value{}, err
						}
					} else {
						newValueOf.Field(j).Set(subRow.FieldByName(newTypeOf.Field(j).Name))
					}
				}
				if err := s.setUsedKeyInvalid(inSlice.Index(subRowIndex)); err != nil {
					return reflect.Value{}, err
				}
				resChildren = reflect.Append(resChildren, newValueOf)
			}
		}
	}
	return resChildren, nil
}
