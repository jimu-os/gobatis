package main

import (
	"errors"
	"gitee.com/aurora-engine/sgo"
	"reflect"
	"strings"
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
	FindUser   func(ctx any) (UserModel, error)
	UserSelect func(ctx any) (map[string]any, error)
}

func main() {
	ctx := map[string]any{
		"id": "3de784d9a29243cdbe77334135b8a282",
	}
	sgo := sgo.New()
	sgo.LoadMapper("/")
	mapper := &UserMapper{}
	Init(mapper, sgo)
	mapper.FindUser(ctx)
}

func Init(mapper any, build *sgo.Build) {
	vf := reflect.ValueOf(mapper)
	if vf.Kind() != reflect.Pointer {
		panic("")
	}
	if vf.Elem().Kind() != reflect.Struct {
		panic("")
	}
	vf = vf.Elem()
	namespace := vf.Type().String()
	namespace = Namespace(namespace)
	for i := 0; i < vf.NumField(); i++ {
		key := make([]string, 0)
		key = append(key, namespace)
		structField := vf.Type().Field(i)
		if !structField.IsExported() {
			continue
		}
		if b, err := MapperCheck(vf.Field(i)); !b {
			panic(err)
		}
		key = append(key, structField.Name)
		sgo.InitMapper(build, key, vf.Field(i))
	}
}

func Namespace(namespace string) string {
	if index := strings.LastIndex(namespace, "."); index != -1 {
		return namespace[index+1:]
	}
	return namespace
}

// MapperCheck 检查 Mapper 函数是否符合规范
func MapperCheck(fun reflect.Value) (bool, error) {
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
