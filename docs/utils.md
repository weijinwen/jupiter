# Jupiter 框架工具包使用指南

本文档介绍 Jupiter 框架中 `pkg/util` 目录下所有工具包的功能、使用方法和最佳实践。

## 目录

- [xattr](#xattr) - 属性键值对管理工具
- [xbuffer](#xbuffer) - 无界缓冲区实现
- [xcast](#xcast) - 类型转换工具
- [xcycle](#xcycle) - 生命周期管理工具
- [xdebug](#xdebug) - 调试辅助工具
- [xdefer](#xdefer) - 延迟执行栈管理
- [xerror](#xerror) - 错误处理工具
- [xfieldmask](#xfieldmask) - 字段掩码工具
- [xfile](#xfile) - 文件操作工具
- [xgo](#xgo) - 并发控制工具
- [xmap](#xmap) - Map 操作工具
- [xnet](#xnet) - 网络相关工具
- [xp2c](#xp2c) - 负载均衡工具
- [xreflect](#xreflect) - 反射工具
- [xretry](#xretry) - 重试机制工具
- [xstring](#xstring) - 字符串处理工具

---

## xattr

### 功能说明

`xattr` 提供了一个简单的键值对属性管理工具，用于存储和检索任意类型的键值对数据。

### 主要方法

- `New(kvs ...interface{}) *Attributes` - 创建一个新的属性对象
- `WithValues(kvs ...interface{}) *Attributes` - 添加键值对并返回新对象
- `Value(key interface{}) interface{}` - 获取指定键的值

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xattr"
)

func main() {
    // 创建属性对象
    attrs := xattr.New("name", "张三", "age", 25, "active", true)
    
    // 获取值
    name := attrs.Value("name")
    age := attrs.Value("age")
    active := attrs.Value("active")
    
    fmt.Printf("姓名: %v, 年龄: %v, 活跃: %v\n", name, age, active)
    
    // 添加新属性
    newAttrs := attrs.WithValues("email", "zhangsan@example.com")
    email := newAttrs.Value("email")
    fmt.Printf("邮箱: %v\n", email)
}
```

### 最佳实践

- 使用 `xattr` 管理配置项或元数据
- 键值对参数必须成对出现，否则会触发 panic
- 适合存储临时属性或配置信息

---

## xbuffer

### 功能说明

`xbuffer` 实现了一个无界缓冲区，用于在生产者和消费者之间传递数据，特别适用于 gRPC 等场景。

### 主要方法

- `NewUnbounded() *Unbounded` - 创建一个新的无界缓冲区
- `Put(t interface{})` - 向缓冲区添加数据
- `Load()` - 将最早的数据加载到读取通道
- `Get() <-chan interface{}` - 获取读取通道

### 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/weijinwen/jupiter/pkg/util/xbuffer"
)

func main() {
    // 创建无界缓冲区
    buffer := xbuffer.NewUnbounded()
    
    // 启动消费者
    go func() {
        for data := range buffer.Get() {
            fmt.Printf("接收到数据: %v\n", data)
            // 每次读取后调用 Load() 加载下一个数据
            buffer.Load()
        }
    }()
    
    // 生产者发送数据
    buffer.Put("消息1")
    buffer.Put("消息2")
    buffer.Put("消息3")
    
    // 等待一段时间让消费者处理
    time.Sleep(100 * time.Millisecond)
}
```

### 最佳实践

- 适用于生产者和消费者速度不匹配的场景
- 每次从通道读取数据后需要调用 `Load()` 加载下一个数据
- 注意避免内存泄漏，及时处理缓冲区中的数据

---

## xcast

### 功能说明

`xcast` 提供了类型转换工具，扩展了标准库的转换功能，特别针对 `[]int64` 类型提供了便捷的转换方法。

### 主要方法

- `ToInt64Slice(i interface{}) []int64` - 将接口转换为 `[]int64` 类型
- `ToInt64SliceE(i interface{}) ([]int64, error)` - 带错误检查的转换

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xcast"
)

func main() {
    // 转换 []int64
    intSlice := []int{1, 2, 3, 4, 5}
    result := xcast.ToInt64Slice(intSlice)
    fmt.Printf("转换结果: %v (类型: %T)\n", result, result)
    
    // 转换 []interface{}
    interfaceSlice := []interface{}{10, 20, 30, 40, 50}
    result2 := xcast.ToInt64Slice(interfaceSlice)
    fmt.Printf("转换结果: %v (类型: %T)\n", result2, result2)
    
    // 带错误检查的转换
    result3, err := xcast.ToInt64SliceE([]string{"1", "2", "3"})
    if err != nil {
        fmt.Printf("转换失败: %v\n", err)
    } else {
        fmt.Printf("转换成功: %v\n", result3)
    }
}
```

### 最佳实践

- 使用 `ToInt64SliceE` 进行带错误检查的转换，避免潜在错误
- 适用于需要将不同类型转换为 `[]int64` 的场景
- 可以处理切片和数组类型的转换

---

## xcycle

### 功能说明

`xcycle` 提供了生命周期管理工具，用于管理多个 goroutine 的生命周期，支持优雅关闭和错误处理。

### 主要方法

- `NewCycle() *Cycle` - 创建一个新的生命周期管理器
- `Run(fn func() error)` - 启动一个新的 goroutine
- `Done() <-chan struct{}` - 返回一个通道，当所有 goroutine 完成时关闭
- `DoneAndClose()` - 等待所有 goroutine 完成并关闭
- `Close()` - 关闭生命周期管理器
- `Wait() <-chan error` - 返回一个错误通道，用于接收错误

### 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/weijinwen/jupiter/pkg/util/xcycle"
)

func main() {
    // 创建生命周期管理器
    cycle := xcycle.NewCycle()
    
    // 启动多个任务
    cycle.Run(func() error {
        for i := 0; i < 5; i++ {
            fmt.Printf("任务1: %d\n", i)
            time.Sleep(100 * time.Millisecond)
        }
        return nil
    })
    
    cycle.Run(func() error {
        for i := 0; i < 3; i++ {
            fmt.Printf("任务2: %d\n", i)
            time.Sleep(150 * time.Millisecond)
        }
        return nil
    })
    
    // 等待所有任务完成
    <-cycle.Done()
    fmt.Println("所有任务已完成")
    
    // 或者使用 DoneAndClose
    // cycle.DoneAndClose()
}
```

### 最佳实践

- 使用 `xcycle` 管理多个相关的 goroutine 生命周期
- 在微服务启动时使用，可以优雅地管理所有服务的生命周期
- 结合 `Close()` 方法实现优雅关闭

---

## xdebug

### 功能说明

`xdebug` 提供了一系列调试辅助工具，包括堆栈跟踪、对象打印、开发模式检测等功能，帮助开发者在开发和测试阶段更好地调试代码。

### 主要方法

- `PanicTrace(kb int) []byte` - 获取 panic 堆栈信息
- `DumpPanicTrace(kb int)` - 打印 panic 堆栈信息
- `PrintObject(message string, obj interface{})` - 打印格式化对象
- `DebugBytes(obj interface{}) string` - 将对象转换为调试格式的字节
- `PrintKV(key string, val string)` - 打印键值对
- `PrintMap(data map[string]interface{})` - 打印 Map
- `IsDevelopmentMode() bool` - 检查是否为开发模式
- `IsTestingMode() bool` - 检查是否为测试模式
- `IfPanic(err error)` - 如果错误不为 nil 则 panic
- `PrettyJsonPrint(message string, obj interface{})` - 打印格式化 JSON

### 使用示例

```go
package main

import (
    "github.com/weijinwen/jupiter/pkg/util/xdebug"
)

type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

func main() {
    user := User{
        Name:  "张三",
        Age:   30,
        Email: "zhangsan@example.com",
    }
    
    // 打印对象
    xdebug.PrintObject("用户信息", user)
    
    // 打印键值对
    xdebug.PrintKV("用户名", "张三")
    xdebug.PrintKV("年龄", "30")
    
    // 打印 Map
    data := map[string]interface{}{
        "name":  "李四",
        "age":   25,
        "email": "lisi@example.com",
    }
    xdebug.PrintMap(data)
    
    // 打印格式化 JSON
    xdebug.PrettyJsonPrint("用户JSON", user)
    
    // 检查开发模式
    if xdebug.IsDevelopmentMode() {
        println("当前处于开发模式")
    }
    
    // 使用 IfPanic 进行错误处理
    err := someFunction()
    xdebug.IfPanic(err)
}

func someFunction() error {
    // 模拟一个错误
    return nil
}
```

### 最佳实践

- 在开发和测试阶段使用 `xdebug` 工具进行调试
- 使用 `IsDevelopmentMode()` 和 `IsTestingMode()` 判断当前环境，避免在生产环境输出调试信息
- 使用 `PrettyJsonPrint` 格式化输出 JSON 数据，便于查看
- 使用 `IfPanic` 简化错误处理逻辑

---

## xdefer

### 功能说明

`xdefer` 提供了一个延迟执行栈管理工具，类似于 Go 的 `defer` 关键字，但提供了更灵活的控制方式，可以动态添加和管理需要延迟执行的函数。

### 主要方法

- `NewStack() *DeferStack` - 创建一个新的延迟执行栈
- `Push(fns ...func() error)` - 向栈中添加延迟执行的函数
- `Clean()` - 按照后进先出的顺序执行栈中的所有函数

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xdefer"
)

func main() {
    // 创建延迟执行栈
    stack := xdefer.NewStack()
    
    // 添加延迟执行的函数
    stack.Push(func() error {
        fmt.Println("清理资源1")
        return nil
    })
    
    stack.Push(func() error {
        fmt.Println("清理资源2")
        return nil
    })
    
    stack.Push(func() error {
        fmt.Println("清理资源3")
        return nil
    })
    
    // 执行栈中的所有函数（后进先出）
    fmt.Println("开始清理...")
    stack.Clean()
    fmt.Println("清理完成")
}
```

### 最佳实践

- 在需要管理多个资源清理的场景中使用 `xdefer`
- 与 Go 的 `defer` 不同，`xdefer` 允许动态添加清理函数
- 适用于需要根据条件决定清理资源的场景

---

## xerror

### 功能说明

`xerror` 提供了一套完整的错误处理机制，包括错误码定义、错误转换、gRPC 错误兼容等功能，帮助开发者更好地管理和处理错误。

### 主要方法

- `New(code codes.Code, message string) *Err` - 创建一个新的错误
- `Convert(err error) *Err` - 将普通错误转换为 `*Err`
- `WithMsg(msg string) *Err` - 设置错误消息
- `WithData(data interface{}) *Err` - 设置错误数据
- `GetEcode() int32` - 获取错误码
- `GetMsg() string` - 获取错误消息
- `GetData() interface{}` - 获取错误数据
- `GRPCStatus() *status.Status` - 获取 gRPC 状态

### 预定义错误

- `OK` - 请求正常
- `Canceled` - 操作取消
- `Unknown` - 未知错误
- `InvalidArgument` - 无效参数
- `DeadlineExceeded` - 处理时间超过最后期限
- `NotFound` - 访问链接不存在
- `AlreadyExists` - 目标已存在
- `PermissionDenied` - 权限不足
- `ResourceExhausted` - 资源耗尽
- `FailedPrecondition` - 前置条件出错
- `Aborted` - 操作中途失败
- `OutOfRange` - 操作超出有效范围
- `Unimplemented` - 当前服务未实现
- `Internal` - 服务内部异常
- `Unavailable` - 服务当前不可用
- `DataLoss` - 数据丢失
- `Unauthenticated` - 未授权错误

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xerror"
)

func main() {
    // 创建自定义错误
    customErr := xerror.New(xerror.InvalidArgument.Ecode, "用户名不能为空")
    fmt.Printf("错误码: %d, 错误消息: %s\n", customErr.GetEcode(), customErr.GetMsg())
    
    // 使用预定义错误
    err := xerror.NotFound.WithMsg("用户不存在")
    fmt.Printf("错误: %s\n", err.Error())
    
    // 添加额外数据
    errWithData := xerror.PermissionDenied.WithData(map[string]interface{}{
        "resource": "user_profile",
        "action":   "read",
    })
    fmt.Printf("错误数据: %v\n", errWithData.GetData())
    
    // 转换普通错误
    normalErr := fmt.Errorf("数据库连接失败")
    convertedErr := xerror.Convert(normalErr)
    fmt.Printf("转换后的错误: %s\n", convertedErr.Error())
    
    // 修改错误消息
    modifiedErr := xerror.Internal.WithMsg("数据库服务不可用")
    fmt.Printf("修改后的错误: %s\n", modifiedErr.Error())
}
```

### 最佳实践

- 使用预定义错误码，保持错误处理的一致性
- 使用 `WithData` 添加上下文信息，便于调试
- 在 API 层使用 `xerror` 统一错误格式
- 使用 `Convert` 将第三方库错误转换为框架错误

---

## xfieldmask

### 功能说明

`xfieldmask` 提供了字段掩码功能，用于过滤或修剪 protobuf 消息中的字段，支持嵌套字段操作，常用于 API 响应中控制返回的字段。

### 主要方法

- `New(path []string) NestedFieldMask` - 创建一个新的字段掩码
- `NewWithPaths(paths ...string) NestedFieldMask` - 使用多个路径创建字段掩码
- `Filter(m proto.Message)` - 过滤消息，只保留掩码中指定的字段
- `Prune(m proto.Message)` - 修剪消息，移除掩码中指定的字段
- `Masked(path string) bool` - 检查路径是否被掩码覆盖

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xfieldmask"
)

func main() {
    // 创建字段掩码
    mask := xfieldmask.NewWithPaths("name", "age", "address.city")
    
    // 检查路径是否被掩码覆盖
    fmt.Printf("路径 'name' 被覆盖: %v\n", mask.Masked("name"))
    fmt.Printf("路径 'email' 被覆盖: %v\n", mask.Masked("email"))
    fmt.Printf("路径 'address.city' 被覆盖: %v\n", mask.Masked("address.city"))
    
    // 注意：Filter 和 Prune 方法需要 protobuf 消息对象
    // 这里仅展示掩码创建和路径检查的用法
}
```

### 最佳实践

- 在 API 响应中使用字段掩码控制返回的字段，减少网络传输
- 使用 `Filter` 只返回客户端需要的字段
- 使用 `Prune` 移除敏感字段或不必要的字段
- 支持嵌套字段路径，如 `"user.profile.name"`

---

## xfile

### 功能说明

`xfile` 提供了一系列文件操作工具，包括文件检查、目录操作、文件锁定等功能，简化了文件系统相关的操作。

### 主要方法

- `Exists(fpath string) bool` - 检查文件是否存在
- `ListFiles(dir string, ext string) []string` - 列出目录中的文件
- `IsFileChanged(src, dest string) (bool, error)` - 检查文件是否已更改
- `IsDirectory(path string) (bool, error)` - 检查路径是否为目录
- `RecursiveFilesLookup(root string, pattern string) ([]string, error)` - 递归查找文件
- `RecursiveDirsLookup(root string, pattern string) ([]string, error)` - 递归查找目录
- `CheckAndGetParentDir(path string) string` - 获取父目录
- `MkdirIfNecessary(createDir string) error` - 必要时创建目录
- `GetCurrentDirectory() string` - 获取当前目录
- `MakeDirectory(dirs ...string) error` - 创建目录

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xfile"
)

func main() {
    // 检查文件是否存在
    filePath := "./example.txt"
    exists := xfile.Exists(filePath)
    fmt.Printf("文件 %s 存在: %v\n", filePath, exists)
    
    // 检查是否为目录
    isDir, err := xfile.IsDirectory("./")
    if err != nil {
        fmt.Printf("检查目录失败: %v\n", err)
    } else {
        fmt.Printf("当前目录是目录: %v\n", isDir)
    }
    
    // 获取当前目录
    currentDir := xfile.GetCurrentDirectory()
    fmt.Printf("当前目录: %s\n", currentDir)
    
    // 列出当前目录中的 .go 文件
    goFiles := xfile.ListFiles("./", ".go")
    fmt.Printf("Go 文件数量: %d\n", len(goFiles))
    
    // 递归查找所有 .md 文件
    mdFiles, err := xfile.RecursiveFilesLookup("./", "*.md")
    if err != nil {
        fmt.Printf("查找 Markdown 文件失败: %v\n", err)
    } else {
        fmt.Printf("Markdown 文件数量: %d\n", len(mdFiles))
    }
    
    // 创建目录
    err = xfile.MakeDirectory("./temp", "./temp/subdir")
    if err != nil {
        fmt.Printf("创建目录失败: %v\n", err)
    } else {
        fmt.Println("目录创建成功")
    }
}
```

### 最佳实践

- 使用 `Exists` 检查文件是否存在，避免文件操作错误
- 使用 `MkdirIfNecessary` 确保目录存在，避免目录不存在导致的错误
- 使用 `RecursiveFilesLookup` 递归查找文件，简化文件搜索逻辑
- 在文件操作前检查路径类型，避免对目录执行文件操作

---

## xgo

### 功能说明

`xgo` 提供了一系列并发控制工具，包括 goroutine 管理、并发执行、错误处理等功能，帮助开发者更好地控制并发行为。

### 主要方法

#### Goroutine 管理
- `Go(fn func())` - 启动一个 goroutine
- `GoDirect(fn interface{}, args ...interface{})` - 启动带参数的 goroutine
- `DelayGo(delay time.Duration, fn func())` - 延迟启动 goroutine
- `SafeGo(fn func(), rec func(error))` - 安全启动 goroutine，带错误处理

#### 串行执行
- `Serial(fns ...func()) func()` - 串行执行多个函数
- `SerialWithError(fns ...func() error) func() error` - 串行执行，带错误处理
- `SerialUntilError(fns ...func() error) func() error` - 串行执行，遇到错误停止
- `SerialWhenError(we WhenError) func(fn ...func() error) func() error` - 串行执行，自定义错误处理策略

#### 并行执行
- `Parallel(fns ...func()) func()` - 并行执行多个函数
- `ParallelWithError(fns ...func() error) func() error` - 并行执行，带错误处理
- `ParallelWithErrorChan(fns ...func() error) chan error` - 并行执行，返回错误通道
- `RestrictParallel(restrict int, fns ...func()) func()` - 限制并发数量的并行执行
- `RestrictParallelWithErrorChan(concurrency int, fns ...func() error) chan error` - 限制并发数量的并行执行，返回错误通道

### 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/weijinwen/jupiter/pkg/util/xgo"
)

func main() {
    // 串行执行
    serial := xgo.Serial(
        func() { fmt.Println("任务1") },
        func() { fmt.Println("任务2") },
        func() { fmt.Println("任务3") },
    )
    serial()
    
    // 并行执行
    parallel := xgo.Parallel(
        func() { fmt.Println("并行任务1") },
        func() { fmt.Println("并行任务2") },
        func() { fmt.Println("并行任务3") },
    )
    parallel()
    
    // 限制并发数量的并行执行
    restricted := xgo.RestrictParallel(2,
        func() { 
            fmt.Println("受限并行任务1")
            time.Sleep(100 * time.Millisecond)
        },
        func() { 
            fmt.Println("受限并行任务2")
            time.Sleep(100 * time.Millisecond)
        },
        func() { 
            fmt.Println("受限并行任务3")
            time.Sleep(100 * time.Millisecond)
        },
        func() { 
            fmt.Println("受限并行任务4")
            time.Sleep(100 * time.Millisecond)
        },
    )
    restricted()
    
    // 带错误处理的串行执行
    serialWithErr := xgo.SerialUntilError(
        func() error { 
            fmt.Println("串行任务1")
            return nil
        },
        func() error { 
            fmt.Println("串行任务2")
            return fmt.Errorf("任务2出错")
        },
        func() error { 
            fmt.Println("串行任务3") // 不会执行
            return nil
        },
    )
    err := serialWithErr()
    fmt.Printf("串行执行错误: %v\n", err)
    
    // 带错误处理的并行执行
    parallelWithErr := xgo.ParallelWithError(
        func() error { 
            fmt.Println("并行任务1")
            return nil
        },
        func() error { 
            fmt.Println("并行任务2")
            return fmt.Errorf("任务2出错")
        },
        func() error { 
            fmt.Println("并行任务3")
            return nil
        },
    )
    err = parallelWithErr()
    fmt.Printf("并行执行错误: %v\n", err)
    
    // 安全启动 goroutine
    xgo.SafeGo(func() {
        fmt.Println("安全 goroutine")
    }, func(err error) {
        fmt.Printf("捕获错误: %v\n", err)
    })
    
    // 延迟启动 goroutine
    xgo.DelayGo(500*time.Millisecond, func() {
        fmt.Println("延迟 goroutine")
    })
    
    // 等待一段时间让所有任务完成
    time.Sleep(600 * time.Millisecond)
}
```

### 最佳实践

- 使用 `xgo.SafeGo` 启动 goroutine，避免 panic 导致程序崩溃
- 使用 `xgo.ParallelWithError` 处理并发任务中的错误
- 使用 `xgo.RestrictParallel` 限制并发数量，避免资源耗尽
- 使用 `xgo.SerialUntilError` 在串行任务中遇到错误时停止执行

---

## xmap

### 功能说明

`xmap` 提供了一系列 Map 操作工具，包括 Map 合并、类型转换、深度搜索等功能，简化了 Map 相关的操作。

### 主要方法

- `MergeStringMap(dest, src map[string]interface{})` - 合并两个 Map
- `ToMapStringInterface(src map[interface{}]interface{}) map[string]interface{}` - 类型转换
- `DeepSearchInMap(m map[string]interface{}, paths ...string) map[string]interface{}` - 深度搜索 Map

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xmap"
)

func main() {
    // 合并 Map
    dest := map[string]interface{}{
        "name": "张三",
        "age":  30,
        "profile": map[string]interface{}{
            "city": "北京",
        },
    }
    
    src := map[string]interface{}{
        "email": "zhangsan@example.com",
        "age":   31, // 会覆盖 dest 中的 age
        "profile": map[string]interface{}{
            "country": "中国",
        },
    }
    
    xmap.MergeStringMap(dest, src)
    fmt.Printf("合并后的 Map: %+v\n", dest)
    
    // 类型转换
    interfaceMap := map[interface{}]interface{}{
        "name": "李四",
        "age":  25,
        100:   "数字键",
    }
    
    stringMap := xmap.ToMapStringInterface(interfaceMap)
    fmt.Printf("转换后的 Map: %+v\n", stringMap)
    
    // 深度搜索
    data := map[string]interface{}{
        "user": map[string]interface{}{
            "profile": map[string]interface{}{
                "name": "王五",
                "contact": map[string]interface{}{
                    "email": "wangwu@example.com",
                    "phone": "13800138000",
                },
            },
        },
    }
    
    // 搜索嵌套路径
    profile := xmap.DeepSearchInMap(data, "user", "profile")
    fmt.Printf("用户配置: %+v\n", profile)
    
    // 搜索不存在的路径，会自动创建
    address := xmap.DeepSearchInMap(data, "user", "address")
    fmt.Printf("用户地址: %+v\n", address)
}
```

### 最佳实践

- 使用 `MergeStringMap` 合并配置，注意嵌套 Map 的合并行为
- 使用 `DeepSearchInMap` 访问嵌套 Map，简化多层访问代码
- 使用 `ToMapStringInterface` 统一 Map 类型，避免类型不一致问题

---

## xnet

### 功能说明

`xnet` 提供了一系列网络相关工具，包括 IP 地址获取、URL 解析、网络地址处理等功能，简化了网络相关的操作。

### 主要方法

#### IP 地址相关
- `GetLocalIP() (string, error)` - 获取本地 IP 地址
- `GetLocalMainIP() (string, int, error)` - 获取本地主 IP 地址和端口
- `GetMacAddrs() []string` - 获取所有 MAC 地址
- `GetIPs() []string` - 获取所有 IP 地址

#### 地址相关
- `Address(listener net.Listener) string` - 从监听器获取服务地址

#### URL 相关
- `ParseURL(raw string) (*URL, error)` - 解析 URL
- `QueryInt(field string, expect int) int` - 获取查询参数整数值
- `QueryInt64(field string, expect int64) int64` - 获取查询参数 int64 值
- `QueryString(field string, expect string) string` - 获取查询参数字符串值
- `QueryDuration(field string, expect time.Duration) time.Duration` - 获取查询参数时间间隔值
- `QueryBool(field string, expect bool) bool` - 获取查询参数布尔值

### 使用示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/weijinwen/jupiter/pkg/util/xnet"
)

func main() {
    // 获取本地 IP
    localIP, err := xnet.GetLocalIP()
    if err != nil {
        fmt.Printf("获取本地 IP 失败: %v\n", err)
    } else {
        fmt.Printf("本地 IP: %s\n", localIP)
    }
    
    // 获取主 IP 和端口
    mainIP, port, err := xnet.GetLocalMainIP()
    if err != nil {
        fmt.Printf("获取主 IP 失败: %v\n", err)
    } else {
        fmt.Printf("主 IP: %s, 端口: %d\n", mainIP, port)
    }
    
    // 获取所有 IP
    ips := xnet.GetIPs()
    fmt.Printf("所有 IP: %v\n", ips)
    
    // 获取所有 MAC 地址
    macAddrs := xnet.GetMacAddrs()
    fmt.Printf("所有 MAC 地址: %v\n", macAddrs)
    
    // 解析 URL
    rawURL := "https://example.com:8080/path?query=value&timeout=30s&debug=true"
    parsedURL, err := xnet.ParseURL(rawURL)
    if err != nil {
        fmt.Printf("解析 URL 失败: %v\n", err)
    } else {
        fmt.Printf("协议: %s\n", parsedURL.Scheme)
        fmt.Printf("主机: %s\n", parsedURL.Host)
        fmt.Printf("路径: %s\n", parsedURL.Path)
        
        // 获取查询参数
        query := parsedURL.Query()
        fmt.Printf("查询参数 query: %s\n", query.Get("query"))
        
        // 使用便捷方法获取查询参数
        timeout := parsedURL.QueryDuration("timeout", 10*time.Second)
        fmt.Printf("超时时间: %v\n", timeout)
        
        debug := parsedURL.QueryBool("debug", false)
        fmt.Printf("调试模式: %v\n", debug)
        
        count := parsedURL.QueryInt("count", 10)
        fmt.Printf("计数: %d\n", count)
    }
}
```

### 最佳实践

- 使用 `GetLocalMainIP` 获取对外服务的 IP 地址
- 使用 `ParseURL` 解析 URL，然后使用便捷方法获取查询参数
- 使用 `QueryDuration` 处理时间间隔参数，简化时间解析逻辑
- 使用 `QueryBool` 处理布尔参数，避免手动解析

---

## xp2c

### 功能说明

`xp2c` 提供了负载均衡工具，实现了 Power of Two Choices (P2C) 算法，包括基础接口和最少负载实现，用于在多个选项中选择负载最低的一个。

### 主要接口

- `P2c` - 负载均衡接口
  - `Next() (interface{}, func(balancer.DoneInfo))` - 获取下一个选项
  - `Add(interface{})` - 添加选项

### 实现类

- `leastloaded` - 最少负载实现
  - `New() P2c` - 创建新的最少负载均衡器

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xp2c"
    "github.com/weijinwen/jupiter/pkg/util/xp2c/leastloaded"
)

func main() {
    // 创建最少负载均衡器
    balancer := leastloaded.New()
    
    // 添加选项
    balancer.Add("服务器1")
    balancer.Add("服务器2")
    balancer.Add("服务器3")
    
    // 获取下一个选项
    for i := 0; i < 5; i++ {
        server, done := balancer.Next()
        fmt.Printf("选择服务器: %v\n", server)
        
        // 模拟使用服务器
        // ...
        
        // 完成使用，通知负载均衡器
        done(balancer.DoneInfo{})
    }
}
```

### 最佳实践

- 使用 `xp2c` 实现服务负载均衡，提高系统性能
- 使用 `leastloaded` 实现最少负载均衡，避免单个服务器过载
- 在使用完服务器后调用 `done` 函数，更新负载信息

---

## xreflect

### 功能说明

`xreflect` 提供了一系列反射工具，包括类型检查、字段覆盖等功能，简化了反射相关的操作。

### 主要方法

- `In(value interface{}, container interface{}) bool` - 检查值是否在容器中
- `Override(left interface{}, right interface{}) error` - 用右对象的字段覆盖左对象的字段

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xreflect"
)

type User struct {
    Name  string
    Age   int
    Email string
}

func main() {
    // 检查值是否在容器中
    slice := []string{"apple", "banana", "orange"}
    inSlice := xreflect.In("banana", slice)
    fmt.Printf("'banana' 在切片中: %v\n", inSlice)
    
    notInSlice := xreflect.In("grape", slice)
    fmt.Printf("'grape' 在切片中: %v\n", notInSlice)
    
    // 检查值是否在 Map 中
    m := map[string]int{"one": 1, "two": 2, "three": 3}
    inMap := xreflect.In("two", m)
    fmt.Printf("'two' 在 Map 中: %v\n", inMap)
    
    // 字段覆盖
    user1 := &User{
        Name:  "张三",
        Age:   30,
        Email: "zhangsan@example.com",
    }
    
    user2 := &User{
        Name:  "李四",
        Age:   0, // 零值不会覆盖
        Email: "lisi@example.com",
    }
    
    err := xreflect.Override(user1, user2)
    if err != nil {
        fmt.Printf("覆盖失败: %v\n", err)
    } else {
        fmt.Printf("覆盖后的用户: %+v\n", *user1)
    }
}
```

### 最佳实践

- 使用 `In` 方法简化值存在性检查，支持切片、数组和 Map
- 使用 `Override` 方法更新对象字段，零值不会覆盖原值
- 注意 `Override` 方法要求参数必须是指针且类型相同

---

## xretry

### 功能说明

`xretry` 提供了简单的重试机制，支持指数退避和随机抖动，用于处理可能失败的操作。

### 主要方法

- `Do(attempts int, sleep time.Duration, f func() error) error` - 执行重试

### 使用示例

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
    "github.com/weijinwen/jupiter/pkg/util/xretry"
)

func main() {
    // 模拟可能失败的操作
    attempt := 0
    unreliableOperation := func() error {
        attempt++
        fmt.Printf("尝试第 %d 次\n", attempt)
        
        // 模拟 70% 的失败率
        if rand.Float32() < 0.7 {
            return fmt.Errorf("操作失败")
        }
        
        return nil
    }
    
    // 使用重试机制
    err := xretry.Do(5, 100*time.Millisecond, unreliableOperation)
    if err != nil {
        fmt.Printf("操作最终失败: %v\n", err)
    } else {
        fmt.Println("操作成功")
    }
}
```

### 最佳实践

- 使用 `xretry.Do` 处理可能失败的操作，如网络请求、数据库连接等
- 设置合理的重试次数和初始间隔，避免过度重试
- 重试间隔会指数增长并添加随机抖动，避免雷群效应

---

## xstring

### 功能说明

`xstring` 提供了一系列字符串处理工具，包括格式转换、JSON 处理、地址转换等功能，简化了字符串相关的操作。

### 主要方法

#### 字符串转换
- `ToSnakeCase(str string) string` - 转换为蛇形命名
- `ToCamelCase(str string) string` - 转换为驼峰命名

#### JSON 处理
- `Json(obj interface{}) string` - 将对象转换为 JSON 字符串
- `JsonBytes(obj interface{}) []byte` - 将对象转换为 JSON 字节
- `PrettyJson(obj interface{}) string` - 将对象转换为格式化的 JSON 字符串
- `PrettyJSONBytes(obj interface{}) []byte` - 将对象转换为格式化的 JSON 字节

#### 地址转换
- `Addr2Hex(str string) (string, error)` - 将地址字符串转换为十六进制
- `Hex2Addr(str string) (string, error)` - 将十六进制转换为地址字符串

#### 字符串工具
- `KickEmpty(ss []string) Strings` - 移除空字符串
- `Kick(ss []string, remove func(item string) bool) Strings` - 根据条件移除元素
- `AnyBlank(ss []string) bool` - 检查是否有空字符串

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xstring"
)

type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

func main() {
    // 字符串格式转换
    snakeCase := xstring.ToSnakeCase("UserName")
    fmt.Printf("蛇形命名: %s\n", snakeCase)
    
    camelCase := xstring.ToCamelCase("user_name")
    fmt.Printf("驼峰命名: %s\n", camelCase)
    
    // JSON 处理
    user := User{
        Name:  "张三",
        Age:   30,
        Email: "zhangsan@example.com",
    }
    
    jsonStr := xstring.Json(user)
    fmt.Printf("JSON 字符串: %s\n", jsonStr)
    
    prettyJson := xstring.PrettyJson(user)
    fmt.Printf("格式化 JSON:\n%s\n", prettyJson)
    
    // 地址转换
    addr := "192.168.1.1:8080"
    hex, err := xstring.Addr2Hex(addr)
    if err != nil {
        fmt.Printf("地址转换失败: %v\n", err)
    } else {
        fmt.Printf("地址十六进制: %s\n", hex)
        
        // 反向转换
        restoredAddr, err := xstring.Hex2Addr(hex)
        if err != nil {
            fmt.Printf("十六进制转换失败: %v\n", err)
        } else {
            fmt.Printf("还原地址: %s\n", restoredAddr)
        }
    }
    
    // 字符串工具
    strings := []string{"apple", "", "banana", "", "cherry"}
    nonEmpty := xstring.KickEmpty(strings)
    fmt.Printf("移除空字符串后: %v\n", nonEmpty)
    
    hasBlank := xstring.AnyBlank(strings)
    fmt.Printf("是否有空字符串: %v\n", hasBlank)
    
    // 根据条件移除元素
    filtered := xstring.Kick(strings, func(item string) bool {
        return item == "banana"
    })
    fmt.Printf("移除 'banana' 后: %v\n", filtered)
}
```

### 最佳实践

- 使用 `ToSnakeCase` 和 `ToCamelCase` 进行命名格式转换，保持命名一致性
- 使用 `PrettyJson` 格式化输出 JSON，便于查看和调试
- 使用 `KickEmpty` 清理字符串数组，避免空值问题
- 使用 `Addr2Hex` 和 `Hex2Addr` 进行网络地址的编码和解码

---

## 总结

Jupiter 框架的 `pkg/util` 目录提供了丰富的工具包，涵盖了开发中常见的各种需求：

1. **基础工具**：`xattr`、`xcast`、`xstring` 等提供了基础数据类型操作
2. **并发控制**：`xgo`、`xcycle` 等提供了强大的并发管理能力
3. **错误处理**：`xerror`、`xretry` 等提供了完善的错误处理机制
4. **网络和文件**：`xnet`、`xfile` 等简化了网络和文件操作
5. **调试辅助**：`xdebug` 提供了丰富的调试工具
6. **高级功能**：`xfieldmask`、`xp2c` 等提供了专业领域的解决方案

合理使用这些工具包可以大大提高开发效率，减少重复代码，提高代码质量和可维护性。在实际开发中，建议根据具体需求选择合适的工具包，并遵循最佳实践，以充分发挥这些工具包的优势。

## 下一步

- 了解 [Worker 命令执行](./worker.md)
- 学习 [日志组件](./xlog.md)
- 探索 [配置系统](./config.md)
- 查看 [MeiliSearch 存储](./meilisearch.md)