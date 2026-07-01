## 架构与基础设施
  - 多租户支持：单库 tenant_id 隔离，JWT 携带租户信息
  - Goose 迁移替代 GORM AutoMigrate，嵌入式 SQL，支持回滚
  - Snowflake ID 自动生成 (BeforeCreate hook)
  - GORM 乐观锁 (optimisticlock.Version)
  - koanf v2 配置管理，camelCase key 显式绑定
  - 服务层解耦：Service 只依赖 context.Context

## 任务队列 (pkg/taskq)
  - EnqueueUnique 幂等入队
  - RetryByType 按任务类型定制重试策略
  - singleflight 缓存击穿保护
  - Alerter 接口 + DeadLetterService 死信落库 MySQL
  - Redis AOF 持久化，重启不丢任务

## 验证码与通知 (pkg/verify_code, pkg/notify)
  - Redis 验证码存储 (5min TTL + 60s 冷却)
  - 阿里云 SMS/Email 发送器，LogSender 默认实现
  - 模板引擎 (embed HTML)，欢迎邮件 + 验证码邮件
  - 登录调用 VerifyCodeService.Validate()

## 限流升级
  - RateLimitByUser (按用户 ID) + RateLimit (按 IP)
  - 非认证接口 IP 10/min，认证接口用户 60/min，管理端 120/min
  - Redis 不可用时本地滑动窗口降级

## RBAC
  - RoleCheckAnyS/RoleCheckAllS 支持动态角色名 (string)
  - 用户/租户停用即时生效 (缓存失效)
  - 登录检查租户 active 状态

## 定时任务 (pkg/cronjob)
  - robfig/cron 封装，Register() 集中注册
  - heartbeat + purge-retried-dead-letters

## 其他
  - GET /healthz (DB+Redis 连通性)
  - Token 自动续期 (剩余 < 1/3 TTL → new-token header)
  - i18n Tf() 简化带参翻译
  - ZapLogger 日志级别守卫 (logger.Check)
  - Accept-Language sync.Map 缓存
  - generate.sh 升级：7文件+自动注册+迁移+go fmt
  - CORS 中间件，限流中间件本地降级
