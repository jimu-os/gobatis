package sgo

import (
	"database/sql"
	"errors"
	"github.com/druidcaesa/ztool"
	"reflect"
	"time"
)

type MapperFunc func([]reflect.Value) []reflect.Value

// Mapper 创建 映射函数
func (build *Build) mapper(id []string, fun reflect.Value, result []reflect.Value) MapperFunc {
	return func(values []reflect.Value) []reflect.Value {
		err := result[1].Interface()
		//value := result[0]
		get, tag, err := build.Get(id, values[0].Interface())
		if err != nil {
			result[1].Set(reflect.ValueOf(err))
			return result
		}
		if b, err := MapperCheck(fun, tag); !b {
			panic(err)
		}

		query, err := build.DB.Query(get)
		if err != nil {
			return nil
		}
		value := resultMapping(query, result[0].Interface())
		if result[0].Kind() != reflect.Slice {
			out := result[0]
			if out.CanSet() {
				out.Set(value.Index(0))
			}
		}
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
			}
		}
		values = append(values, outValue)
	}
	f := reflect.MakeFunc(fun.Type(), build.mapper(id, fun, values))
	fun.Set(f)
}

func resultMapping(row *sql.Rows, resultType any) reflect.Value {
	of := reflect.ValueOf(row)
	// 确定数据库 列顺序 排列扫描顺序
	columns, err := row.Columns()
	if err != nil {
		panic(err.Error())
	}
	// 校验 resultType 是否覆盖了结果集
	if b, err := SelectCheck(columns, resultType); !b {
		panic(err)
	}
	// 解析结构体 映射字段
	// 拿到 scan 方法
	scan := of.MethodByName("Scan")
	next := of.MethodByName("Next")
	t := reflect.SliceOf(reflect.TypeOf(resultType))
	result := reflect.MakeSlice(t, 0, 0)
	mapping := ResultMapping(resultType)
	for (next.Call(nil))[0].Interface().(bool) {
		var value, unValue reflect.Value
		if reflect.TypeOf(resultType).Kind() == reflect.Pointer {
			//创建一个 接收结果集的变量
			value = reflect.New(reflect.TypeOf(resultType).Elem())
			unValue = value.Elem()
		} else {
			value = reflect.New(reflect.TypeOf(resultType))
			value = value.Elem()
			unValue = value
		}
		// 创建 接收器
		values, fieldIndexMap := buildScan(unValue, columns, mapping)
		// 执行扫描, 执行结果扫描，不处理error 扫码结果类型不匹配，默认为零值
		scan.Call(values)
		// 迭代是否有特殊结构体 主要对 时间类型做了处理
		scanWrite(values, fieldIndexMap)
		// 添加结果集
		result = reflect.Append(result, value)
	}
	return result
}

// 构建结构体接收器
func buildScan(value reflect.Value, columns []string, resultColumn map[string]string) ([]reflect.Value, map[int]reflect.Value) {
	// Scan 函数调用参数列表,接收器存储的都是指针类 反射的指针类型
	values := make([]reflect.Value, 0)
	// 存储的 也将是指针的反射形式
	fieldIndexMap := make(map[int]reflect.Value)
	// 创建 接收器
	for _, column := range columns {
		// 通过结构体映射找到 数据库映射到结构体的字段名
		name := resultColumn[column]
		// 找到对应的字段
		byName := value.FieldByName(name)
		// 检查 接收参数 如果是特殊参数 比如结构体，时间类型的情况需要特殊处理 当前仅对时间进行特殊处理 ,获取当前 参数的 values 索引 并保存替换
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
			fieldIndexMap[index] = byName.Addr()
			// 替换 使用空字符串去接收
			values = append(values, reflect.New(reflect.TypeOf("")))
			continue
		}
		values = append(values, byName.Addr())
	}
	return values, fieldIndexMap
}

// 对 buildScan 函数构建阶段存在特殊字段的处理 进行回写到指定的结构体位置
func scanWrite(values []reflect.Value, fieldIndexMap map[int]reflect.Value) {
	// 迭代是否有特殊结构体 主要对 时间类型做了处理
	for k, v := range fieldIndexMap {
		// 拿到 特殊结构体对应的 值
		mapV := values[k]
		structField := v.Interface()
		switch structField.(type) {
		case *time.Time:
			// 吧把对应的 mappv 转化为 time.Time
			mappvalueString := mapV.Elem().Interface().(string)
			parse, err := ztool.DateUtils.Parse(mappvalueString)
			if err != nil {
				panic(err)
			}
			t2 := parse.Time()
			valueOf := reflect.ValueOf(t2)
			//设置该指针指向的值
			v.Elem().Set(valueOf)
		case **time.Time:
			// 吧把对应的 mappv 转化为 time.Time
			mappvalueString := mapV.Elem().Interface().(string)
			parse, err := ztool.DateUtils.Parse(mappvalueString)
			if err != nil {
				panic(err)
			}
			t2 := parse.Time()
			valueOf := reflect.ValueOf(&t2)
			//设置该指针指向的值
			v.Elem().Set(valueOf)
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
	case reflect.Map:
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
			return false, errors.New("")
		}
	}
	return true, nil
}

// MapperCheck 检查 不同类别的sql标签 Mapper 函数是否符合规范
func MapperCheck(fun reflect.Value, tag string) (bool, error) {
	switch tag {
	case Select:
		//TODO
	case Insert:
		//TODO
	case Update:
		//TODO
	case Delete:
		//TODO
	}
	if fun.Type().NumIn() != 1 {
		return false, errors.New("there can only be one argument")
	}
	if fun.Type().NumOut() != 2 {
		return false, errors.New("there can only be two return values")
	}
	out := fun.Type().Out(1)
	if !out.Implements(reflect.TypeOf(new(error)).Elem()) {
		return false, errors.New("the second return value must be error")
	}
	return true, nil
}
