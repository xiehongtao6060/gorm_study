package main

import (
	"fmt"
	"gorm_study/demo/constant"
	"gorm_study/task"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User 定义用户表
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
}

// Post 定义文章表
type Post struct {
	ID     uint   `gorm:"primaryKey"`
	Title  string `gorm:"column:title"`
	Body   string `gorm:"column:body"`
	UserID uint   `gorm:"column:user_id"`
}

// InitDB 初始化数据库
func InitDB() *gorm.DB {
	db := ConnectDB()
	err := db.AutoMigrate(&User{}, &Post{}, &task.Student{}, &task.Account{}, &task.Transaction{})
	if err != nil {
		panic(err)
	}
	return db
}

// ConnectDB 连接数据库
func ConnectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(constant.DBPATH), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func runDemo() {
	db := InitDB()
	fmt.Printf("使用的数据库文件：%s\n", constant.DBPATH)

	// 插入数据
	db.Create(&User{Name: "Alice", Email: "alice@example.com"})
	db.Create(&User{Name: "Bob", Email: "bob@example.com"})

	// 查询数据
	var users []User
	db.Find(&users)
	fmt.Println("所有用户：", users)

	// 更新数据
	db.Model(&users[0]).Update("Email", "alice@newdomain.com")

	// 删除数据
	db.Delete(&users[1])

	// 查询剩余
	var remaining []User
	db.Find(&remaining)
	fmt.Println("剩余用户：", remaining)

	// 🔥 关键：关闭连接，强制 flush，否则不会写入到硬盘
	sqlDB, err := db.DB()
	if err != nil {
		panic("获取底层数据库连接失败：" + err.Error())
	}
	err = sqlDB.Close()
	if err != nil {
		panic(err)
	}
}

func testTask1() {

}
func main() {
	//db := InitDB()
	task.Task3()
}
