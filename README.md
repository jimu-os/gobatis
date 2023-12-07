# GoBatis
[![Go Report Card](https://goreportcard.com/badge/gitee.com/aurora-engine/sgo)](https://goreportcard.com/report/gitee.com/aurora-engine/sgo)<br>
## version
```shell
go1.21
```
`GoBatis` 是 `MyBatis` go语言实现，`GoBatis`提供对 mapper 的上下文数据解析填充，并不保证对 sql 语句的语法检查,支持自定义数据映射.

## XML 解析规则

`gobatis` 解析 xml 文件中的sql语句，会严格检查上下文中的数据类型，字符串类型参数会自定添加 ` '' `
单引号，其他基础数据类型不会添加，对于复杂数据结构(复合结构，泛型结构体等)会持续跟进
，目前仅支持基础数据类型。

### 上下文数据

上下文数据是由用户调用时候传递接，仅接受 map 或者结构体.

### 标签详情

| 标签         | 描述       | 功能                                                             |
|:-----------|:---------|:---------------------------------------------------------------|
| `<mapper>` | 根节点      |                                                                |
| `<insert>` | insert语句 | 生成插入语句                                                         |
| `<select>` | select语句 | 生成查询语句                                                         |
| `<update>` | update语句 | 生成更新语句                                                         |
| `<delete>` | delete语句 | 生成删除语句                                                         |
| `<where>`  | where语句  | where 标签内的标签将被解析，如果条件存在成立会自动补全 where关键字，若子标签完全不成立则不会补全where关键字 |
| `<for>`    | for迭代    | 生成IN语句，指定需要生成IN条件的字段，可以生成对应的IN条件                               |
| `<if>`     | if条件     | 判断是否满足属性表达式的条件，满足条件就对标签内的sql进行解析                               |

## 定义 Mapper

`gobatis` 中的 `mapper` 定义是基于结构体 和`匿名函数`字段来实现的(匿名函数字段，需要遵循一些规则):

- 上下文参数，只能是结构体，指针结构体或者map
- 至少有一个返回值，一个返回值只能是 `error`

## 快速入门

### 创建 table

创建一张表，用于测试

```sql
create table student
(
    id          int auto_increment primary key,
    name        varchar(20) null,
    age         int         null,
    create_time datetime    null
);
```

### 创建 映射模型

更具 数据库表，或者sql查询结果集 创建一个结构体用于接收查询数据，`column` 属性的值对应者 sql 表定义的列名

```go
// Student 用户模型
type Student struct {
	Id         int    `column:"id"json:"id,omitempty"`
	Name       string `column:"name"json:"name,omitempty"`
	Age        int    `column:"age"json:"age,omitempty"`
	CreateTime string `column:"create_time"json:"create_time,omitempty"`
}
```

### 准备 运行环境

#### 创建 mapper 结构体

```go
type StudentMapper struct {
}
```

#### 创建 mapper 文件

更具 mapper 结构体的名称创建一个 mapper xml文件

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">

</mapper>
```

#### 项目目录结构

```txt
|--root
|--model
|   |--student.go
|--mapper_test.go
|--StudentMapper.go
|--text.xml
```

#### 初始化 gobatis

```go
var err error
var open *sql.DB
var studentMapper *StudentMapper

func init() {
	studentMapper = &StudentMapper{}
	open, err = sql.Open("mysql", "xxx:xx@xx@tcp(localhost:3306)/test?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return
	}
	batis := gobatis.New(open)
	batis.Source("/")
	batis.ScanMappers(studentMapper)
}
```

## 数据插入数据

#### 添加 mapper 方法

此时你的 `mapper` 应该是下面的样子

```go
type StudentMapper struct {
	AddOne func(student model.Student) error
}
```

#### 添加 xml insert 标签

根据定义的字段名称,对应在 mapper 文件中添加一个 `id="AddOne"` insert 标签，
标签内书写需要执行的`sql`语句，`sql`语句中的变量通过 `{}` 的形式去加载解析

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">
    <insert id="AddOne">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
</mapper>
```

#### 调用执行插入数据

下面通过测试对刚刚定义的插入方法进行执行，所有的前置步骤都在上面的初始化中准备好了，直接调用 `AddOne` 字段即可实现数据插入

```go
func TestInsert(t *testing.T) {
	s := model.Student{
		Name:       "test",
		Age:        1,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	if err = studentMapper.AddOne(s); err != nil {
		t.Error(err.Error())
		return
	}
}
```

## 执行行数 和 自增主键

定义mapper字段 `InsertId`,它有3个返回值，第一个返回值是执行sql返回的影响行数，第二个返回值是返回自增长逐渐值，
默认第一个参数是返回影响行数。

```go
type StudentMapper struct {
	AddOne   func(student model.Student) error
	InsertId func(student model.Student) (int64, int64, error)
}
```

#### 定义xml

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">
    <insert id="AddOne">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
    <insert id="InsertId">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
</mapper>
```

#### 执行测试

```go
func TestInsertId(t *testing.T) {
	var count, id int64
	s = model.Student{
		Name:       "test",
		Age:        2,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	if count, id, err = studentMapper.InsertId(s); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("count:", count, "id:", id)
}
```

## 实现批量插入

添加新的方法，此时你的 `mapper` 应该是下面的样子

```go
type StudentMapper struct {
	AddOne   func(student model.Student) error
	InsertId func(student model.Student) (int64, int64, error)
	Adds     func(ctx any) error
}
```

#### 定义新的 mapper insert

此时你的的 mapper 应该是如下，添加了一个新的 insert 标签 `id="Adds"`，其中 使用`<for></for>` 标签对传递的数组数据进行了解析
`slice="{arr}"` 属性指定了属性名称为 arr 的数据，`item="stu"`表示的是迭代过程中的对象参数，更具数据元素来定，如果是基础数据，那么代表数据本身

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">
    <insert id="AddOne">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
    <insert id="InsertId">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
    <insert id="Adds">
        insert into student (name, age, create_time) values
        <for slice="{arr}" item="stu">
            ({stu.Name},{stu.Age},{stu.CreateTime})
        </for>
    </insert>
</mapper>
```

#### 调用执行

```go
func TestSliceInsert(t *testing.T) {
	var arr []model.Student
	for i := 0; i < 10; i++ {
		s := model.Student{
			Name:       fmt.Sprintf("test_%d", i),
			Age:        i + 2,
			CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		}
		arr = append(arr, s)
	}
	err = studentMapper.Adds(
		map[string]any{
			"arr": arr,
		},
	)
	if err != nil {
		t.Error(err.Error())
		return
	}
}
```

## 数据查询

#### 定义查询

定义了mapper字段 `QueryAll` 查询全部采用对应的切片模型进行接收即可，查询多条数据结果集的时候任然可以使用单个模型接收，
只是单个模型的数据仅仅取到结果集的第一条数据。

```go
type StudentMapper struct {
	AddOne   func(student model.Student) error
	InsertId func(student model.Student) (int64, int64, error)
	Adds     func(ctx any) error
	
	QueryAll func() ([]model.Student, error)
}
```

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">
    <insert id="AddOne">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
    <insert id="InsertId">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
    <insert id="Adds">
        insert into student (name, age, create_time) values
        <for slice="{arr}" item="stu">
            ({stu.Name},{stu.Age},{stu.CreateTime})
        </for>
    </insert>

    <select id="QueryAll">
        select * from student
    </select>
</mapper>
```

#### 执行

```go
func TestQueryAll(t *testing.T) {
	var stus []model.Student
	if stus, err = studentMapper.QueryAll(); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(stus)
}
```

## 分页查询

添加分页 mapper 字段 `QueryPage`，作为测试我们不进行参数传递，它返回3个参数，第一个参数是分页数据，第二个参数，是`sql`
条件所统计的总数，
查询mapper不返回 `int64` 的参数就不会自动统计数量

```go
type StudentMapper struct {
	AddOne   func(student model.Student) error
	InsertId func(student model.Student) (int64, int64, error)
	Adds     func(ctx any) error

	QueryAll  func() ([]model.Student, error)
	QueryPage func() ([]model.Student, int64, error)
}
```

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">
    <insert id="AddOne">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
    <insert id="InsertId">
        insert into student (name, age, create_time)
        values ({Name},{Age},{CreateTime});
    </insert>
    <insert id="Adds">
        insert into student (name, age, create_time) values
        <for slice="{arr}" item="stu">
            ({stu.Name},{stu.Age},{stu.CreateTime})
        </for>
    </insert>

    <select id="QueryAll">
        select * from student
    </select>

    <select id="QueryPage">
        select * from student limit 2 offset 0
    </select>
</mapper>
```

#### 执行

```go
func TestQueryPage(t *testing.T) {
	var stus []model.Student
	var count int64
	if stus, count, err = studentMapper.QueryPage(); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log("rows:", stus, "count:", count)
}
```

## 事务支持

定义一个数据修改操作，通过外部传递一个事务 `tx` 由它来完成数据库操作后的提交或是回滚，我们定义一个 `Update` 第二个参数传递事务

```go
type StudentMapper struct {
	AddOne   func(student model.Student) error
	InsertId func(student model.Student) (int64, int64, error)
	Adds     func(ctx any) error

	QueryAll  func() ([]model.Student, error)
	QueryPage func() ([]model.Student, int64, error)

	Update func(student model.Student, tx *sql.Tx) (int64, error)
}
```

编写sql语句，修改年龄大于5的数据姓名修改为AAA

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">
    <!--  略 ..  -->
    <update id="Update">
        update student set name={Name} where age>{Age}
    </update>
</mapper>
```

#### 运行测试

```go
func TestUpdate(t *testing.T) {
	var begin *sql.Tx
	var count int64
	begin, err = open.Begin()
	if err != nil {
		t.Error(err.Error())
		return
	}
	u := model.Student{
		Name: "AAA",
		Age:  5,
	}
	count, err = studentMapper.Update(u, begin)
	if err != nil {
		t.Error(err.Error())
		return
	}
	begin.Commit()
	t.Log(count)
}
```

## if 标签的使用

编写 xml,`QueryIf` 查询中使用了`where`标签，在`where`标签中，使用if来对上下文参数进行判断，如果存在 if标签将被解析到语句中

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<mapper namespace="StudentMapper">
    <!--  略 ..  -->
    <select id="QueryIf">
        select * from student
        <where>
            <if expr="{name!=nil}">
                name={name}
            </if>
        </where>
    </select>
</mapper>
```

定义mapper字段

```go
type StudentMapper struct {
	AddOne   func(student model.Student) error
	InsertId func(student model.Student) (int64, int64, error)
	Adds     func(ctx any) error

	QueryAll  func() ([]model.Student, error)
	QueryPage func() ([]model.Student, int64, error)

	Update func(student model.Student, tx *sql.Tx) (int64, error)

	QueryIf func(any) (model.Student, error)
}
```
运行测试
```go
func TestIf(t *testing.T) {
	var stu model.Student
	args := map[string]any{
		"id": 1,
		"name": "test_0",
	}
	if stu, err = studentMapper.QueryIf(args); err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(stu)
}
```

## 自定义映射数据

`GoBatis` 提供了上下文参数中复杂数据类型如何解析对应到 SQL 中对应的参数以及 SQL 中的查询结果集如何映射到自定义的复杂数据类型中。

### Go参数解析到 SQL

```go
// ToDatabase mapper 中sql解析模板对应的复杂数据据类型解析器
// data : 对应的数据本身
// 对应需要返回一个非结构体的基础数据类型（int float，bool，string） 更具需要构成的实际sql决定，后续的sql解析将自动匹配数据类
type ToDatabase func(data any) (any, error)

// DatabaseType 对外提供添加 自定义sql语句数据类型解析支持
func DatabaseType(key string, dataType ToDatabase) 
```

需要注册一个 `ToDatabase` 的解析器，`GoBatis`中对时间类型做了内置支持如下

```go
func ToDatabaseTime(data any) (any, error) {
	t := data.(time.Time)
	return t.Format("2006-01-02 15:04:05"), nil
}

func ToDatabaseTimePointer(data any) (any, error) {
	t := data.(*time.Time)
	return t.Format("2006-01-02 15:04:05"), nil
}
```

### SQL结果集解析到Go

```go
// ToGolang 处理数据库从查询结果集中的复杂数据类型的赋值
// value : 是在一个结构体内的字段反射，通过该函数可以对这个字段进行初始化赋值
// data  : 是value对应的具体参数值，可能是字符串，切片，map
type ToGolang func(value reflect.Value, data any) error

// GolangType 对外提供添加 自定义结果集数据类型解析支持
// key 需要通过 TypeKey 函数获取一个全局唯一的标识符
// dataType 需要提供 对应数据解析逻辑细节可以参考 TimeData 或者 TimeDataPointer
func GolangType(key string, dataType ToGolang)
```

需要注册一个 `ToGolang` 的解析器，`GoBatis`中对时间类型做了内置支持如下

```go
// TimeData 时间类型数据
func TimeData(value reflect.Value, data any) error {
	t := data.(string)
	parse, err := time.Parse("2006-04-02 15:04:05", t)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(parse))
	return nil
}

func TimeDataPointer(value reflect.Value, data any) error {
	t := data.(string)
	parse, err := time.Parse("2006-04-02 15:04:05", t)
	if err != nil {
		return err
	}
	if value.CanSet() {
		value.Set(reflect.ValueOf(&parse))
	}
	return nil
}
```