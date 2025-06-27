.PHONY: help build up down restart logs clean

# 默认目标
help:
	@echo "Usage:"
	@echo "  make build    - 构建Docker镜像"
	@echo "  make up       - 启动所有服务"
	@echo "  make down     - 停止所有服务"
	@echo "  make restart  - 重启所有服务"
	@echo "  make logs     - 查看日志"
	@echo "  make clean    - 清理容器和镜像"
	@echo "  make admin    - 创建管理员账号"

# 构建镜像
build:
	docker-compose build

# 启动服务
up:
	docker-compose up -d

# 停止服务
down:
	docker-compose down

# 重启服务
restart:
	docker-compose restart

# 查看日志
logs:
	docker-compose logs -f

# 查看后端日志
logs-backend:
	docker-compose logs -f backend

# 查看MySQL日志
logs-mysql:
	docker-compose logs -f mysql

# 清理
clean:
	docker-compose down -v
	docker system prune -f

# 创建管理员账号
admin:
	docker-compose exec backend ./main -create-admin

# 进入MySQL控制台
mysql:
	docker-compose exec mysql mysql -u root -pMysql()P@$$w0rd binance

# 备份数据库
backup:
	docker-compose exec mysql mysqldump -u root -pMysql()P@$$w0rd binance > backup-$(shell date +%Y%m%d-%H%M%S).sql

# 恢复数据库
restore:
	@echo "Usage: make restore FILE=backup-xxx.sql"
	docker-compose exec -T mysql mysql -u root -pMysql()P@$$w0rd binance < $(FILE)