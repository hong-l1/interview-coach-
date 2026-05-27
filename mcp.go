//package main
//
//import (
//	"context"
//	"fmt"
//
//	"github.com/mark3labs/mcp-go/mcp"
//	"github.com/mark3labs/mcp-go/server"
//)
//
//func main() {
//	// Create a new MCP server
//	s := server.NewMCPServer(
//		"Calculator Demo",
//		"1.0.0",
//		server.WithToolCapabilities(false),
//		server.WithRecovery(),
//	)
//
//	// Add a calculator tool
//	calculatorTool := mcp.NewTool("calculate",
//		mcp.WithDescription("Perform basic arithmetic operations"),
//		mcp.WithString("operation",
//			mcp.Required(),
//			mcp.Description("The operation to perform (add, subtract, multiply, divide)"),
//			mcp.Enum("add", "subtract", "multiply", "divide"),
//		),
//		mcp.WithNumber("x",
//			mcp.Required(),
//			mcp.Description("First number"),
//		),
//		mcp.WithNumber("y",
//			mcp.Required(),
//			mcp.Description("Second number"),
//		),
//	)
//
//	// Add the calculator handler
//	s.AddTool(calculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
//		// Using helper functions for type-safe argument access
//		op, err := request.RequireString("operation")
//		if err != nil {
//			return mcp.NewToolResultError(err.Error()), nil
//		}
//
//		x, err := request.RequireFloat("x")
//		if err != nil {
//			return mcp.NewToolResultError(err.Error()), nil
//		}
//
//		y, err := request.RequireFloat("y")
//		if err != nil {
//			return mcp.NewToolResultError(err.Error()), nil
//		}
//
//		var result float64
//		switch op {
//		case "add":
//			result = x + y
//		case "subtract":
//			result = x - y
//		case "multiply":
//			result = x * y
//		case "divide":
//			if y == 0 {
//				return mcp.NewToolResultError("cannot divide by zero"), nil
//			}
//			result = x / y
//		}
//
//		return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
//	})
//
//	// Start the server
//	if err := server.ServeStdio(s); err != nil {
//		fmt.Printf("Server error: %v\n", err)
//	}
//}
package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

// 全局Redis客户端
var redisClient *redis.Client
var globalCtx = context.Background()

// Lock 分布式锁结构体（持有看门狗控制柄）
type Lock struct {
	key       string             // 锁key
	uniqueVal string             // 锁唯一值（防误删）
	cancel    context.CancelFunc // 停止看门狗
	expire    time.Duration      // 锁过期时间
}

// 初始化Redis
func initRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "43.142.57.35:6379",
		DB:   0,
	})
	// 测试连接
	if _, err := redisClient.Ping(globalCtx).Result(); err != nil {
		panic(err)
	}
	fmt.Println("Redis 连接成功")
}

// ---------------------- 加锁 + 自动启动看门狗 ----------------------
// key 锁名
// expire 锁过期时间（建议10~30秒）
func LockWithWatchDog(key string, expire time.Duration) (*Lock, error) {
	uniqueVal := uuid.NewString()

	// 1. 原子加锁（NX=互斥 EX=过期）
	ok, err := redisClient.SetNX(globalCtx, key, uniqueVal, expire).Result()
	if err != nil || !ok {
		return nil, fmt.Errorf("加锁失败: %v", err)
	}

	// 2. 创建上下文，用于停止看门狗
	ctx, cancel := context.WithCancel(globalCtx)

	// 3. 构造锁对象
	lock := &Lock{
		key:       key,
		uniqueVal: uniqueVal,
		cancel:    cancel,
		expire:    expire,
	}

	// 4. 启动看门狗协程（自动续期）
	go watchDog(ctx, lock)

	fmt.Printf("加锁成功：key=%s, 唯一值=%s\n", key, uniqueVal)
	return lock, nil
}

// ---------------------- 看门狗：自动续期 ----------------------
// 每 expire/3 秒续期一次，直到ctx取消
func watchDog(ctx context.Context, lock *Lock) {
	// 续期间隔 = 过期时间的 1/3（标准策略）
	ticker := time.NewTicker(lock.expire / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 收到停止信号，退出看门狗
			fmt.Println("看门狗已停止")
			return
		case <-ticker.C:
			// 续期：只有锁是自己的，才重置过期时间
			renewal(lock)
		}
	}
}

// ---------------------- 续期逻辑（原子操作） ----------------------
func renewal(lock *Lock) {
	// Lua脚本：判断锁归属，是自己的就续期
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("expire", KEYS[1], ARGV[2])
	else
		return 0
	end
`
	// 执行续期
	result, err := redisClient.Eval(
		globalCtx,
		script,
		[]string{lock.key},
		lock.uniqueVal,
		int(lock.expire.Seconds()),
	).Result()

	if err != nil {
		fmt.Printf("续期失败: %v\n", err)
		return
	}
	if result.(int64) == 1 {
		fmt.Printf("自动续期成功：key=%s, 重置超时=%v\n", lock.key, lock.expire)
	}
}

// ---------------------- 解锁（停止看门狗 + 删除锁） ----------------------
func (l *Lock) Unlock() {
	// 1. 停止看门狗
	l.cancel()

	// 2. Lua脚本原子解锁
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`
	_, err := redisClient.Eval(globalCtx, script, []string{l.key}, l.uniqueVal).Result()
	if err != nil {
		fmt.Printf("解锁失败: %v\n", err)
		return
	}
	fmt.Println("解锁成功")
}

// ---------------------- 测试：业务超时也不会丢锁 ----------------------
func main() {
	initRedis()
	lockKey := "stock:lock:1001"

	// 加锁：30秒过期 + 自动续期
	lock, err := LockWithWatchDog(lockKey, 30*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 模拟业务：执行50秒（远超锁过期时间）
	fmt.Println("业务执行中...(50秒)")
	time.Sleep(50 * time.Second)

	// 解锁
	lock.Unlock()
	fmt.Println("业务执行完成")
}
