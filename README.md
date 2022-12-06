# GoBatis
[![Go Report Card](https://goreportcard.com/badge/gitee.com/aurora-engine/sgo)](https://goreportcard.com/report/gitee.com/aurora-engine/sgo)<br>
## version
```shell
go1.19
```
`GoBatis` 是参考 `MyBatis` 编写的sql标签解析，`GoBatis`仅提供对 mapper 的上下文数据解析填充，并不保证对 sql 语句的语法检查。
## XML 解析规则
`GoBatis` 解析 xml 文件中的sql语句，会严格检查上下文中的数据类型，字符串类型参数会自定添加 ` '' ` 单引号，其他基础数据类型不会添加，对于复杂数据结构(复合结构，泛型结构体等)会持续跟进
，目前仅支持基础数据类型。
### 上下文数据
上下文数据是由用户调用时候传递接，仅接受 map 或者结构体如下：
### 标签详情
|标签|描述|功能|
|:-|:-|:-|
|`<mapper>`|根节点||
|`<insert>`|insert语句|生成插入语句|
|`<select>`|select语句|生成查询语句|
|`<update>`|update语句|生成更新语句|
|`<delete>`|delete语句|生成删除语句|
|`<for>`|for迭代|生成IN语句，指定需要生成IN条件的字段，可以生成对应的IN条件|
|`<if>`|if条件|判断是否满足属性表达式的条件，满足条件就对标签内的sql进行解析|

## demo

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE mapper SYSTEM "http://aurora-engine.com/GoBatis.dtd">

<mapper namespace="user">
    <select id="find">
        select * from student where sss={name}
        <if expr="{arr}!=nil and {len(arr)}>0">
            and
            <for slice="{arr}" item="obj" column="id" open="(" separator="," close=")">
                {obj}
            </for>
        </if>

        <if expr="{name}=='aaa'">
            and abc = 1
            <if expr="1==1">
                and 1=1
                <if expr="1!=1">
                    or 1!=1
                </if>
            </if>
            or cba=1
        </if>
        or name = {name} and 1=1
    </select>
</mapper>
```
### xml解析
#### 第一层
`<mapper>`标签是整个xml的根 `namespace` 属性定义了 xml的标识符，调用阶段 `namespace`的属性至关重要
#### 第二层
`<select>`标签定义了 `id` 属性， `id` 属性是唯一标识，结合 `namespace` 能够定位，标签内的所有 `{xx}` 数据都来自于上下文数据，`{xx}` 将被解析成为具体的值
#### 第三层
`<if>` 标签 定义了 `expr` 属性， `expr` 属性的值为一串表达式，表达式应返回一个 `true` 或者 `false`，表示 `<if>` 标签内的内容是否可以被解析，表达式中使用到上下文数据可以通过点直接调用属性(注意属性名不要和关键字同名)

## 定义 Mapper
`GoBatis` 中的 `mapper` 定义是基于结构体 和匿名函数字段来实现的(匿名函数字段，需要遵循一些规则):

- 上下文参数，只能是结构体，指针结构体或者map
- 至少有一个返回值，一个返回值只能是 error

## 快速入门

### 创建 table
创建一张表，用于测试
```sql
## 用户设计
create table comm_user(
    user_id varchar(50) primary key         comment '主键',
    user_account varchar(50)                comment '账号',
    user_email varchar(50)                  comment '邮箱',
    user_password varchar(200)              comment '密码',
    user_name varchar(50) not null          comment '昵称',
    user_age int default 0                  comment '年龄',
    user_birthday datetime                  comment '生日',
    user_head_picture varchar(100)          comment '头像',
    user_create_time timestamp              comment '用户创建时间'
) comment '用户设计';
```
### 创建 映射模型
更具 数据库表，或者sql查询结果集 创建一个结构体用于接收查询数据
```go
// UserModel 用户模型
type UserModel struct {
	UserId          string `column:"user_id"`
	UserAccount     string `column:"user_account"`
	UserEmail       string `column:"user_email"`
	UserPassword    string `column:"user_password"`
	UserName        string `column:"user_name"`
	UserAge         int    `column:"user_age"`
	UserBirthday    string `column:"user_birthday"`
	UserHeadPicture string `column:"user_head_picture"`
	UserCreateTime  string `column:"user_create_time"`
}
```

### 创建 Mapper
更具上述的规范，创建 `Mapper`
```go
// UserMapper s
type UserMapper struct {
	FindUser   func(ctx any) (UserModel, error)
	UserSelect func(ctx any) (map[string]any, error)
}
```

### 创建 XML

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE mapper
        SYSTEM "http://aurora-engine.com/GoBatis.dtd">

<mapper namespace="UserMapper">
    <select id="FindUser">
        select * from comm_user where user_id={id}
    </select>
    <select id="UserSelect">
        select * from comm_user where user_id={id}
    </select>
</mapper>
```
创建的 xml 文件 `<mapper>` `namespace` 属性一定要和 Mapper 结构体的名称一样，区分大小写，`<select>` `id` 属性要和 Mapper 结构体 函数字段名称匹配。

### 创建并使用
```go
package main

import (
	"fmt"
	"gitee.com/aurora-engine/gobatis"
)
func main() {
	ctx := map[string]any{
		"id": "3de784d9a29243cdbe77334135b8a282",
	}
	open, err := sql.Open("mysql", "root:xxx@2022@tcp(82.xx.xx.xx:xx)/xx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	build := sgo.New(open)
	build.Source("/")
	mapper := &UserMapper{}
	build.ScanMappers(mapper)
	user, err := mapper.FindUser(ctx)
	if err != nil {
		return
	}
	fmt.Println(user)
}
```

## 创建数据库表
创建一个学生表
```sql
create table student
(
    id          int         null,
    name        varchar(20) null,
    age         int         null,
    create_time datetime    null
);
```

## 创建映射对象
对应在go代码中创建表的对应映射结构, `column` tag 设置字段的一一对应关系
```go
type Student struct {
	Id         string `column:"id"`
	Name       string `column:"name"`
	Age        int    `column:"age"`
	CreateTime string `column:"create_time"`
}
```

## 创建 Mapper 和 XML

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE mapper SYSTEM "http://aurora-engine.com/GoBatis.dtd">
<mapper namespace="StudentMapper">

</mapper>
```
解析： `http://aurora-engine.com/GoBatis.dtd` 文档约束是通过 编辑器设置的，项目文件夹下的 GoBatis.dtd 文件导入即可。<br>
更具 mapper xml 文件定义的 命名空间定义一个结构体类型名称一致的 Mapper 结构体（和普通结构体没什么区别只是一个叫法）
```go
type StudentMapper struct {
	
}
```

前置工作已经准备就绪，xml 和 mapper 结构体里面的内容会在下面的案例中一步一步的添加进去。

## Insert
### Insert 插入数据
对学生表进行新增数据,我们先从定义 mapper 函数开始。
```go
type StudentMapper struct {
	InsertOne func(any) (int, error)
}
```
开始定义 xml 元素，insert 中的模板参数，均来自于 mapper 函数的上下文参数中

```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
<!DOCTYPE mapper SYSTEM "http://aurora-engine.com/GoBatis.dtd">

<mapper namespace="StudentMapper">
    <insert id="InsertOne">
        insert into student(id,name,age,create_time) value({id},{name},{age},{time})
    </insert>
</mapper>
```
创建 GoBatis 并调用执行
```go
package main

import (
	"database/sql"
	"fmt"
	"gitee.com/aurora-engine/gobatis"
	_ "github.com/go-sql-driver/mysql"
	"time"
)
type Student struct {
	Id         string `column:"id"`
	Name       string `column:"name"`
	Age        int    `column:"age"`
	CreateTime string `column:"create_time"`
}

type StudentMapper struct {
	InsertOne func(any) (int64, error)
}

func main() {
	ctx := map[string]any{
		"id":   "1",
		"name": "test1",
		"age":  19,
		"time": time.Now().Format("2006-01-02 15:04:05"),
	}
	open, err := sql.Open("mysql", "xxxx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err != nil {
		return
	}
	build := gobatis.New(open)
	build.Source("/")
	mapper := &StudentMapper{}
	build.ScanMappers(mapper)
	count, err := mapper.InsertOne(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(count)
}
```
### 批量插入数据
我们现在继续向 Mapper 结构体中添加定义
```go
type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)
}
```
我们添加了 `InsertArr func(any) (int64, error)` 字段定义，在xml里面我们编写对应的sql语句,`<insert>`.
```xml
<mapper namespace="StudentMapper">
    <insert id="InsertOne">
        insert into student(id,name,age,create_time) value({id},{name},{age},{time})
    </insert>

    <insert id="InsertArr">
        insert into student(id,name,age,create_time) values
        <for slice="{arr}" item="obj">
            ({obj.id},{obj.name},{obj.age},{obj.time})
        </for>
    </insert>

</mapper>
```
`arr` 是上下文中的属性，`obj` 是作为 for 标签内的上下文数据，for 内是无法使用全局赏析问数据的。编写代码执行批量插入。
```go
package main

import (
	"database/sql"
	"fmt"
	"gitee.com/aurora-engine/gobatis"
	_ "github.com/go-sql-driver/mysql"
	"time"
)
type Student struct {
	Id         string `column:"id"`
	Name       string `column:"name"`
	Age        int    `column:"age"`
	CreateTime string `column:"create_time"`
}

type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)
}

func main() {
	ctx := map[string]any{
		"arr": []map[string]any{
			{
				"id":   "1",
				"name": "test1",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
			{
				"id":   "2",
				"name": "test2",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
			{
				"id":   "3",
				"name": "test3",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
		},
	}
	open, err := sql.Open("mysql", "xxxxx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	build := gobatis.New(open)
	build.Source("/")
	mapper := &StudentMapper{}
	build.ScanMappers(mapper)
	count, err := mapper.InsertArr(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(count)
}
```
::: tip
Insert,Update,Delete，定义的返回值只能返回数据库处理记录，第一个参数返回类型不正确将会返回错误信息，Insert 相对特殊，第二个参数可以返回，自增长id。
:::

## Select
### 查询一条记录
添加 查询定义如下：
```go
type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)

	SelectById func(any) (Student, error)
}
```
```xml
<select id="SelectById">
        select * from student where id={id}
</select>
```
```go
package main

import (
	"database/sql"
	"fmt"
	"gitee.com/aurora-engine/gobatis"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type Student struct {
	Id         string `column:"id"`
	Name       string `column:"name"`
	Age        int    `column:"age"`
	CreateTime string `column:"create_time"`
}

type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)

	SelectById  func(any) (Student, error)
}

func main() {
	ctx := map[string]any{
		"arr": []map[string]any{
			{
				"id":   "1",
				"name": "test1",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
			{
				"id":   "2",
				"name": "test2",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
			{
				"id":   "3",
				"name": "test3",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
		},
		"id":  "1",
		"ids": []string{"1", "2"},
	}
	open, err := sql.Open("mysql", "xxxx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	build := gobatis.New(open)
	build.Source("/")
	mapper := &StudentMapper{}
	build.ScanMappers(mapper)
	stu, err := mapper.SelectById(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(stu)
}

```

### 查询多条数据
```go
type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)

	SelectById func(any) (Student, error)
	SelectAll   func() ([]Student, error)
}
```
```xml
<select id="SelectAll">
    select * from student
</select>
```
```go
package main

import (
	"database/sql"
	"fmt"
	"gitee.com/aurora-engine/gobatis"
	_ "github.com/go-sql-driver/mysql"
	"time"
)
type Student struct {
	Id         string `column:"id"`
	Name       string `column:"name"`
	Age        int    `column:"age"`
	CreateTime string `column:"create_time"`
}

type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)

	SelectById  func(any) (Student, error)
	SelectAll   func() ([]Student, error)
}

func main() {
	open, err := sql.Open("mysql", "xxxx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	build := gobatis.New(open)
	build.Source("/")
	mapper := &StudentMapper{}
	build.ScanMappers(mapper)
	stu, err := mapper.SelectAll()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(stu)
}

```

### 批量查询
```go
type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)

	SelectById func(any) (Student, error)
	SelectAll   func() ([]Student, error)
	SelectByIds func(any) ([]Student, error)
}
```
```xml
<select id="SelectByIds">
    select * from student where id in
    <for slice="{ids}" item="id" open="(" separator="," close=")">
        {id}
    </for>
</select>
```
```go
package main

import (
	"database/sql"
	"fmt"
	"gitee.com/aurora-engine/gobatis"
	_ "github.com/go-sql-driver/mysql"
	"time"
)
type Student struct {
	Id         string `column:"id"`
	Name       string `column:"name"`
	Age        int    `column:"age"`
	CreateTime string `column:"create_time"`
}

type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)

	SelectById  func(any) (Student, error)
	SelectAll   func() ([]Student, error)
	SelectByIds func(any) ([]Student, error)
}

func main() {
	ctx := map[string]any{
		"arr": []map[string]any{
			{
				"id":   "1",
				"name": "test1",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
			{
				"id":   "2",
				"name": "test2",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
			{
				"id":   "3",
				"name": "test3",
				"age":  19,
				"time": time.Now().Format("2006-01-02 15:04:05"),
			},
		},
		"id":  "1",
		"ids": []string{"1", "2"},
	}
	open, err := sql.Open("mysql", "xxxxxxx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	build := gobatis.New(open)
	build.Source("/")
	mapper := &StudentMapper{}
	build.ScanMappers(mapper)
	stu, err := mapper.SelectByIds(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(stu)
}
```

## Update
同上...

## Delete
同上...