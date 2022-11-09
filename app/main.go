package main

import (
	"gitee.com/aurora-engine/sgo"
	"reflect"
)

// UserModel 用户模型
type UserModel struct {
	UserId          string `column:"user_id" gorm:"primaryKey;column:user_id;type:varchar" json:"userId,omitempty"`
	UserAccount     string `column:"user_account" gorm:"column:user_account;type:varchar" json:"userAccount,omitempty"`
	UserEmail       string `column:"user_email" gorm:"column:user_email;type:varchar" model:"false" json:"userEmail,omitempty"`
	UserPassword    string `column:"user_password" gorm:"column:user_password;type:varchar" model:"false" json:"userPassword,omitempty"`
	UserName        string `column:"user_name" gorm:"column:user_name;type:varchar" json:"userName,omitempty"`
	UserAge         int    `column:"user_age" gorm:"column:user_age;type:int" json:"userAge,omitempty"`
	UserBirthday    string `column:"user_birthday" gorm:"column:user_birthday;type:datetime" json:"userBirthday,omitempty"`
	UserHeadPicture string `column:"user_head_picture" gorm:"column:user_head_picture;type:varchar" json:"userHeadPicture,omitempty"`
	UserCreateTime  string `column:"user_create_time" gorm:"column:user_create_time;type:datetime" json:"userCreateTime,omitempty"`
}

type UserMapper struct {
	FindUser   func(ctx any) UserModel
	UserSelect func(ctx any) map[string]any
}

func main() {
	ctx := map[string]any{
		"id": "3de784d9a29243cdbe77334135b8a282",
	}
	sgo := sgo.NewSgo()
	sgo.LoadXml("/")
	mapper := UserMapper{}
	Init(&mapper, sgo)
	mapper.FindUser(ctx)
}

func Init(mapper any, build *sgo.Sgo) {
	of := reflect.ValueOf(mapper).Elem()
	namespace := of.Type().String()
	for i := 0; i < of.NumField(); i++ {
		key := make([]string, 0)
		key = append(key, namespace)
		structField := of.Type().Field(i)
		key = append(key, structField.Name)
		sgo.InitMapper(build, key, of.Field(i))
	}
}
