package model

type Student struct {
	Id         int    `column:"id"json:"id,omitempty"`
	Name       string `column:"name"json:"name,omitempty"`
	Age        int    `column:"age"json:"age,omitempty"`
	CreateTime string `column:"create_time"json:"create_time,omitempty"`
}
