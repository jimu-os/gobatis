# sgo

## 介绍
sgo 是 go 版轻量级 mybatis，sgo仅完成对 sql 标签的解析。

## 快速开始
### 创建 mapper 文件
项目根路径下创建xml文件 写入下面的内容
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
### 加载xml并使用
```go
package main

import (
	"fmt"
	"gitee.com/aurora-engine/sgo"
)

func main() {
	ctx := map[string]any{
		"arr":  []int{1, 2, 3, 4},
		"name": "test",
	}
	sgo := sgo.NewSgo()
	// 根路径
	sgo.LoadXml("/")
	sql, err := sgo.Sql("user.find", ctx)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(sql)
}
```