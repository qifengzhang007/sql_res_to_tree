package sql_res_to_tree

//
//// ScanToTreeDataOld  为原函数 ScanToTreeData 的备份
//func (s *sqlResFormatTree) ScanToTreeDataOld(inSlice interface{}, destSlicePtr interface{}) (err error) {
//	inValueOf := reflect.ValueOf(inSlice)
//	if inValueOf.Kind() != reflect.Slice {
//		return errors.New(inSliceErrMustValidSlice)
//	}
//
//	s.inSliceValueOf = inValueOf // sql原始值的 valueOf 存储起来
//	s.inSliceLen = inValueOf.Len()
//
//	if s.inSliceLen == 0 {
//		return errors.New(inSliceErrMustValidSlice)
//	}
//
//	destValueOf := reflect.ValueOf(destSlicePtr)
//	if destValueOf.Kind() != reflect.Ptr || destValueOf.Elem().Kind() != reflect.Slice {
//		return errors.New(destSlicePtrErrMustPtr)
//	}
//
//	destSlice := destValueOf.Elem()
//	destTmpSlice := reflect.MakeSlice(destSlice.Type(), 0, 0)
//
//	destStructElem := destSlice.Type().Elem()
//	primaryKeyName := s.getCurStructPrimaryKeyName(destStructElem)
//	if primaryKeyName == "" {
//		return errors.New(structErrMustPrimaryKey)
//	}
//	s.storePrimaryKey(primaryKeyName)
//
//	tmpDestStructElem := reflect.New(destStructElem)
//	structElemTypeOf := tmpDestStructElem.Elem().Type()
//	structElemValueOf := tmpDestStructElem.Elem()
//	fieldNum := structElemTypeOf.NumField()
//
//	for rowIndex := 0; rowIndex < s.inSliceLen; rowIndex++ {
//		s.counts++
//		row := inValueOf.Index(rowIndex)
//		if !s.destStructFieldIsExists(row.Type(), primaryKeyName) {
//			return errors.New(destStructFieldNotExists + primaryKeyName)
//		}
//
//		primaryKeyDataType, err := s.curPrimaryKeyDataType(row, primaryKeyName)
//		if err != nil {
//			return err
//		}
//
//		mainKeyField := row.FieldByName(primaryKeyName)
//		var primaryKeyIdInterf interface{}
//
//		switch primaryKeyDataType {
//		case 1:
//			if primaryKeyIdInt := mainKeyField.Int(); primaryKeyIdInt > 0 {
//				primaryKeyIdInterf = primaryKeyIdInt
//			} else {
//				continue
//			}
//		case 2:
//			if primaryKeyIdStr := mainKeyField.String(); strings.TrimSpace(primaryKeyIdStr) != "" {
//				primaryKeyIdInterf = primaryKeyIdStr
//			} else {
//				continue
//			}
//		}
//
//		if primaryKeyIdInterf != nil {
//			for i := 0; i < fieldNum; i++ {
//				field := destStructElem.Field(i)
//				if field.Name == "Children" {
//					if field.Type.Kind() == reflect.Slice {
//						if val, err := s.analysisChildren(int64(rowIndex), row, field.Type); err == nil {
//							structElemValueOf.Field(i).Set(val)
//						} else {
//							return err
//						}
//					} else if field.Type.Kind() == reflect.Ptr {
//						if val, err := s.analysisChildren(int64(rowIndex), row, field.Type.Elem()); err == nil {
//							tmpVal := reflect.New(val.Type())
//							tmpVal.Elem().Set(val)
//							structElemValueOf.Field(i).Set(tmpVal)
//						} else {
//							return err
//						}
//					}
//				} else if s.destStructFieldIsSame(row.Type(), field) {
//					structElemValueOf.Field(i).Set(row.FieldByName(field.Name))
//				} else if val, ok := s.setFieldDefaultValue(structElemTypeOf, field.Name); ok {
//					structElemValueOf.Field(i).Set(val)
//				}
//			}
//			destTmpSlice = reflect.Append(destTmpSlice, structElemValueOf)
//		}
//	}
//	destSlice.Set(destTmpSlice)
//	return nil
//}
