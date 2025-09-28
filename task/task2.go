package task

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

/*
题目2：事务语句
假设有两个表： accounts 表（包含字段 id 主键， balance 账户余额）
和 transactions 表（包含字段 id 主键， from_account_id 转出账户ID， to_account_id 转入账户ID， amount 转账金额）。
要求 ：
编写一个事务，实现从账户 A 向账户 B 转账 100 元的操作。
在事务中，需要先检查账户 A 的余额是否足够，如果足够则从账户 A 扣除 100 元，
向账户 B 增加 100 元，并在 transactions 表中记录该笔转账信息。如果余额不足，则回滚事务。
*/

// Account 使用单数命名，并优化了字段和标签
type Account struct {
	ID      uint            `gorm:"primaryKey"`
	Balance decimal.Decimal `gorm:"column:balance;type:decimal(10,2)"`
}

// Transaction 使用单数命名，并优化了字段和标签
type Transaction struct {
	ID            uint            `gorm:"primaryKey"`
	FromAccountID uint            `gorm:"column:from_account_id"`
	ToAccountID   uint            `gorm:"column:to_account_id"`
	Amount        decimal.Decimal `gorm:"column:amount;type:decimal(10,2)"`
}

// Transfer 封装了转账的核心逻辑，这是最佳实践
func Transfer(db *gorm.DB, fromID, toID uint, amount decimal.Decimal) error {
	// 1. 使用 GORM 的 Transaction 方法，它会自动处理提交和回滚
	return db.Transaction(func(tx *gorm.DB) error {
		// 检查转账金额是否为正数
		if amount.LessThanOrEqual(decimal.Zero) {
			return errors.New("转账金额必须为正数")
		}

		// 2. 锁定行以防止并发问题 (SELECT ... FOR UPDATE)
		// 在事务中查询转出和转入账户，并使用悲观锁锁定这两行
		var fromAccount, toAccount Account

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&fromAccount, fromID).Error; err != nil {
			return fmt.Errorf("找不到转出账户: %w", err)
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&toAccount, toID).Error; err != nil {
			return fmt.Errorf("找不到转入账户: %w", err)
		}

		// 3. 检查余额是否足够
		if fromAccount.Balance.LessThan(amount) {
			return errors.New("余额不足")
		}

		// 4. 更新账户余额
		fromAccount.Balance = fromAccount.Balance.Sub(amount)
		toAccount.Balance = toAccount.Balance.Add(amount)

		if err := tx.Save(&fromAccount).Error; err != nil {
			return fmt.Errorf("更新转出账户失败: %w", err)
		}
		if err := tx.Save(&toAccount).Error; err != nil {
			return fmt.Errorf("更新转入账户失败: %w", err)
		}

		// 5. 创建交易记录
		transaction := Transaction{
			FromAccountID: fromID,
			ToAccountID:   toID,
			Amount:        amount,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("创建交易记录失败: %w", err)
		}

		// 6. 如果到这里都没有错误，GORM 会自动提交事务
		// 如果任何一步返回了 error，GORM 会自动回滚
		return nil
	})
}

// Task2 负责执行题目要求的场景
func Task2(db *gorm.DB) {
	// 自动迁移，确保表已创建
	db.AutoMigrate(&Account{}, &Transaction{})

	// --- 准备初始数据 ---
	// 为了让示例可重复运行，我们先清空表，然后创建两个账户
	db.Exec("DELETE FROM accounts")
	db.Exec("DELETE FROM transactions")

	accounts := []Account{
		{ID: 1, Balance: decimal.NewFromInt(1000)}, // 账户1有1000元
		{ID: 2, Balance: decimal.NewFromInt(500)},  // 账户2有500元
	}
	db.Create(&accounts)
	fmt.Println("--- 初始状态 ---")
	fmt.Printf("账户1余额: %s\n", accounts[0].Balance)
	fmt.Printf("账户2余额: %s\n", accounts[1].Balance)
	fmt.Println("-----------------")

	// --- 场景一：成功转账 ---
	fmt.Println("\n>>> 尝试从账户1向账户2转账 100 元...")
	transferAmount := decimal.NewFromInt(100)
	err := Transfer(db, 1, 2, transferAmount)
	if err != nil {
		fmt.Printf("转账失败: %v\n", err)
	} else {
		fmt.Println("转账成功!")
	}

	// 查询并显示转账后的余额
	var acc1, acc2 Account
	db.First(&acc1, 1)
	db.First(&acc2, 2)
	fmt.Println("--- 转账后状态 ---")
	fmt.Printf("账户1余额: %s\n", acc1.Balance)
	fmt.Printf("账户2余额: %s\n", acc2.Balance)
	fmt.Println("-----------------")

	// --- 场景二：余额不足，事务回滚 ---
	fmt.Println("\n>>> 尝试从账户2向账户1转账 600 元 (余额不足)...")
	insufficientAmount := decimal.NewFromInt(600)
	err = Transfer(db, 2, 1, insufficientAmount)
	if err != nil {
		fmt.Printf("转账失败，事务已回滚: %v\n", err)
	} else {
		fmt.Println("转账成功!") // 这行不会被执行
	}

	// 再次查询余额，确认没有变化
	db.First(&acc1, 1)
	db.First(&acc2, 2)
	fmt.Println("--- 回滚后状态 ---")
	fmt.Printf("账户1余额: %s\n", acc1.Balance)
	fmt.Printf("账户2余额: %s\n", acc2.Balance)
	fmt.Println("-----------------")
}
