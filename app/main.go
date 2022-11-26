package main

import (
	"database/sql"
	"fmt"
	"gitee.com/aurora-engine/sgo"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

// UserModel 用户模型
type UserModel struct {
	UserId          string     `column:"user_id"`
	UserAccount     string     `column:"user_account"`
	UserEmail       string     `column:"user_email"`
	UserPassword    string     `column:"user_password"`
	UserName        string     `column:"user_name"`
	UserAge         int        `column:"user_age"`
	UserBirthday    string     `column:"user_birthday"`
	UserHeadPicture string     `column:"user_head_picture"`
	UserCreateTime  *time.Time `column:"user_create_time"`
}

// UserMapper s
type UserMapper struct {
	Find       func(any) error
	FindUser   func(any) (UserModel, error)
	UserSelect func(any) (map[string]any, error)
}

type Student struct {
	Id         string `column:"id"`
	Name       string `column:"name"`
	Age        int    `column:"age"`
	CreateTime string `column:"create_time"`
}

type StudentMapper struct {
	InsertOne func(any) (int64, error)
	InsertArr func(any) (int64, error)

	SelectById  func(any) (error, Student)
	SelectAll   func() ([]Student, error)
	SelectByIds func(any) ([]Student, error)
}

func main() {
	now := time.Now()
	ctx := map[string]any{
		"arr": []map[string]any{
			{
				"id":   "22",
				"name": "test22",
				"age":  22,
				"time": &now,
			},
		},
		"id":  "1",
		"ids": []string{"1", "2", "3", "4", "5", "6"},
	}
	open, err := sql.Open("mysql", "xxxx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	build := sgo.New(open)
	build.Source("/")
	mapper := &StudentMapper{}
	build.ScanMappers(mapper)
	count, err := mapper.SelectByIds(ctx)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(count)
}
