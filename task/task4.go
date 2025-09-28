package task

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // 导入 sqlite3 驱动
)

/*
 题目2：实现类型安全映射
- 假设有一个 books 表，包含字段 id 、 title 、 author 、 price 。
- 要求 ：
- 定义一个 Book 结构体，包含与 books 表对应的字段。
- 编写Go代码，使用Sqlx执行一个复杂的查询，例如查询价格大于 50 元的书籍，
并将结果映射到 Book 结构体切片中，确保类型安全。
*/

// Book 结构体定义，用于映射 books 表。
// `db` 标签确保了数据库列名和 Go 结构体字段之间的正确映射。
type Book struct {
	ID     int     `db:"id"`
	Title  string  `db:"title"`
	Author string  `db:"author"`
	Price  float64 `db:"price"`
}

// Task4 演示了如何进行类型安全的查询和映射。
func Task4() {
	// 1. 使用 sqlx.Connect 连接到内存中的 SQLite 数据库，方便测试。
	db, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 2. 初始化数据库：创建表并插入一些示例数据。
	setupBookDatabase(db)

	// 3. 执行查询：查询价格大于 50 的书籍。
	fmt.Println("--- 查询价格大于 50 元的书籍 ---")
	var expensiveBooks []Book
	priceThreshold := 50.0

	// db.Select 会执行查询，并将结果安全地映射到 `expensiveBooks` 切片中。
	// sqlx 在此过程中会检查数据库列类型和结构体字段类型是否兼容，
	// 如果不兼容，将返回一个错误，从而保证了类型安全。
	err = db.Select(&expensiveBooks, "SELECT * FROM books WHERE price > ?", priceThreshold)
	if err != nil {
		log.Fatalf("查询书籍失败: %v", err)
	}

	// 4. 打印查询结果。
	if len(expensiveBooks) == 0 {
		fmt.Println("没有找到价格大于 50 元的书籍。")
	} else {
		fmt.Println("查询结果:")
		for _, book := range expensiveBooks {
			fmt.Printf("ID: %d, 书名: %s, 作者: %s, 价格: %.2f\n", book.ID, book.Title, book.Author, book.Price)
		}
	}
}

// setupBookDatabase 用于创建 books 表并插入示例数据。
func setupBookDatabase(db *sqlx.DB) {
	schema := `
    CREATE TABLE books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        author TEXT,
        price REAL
    );`
	// MustExec 用于执行必须成功的 SQL，失败则 panic。
	db.MustExec(schema)

	// 插入一些示例书籍数据。
	books := []Book{
		{Title: "Go 语言编程", Author: "Alan A. A. Donovan", Price: 89.0},
		{Title: "深入理解计算机系统", Author: "Randal E. Bryant", Price: 128.0},
		{Title: "代码整洁之道", Author: "Robert C. Martin", Price: 45.5},
		{Title: "算法导论", Author: "Thomas H. Cormen", Price: 150.0},
		{Title: "Effective Java", Author: "Joshua Bloch", Price: 49.9},
	}

	// 使用 NamedExec 批量插入数据，代码更简洁。
	_, err := db.NamedExec(`INSERT INTO books (title, author, price) VALUES (:title, :author, :price)`, books)
	if err != nil {
		log.Fatalf("插入示例数据失败: %v", err)
	}
	fmt.Println("数据库初始化成功，已插入示例书籍数据。")
}
