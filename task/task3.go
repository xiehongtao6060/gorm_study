package task

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // 导入 sqlite3 驱动
)

/*
题目1：使用SQL扩展库进行查询
假设你已经使用Sqlx连接到一个数据库，并且有一个 employees 表，包含字段 id 、 name 、 department 、 salary 。
要求 ：
1. 编写Go代码，使用Sqlx查询 employees 表中所有部门为 "技术部" 的员工信息，并将结果映射到一个自定义的 Employee 结构体切片中。
2. 编写Go代码，使用Sqlx查询 employees 表中工资最高的员工信息，并将结果映射到一个 Employee 结构体中。
*/

// Employee 结构体用于映射数据库中的 employees 表
// `db` 标签是 sqlx 用于将列名映射到结构体字段的关键
type Employee struct {
	ID         int     `db:"id"`
	Name       string  `db:"name"`
	Department string  `db:"department"`
	Salary     float64 `db:"salary"`
}

// Task3 包含了完整的解题步骤
func Task3() {
	// 1. 连接数据库
	// 为了方便演示，我们使用内存中的 SQLite 数据库。
	// ":memory:" 表示数据只存在于本次运行的内存中，程序结束即消失。
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 2. 准备数据：创建表并插入一些示例数据
	setupDatabase(db)

	fmt.Println("--- 问题1：查询所有'技术部'���员工 ---")
	// 3. 使用 db.Select 查询多行数据
	var techEmployees []Employee
	err = db.Select(&techEmployees, "SELECT * FROM employees WHERE department = ?", "技术部")
	if err != nil {
		log.Fatalf("查询技术部员工失败: %v", err)
	}

	fmt.Println("查询结果:")
	for _, emp := range techEmployees {
		fmt.Printf("ID: %d, 姓名: %s, 部门: %s, 工资: %.2f\n", emp.ID, emp.Name, emp.Department, emp.Salary)
	}

	fmt.Println("\n--- 问题2：查询工资最高的员工 ---")
	// 4. 使用 db.Get 查询单行数据
	var highestPaidEmployee Employee
	err = db.Get(&highestPaidEmployee, "SELECT * FROM employees ORDER BY salary DESC LIMIT 1")
	if err != nil {
		log.Fatalf("查询工资最高的员工失败: %v", err)
	}

	fmt.Println("查询结果:")
	fmt.Printf("ID: %d, 姓名: %s, 部门: %s, 工资: %.2f\n", highestPaidEmployee.ID, highestPaidEmployee.Name, highestPaidEmployee.Department, highestPaidEmployee.Salary)
}

// setupDatabase 用于创建表结构并插入测试数据
func setupDatabase(db *sqlx.DB) {
	schema := `
    CREATE TABLE employees (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT,
        department TEXT,
        salary REAL
    );`
	// MustExec 在执行失败时会直接 panic，适用于初始化阶段
	db.MustExec(schema)

	// 插入一些示例员工数据
	employees := []Employee{
		{Name: "Alice", Department: "技术部", Salary: 8000},
		{Name: "Bob", Department: "技术部", Salary: 9500},
		{Name: "Charlie", Department: "市场部", Salary: 6000},
		{Name: "David", Department: "技术部", Salary: 8500},
		{Name: "Eve", Department: "财务部", Salary: 7000},
	}

	// 使用 NamedExec 可以方便地通过结构体字段名进行批量插入
	_, err := db.NamedExec(`INSERT INTO employees (name, department, salary) VALUES (:name, :department, :salary)`, employees)
	if err != nil {
		log.Fatalf("插入示例数据失败: %v", err)
	}
	fmt.Println("数据库初始化成功，已插入示例数据。")
}
