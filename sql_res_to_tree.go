package sql_res_to_tree

import (
	"errors"
	"reflect"
	"strconv"
)

func CreateSqlResFormatFactory() *sqlResFormatTree {
	return &sqlResFormatTree{primaryKey: make(map[string]string), counts: 0}
}

type sqlResFormatTree struct {
	primaryKey     map[string]string
	counts         int // 统计程序秭归计算的次数
	inSliceValueOf reflect.Value
	inSliceLen     int
}

// ScanToTreeData 是 ScanToTreeDataOld 的优化版本，代码更简洁，性能更好
func (s *sqlResFormatTree) ScanToTreeData(inSlice interface{}, destSlicePtr interface{}) (err error) {
	inValueOf := reflect.ValueOf(inSlice)
	if inValueOf.Kind() != reflect.Slice {
		return errors.New(inSliceErrMustValidSlice)
	}

	s.inSliceValueOf = inValueOf
	s.inSliceLen = inValueOf.Len()

	if s.inSliceLen == 0 {
		return errors.New(inSliceErrMustValidSlice)
	}

	destValueOf := reflect.ValueOf(destSlicePtr)
	if destValueOf.Kind() != reflect.Ptr || destValueOf.Elem().Kind() != reflect.Slice {
		return errors.New(destSlicePtrErrMustPtr)
	}

	destSlice := destValueOf.Elem()
	destElemType := destSlice.Type().Elem()

	primaryKeyName := s.getCurStructPrimaryKeyName(destElemType)
	if primaryKeyName == "" {
		return errors.New(structErrMustPrimaryKey)
	}

	processed := make([]bool, s.inSliceLen)
	roots := reflect.MakeSlice(destSlice.Type(), 0, 0)

	for i := 0; i < s.inSliceLen; i++ {
		if processed[i] {
			continue
		}
		row := inValueOf.Index(i)
		primaryKeyField := row.FieldByName(primaryKeyName)
		if primaryKeyField.IsZero() {
			continue
		}

		node, err := s.convertRowToNode(row, destElemType)
		if err != nil {
			return err
		}

		if err := s.attachChildren2(node, i, processed); err != nil {
			return err
		}

		roots = reflect.Append(roots, node)
		processed[i] = true
	}

	destSlice.Set(roots)
	return nil
}

func (s *sqlResFormatTree) storePrimaryKey(keyName string) {
	if s.primaryKey[keyName] != keyName {
		s.primaryKey[keyName] = keyName
	}
}

func (s *sqlResFormatTree) getCurStructPrimaryKeyName(rTypeOf reflect.Type) string {
	for i := 0; i < rTypeOf.NumField(); i++ {
		if rTypeOf.Field(i).Tag.Get("primaryKey") == "yes" {
			return rTypeOf.Field(i).Name
		}
	}
	return ""
}

func (s *sqlResFormatTree) destStructFieldIsExists(inSliceStruct reflect.Type, destFieldStructName string) bool {
	for i := 0; i < inSliceStruct.NumField(); i++ {
		if inSliceStruct.Field(i).Name == destFieldStructName {
			return true
		}
	}
	return false
}

func (s *sqlResFormatTree) destStructFieldIsSame(inSliceStruct reflect.Type, destFieldStruct reflect.StructField) bool {
	for i := 0; i < inSliceStruct.NumField(); i++ {
		if inSliceStruct.Field(i).Name == destFieldStruct.Name && inSliceStruct.Field(i).Type == destFieldStruct.Type {
			return true
		}
	}
	return false
}

