package sgo

import (
	"database/sql"
	"fmt"
	"github.com/druidcaesa/ztool"
	"reflect"
	"time"
)

type MapperFunc func([]reflect.Value) []reflect.Value

// Mapper 创建 映射函数
func (build *Build) mapper(id []string, result []reflect.Value) MapperFunc {
	return func(values []reflect.Value) []reflect.Value {
		fmt.Printf("%s %s \n", result[0].Type().String(), result[1].Type().String())
		err := result[1].Interface()
		//value := result[0]
		get, err := build.Get(id, values[0].Interface())
		if err != nil {
			result[1].Set(reflect.ValueOf(err))
			return result
		}
		fmt.Println(get)
		return result
	}
}

func (build *Build) initMapper(id []string, fun reflect.Value) {
	numOut := fun.Type().NumOut()
	values := make([]reflect.Value, 0)
	for i := 0; i < numOut; i++ {
		out := fun.Type().Out(i)
		elem := reflect.New(out).Elem()
		values = append(values, elem)
	}
	f := reflect.MakeFunc(fun.Type(), build.mapper(id, values))
	fun.Set(f)
}

func resultMapping(row *sql.Rows, resultType any) []any {
	of := reflect.ValueOf(row)
	// 确定数据库 列顺序 排列扫描顺序
	columns, err := row.Columns()
	if err != nil {
		panic(err.Error())
	}
	// 校验 resultType 是否覆盖了结果集
	// 解析结构体 映射字段
	// 拿到 scan 方法
	scan := of.MethodByName("Scan")
	next := of.MethodByName("Next")
	t := reflect.SliceOf(reflect.TypeOf(resultType))
	result := reflect.MakeSlice(t, 0, 0)
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
		mapping := structMapping(resultType)
		// 创建 接收器
		values, fieldIndexMap := buildScan(unValue, columns, mapping)
		// 执行扫描, 执行结果扫描，不处理error 扫码结果类型不匹配，默认为零值
		scan.Call(values)
		// 迭代是否有特殊结构体 主要对 时间类型做了处理
		scanWrite(values, fieldIndexMap)
		// 添加结果集
		result = reflect.Append(result, value)
	}
	return result.Interface().([]any)
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

// 生成结构体 映射匹配
func structMapping(s any) map[string]string {
	mapp := make(map[string]string)
	of := reflect.TypeOf(s)
	if of.Kind() == reflect.Pointer {
		tf := reflect.ValueOf(s).Elem()
		return structMapping(tf.Interface())
	}
	for i := 0; i < of.NumField(); i++ {
		field := of.Field(i)
		mapp[field.Name] = field.Name
		if get := field.Tag.Get("column"); get != "" {
			mapp[get] = field.Name
		}
	}
	return mapp
}

func ormCheck() {

}
