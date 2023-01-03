package gobatis

import (
	"context"
	"database/sql"
	"errors"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
	"time"
)

type MapperFunc func([]reflect.Value) []reflect.Value

// Mapper 创建 映射函数
func (batis *GoBatis) mapper(id []string, returns []reflect.Value) MapperFunc {
	return func(values []reflect.Value) []reflect.Value {
		result := make([]reflect.Value, len(returns))
		copy(result, returns)
		var errType, Exec, BeginCall reflect.Value
		var ctx any
		errType = reflect.New(reflect.TypeOf(new(error)).Elem()).Elem()
		db := batis.db
		c, ctx, db, auto := Args(db, values)
		results := Return(result)
		statements, tag, templateSql, params, err := batis.get(id, ctx)
		if err != nil {
			errType = reflect.ValueOf(err)
			results[len(results)-1].Set(errType)
			return result
		}
		switch tag {
		case Select:
			err := batis.selectStatement(db, c, statements, templateSql, params, results)
			if errType = err; !errType.IsZero() {
				goto end
			}
		case Insert, Update, Delete:
			errType = batis.execStatement(db, c, Exec, &BeginCall, auto, statements, templateSql, params, results)
			if !errType.IsZero() {
				goto end
			}
		}
	end:
		End(auto, results, errType, BeginCall)
		return result
	}
}

// Args 参数赋值处理
// 处理定义函数的入参，返回一个参数序列给到后面的函数调用入参
func Args(db reflect.Value, values []reflect.Value) (ctx reflect.Value, args any, tx reflect.Value, auto bool) {
	params := make(map[string]any)
	tx = db
	// 是否启用自动提交事务
	auto = true
	ctx = reflect.ValueOf(context.Background())
	ctxType := reflect.TypeOf(new(context.Context)).Elem()
	txType := reflect.TypeOf(&sql.Tx{})
	length := len(values)
	for i := 0; i < length; i++ {
		arg := values[i]
		argType := arg.Type()
		if argType.AssignableTo(ctxType) {
			ctx = arg
			continue
		}
		if argType.AssignableTo(txType) {
			if arg.IsZero() {
				// 针对 txType 为空，我们不选择采用 外部提供的事务，主要方便支持一个定义灵活调用情况
				continue
			}
			tx = arg
			// 外部提供 事务，SGO 内部不自动提交
			auto = false
			continue
		}
		args = arg.Interface()
		m := toMap(args)
		mergeMap(params, m)
	}
	args = params
	return
}

// Return 处理返回值排序
func Return(result []reflect.Value) (ret []reflect.Value) {
	var err reflect.Value
	errType := reflect.TypeOf(new(error)).Elem()
	for i := 0; i < len(result); i++ {
		r := result[i]
		rType := r.Type()
		if rType.AssignableTo(errType) {
			err = r
			continue
		}
		ret = append(ret, r)
	}
	ret = append(ret, err)
	return
}

// SelectStatement 执行查询
func (batis *GoBatis) selectStatement(db, ctx reflect.Value, statements, templateSql string, params []any, result []reflect.Value) reflect.Value {
	var resultType reflect.Value
	star := time.Now()
	Query := db.MethodByName("QueryContext")
	call := Query.CallSlice([]reflect.Value{
		ctx,
		reflect.ValueOf(templateSql),
		reflect.ValueOf(params),
	})
	if !call[1].IsZero() {
		return call[1]
	}
	if result[0].Kind() == reflect.Slice {
		// 拿到 切片元素类型
		sliceType := result[0].Type().Elem()
		// 切片是没有具体元素，所以需要创建一个具体的元素
		switch sliceType.Kind() {
		case reflect.Pointer:
			// 创建 指针
			resultType = reflect.New(sliceType).Elem()
			// 创建 指针 *
			elem := reflect.New(sliceType.Elem())
			// 给指针 设置一个对应的指针值
			resultType.Set(elem)
		case reflect.Struct:
			resultType = reflect.New(sliceType).Elem()
		}
	} else {
		resultType = result[0]
	}
	value, err := resultMapping(call[0], resultType.Interface())
	if !err.IsZero() {
		return err
	}
	QueryResultMapper(value, result)
	end := time.Now()
	batis.Info("SQL Query Statements ==>", statements, "SQL Template ==> ", templateSql, ",Parameter:", params, "Count:", value.Len(), "Time:", end.Sub(star).String())
	return err
}

