package sgo

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"
)

type MapperFunc func([]reflect.Value) []reflect.Value

// Mapper 创建 映射函数
func (build *Build) mapper(id []string, result []reflect.Value) MapperFunc {
	return func(values []reflect.Value) []reflect.Value {
		var value, resultType, errType, Query, Exec reflect.Value
		db := build.db
		length := len(values)
		if length > 1 && values[length-1].Type().AssignableTo(db.Type()) {
			db.Set(values[length-1])
		}
		star := time.Now()
		ctx := values[0].Interface()
		statements, tag, templateSql, params, err := build.Get(id, ctx)
		if err != nil {
			errType = reflect.ValueOf(err)
			goto end
		}
		switch tag {
		case Select:
			Query = db.MethodByName("Query")
			call := Query.CallSlice([]reflect.Value{
				reflect.ValueOf(templateSql),
				reflect.ValueOf(params),
			})
			if !call[1].IsZero() {
				errType = call[1]
				goto end
			}

			if result[0].Kind() == reflect.Slice {
				resultType = reflect.New(result[0].Type().Elem()).Elem()
			} else {
				resultType = result[0]
			}
			value, errType = resultMapping(call[0], resultType.Interface())
			if !errType.IsZero() {
				goto end
			}
			QueryResultMapper(value, result)
		case Insert, Update, Delete:
			Exec = db.MethodByName("Exec")
			call := Exec.CallSlice([]reflect.Value{
				reflect.ValueOf(templateSql),
				reflect.ValueOf(params),
			})
			if !call[1].IsZero() {
				errType = call[1]
				goto end
			}
			err = ExecResultMapper(result, call[0].Interface().(sql.Result))
		}
	end:
		if (len(result) > 1 || len(result) == 1) && err != nil {
			outEnd := result[len(result)-1]
			if errType.Type().AssignableTo(outEnd.Type()) {
				outEnd.Set(errType)
			}
		}
		end := time.Now()
		Info(" \nSQL Statements ==>", statements, "\nSQL Template ==> ", templateSql, ",\nContext Parameter:", ctx, " \nCount:", value.Len(), "\nTime:", end.Sub(star).String())
		return result
	}
}

func (build *Build) initMapper(id []string, fun reflect.Value) {
	numOut := fun.Type().NumOut()
	values := make([]reflect.Value, 0)
	var outValue reflect.Value
	for i := 0; i < numOut; i++ {
		out := fun.Type().Out(i)
		outValue = reflect.New(out).Elem()
		if out.Kind() == reflect.Pointer {
			elem := reflect.New(out.Elem())
			if outValue.CanSet() {
				outValue.Set(elem)
				initField(outValue)
			}
		}
		initField(outValue)
		values = append(values, outValue)
	}
	f := reflect.MakeFunc(fun.Type(), build.mapper(id, values))
	fun.Set(f)
}

func initField(value reflect.Value) {
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
		initField(value)
	}
	if value.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if !field.CanSet() {
			continue
		}
		if field.Kind() == reflect.Pointer {
			if field.IsZero() {
				if !field.CanSet() {
					elem := reflect.New(field.Type()).Elem()
					field.Set(elem)
				}
				if !field.Elem().CanSet() {
					elem := reflect.New(field.Type()).Elem()
					fmt.Println(elem.Type().String())
					field.Set(elem)
				}
			}
			initField(field)
			continue
		}
		if field.Kind() == reflect.Struct {
			initField(field)
			continue
		}
	}
}

func resultMapping(row reflect.Value, resultType any) (reflect.Value, reflect.Value) {
	var err error
	var flag bool
	var column []string
	t := reflect.SliceOf(reflect.TypeOf(resultType))
	result := reflect.MakeSlice(t, 0, 0)
	// 确定数据库 列顺序 排列扫描顺序
	columns := row.MethodByName("Columns").Call(nil)
	if !columns[1].IsZero() {
		return result, columns[1]
	}
	if column, flag = columns[0].Interface().([]string); !flag {
		return result, reflect.ValueOf(errors.New("get row column error"))
	}
	// 校验 resultType 是否覆盖了结果集
	if flag, err = SelectCheck(column, resultType); !flag {
		return result, reflect.ValueOf(err)
	}
	// 解析结构体 映射字段
	// 拿到 scan 方法
	scan := row.MethodByName("Scan")
	next := row.MethodByName("Next")

	mapping := ResultMapping(resultType)
	for (next.Call(nil))[0].Interface().(bool) {
		var value, unValue reflect.Value
		if reflect.TypeOf(resultType).Kind() == reflect.Pointer {
			//创建一个 接收结果集的变量
			value = reflect.New(reflect.TypeOf(resultType).Elem())
			unValue = value.Elem()
		} else if reflect.TypeOf(resultType).Kind() == reflect.Map {
			value = reflect.MakeMap(reflect.TypeOf(resultType))
			unValue = value
		} else {
			value = reflect.New(reflect.TypeOf(resultType))
			value = value.Elem()
			unValue = value
		}
		// 创建 接收器
		values, fieldIndexMap, MapKey := buildScan(unValue, column, mapping)
		// 执行扫描, 执行结果扫描，不处理error 扫码结果类型不匹配，默认为零值
		scan.Call(values)
		// 迭代是否有特殊结构体 主要对 时间类型做了处理
		scanWrite(values, fieldIndexMap)
		scanMap(unValue, values, MapKey)
		// 添加结果集
		result = reflect.Append(result, value)
	}
	return result, reflect.New(reflect.TypeOf(new(error)).Elem()).Elem()
}

