package task

import (
	"fmt"

	"gorm.io/gorm"
)

/*
题目1：基本CRUD操作
假设有一个名为 students 的表，包含字段 id （主键，自增）、 name （学生姓名，字符串类型）、 age （学生年龄，整数类型）、 grade （学生年级，字符串类型）。
要求 ：
编写SQL语句向 students 表中插入一条新记录，学生姓名为 "张三"，年龄为 20，年级为 "三年级"。
编写SQL语句查询 students 表中所有年龄大于 18 岁的学生信息。
编写SQL语句将 students 表中姓名为 "张三" 的学生年级更新为 "四年级"。
编写SQL语句删除 students 表中年龄小于 15 岁的学生记录。

*/
// Student 定义学生表
type Student struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"column:name"`
	Age   int    `gorm:"column:age"`
	Grade string `gorm:"column:grade"`
}

func RunTask1(db *gorm.DB) {
	// 插入数据
	db.Create(&Student{Name: "张三", Age: 20, Grade: "三年级"})

	// 查询数据
	var students []Student
	db.Where("age > ?", 18).Find(&students)
	fmt.Println("所有年龄大于 18 岁的学生：", students)

	// 更新数据
	db.Model(&students[0]).Update("Grade", "四年级")

	// 删除数据
	db.Where("age < ?", 15).Delete(&Student{})
}
