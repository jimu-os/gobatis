# sgo
[![Go Report Card](https://goreportcard.com/badge/gitee.com/aurora-engine/sgo)](https://goreportcard.com/report/gitee.com/aurora-engine/sgo)<br>
`sgo` 是参考 `mybatis` 编写的sql标签解析，`sgo`仅提供对 sql 的上下文数据解析填充，并不保证对 sql 语句的语法检查。
## XML 解析规则
`sgo` 解析 xml 文件中的sql语句，会严格检查上下文中的数据类型，字符串类型参数会自定添加 ` '' ` 单引号，其他基础数据类型不会添加，对于复杂数据结构(复合结构，泛型结构体等)会持续跟进
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
<!DOCTYPE mapper SYSTEM "https://gitee.com/aurora-engine/sgo/blob/master/sgo.dtd">

<mapper namespace="user">
    <select id="find">
        select * from student where sss={name}
        <if expr="{arr}!=nil and {len(arr)}>0">
            and
            <for slice="{arr}" item="obj" column="id" open="("  separator="," close=")" >
                {obj}
            </for>
        </if>

        <if expr="{name}=='aaa'" >
            and abc = 1
            <if expr="1==1">
                and 1=1
                <if expr="1!=1">
                    or 1!=1
                </if>
            </if>
            or cba=1
        </if>
        or  name = {name} and 1=1
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
`sgo` 中的 `mapper` 定义是基于结构体 和匿名函数字段来实现的(匿名函数字段，需要遵循一些规则):

- 只有一个入参，并且只能是结构体，指针结构体或者map
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
        SYSTEM "https://gitee.com/aurora-engine/sgo/blob/master/sgo.dtd">

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
	"gitee.com/aurora-engine/sgo"
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