func (s *sqlResFormatTree) analysisChildren(parentRowIndex int64, parentField reflect.Value, childrenType reflect.Type) (reflect.Value, error) {
	s.counts++
	resChildren := reflect.MakeSlice(childrenType, 0, 0)

	if s.counts > allowMaxRows {
		return resChildren, errors.New(overAllowMaxRows)
	}

	vType := childrenType.Elem()
	newStruct := reflect.New(vType)
	newTypeOf := newStruct.Elem().Type()
	newValueOf := newStruct.Elem()
	fieldNum := newTypeOf.NumField()

	parentKeyName := s.getCurStructParentKeyName(newTypeOf)
	if parentKeyName == "" {
		return reflect.Value{}, errors.New(structErrMustFid)
	}

	curStructPrimaryKeyName := s.getCurStructPrimaryKeyName(newTypeOf)
	if curStructPrimaryKeyName == "" {
		return reflect.Value{}, errors.New(structErrMustPrimaryKey)
	}

	if !s.destStructFieldIsExists(parentField.Type(), parentKeyName) {
		return reflect.Value{}, errors.New(destStructFidFieldNotExists + parentKeyName)
	}

	s.storePrimaryKey(curStructPrimaryKeyName)

	ParentDataType, err := s.curPrimaryKeyDataType(parentField, parentKeyName)
	if err != nil {
		return reflect.Value{}, err
	}

	switch ParentDataType {
	case 1:
		ParentIdInt := parentField.FieldByName(parentKeyName).Int()
		if ParentIdInt > 0 {
			for subRowIndex := int(parentRowIndex); subRowIndex < s.inSliceLen; subRowIndex++ {
				subRow := s.inSliceValueOf.Index(subRowIndex)
				subFKeyName := s.getCurStructSubFKeyName(newTypeOf)
				if subFKeyName == "" {
					return reflect.Value{}, errors.New(structErrMustFid + subFKeyName)
				}

				subFKeyField := subRow.FieldByName(subFKeyName)
				subPrimaryKeyName := s.getCurStructPrimaryKeyName(newTypeOf)
				if subPrimaryKeyName == "" {
					return reflect.Value{}, errors.New(structErrMustPrimaryKey)
				}

				subPrimaryKeyField := subRow.FieldByName(subPrimaryKeyName)

				if val, _ := s.curPrimaryKeyDataType(subRow, subFKeyName); val != 1 {
					return reflect.Value{}, errors.New(subKeyDataTypeIsNotIntError + subFKeyName)
				}

				subFKeyId := subFKeyField.Int()
				s.storePrimaryKey(subPrimaryKeyName)

				if subFKeyId > 0 && subFKeyId == ParentIdInt && !subPrimaryKeyField.IsZero() {
					if tmpChildren, err := s.getLevelGe2Children(fieldNum, resChildren, newTypeOf, parentRowIndex, subRowIndex, ParentIdInt, subFKeyName, subPrimaryKeyName, subRow, newValueOf); err != nil {
						return reflect.Value{}, err
					} else {
						resChildren = tmpChildren
					}
				}
			}
		}
	case 2:
		ParentIdStr := parentField.FieldByName(parentKeyName).String()
		if ParentIdStr != "" {
			for subRowIndex := int(parentRowIndex); subRowIndex < s.inSliceLen; subRowIndex++ {
				subRow := s.inSliceValueOf.Index(subRowIndex)
				subFKeyName := s.getCurStructSubFKeyName(newTypeOf)
				if subFKeyName == "" {
					return reflect.Value{}, errors.New(structErrMustFid + subFKeyName)
				}

				subFKeyField := subRow.FieldByName(subFKeyName)
				subPrimaryKeyName := s.getCurStructPrimaryKeyName(newTypeOf)
				if subPrimaryKeyName == "" {
					return reflect.Value{}, errors.New(structErrMustPrimaryKey)
				}

				subPrimaryKeyField := subRow.FieldByName(subPrimaryKeyName)

				if val, _ := s.curPrimaryKeyDataType(subRow, subFKeyName); val != 2 {
					return reflect.Value{}, errors.New(subKeyDataTypeIsNotStringError + subFKeyName)
				}

				subFKeyId := subFKeyField.String()
				s.storePrimaryKey(subPrimaryKeyName)

				if subFKeyId != "" && subFKeyId == ParentIdStr && !subPrimaryKeyField.IsZero() {
					if tmpChildren, err := s.getLevelGe2Children(fieldNum, resChildren, newTypeOf, parentRowIndex, subRowIndex, ParentIdStr, subFKeyName, subPrimaryKeyName, subRow, newValueOf); err != nil {
						return reflect.Value{}, err
					} else {
						resChildren = tmpChildren
					}
				}
			}
		}
	}

	return resChildren, nil
}