// 构建结构体接收器
// value 接收数据库结果对应的参数，可能是结构体也可能是 map
// columns 数据库结果集的列名
// resultColumn 对应 value(结构体类型)参数 和 columns 参数的 映射关系，value(map类型)时候 该值为空
func buildScan(value reflect.Value, columns []string, resultColumn map[string]string) ([]reflect.Value, map[int]reflect.Value, map[int]string) {
	// Scan 函数调用参数列表,接收器存储的都是指针类 反射的指针类型
	values := make([]reflect.Value, 0)
	// 存储的 也将是指针的反射形式
	fieldIndexMap := make(map[int]reflect.Value)
	MapKey := make(map[int]string)
	if len(columns) == 1 {
		values = append(values, value.Addr())
		return values, fieldIndexMap, MapKey
	}
	// 创建 接收器
	for index, column := range columns {
		if resultColumn == nil {
			MapKey[index] = column
			values = append(values, reflect.New(reflect.TypeOf("")))
			continue
		}
		// 通过结构体映射找到 数据库映射到结构体的字段名
		name := resultColumn[column]
		// 找到对应的字段
		byName := value.FieldByName(name)
		if byName == (reflect.Value{}) {
			// 没有找到对应的
			panic("The type of the returned value does not match the result set of the sql query, and the mapping fails. Check whether the structure field name or 'column' tag matches the mapping relationship of the query data set")
		}
		// 检查 接收参数 如果是特殊参数 比如结构体，时间类型的情况需要特殊处理 当前仅对时间进行特殊处理 ,获取当前 参数的 values 索引 并保存替换
		// fieldIndexMap 存储的是对应字段的地址，若字段类型为指针，则要为指针分配地址后进行保存
		field := byName.Interface()
		switch field.(type) {
		case time.Time:
			// 记录特殊 值的索引 并且替换掉
			index := len(values)
			fieldIndexMap[index] = byName.Addr()
			// 替换 默认使用空字符串去接收
			values = append(values, reflect.New(reflect.TypeOf("")))
			continue
		case *time.Time:
			// 记录特殊 值的索引 并且替换掉
			index := len(values)
			fieldIndexMap[index] = byName
			// 替换 使用空字符串去接收
			values = append(values, reflect.New(reflect.TypeOf("")))
			continue
		}
		values = append(values, byName.Addr())
	}
	return values, fieldIndexMap, MapKey
}

// 对 buildScan 函数构建阶段存在特殊字段的处理 进行回写到指定的结构体位置
// values 数据结果集 一行记录
func scanWrite(values []reflect.Value, fieldIndexMap map[int]reflect.Value) {
	// 迭代是否有特殊结构体 主要对 时间类型做了处理
	for k, v := range fieldIndexMap {
		// 拿到 特殊结构体对应的 值
		mapV := values[k]
		key := BaseTypeKey(v)
		err := BaseType[key](v, mapV.Elem().Interface())
		if err != nil {
			Panic(err)
		}
	}
}

func scanMap(value reflect.Value, values []reflect.Value, MapKey map[int]string) {
	if len(MapKey) > 0 {
		for i := 0; i < len(values); i++ {
			key := MapKey[i]
			value.SetMapIndex(reflect.ValueOf(key), values[i].Elem())
		}
	}
}

// ResultMapping
// 解析结构体上的column标签 生成数据库字段到结构体字段的映射匹配
// value 如果传递的是结构体，则会解析对应的 column 标签 或者字段本身
// value 如果是 map 则不会做任何处理，返回 nil
func ResultMapping(value any) map[string]string {
	mapp := make(map[string]string)
	of := reflect.TypeOf(value)
	if of.Kind() == reflect.Pointer {
		tf := reflect.ValueOf(value).Elem()
		return ResultMapping(tf.Interface())
	}
	switch of.Kind() {
	case reflect.Struct:
		for i := 0; i < of.NumField(); i++ {
			field := of.Field(i)
			mapp[field.Name] = field.Name
			if get := field.Tag.Get("column"); get != "" {
				mapp[get] = field.Name
			}
		}
	default:
		return nil
	}
	return mapp
}

func SelectCheck(columns []string, resultType any) (bool, error) {
	rf := reflect.ValueOf(resultType)
	if rf.Kind() == reflect.Pointer {
		return SelectCheck(columns, rf.Elem())
	}
	if len(columns) > 1 {
		if rf.Kind() != reflect.Struct && rf.Kind() != reflect.Map {
			return false, errors.New("the return type is incorrect and requires either a structure type or a map to receive")
		}
	}
	return true, nil
}

// QueryResultMapper SQL 查询结果赋值
func QueryResultMapper(value reflect.Value, result []reflect.Value) {
	var itemT reflect.Type
	var itemV reflect.Value
	length := len(result)
	if value.Len() == 0 {
		return
	}
	itemT = value.Type().Elem()
	itemV = value.Index(0)
	for i := 0; i < length-1; i++ {
		out := result[i]
		if value.Type().AssignableTo(out.Type()) {
			out.Set(value)
		} else if itemT.AssignableTo(out.Type()) {
			out.Set(itemV)
		}
	}
}

// ExecResultMapper SQL执行结果赋值
// 规则:
// insert,update,delete,默认第一个返回值为 执行sql 影响的具体行数
// insert 第二个返回参数是 自增长主键
func ExecResultMapper(result []reflect.Value, exec sql.Result) error {
	length := len(result)
	for i := 0; i < length-1; i++ {
		if i == 0 {
			affected, err := exec.RowsAffected()
			if err != nil {
				return err
			}
			result[i].Set(reflect.ValueOf(affected))
		}
		if i == 1 {
			affected, err := exec.LastInsertId()
			if err != nil {
				return err
			}
			result[i].Set(reflect.ValueOf(affected))
		}
		if i > 1 {
			break
		}
	}
	return nil
}
