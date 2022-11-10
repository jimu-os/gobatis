package main

import (
	"database/sql"
	"fmt"
	"gitee.com/aurora-engine/sgo"
	_ "github.com/go-sql-driver/mysql"
)

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

// UserMapper s
type UserMapper struct {
	FindUser   func(ctx any) (string, error)
	UserSelect func(ctx any) (map[string]any, error)
}

func main() {
	ctx := map[string]any{
		"id": "3de784d9a29243cdbe77334135b8a282",
	}
	open, err := sql.Open("mysql", "root:Aurora@2022@tcp(82.157.160.117:3306)/community")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	build := sgo.New(open)
	build.LoadMapper("/")
	mapper := &UserMapper{}
	build.ScanMappers(mapper)
	user, err := mapper.FindUser(ctx)
	if err != nil {
		return
	}
	fmt.Println(user)

}