// ExecStatement 执行修改
func (batis *GoBatis) execStatement(db, ctx, Exec reflect.Value, BeginCall *reflect.Value, auto bool, statements, templateSql string, params []any, result []reflect.Value) reflect.Value {
	star := time.Now()
	errType := reflect.New(reflect.TypeOf(new(error)).Elem()).Elem()
	if !auto {
		Exec = db.MethodByName("ExecContext")
	} else {
		BeginFunc := db.MethodByName("Begin")
		call := BeginFunc.Call(nil)
		if !call[1].IsZero() {
			return call[1]
		}
		*BeginCall = call[0]
		Exec = BeginCall.MethodByName("ExecContext")
	}
	call := Exec.CallSlice([]reflect.Value{
		ctx,
		reflect.ValueOf(templateSql),
		reflect.ValueOf(params),
	})
	if !call[1].IsZero() {
		return call[1]
	}
	var count int64
	count, err := ExecResultMapper(result, call[0].Interface().(sql.Result))
	if err != nil {
		errType.Set(reflect.ValueOf(err))
		return errType
	}
	end := time.Now()
	batis.Info("SQL Exec Statements ==>", statements, "SQL Template ==> ", templateSql, ",Parameter:", params, "Count:", count, "Time:", end.Sub(star).String())
	return errType
}

// End 错误提交及回滚
func End(auto bool, result []reflect.Value, errType, BeginCall reflect.Value) {
	length := len(result)
	outEnd := (result)[length-1]
	if errType.Type().AssignableTo(outEnd.Type()) && !errType.IsZero() {
		outEnd.Set(errType)
		if auto {
			RollbackFunc := BeginCall.MethodByName("Rollback")
			Rollback := RollbackFunc.Call(nil)
			if !Rollback[0].IsZero() {
				outEnd.Set(Rollback[0])
			}
		}
	} else if BeginCall != (reflect.Value{}) && auto {
		CommitFunc := BeginCall.MethodByName("Commit")
		Commit := CommitFunc.Call(nil)
		if !Commit[0].IsZero() {
			outEnd.Set(Commit[0])
		}
	}
}

func (batis *GoBatis) initMapper(id []string, fun reflect.Value) {
	numOut := fun.Type().NumOut()
	values := make([]reflect.Value, 0)
	var outValue reflect.Value
	for i := 0; i < numOut; i++ {
		out := fun.Type().Out(i)
		switch out.Kind() {
		case reflect.Pointer:
			outValue = reflect.New(out).Elem()
			elem := reflect.New(out.Elem())
			if outValue.CanSet() {
				outValue.Set(elem)
				initField(outValue)
			}
		case reflect.Slice:
			slice := reflect.MakeSlice(out, 0, 0)
			outValue = reflect.New(out).Elem()
			outValue.Set(slice)
		default:
			outValue = reflect.New(out).Elem()
		}
		initField(outValue)
		values = append(values, outValue)
	}
	f := reflect.MakeFunc(fun.Type(), batis.mapper(id, values))
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
				if !field.Elem().CanSet() {
					elem := reflect.New(field.Type()).Elem()
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
			// 初始化 内部指针
			initField(unValue)
		} else if reflect.TypeOf(resultType).Kind() == reflect.Map {
			value = reflect.MakeMap(reflect.TypeOf(resultType))
			unValue = value
		} else {
			value = reflect.New(reflect.TypeOf(resultType))
			value = value.Elem()
			unValue = value
			// 初始化 内部指针
			initField(unValue)
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
		Field := value.FieldByName(name)
		if Field == (reflect.Value{}) {
			// 没有找到对应的
			panic("The '" + column + "' of the result set does not match the structure '" + value.Type().String() + "',the type of the returned value does not match the result set of the sql query, and the mapping fails. Check whether the structure field name or 'column' tag matches the mapping relationship of the query data set")
		}
		// 检查 接收参数 如果是特殊参数 比如结构体，时间类型的情况需要特殊处理 当前仅对时间进行特殊处理 ,获取当前 参数的 values 索引 并保存替换
		// fieldIndexMap 存储的是对应字段的地址，若字段类型为指针，则要为指针分配地址后进行保存
		switch Field.Kind() {
		case reflect.Struct:
			// indexV (在调用 scan(。。)方法参数的索引位置) 记录特殊 值的索引 并且替换掉，将会在 scanWrite 方法中执行替换数据
			indexV := len(values)
			fieldIndexMap[indexV] = Field
			// 替换 默认使用空字符串去接收
			values = append(values, reflect.New(reflect.TypeOf("")))
			continue
		case reflect.Pointer:
			// indexV (在调用 scan(。。)方法参数的索引位置) 记录特殊 值的索引 并且替换掉 ，将会在 scanWrite 方法中执行替换数据
			indexV := len(values)
			fieldIndexMap[indexV] = Field
			// 替换 使用空字符串去接收
			values = append(values, reflect.New(reflect.TypeOf("")))
			continue
		}
		values = append(values, Field.Addr())
	}
	return values, fieldIndexMap, MapKey
}

