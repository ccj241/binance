package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 初始化配置
	cfg := config.NewConfig()

	// 创建测试用户
	var testUser models.User
	if err := cfg.DB.Where("username = ?", "testuser").First(&testUser).Error; err != nil {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
		testUser = models.User{
			Username: "testuser",
			Password: string(hashedPassword),
		}
		cfg.DB.Create(&testUser)
		log.Println("创建测试用户: testuser")
	}

	// 创建模拟交易记录
	createMockTrades(cfg, testUser.ID)

	// 创建模拟订单历史
	createMockOrders(cfg, testUser.ID)

	// 创建模拟提币历史
	createMockWithdrawals(cfg, testUser.ID)

	log.Println("模拟数据创建完成!")
}

func createMockTrades(cfg *config.Config, userID uint) {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT"}

	// 为每个交易对创建一些交易记录
	for _, symbol := range symbols {
		// 创建10-20条交易记录
		numTrades := rand.Intn(10) + 10
		for i := 0; i < numTrades; i++ {
			// 随机生成过去30天内的时间
			hoursAgo := rand.Intn(720) // 30天 * 24小时
			tradeTime := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

			// 随机生成价格和数量
			var price, qty float64
			switch symbol {
			case "BTCUSDT":
				price = 40000 + rand.Float64()*20000 // 40000-60000
				qty = rand.Float64() * 0.1           // 0-0.1 BTC
			case "ETHUSDT":
				price = 2000 + rand.Float64()*2000 // 2000-4000
				qty = rand.Float64() * 1           // 0-1 ETH
			case "BNBUSDT":
				price = 300 + rand.Float64()*200 // 300-500
				qty = rand.Float64() * 5         // 0-5 BNB
			case "SOLUSDT":
				price = 50 + rand.Float64()*100 // 50-150
				qty = rand.Float64() * 10       // 0-10 SOL
			}

			trade := models.Trade{
				UserID:    userID,
				Symbol:    symbol,
				Price:     price,
				Qty:       qty,
				Time:      tradeTime.UnixMilli(),
				CreatedAt: tradeTime,
			}

			if err := cfg.DB.Create(&trade).Error; err != nil {
				log.Printf("创建交易记录失败: %v", err)
			}
		}

		log.Printf("为 %s 创建了 %d 条交易记录", symbol, numTrades)
	}
}

func createMockOrders(cfg *config.Config, userID uint) {
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT"}
	statuses := []string{"filled", "cancelled", "pending"}
	sides := []string{"BUY", "SELL"}

	// 创建20-30个订单
	numOrders := rand.Intn(10) + 20
	for i := 0; i < numOrders; i++ {
		symbol := symbols[rand.Intn(len(symbols))]
		status := statuses[rand.Intn(len(statuses))]
		side := sides[rand.Intn(len(sides))]

		// 随机生成时间
		hoursAgo := rand.Intn(168) // 过去7天
		orderTime := time.Now().Add(-time.Duration(hoursAgo) * time.Hour)

		// 随机生成价格和数量
		var price, qty float64
		switch symbol {
		case "BTCUSDT":
			price = 40000 + rand.Float64()*20000
			qty = rand.Float64() * 0.1
		case "ETHUSDT":
			price = 2000 + rand.Float64()*2000
			qty = rand.Float64() * 1
		case "BNBUSDT":
			price = 300 + rand.Float64()*200
			qty = rand.Float64() * 5
		case "SOLUSDT":
			price = 50 + rand.Float64()*100
			qty = rand.Float64() * 10
		}

		order := models.Order{
			UserID:      userID,
			Symbol:      symbol,
			Side:        side,
			Price:       price,
			Quantity:    qty,
			OrderID:     int64(1000000 + i), // 模拟订单ID
			Status:      status,
			CancelAfter: orderTime.Add(2 * time.Hour),
			CreatedAt:   orderTime,
			UpdatedAt:   orderTime,
		}

		// 如果是已完成或已取消的订单，更新时间应该更晚
		if status != "pending" {
			updateTime := orderTime.Add(time.Duration(rand.Intn(120)) * time.Minute)
			order.UpdatedAt = updateTime
		}

		if err := cfg.DB.Create(&order).Error; err != nil {
			log.Printf("创建订单失败: %v", err)
		}
	}

	log.Printf("创建了 %d 个订单", numOrders)
}

func createMockWithdrawals(cfg *config.Config, userID uint) {
	assets := []string{"BTC", "ETH", "USDT", "BNB"}
	statuses := []string{"6", "5", "4", "2", "1"} // 币安提币状态码

	// 创建5-10条提币记录
	numWithdrawals := rand.Intn(5) + 5
	for i := 0; i < numWithdrawals; i++ {
		asset := assets[rand.Intn(len(assets))]
		status := statuses[rand.Intn(len(statuses))]

		// 随机生成时间
		daysAgo := rand.Intn(90) // 过去90天
		withdrawalTime := time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour)

		// 随机生成金额
		var amount float64
		switch asset {
		case "BTC":
			amount = rand.Float64() * 0.5 // 0-0.5 BTC
		case "ETH":
			amount = rand.Float64() * 5 // 0-5 ETH
		case "USDT":
			amount = rand.Float64() * 10000 // 0-10000 USDT
		case "BNB":
			amount = rand.Float64() * 20 // 0-20 BNB
		}

		// 生成模拟地址
		address := fmt.Sprintf("0x%x", rand.Int63())
		if asset == "BTC" {
			address = fmt.Sprintf("bc1q%x", rand.Int63())
		}

		withdrawal := models.WithdrawalHistory{
			UserID:       userID,
			Asset:        asset,
			Amount:       amount,
			Address:      address,
			WithdrawalID: fmt.Sprintf("WD%d", 1000000+i),
			TxID:         fmt.Sprintf("0x%x", rand.Int63()),
			Status:       status,
			CreatedAt:    withdrawalTime,
			UpdatedAt:    withdrawalTime,
		}

		if err := cfg.DB.Create(&withdrawal).Error; err != nil {
			log.Printf("创建提币记录失败: %v", err)
		}
	}

	log.Printf("创建了 %d 条提币记录", numWithdrawals)
}
