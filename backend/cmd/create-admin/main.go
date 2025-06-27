package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	// 命令行参数
	var (
		username       = flag.String("username", "", "管理员用户名")
		password       = flag.String("password", "", "管理员密码")
		nonInteractive = flag.Bool("non-interactive", false, "非交互模式")
	)
	flag.Parse()

	// 初始化配置
	cfg := config.NewConfig()

	// 确保数据库表已创建
	if err := models.MigrateDB(cfg.DB); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	var adminUsername, adminPassword string

	if *nonInteractive {
		// 非交互模式，从命令行参数获取
		if *username == "" || *password == "" {
			log.Fatal("非交互模式下必须提供 -username 和 -password 参数")
		}
		adminUsername = *username
		adminPassword = *password
	} else {
		// 交互模式
		fmt.Println("=== 创建管理员账号 ===")
		fmt.Println()

		// 获取用户名
		adminUsername = getInput("请输入管理员用户名", *username)

		// 获取密码
		if *password != "" {
			adminPassword = *password
			fmt.Println("使用命令行提供的密码")
		} else {
			adminPassword = getPassword("请输入管理员密码")
			confirmPassword := getPassword("请确认管理员密码")

			if adminPassword != confirmPassword {
				log.Fatal("两次输入的密码不匹配")
			}
		}
	}

	// 验证输入
	if len(adminUsername) < 3 {
		log.Fatal("用户名长度至少为3个字符")
	}

	if len(adminPassword) < 6 {
		log.Fatal("密码长度至少为6个字符")
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := cfg.DB.Where("username = ?", adminUsername).First(&existingUser).Error; err == nil {
		if !*nonInteractive {
			fmt.Printf("\n用户 '%s' 已存在，是否更新为管理员? (y/N): ", adminUsername)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("操作已取消")
				return
			}
		}

		// 更新现有用户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("密码加密失败: %v", err)
		}

		existingUser.Password = string(hashedPassword)
		existingUser.Role = "admin"
		existingUser.Status = "active"

		if err := cfg.DB.Save(&existingUser).Error; err != nil {
			log.Fatalf("更新用户失败: %v", err)
		}

		fmt.Printf("\n✅ 用户 '%s' 已成功更新为管理员\n", adminUsername)
	} else {
		// 创建新用户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("密码加密失败: %v", err)
		}

		adminUser := models.User{
			Username: adminUsername,
			Password: string(hashedPassword),
			Role:     "admin",
			Status:   "active",
		}

		if err := cfg.DB.Create(&adminUser).Error; err != nil {
			log.Fatalf("创建管理员用户失败: %v", err)
		}

		// 重要修改：确保角色确实被设置为 admin
		// 由于 GORM 可能有默认值或钩子函数，我们需要再次查询并更新
		var createdUser models.User
		if err := cfg.DB.Where("username = ?", adminUsername).First(&createdUser).Error; err != nil {
			log.Fatalf("查询新创建的用户失败: %v", err)
		}

		// 如果角色不是 admin，强制更新为 admin
		if createdUser.Role != "admin" {
			log.Printf("检测到角色不是 admin (%s)，正在修正...", createdUser.Role)
			if err := cfg.DB.Model(&createdUser).Updates(map[string]interface{}{
				"role":   "admin",
				"status": "active",
			}).Error; err != nil {
				log.Fatalf("更新用户角色失败: %v", err)
			}
		}

		// 再次验证角色是否正确设置
		var finalUser models.User
		if err := cfg.DB.Where("username = ?", adminUsername).First(&finalUser).Error; err != nil {
			log.Fatalf("最终验证用户失败: %v", err)
		}

		if finalUser.Role != "admin" {
			log.Fatalf("错误：无法将用户角色设置为 admin，当前角色为: %s", finalUser.Role)
		}

		fmt.Printf("\n✅ 管理员账号创建成功！\n")
		fmt.Printf("验证信息 - 用户ID: %d, 角色: %s, 状态: %s\n", finalUser.ID, finalUser.Role, finalUser.Status)
	}

	fmt.Printf("\n管理员用户名: %s\n", adminUsername)
	fmt.Printf("请妥善保管密码\n")
	fmt.Println("\n现在可以使用此账号登录系统")
}

// getInput 获取用户输入
func getInput(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)

	if defaultValue != "" {
		fmt.Printf("%s (默认: %s): ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("读取输入失败: %v", err)
	}

	input = strings.TrimSpace(input)

	if input == "" && defaultValue != "" {
		return defaultValue
	}

	return input
}

// getPassword 安全地获取密码输入（不显示在屏幕上）
func getPassword(prompt string) string {
	fmt.Printf("%s: ", prompt)

	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatalf("读取密码失败: %v", err)
	}

	fmt.Println() // 换行
	return string(password)
}