// 对 buildScan 函数构建阶段存在特殊字段的处理 进行回写到指定的结构体位置
// values 数据结果集 一行记录
func scanWrite(values []reflect.Value, fieldIndexMap map[int]reflect.Value) {
	var fun ToGolang
	var b bool
	// 迭代是否有特殊结构体 主要对 时间类型做了处理
	for k, v := range fieldIndexMap {
		// 拿到 特殊结构体对应的 值
		mapV := values[k]
		key := BaseTypeKey(v)
		if fun, b = databaseToGolang[key]; !b {
			// 进行自定义 数据映射期间找不到对应的匹配处理器，将产生恐慌提示用户对这个数据类型应该提供一个处理注册
			// 没有找到对应的数据处理，可以通过 gobatis.GolangType 方法对 具体类型进行注册
			Panic("The data processor corresponding to the '" + key + "' is not occupied. You need to register GolangType to support this type")
		}
		err := fun(v, mapV.Elem().Interface())
		if err != nil {
			Panic(err)
		}
	}
}

// scanMap select 查询返回的结果集是 map 的处理方式
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
			name := field.Name
			// 添加多种字段名匹配情况

			// 蛇形
			mapp[strcase.ToSnake(name)] = name

			// 驼峰
			mapp[strcase.ToCamel(name)] = name
			mapp[strcase.ToLowerCamel(name)] = name

			// 全小写
			mapp[strings.ToLower(name)] = name
			// 全大写
			mapp[strings.ToUpper(name)] = name
			// 自定义
			if get := field.Tag.Get("column"); get != "" {
				mapp[get] = name
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
// value 对应查询结果集
// result 对应mapper函数返回值
func QueryResultMapper(value reflect.Value, result []reflect.Value) {
	var itemT reflect.Type
	var itemV reflect.Value
	length := len(result)
	if value.Len() == 0 {
		return
	}
	// 拿到 结果集切片内部存储的数据类型
	itemT = value.Type().Elem()
	// 默认取所有结果集中的第一个参数，以变单条数据查询赋值
	itemV = value.Index(0)
	for i := 0; i < length-1; i++ {
		out := result[i]
		if value.Type().AssignableTo(out.Type()) {
			// 校验 切片是否可以赋值给 返回值
			out.Set(value)
		} else if itemT.AssignableTo(out.Type()) {
			// 校验 切片第一个参数是否可以赋值给 返回值
			out.Set(itemV)
		}
	}
}

// ExecResultMapper SQL执行结果赋值
// 规则:
// insert,update,delete,默认第一个返回值为 执行sql 影响的具体行数
// insert 第二个返回参数是 自增长主键
func ExecResultMapper(result []reflect.Value, exec sql.Result) (count int64, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(e.(string))
			count = -1
		}
	}()
	var lid int64
	length := len(result)
	if length-1 == 0 {
		count, err = exec.RowsAffected()
		if err != nil {
			return
		}
	}
	for i := 0; i < length-1; i++ {
		if i == 0 {
			count, err = exec.RowsAffected()
			if err != nil {
				return
			}
			result[i].Set(reflect.ValueOf(count))
		}
		if i == 1 {
			lid, err = exec.LastInsertId()
			if err != nil {
				return
			}
			result[i].Set(reflect.ValueOf(lid))
		}
		if i > 1 {
			break
		}
	}
	return
}
