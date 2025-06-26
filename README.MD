# 币安自动交易系统

一个功能齐全的加密货币自动交易系统，支持多种交易策略、双币投资、自动提币等功能。

## 功能特性

### 核心功能
- 🔐 **用户认证与授权**：JWT认证，支持管理员和普通用户角色
- 📊 **实时价格监控**：WebSocket实时获取币价，支持多交易对同时监控
- 🎯 **多种交易策略**：
    - 简单策略：单点位下单
    - 冰山策略：自动分层下单
    - 自定义策略：支持深度级别和万分比偏移配置
- 📈 **订单管理**：自动下单、订单状态跟踪、批量取消
- 💰 **双币投资**：
    - 支持单次投资、自动复投、梯度投资、价格触发等策略
    - 实时产品同步和收益计算
    - 投资统计和历史记录
- 🏦 **自动提币**：设置阈值自动提币到指定地址
- 👥 **用户管理**：管理员可审核用户、管理权限

### 安全特性
- API密钥加密存储（AES-256-GCM）
- 用户密码bcrypt加密
- JWT Token认证
- 请求频率限制
- 并发控制机制

## 技术栈

### 后端
- **语言**：Go 1.24.4
- **Web框架**：Gin
- **数据库**：MySQL (GORM)
- **API集成**：币安官方API
- **WebSocket**：实时价格推送
- **认证**：JWT

### 前端
- **框架**：Vue 3
- **构建工具**：Vite
- **路由**：Vue Router
- **HTTP客户端**：Axios
- **UI设计**：自定义CSS变量系统

## 快速开始

### 环境要求
- Go 1.24.4+
- Node.js 16+
- MySQL 5.7+
- 币安API密钥

### 后端安装

1. 克隆项目
```bash
git clone https://github.com/ccj241/binance.git
cd binance/backend
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境变量
```bash
export DATABASE_DSN="root:password@tcp(127.0.0.1:3306)/binance?charset=utf8mb4&parseTime=True&loc=Local"
export JWT_SECRET="your-secret-key-change-in-production"
export ENCRYPTION_KEY="your-32-byte-encryption-key-here"
```

4. 创建数据库
```sql
CREATE DATABASE binance CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

5. 运行数据库迁移（自动执行）

6. 创建管理员账号
```bash
# 交互式创建
go run cmd/create-admin/main.go

# 或命令行参数创建
go run cmd/create-admin/main.go -username admin -password your_password -non-interactive
```

7. 启动后端服务
```bash
go run main.go
```

服务将在 http://localhost:8081 启动

### 前端安装

1. 进入前端目录
```bash
cd ../frontend/binance-frontend
```

2. 安装依赖
```bash
npm install
```

3. 启动开发服务器
```bash
npm run dev
```

前端将在 http://localhost:8080 启动

### 生产部署

1. 构建前端
```bash
npm run build
```

2. 编译后端
```bash
go build -o binance-trading main.go
```

3. 使用进程管理器（如systemd或supervisor）运行

## 使用指南

### 初始设置

1. 使用创建的管理员账号登录系统
2. 进入设置页面，配置币安API密钥
3. 添加需要监控的交易对

### 创建交易策略

1. 进入策略管理页面
2. 选择策略类型：
    - **简单策略**：在触发价格时下单
    - **冰山策略**：自动分层下单，降低市场影响
    - **自定义策略**：灵活配置价格偏移和数量分配

3. 设置策略参数：
    - 交易对
    - 触发价格
    - 总交易量
    - 订单自动取消时间

### 双币投资

1. 查看可投资产品列表
2. 创建投资策略：
    - **单次投资**：手动选择产品投资
    - **自动复投**：到期自动再投资
    - **梯度投资**：按深度梯度分散投资
    - **价格触发**：达到指定价格时投资

3. 查看投资统计和历史记录

### 自动提币设置

1. 进入自动提币页面
2. 添加提币规则：
    - 选择币种
    - 设置触发阈值
    - 输入提币地址
    - 设置提币金额（0表示全部提取）

## API接口文档

### 认证相关

#### 用户注册
```http
POST /register
Content-Type: application/json

{
  "username": "user123",
  "password": "password123"
}
```

#### 用户登录
```http
POST /login
Content-Type: application/json

{
  "username": "user123",
  "password": "password123"
}
```

### 交易相关

#### 获取余额
```http
GET /balance
Authorization: Bearer {token}
```

#### 创建订单
```http
POST /order
Authorization: Bearer {token}
Content-Type: application/json

{
  "symbol": "BTCUSDT",
  "side": "BUY",
  "price": 50000,
  "quantity": 0.001
}
```

### 策略相关

#### 创建策略
```http
POST /strategy
Authorization: Bearer {token}
Content-Type: application/json

{
  "symbol": "BTCUSDT",
  "strategyType": "iceberg",
  "side": "BUY",
  "price": 50000,
  "totalQuantity": 0.1,
  "cancelAfterMinutes": 120
}
```

## 配置说明

### 数据库配置
通过环境变量 `DATABASE_DSN` 配置数据库连接：
```bash
export DATABASE_DSN="用户名:密码@tcp(主机:端口)/数据库名?charset=utf8mb4&parseTime=True&loc=Local"
```

### JWT配置
设置 `JWT_SECRET` 环境变量用于签名Token：
```bash
export JWT_SECRET="your-secure-secret-key"
```

### 加密密钥
设置 `ENCRYPTION_KEY` 用于加密API密钥：
```bash
export ENCRYPTION_KEY="32-byte-encryption-key-for-aes256"
```

## 注意事项

1. **API密钥安全**：
    - 请确保使用只有交易权限的API密钥
    - 建议限制API密钥的IP白名单
    - 定期更换API密钥

2. **风险提示**：
    - 自动交易存在风险，请谨慎设置策略参数
    - 建议先在测试网环境测试
    - 设置合理的止损和仓位控制

3. **性能优化**：
    - 使用WebSocket减少API调用
    - 实现了请求频率限制
    - 数据库查询优化

## 故障排查

### 常见问题

1. **无法连接数据库**
    - 检查MySQL服务是否启动
    - 验证数据库连接字符串
    - 确认数据库用户权限

2. **API密钥错误**
    - 确认API密钥已启用
    - 检查IP白名单设置
    - 验证权限设置（需要读取和交易权限）

3. **WebSocket连接失败**
    - 检查网络连接
    - 确认防火墙设置
    - 查看币安API状态

## 贡献指南

欢迎提交Issue和Pull Request！

提交PR前请确保：
1. 代码通过所有测试
2. 遵循Go代码规范
3. 更新相关文档

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交Issue或联系维护者。

---

**免责声明**：本软件仅供学习和研究使用。使用本软件进行实际交易所产生的任何损失，作者不承担任何责任。请谨慎使用自动交易功能。