func (s *sqlResFormatTree) getLevelGe2Children(fieldNum int, resChildren reflect.Value, newTypeOf reflect.Type, parentRowIndex int64, subRowIndex int, ParentId interface{}, subFKeyName, subPrimaryKeyName string, subRow, newValueOf reflect.Value) (reflect.Value, error) {
	for j := 0; j < fieldNum; j++ {
		field := newTypeOf.Field(j)
		if field.Type.Kind() == reflect.Slice && field.Name == "Children" {
			if s.curItemHasSubLists(parentRowIndex, ParentId, subFKeyName) {
				if dataType, err := s.curPrimaryKeyDataType(subRow, subPrimaryKeyName); err == nil {
					switch dataType {
					case 1, 2:
						if val, err := s.analysisChildren(int64(subRowIndex), subRow, field.Type); err == nil {
							newValueOf.Field(j).Set(val)
						} else {
							return reflect.Value{}, err
						}
					}
				}
			} else {
				return resChildren, nil
			}
		} else if field.Type.Kind() == reflect.Ptr && field.Name == "Children" {
			if s.curItemHasSubLists(parentRowIndex, ParentId, subFKeyName) {
				if dataType, err := s.curPrimaryKeyDataType(subRow, subPrimaryKeyName); err == nil {
					switch dataType {
					case 1, 2:
						if val, err := s.analysisChildren(int64(subRowIndex), subRow, field.Type.Elem()); err == nil {
							tmpVal := reflect.New(val.Type())
							tmpVal.Elem().Set(val)
							newValueOf.Field(j).Set(tmpVal)
						} else {
							return reflect.Value{}, err
						}
					}
				}
			} else {
				return resChildren, nil
			}
		} else if s.destStructFieldIsExists(subRow.Type(), field.Name) {
			newValueOf.Field(j).Set(subRow.FieldByName(field.Name))
		} else if val, ok := s.setFieldDefaultValue(newTypeOf, field.Name); ok {
			newValueOf.Field(j).Set(val)
		}
	}
	if err := s.setUsedKeyInvalid(subRowIndex); err != nil {
		return reflect.Value{}, err
	}
	resChildren = reflect.Append(resChildren, newValueOf)
	return resChildren, nil
}

func (s *sqlResFormatTree) setUsedKeyInvalid(subRowIndex int) error {
	for primaryKeyName := range s.primaryKey {
		if s.inSliceValueOf.Index(subRowIndex).FieldByName(primaryKeyName).CanSet() {
			if dataType, err := s.curPrimaryKeyDataType(s.inSliceValueOf.Index(subRowIndex), primaryKeyName); err == nil {
				switch dataType {
				case 1:
					s.inSliceValueOf.Index(subRowIndex).FieldByName(primaryKeyName).SetInt(0)
				case 2:
					s.inSliceValueOf.Index(subRowIndex).FieldByName(primaryKeyName).SetString("")
				}
			} else {
				return errors.New(primaryKeyDataTypeError + primaryKeyName)
			}
		} else {
			return errors.New(structPrimaryKeyMustUpper)
		}
	}
	return nil
}

func (s *sqlResFormatTree) setFieldDefaultValue(fieldType reflect.Type, fieldName string) (reflect.Value, bool) {
	if f, ok := fieldType.FieldByName(fieldName); ok {
		if val, ok2 := f.Tag.Lookup("default"); ok2 {
			switch f.Type.Kind() {
			case reflect.String:
				return reflect.ValueOf(val), true
			case reflect.Float32:
				if tmp, err := strconv.ParseFloat(val, 32); err == nil {
					return reflect.ValueOf(tmp), true
				}
			case reflect.Float64:
				if tmp, err := strconv.ParseFloat(val, 64); err == nil {
					return reflect.ValueOf(tmp), true
				}
			case reflect.Int:
				if tmp, err := strconv.Atoi(val); err == nil {
					return reflect.ValueOf(tmp), true
				}
			case reflect.Int32:
				if tmp, err := strconv.ParseInt(val, 10, 32); err == nil {
					return reflect.ValueOf(tmp), true
				}
			case reflect.Int64:
				if tmp, err := strconv.ParseInt(val, 10, 64); err == nil {
					return reflect.ValueOf(tmp), true
				}
			case reflect.Bool:
				if tmp, err := strconv.ParseBool(val); err == nil {
					return reflect.ValueOf(tmp), true
				}
			default:
				break
			}
		}
	}
	return reflect.Value{}, false
}

func (s *sqlResFormatTree) curItemHasSubLists(curIndex int64, curMainId interface{}, subFKeyName string) bool {
	for i := int(curIndex); i < s.inSliceLen; i++ {
		tmpField := s.inSliceValueOf.Index(i)
		if pDataType, err := s.curPrimaryKeyDataType(tmpField, subFKeyName); err == nil {
			switch pDataType {
			case 1:
				if curMainId.(int64) == tmpField.FieldByName(subFKeyName).Int() {
					return true
				}
			case 2:
				if curMainId.(string) == tmpField.FieldByName(subFKeyName).String() {
					return true
				}
			}
		}
	}
	return false
}

func (s *sqlResFormatTree) curPrimaryKeyDataType(rValue reflect.Value, keyName string) (int, error) {
	switch rValue.FieldByName(keyName).Kind() {
	case reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return 1, nil
	case reflect.String:
		return 2, nil
	default:
		return 0, errors.New(primaryKeyDataTypeError + keyName)
	}
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

func (s *sqlResFormatTree) convertRowToNode(row reflect.Value, destType reflect.Type) (reflect.Value, error) {
	node := reflect.New(destType).Elem()
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		if field.Name == "Children" {
			continue
		}
		if srcField := row.FieldByName(field.Name); srcField.IsValid() && srcField.Type() == field.Type {
			node.FieldByName(field.Name).Set(srcField)
		} else if val, ok := s.setFieldDefaultValue(destType, field.Name); ok {
			node.FieldByName(field.Name).Set(val)
		}
	}
	return node, nil
}

