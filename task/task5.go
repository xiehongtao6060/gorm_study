package task

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
题目1：模型定义
假设你要开发一个博客系统，有以下几个实体： User （用户）、 Post （文章）、 Comment （评论）。
要求 ：
- 使用Gorm定义 User 、 Post 和 Comment 模型，其中 User 与 Post 是一对多关系（一个用户可以发布多篇文章）， Post 与 Comment 也是一对多关系（一篇文章可以有多个评论）。
- 编写Go代码，使用Gorm创建这些模型对应的数据库表。

题目2：关联查询
基于上述博客系统的模型定义。
要求 ：
- 编写Go代码，使用Gorm查询某个用户发布的所有文章及其对应的评论信息。
- 编写Go代码，使用Gorm查询评论数量最多的文章信息。

题目3：钩子函数
继续使用博客系统的模型。
要求 ：
- 为 Post 模型添加一个钩子函数，在文章创建时自动更新用户的文章数量统计字段。
- 为 Comment 模型添加一个钩子函数，在评论删除时检查文章的评论数量，如果评论数量为 0，则更新文章的评论状态为 "无评论"。
*/

// --- 题目1：模型定义 ---

// User 用户模型
type User struct {
	gorm.Model
	Name      string
	PostCount int    // 用于题目3：统计用户文章数量
	Posts     []Post // 一对多关系：一个用户可以有多篇文章
}

// Post 文章模型
type Post struct {
	gorm.Model
	Title         string
	Content       string
	UserID        uint      // 外键，关联到 User
	CommentStatus string    `gorm:"default:'有评论'"` // 用于题目3
	Comments      []Comment // 一对多关系：一篇文章可以有多条评论
}

// Comment 评论模型
type Comment struct {
	gorm.Model
	Content string
	PostID  uint // 外键，关联到 Post
}

// --- 题目3：钩子函数 ---

// BeforeCreate Post模型的钩子函数，在创建文章前触发
func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	fmt.Println("[Hook] BeforeCreate Post: 准备更新用户文章数量...")
	// 使用 tx.Model(&User{}) 来在当前事务中更新 User 表
	// 找到对应的用户，并将其 PostCount 字段加 1
	err = tx.Model(&User{}).Where("id = ?", p.UserID).UpdateColumn("post_count", gorm.Expr("post_count + 1")).Error
	if err != nil {
		fmt.Printf("[Hook] BeforeCreate Post: 更新用户文章数量失败, %v\n", err)
	} else {
		fmt.Println("[Hook] BeforeCreate Post: 更新用户文章数量成功!")
	}
	return
}

// AfterDelete Comment模型的钩子函数，在删除评论后触发
func (c *Comment) AfterDelete(tx *gorm.DB) (err error) {
	fmt.Printf("[Hook] AfterDelete Comment: 评论 (ID: %d) 已删除，检查文章 (PostID: %d) 的剩余评论数...\n", c.ID, c.PostID)
	var count int64
	// 统计该文章还剩下多少评论
	tx.Model(&Comment{}).Where("post_id = ?", c.PostID).Count(&count)
	fmt.Printf("[Hook] AfterDelete Comment: 文章 (PostID: %d) 剩余 %d 条评论。\n", c.PostID, count)

	// 如果评论数量为0，则更新文章状态
	if count == 0 {
		fmt.Printf("[Hook] AfterDelete Comment: 评论已清空，更新文章 (PostID: %d) 状态为 '无评论'。\n", c.PostID)
		err = tx.Model(&Post{}).Where("id = ?", c.PostID).Update("comment_status", "无评论").Error
		if err != nil {
			fmt.Printf("[Hook] AfterDelete Comment: 更新文章状态失败, %v\n", err)
		}
	}
	return
}