func (s *sqlResFormatTree) attachChildren2(parent reflect.Value, parentIndex int, processed []bool) error {
	s.counts++
	if s.counts > allowMaxRows {
		return errors.New(overAllowMaxRows)
	}

	parentType := parent.Type()
	childrenField, hasChildren := parentType.FieldByName("Children")
	if !hasChildren {
		return nil
	}

	childrenType := childrenField.Type
	var childElemType reflect.Type
	var sliceType reflect.Type

	if childrenType.Kind() == reflect.Slice {
		childElemType = childrenType.Elem()
		sliceType = childrenType
	} else if childrenType.Kind() == reflect.Ptr && childrenType.Elem().Kind() == reflect.Slice {
		childElemType = childrenType.Elem().Elem()
		sliceType = childrenType.Elem()
	} else {
		return nil
	}

	childPrimaryKey := s.getCurStructPrimaryKeyName(childElemType)
	if childPrimaryKey == "" {
		return errors.New(structErrMustPrimaryKey)
	}

	parentKeyName := s.getCurStructParentKeyName(childElemType)
	if parentKeyName == "" {
		return errors.New(structErrMustFid)
	}

	// 获取子节点的外键字段名（fid标签所在的字段）
	subFKeyName := s.getCurStructSubFKeyName(childElemType)
	if subFKeyName == "" {
		return errors.New(structErrMustFid + " (找不到fid标签字段)")
	}

	parentKeyField := parent.FieldByName(parentKeyName)
	if !parentKeyField.IsValid() {
		return errors.New(destStructFidFieldNotExists + parentKeyName)
	}

	parentKeyVal := parentKeyField.Interface()
	children := reflect.MakeSlice(sliceType, 0, 0)

	for i := parentIndex; i < s.inSliceLen; i++ {
		if processed[i] {
			continue
		}
		row := s.inSliceValueOf.Index(i)
		childKeyField := row.FieldByName(subFKeyName) // 使用子外键字段名
		if !childKeyField.IsValid() || childKeyField.IsZero() {
			continue
		}

		// 检查是否找到自身（主键相同）
		childPrimaryKeyField := row.FieldByName(childPrimaryKey)
		if childPrimaryKeyField.IsValid() {
			parentPrimaryKeyField := parent.FieldByName(childPrimaryKey)
			if parentPrimaryKeyField.IsValid() && childPrimaryKeyField.Interface() == parentPrimaryKeyField.Interface() {
				// 找到自身，跳过
				continue
			}
		}

		// 直接比较interface{}值
		childKeyVal := childKeyField.Interface()
		if parentKeyVal != childKeyVal {
			continue
		}

		childNode, err := s.convertRowToNode(row, childElemType)
		if err != nil {
			return err
		}

		if err := s.attachChildren2(childNode, i, processed); err != nil {
			return err
		}

		processed[i] = true
		children = reflect.Append(children, childNode)
	}

	// 总是设置Children字段，即使为空
	if childrenType.Kind() == reflect.Ptr {
		ptrChildren := reflect.New(childrenType.Elem())
		ptrChildren.Elem().Set(children)
		parent.FieldByName("Children").Set(ptrChildren)
	} else {
		parent.FieldByName("Children").Set(children)
	}
	return nil
}

// valuesEqual2 比较两个反射值是否相等，支持整数和字符串类型
func (s *sqlResFormatTree) valuesEqual2(v1, v2 reflect.Value) bool {
	if !v1.IsValid() || !v2.IsValid() {
		return false
	}

	// 转换为interface{}比较
	val1 := v1.Interface()
	val2 := v2.Interface()

	switch v1.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v2.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflect.ValueOf(val1).Int() == reflect.ValueOf(val2).Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return reflect.ValueOf(val1).Int() == int64(reflect.ValueOf(val2).Uint())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch v2.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(reflect.ValueOf(val1).Uint()) == reflect.ValueOf(val2).Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return reflect.ValueOf(val1).Uint() == reflect.ValueOf(val2).Uint()
		}
	case reflect.String:
		if v2.Kind() == reflect.String {
			return reflect.ValueOf(val1).String() == reflect.ValueOf(val2).String()
		}
	}

	// 回退到简单比较
	return val1 == val2
}