// Task5 是所有问题的统一解决方案入口
func Task5() {
	// 使用内存数据库进行演示
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// --- 题目1：创建表 ---
	fmt.Println("--- 题目1：模型定义与数据库表创建 ---")
	// 自动迁移，根据模型创建数据库表
	err = db.AutoMigrate(&User{}, &Post{}, &Comment{})
	if err != nil {
		log.Fatalf("数据库表迁移失败: %v", err)
	}
	fmt.Println("数据库表 User, Post, Comment 创建成功！")

	// 准备一些测试数据
	prepareData(db)

	// --- 题目2：关联查询 ---
	fmt.Println("\n--- 题目2：关联查询 ---")
	// 2.1 查询某个用户的所有文章及其评论
	fmt.Println("\n--- 2.1 查询用户 'Alice' 的所有文章和评论 ---")
	var user User
	// 使用 Preload("Posts.Comments") 来预加载文章以及文章下的评论
	db.Preload("Posts.Comments").First(&user, "name = ?", "Alice")
	fmt.Printf("查询到用户: %s (文章数: %d)\n", user.Name, len(user.Posts))
	for _, post := range user.Posts {
		fmt.Printf("  - 文章: '%s' (评论数: %d)\n", post.Title, len(post.Comments))
		for _, comment := range post.Comments {
			fmt.Printf("    - 评论: '%s'\n", comment.Content)
		}
	}

	// 2.2 查询评论数量最多的文章
	fmt.Println("\n--- 2.2 查询评论数量最多的文章 ---")
	var mostCommentedPost Post
	// 使用子查询来计算每个文章的评论数，并按此排序
	db.Model(&Post{}).
		Select("posts.*, (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id) as comment_count").
		Order("comment_count DESC").
		First(&mostCommentedPost)
	fmt.Printf("评论最多的文章是: '%s' (ID: %d)\n", mostCommentedPost.Title, mostCommentedPost.ID)

	// --- 题目3：钩子函数演示 ---
	fmt.Println("\n--- 题目3：钩子函数演示 ---")
	// 3.1 演示 Post 的 BeforeCreate 钩子
	fmt.Println("\n--- 3.1 演示创建文章时触发钩子 ---")
	var charlie User
	db.First(&charlie, "name = ?", "Charlie")
	fmt.Printf("创建文章前, 用户 '%s' 的 PostCount: %d\n", charlie.Name, charlie.PostCount)
	// 创建一篇新文章，这将触发 BeforeCreate 钩子
	db.Create(&Post{Title: "Charlie 的新文章", Content: "...", UserID: charlie.ID})
	// 重新查询 Charlie 的信息以验证 PostCount 是否已更新
	db.First(&charlie, "name = ?", "Charlie")
	fmt.Printf("创建文章后, 用户 '%s' 的 PostCount: %d\n", charlie.Name, charlie.PostCount)

	// 3.2 演示 Comment 的 AfterDelete 钩子
	fmt.Println("\n--- 3.2 演示删除评论时触发钩子 ---")
	// 找到只有一条评论的文章 "GORM 基础"
	var postWithOneComment Post
	db.Preload("Comments").First(&postWithOneComment, "title = ?", "GORM 基础")
	fmt.Printf("删除评论前, 文章 '%s' 的状态是: '%s', 评论数: %d\n", postWithOneComment.Title, postWithOneComment.CommentStatus, len(postWithOneComment.Comments))
	// 删除这条唯一的评论
	db.Delete(&postWithOneComment.Comments[0])
	// 重新查询文章信息，验证状态是否已更新
	db.First(&postWithOneComment, postWithOneComment.ID)
	fmt.Printf("删除评论后, 文章 '%s' 的状态是: '%s'\n", postWithOneComment.Title, postWithOneComment.CommentStatus)
}

// prepareData 用于向数据库填充一些初始数据
func prepareData(db *gorm.DB) {
	fmt.Println("\n--- 准备测试数据 ---")
	// 创建用户
	users := []User{
		{Name: "Alice"},
		{Name: "Bob"},
		{Name: "Charlie"},
	}
	db.Create(&users)

	// 为 Alice 创建文章和评论
	post1 := Post{Title: "GORM 探索", Content: "...", UserID: users[0].ID}
	db.Create(&post1)
	db.Create(&Comment{Content: "写得好！", PostID: post1.ID})
	db.Create(&Comment{Content: "学习了。", PostID: post1.ID})
	db.Create(&Comment{Content: "期待续集。", PostID: post1.ID})

	post2 := Post{Title: "Go 语言技巧", Content: "...", UserID: users[0].ID}
	db.Create(&post2)
	db.Create(&Comment{Content: "非常实用！", PostID: post2.ID})

	// 为 Bob 创建文章
	post3 := Post{Title: "GORM 基础", Content: "...", UserID: users[1].ID}
	db.Create(&post3)
	db.Create(&Comment{Content: "入门好文。", PostID: post3.ID}) // 这篇文章只有一条评论，用于测试钩子

	fmt.Println("测试数据创建完成！")
}
