# Jupiter 框架工具包使用指南

本文档介绍 Jupiter 框架中 `pkg/util` 目录下所有工具包的功能、使用方法和最佳实践。

## 目录

- [xattr](#xattr) - 属性键值对管理工具
- [xbuffer](#xbuffer) - 无界缓冲区实现
- [xcast](#xcast) - 类型转换工具
- [xcrypto](#xcrypto) - 加密解密工具
- [xcycle](#xcycle) - 生命周期管理工具
- [xdebug](#xdebug) - 调试辅助工具
- [xdefer](#xdefer) - 延迟执行栈管理
- [xerror](#xerror) - 错误处理工具
- [xfieldmask](#xfieldmask) - 字段掩码工具
- [xfile](#xfile) - 文件操作工具
- [xgo](#xgo) - 并发控制工具
- [xmap](#xmap) - Map 操作工具
- [xmath](#xmath) - 数学运算工具
- [xnet](#xnet) - 网络相关工具
- [xp2c](#xp2c) - 负载均衡工具
- [xrand](#xrand) - 随机数生成工具
- [xreflect](#xreflect) - 反射工具
- [xretry](#xretry) - 重试机制工具
- [xslice](#xslice) - 切片操作工具
- [xstring](#xstring) - 字符串处理工具
- [xstruct](#xstruct) - 结构体操作工具
- [xtime](#xtime) - 时间处理工具

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

## xcrypto

### 功能说明

`xcrypto` 提供了丰富的加密解密工具，包括哈希计算、对称加密、非对称加密、编码解码等功能。

### 主要方法

#### 哈希计算
- `MD5(s string) string` - 计算 MD5 哈希
- `MD5Short(s string) string` - 返回 MD5 哈希的前 8 位
- `SHA1(s string) string` - 计算 SHA1 哈希
- `SHA256(s string) string` - 计算 SHA256 哈希
- `SHA512(s string) string` - 计算 SHA512 哈希

#### HMAC
- `HmacMD5(message, key string) string` - 计算 HMAC-MD5
- `HmacSHA1(message, key string) string` - 计算 HMAC-SHA1
- `HmacSHA256(message, key string) string` - 计算 HMAC-SHA256

#### Base64 编码
- `Base64Encode(s string) string` - Base64 编码
- `Base64Decode(s string) (string, error)` - Base64 解码
- `Base64URLEncode(s string) string` - Base64 URL 编码
- `Base64URLDecode(s string) (string, error)` - Base64 URL 解码

#### AES 加密
- `AESEncrypt(key, plaintext string) (string, error)` - AES-CBC 加密
- `AESDecrypt(key, ciphertext string) (string, error)` - AES-CBC 解密
- `AESGCMEncrypt(key, plaintext string) (string, error)` - AES-GCM 加密
- `AESGCMDecrypt(key, ciphertext string) (string, error)` - AES-GCM 解密

#### RSA 加密
- `GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error)` - 生成 RSA 密钥对
- `RSAEncrypt(pub *rsa.PublicKey, plaintext string) (string, error)` - RSA 加密
- `RSADecrypt(priv *rsa.PrivateKey, ciphertext string) (string, error)` - RSA 解密

#### 随机数生成
- `GenerateRandomString(length int) string` - 生成随机字符串
- `GenerateRandomBytes(length int) ([]byte, error)` - 生成随机字节

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xcrypto"
)

func main() {
    // 哈希计算
    md5Hash := xcrypto.MD5("hello")
    fmt.Printf("MD5: %s\n", md5Hash)

    sha256Hash := xcrypto.SHA256("hello")
    fmt.Printf("SHA256: %s\n", sha256Hash)

    // Base64 编码
    encoded := xcrypto.Base64Encode("hello world")
    fmt.Printf("Base64: %s\n", encoded)

    decoded, _ := xcrypto.Base64Decode(encoded)
    fmt.Printf("解码: %s\n", decoded)

    // AES 加密
    key := "0123456789abcdef"
    encrypted, err := xcrypto.AESEncrypt(key, "hello")
    if err == nil {
        fmt.Printf("加密: %s\n", encrypted)
        decrypted, _ := xcrypto.AESDecrypt(key, encrypted)
        fmt.Printf("解密: %s\n", decrypted)
    }

    // 生成随机字符串
    randomStr := xcrypto.GenerateRandomString(16)
    fmt.Printf("随机字符串: %s\n", randomStr)
}
```

---

## xmath

### 功能说明

`xmath` 提供了泛型数学运算工具，支持各种数学函数和统计计算。

### 主要方法

#### 基础运算
- `Abs[T](x T) T` - 绝对值
- `Max[T](x, y T) T` - 最大值
- `Min[T](x, y T) T` - 最小值
- `Clamp[T](value, min, max T) T` - 范围限制
- `Sum[T](numbers ...T) T` - 求和
- `Average[T](numbers ...T) float64` - 平均值

#### 统计函数
- `Median[T](numbers ...T) float64` - 中位数
- `Mode[T](numbers ...T) []T` - 众数

#### 数学函数
- `Pow[T](x, y T) float64` - 幂运算
- `Sqrt[T](x T) float64` - 平方根
- `Log[T](x T) float64` - 自然对数
- `Exp[T](x T) float64` - 指数函数
- `Ceil[T](x T) float64` - 向上取整
- `Floor[T](x T) float64` - 向下取整
- `Round[T](x T) float64` - 四舍五入

#### 数论函数
- `GCD[T](a, b T) T` - 最大公约数
- `LCM[T](a, b T) T` - 最小公倍数
- `Factorial[T](n T) uint64` - 阶乘
- `Fibonacci[T](n T) uint64` - 斐波那契数
- `IsPrime[T](n T) bool` - 素数判断
- `IsEven[T](n T) bool` - 偶数判断

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xmath"
)

func main() {
    // 基础运算
    fmt.Printf("Abs(-5): %d\n", xmath.Abs(-5))
    fmt.Printf("Max(3, 7): %d\n", xmath.Max(3, 7))
    fmt.Printf("Min(3, 7): %d\n", xmath.Min(3, 7))
    fmt.Printf("Clamp(15, 0, 10): %d\n", xmath.Clamp(15, 0, 10))

    // 统计函数
    numbers := []int{1, 2, 3, 4, 5, 5, 6}
    fmt.Printf("Sum: %d\n", xmath.Sum(numbers...))
    fmt.Printf("Average: %.2f\n", xmath.Average(numbers...))
    fmt.Printf("Median: %.2f\n", xmath.Median(numbers...))
    fmt.Printf("Mode: %v\n", xmath.Mode(numbers...))

    // 数学函数
    fmt.Printf("Pow(2, 3): %.2f\n", xmath.Pow(2, 3))
    fmt.Printf("Sqrt(16): %.2f\n", xmath.Sqrt(16))
    fmt.Printf("GCD(12, 8): %d\n", xmath.GCD(12, 8))
    fmt.Printf("IsPrime(7): %v\n", xmath.IsPrime(7))
}
```

---

## xrand

### 功能说明

`xrand` 提供了随机数和随机字符串生成工具，支持普通随机数和加密安全的随机数生成。

### 主要方法

#### 随机数生成
- `Int() int` - 非负随机整数
- `Intn(n int) int` - [0, n) 范围内的随机整数
- `IntRange(min, max int) int` - [min, max] 范围内的随机整数
- `Float64() float64` - [0.0, 1.0) 范围内的随机浮点数
- `Float64Range(min, max float64) float64` - [min, max) 范围内的随机浮点数
- `Bool() bool` - 随机布尔值

#### 随机字符串
- `String(length int) string` - 随机字母字符串
- `AlphanumericString(length int) string` - 随机字母数字字符串
- `NumericString(length int) string` - 随机数字字符串
- `Password(length int) string` - 随机密码字符串（含特殊字符）
- `Hex(length int) string` - 随机十六进制字符串
- `UUID() string` - UUID v4 字符串

#### 切片操作
- `Pick[T any](slice []T) T` - 随机选择一个元素
- `PickN[T any](slice []T, n int) []T` - 随机选择 n 个元素
- `Shuffle[T any](slice []T) []T` - 随机打乱切片
- `WeightedPick[T any](items []T, weights []int) T` - 根据权重随机选择

#### 加密安全随机数
- `SecureInt(max int64) (int64, error)` - 加密安全的随机整数
- `SecureBytes(length int) ([]byte, error)` - 加密安全的随机字节
- `SecureString(length int) (string, error)` - 加密安全的随机字符串

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xrand"
)

func main() {
    // 随机数
    fmt.Printf("随机整数: %d\n", xrand.Int())
    fmt.Printf("[0, 100) 随机: %d\n", xrand.Intn(100))
    fmt.Printf("[1, 10] 随机: %d\n", xrand.IntRange(1, 10))

    // 随机字符串
    fmt.Printf("随机字母: %s\n", xrand.String(10))
    fmt.Printf("随机数字: %s\n", xrand.NumericString(6))
    fmt.Printf("UUID: %s\n", xrand.UUID())

    // 切片操作
    fruits := []string{"apple", "banana", "cherry", "date"}
    fmt.Printf("随机选择: %s\n", xrand.Pick(fruits))
    fmt.Printf("随机打乱: %v\n", xrand.Shuffle(fruits))

    // 概率判断
    if xrand.Chance(0.3) {
        fmt.Println("30% 概率触发")
    }
}
```

---

## xslice

### 功能说明

`xslice` 提供了泛型切片操作工具，支持函数式编程风格的切片处理。

### 主要方法

#### 查询操作
- `Contains[T comparable](slice []T, target T) bool` - 检查元素是否存在
- `ContainsFunc[T any](slice []T, predicate func(T) bool) bool` - 检查是否满足条件
- `Index[T comparable](slice []T, target T) int` - 查找元素索引
- `IndexFunc[T any](slice []T, predicate func(T) bool) int` - 查找满足条件的索引
- `First[T any](slice []T, predicate func(T) bool) (T, bool)` - 查找第一个满足条件的元素
- `Last[T any](slice []T, predicate func(T) bool) (T, bool)` - 查找最后一个满足条件的元素

#### 转换操作
- `Map[T any, R any](slice []T, transform func(T) R) []R` - 映射转换
- `FlatMap[T any, R any](slice []T, transform func(T) []R) []R` - 扁平映射
- `Filter[T any](slice []T, predicate func(T) bool) []T` - 过滤
- `Reduce[T any, R any](slice []T, initial R, reducer func(R, T) R) R` - 归约

#### 聚合操作
- `Sum[T ~int | ~float64](slice []T) T` - 求和
- `Min[T ~int | ~float64](slice []T) (T, bool)` - 最小值
- `Max[T ~int | ~float64](slice []T) (T, bool)` - 最大值
- `Count[T any](slice []T, predicate func(T) bool) int` - 计数
- `All[T any](slice []T, predicate func(T) bool) bool` - 全部满足
- `Any[T any](slice []T, predicate func(T) bool) bool` - 任一满足

#### 分组操作
- `Unique[T comparable](slice []T) []T` - 去重
- `UniqueFunc[T any, K comparable](slice []T, keyFn func(T) K) []T` - 按键去重
- `GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T` - 分组
- `Partition[T any](slice []T, predicate func(T) bool) ([]T, []T)` - 分割

#### 切片操作
- `Chunk[T any](slice []T, size int) [][]T` - 分块
- `Flatten[T any](slice [][]T) []T` - 扁平化
- `Reverse[T any](slice []T)` - 反转
- `Copy[T any](slice []T) []T` - 复制
- `Compact[T comparable](slice []T) []T` - 移除零值

### 使用示例

```go
package main

import (
    "fmt"
    "strings"
    "github.com/weijinwen/jupiter/pkg/util/xslice"
)

func main() {
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    // 查询
    fmt.Printf("包含 5: %v\n", xslice.Contains(numbers, 5))
    fmt.Printf("第一个偶数: %v\n", xslice.First(numbers, func(n int) bool { return n%2 == 0 }))

    // 转换
    doubled := xslice.Map(numbers, func(n int) int { return n * 2 })
    fmt.Printf("翻倍: %v\n", doubled)

    evens := xslice.Filter(numbers, func(n int) bool { return n%2 == 0 })
    fmt.Printf("偶数: %v\n", evens)

    // 聚合
    fmt.Printf("总和: %d\n", xslice.Sum(numbers))
    min, _ := xslice.Min(numbers)
    fmt.Printf("最小值: %d\n", min)

    // 分组
    words := []string{"apple", "banana", "avocado", "blueberry", "cherry"}
    grouped := xslice.GroupBy(words, func(s string) string {
        return strings.ToUpper(s[:1])
    })
    fmt.Printf("按首字母分组: %v\n", grouped)

    // 分块
    chunks := xslice.Chunk(numbers, 3)
    fmt.Printf("分块: %v\n", chunks)
}
```

---

## xstruct

### 功能说明

`xstruct` 提供了结构体操作工具，支持结构体深拷贝、字段操作、类型转换等功能。

### 主要方法

#### 拷贝操作
- `Copy(src, dst any) error` - 结构体深拷贝
- `Clone[T any](src T) (T, error)` - 克隆结构体
- `MustClone[T any](src T) T` - 克隆结构体（出错 panic）

#### 字段操作
- `GetField(obj any, field string) (any, error)` - 获取字段值
- `SetField(obj any, field string, value any) error` - 设置字段值
- `GetFieldNames(obj any) []string` - 获取所有字段名
- `GetFieldValues(obj any) []any` - 获取所有字段值
- `GetFieldMap(obj any) map[string]any` - 获取字段名到值的映射
- `HasField(obj any, field string) bool` - 检查字段是否存在

#### 类型转换
- `ToMap(obj any, tagName ...string) map[string]any` - 结构体转 map
- `MapToStruct(m map[string]any, obj any) error` - map 转结构体

#### 合并操作
- `Merge(dst, src any) error` - 合并结构体非零值字段
- `PickFields(src any, fields ...string) map[string]any` - 选择指定字段
- `OmitFields(src any, fields ...string) map[string]any` - 排除指定字段

#### 类型检查
- `IsStruct(v any) bool` - 检查是否为结构体
- `IsSlice(v any) bool` - 检查是否为切片
- `IsMap(v any) bool` - 检查是否为 map
- `TypeName(v any) string` - 获取类型名称
- `FullName(v any) string` - 获取完整类型名

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xstruct"
)

type User struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

func main() {
    user := User{Name: "张三", Age: 30, Email: "zhangsan@example.com"}

    // 深拷贝
    copied, _ := xstruct.Clone(user)
    fmt.Printf("拷贝: %+v\n", copied)

    // 字段操作
    name, _ := xstruct.GetField(user, "Name")
    fmt.Printf("Name: %v\n", name)

    // 结构体转 map
    m := xstruct.ToMap(user)
    fmt.Printf("Map: %v\n", m)

    // 选择字段
    picked := xstruct.PickFields(user, "Name", "Age")
    fmt.Printf("选择: %v\n", picked)

    // 排除字段
    omitted := xstruct.OmitFields(user, "Email")
    fmt.Printf("排除: %v\n", omitted)
}
```

---

## xtime

### 功能说明

`xtime` 提供了时间处理工具，包括时间格式化、解析、计算等功能。

### 主要方法

#### 时间获取
- `Now() time.Time` - 当前时间
- `NowUnix() int64` - 当前 Unix 时间戳（秒）
- `NowUnixMilli() int64` - 当前 Unix 时间戳（毫秒）
- `Today() time.Time` - 今天零点
- `Tomorrow() time.Time` - 明天零点
- `Yesterday() time.Time` - 昨天零点

#### 时间格式化
- `Format(t time.Time, layout string) string` - 格式化时间
- `FormatNow(layout string) string` - 格式化当前时间
- `Parse(layout, value string) (time.Time, error)` - 解析时间字符串
- `MustParse(layout, value string) time.Time` - 解析时间字符串（出错 panic）

#### 时间戳转换
- `UnixToDate(timestamp int64) string` - 时间戳转日期
- `UnixToDateTime(timestamp int64) string` - 时间戳转日期时间
- `DateToUnix(date string) (int64, error)` - 日期转时间戳
- `DateTimeToUnix(datetime string) (int64, error)` - 日期时间转时间戳

#### 时间计算
- `Duration(start, end time.Time) time.Duration` - 计算时间间隔
- `Age(birthday time.Time) int` - 计算年龄

#### 时间区间
- `StartOfDay(t time.Time) time.Time` - 一天开始
- `EndOfDay(t time.Time) time.Time` - 一天结束
- `StartOfWeek(t time.Time) time.Time` - 一周开始（周一）
- `EndOfWeek(t time.Time) time.Time` - 一周结束
- `StartOfMonth(t time.Time) time.Time` - 一月开始
- `EndOfMonth(t time.Time) time.Time` - 一月结束
- `StartOfYear(t time.Time) time.Time` - 一年开始
- `EndOfYear(t time.Time) time.Time` - 一年结束

#### 时间判断
- `IsToday(t time.Time) bool` - 是否是今天
- `IsTomorrow(t *time.Time) bool` - 是否是明天
- `IsYesterday(t time.Time) bool` - 是否是昨天
- `IsLeapYear(year int) bool` - 是否是闰年
- `Between(t, start, end time.Time) bool` - 是否在时间范围内

### 常量

```go
const (
    FormatDateTime     = "2006-01-02 15:04:05"
    FormatDate         = "2006-01-02"
    FormatTime         = "15:04:05"
    FormatDateTimeNano = "2006-01-02 15:04:05.999999999"
    FormatMonth        = "2006-01"
    FormatYear         = "2006"
)
```

### 使用示例

```go
package main

import (
    "fmt"
    "github.com/weijinwen/jupiter/pkg/util/xtime"
)

func main() {
    // 当前时间
    fmt.Printf("当前时间: %s\n", xtime.FormatNow(xtime.FormatDateTime))
    fmt.Printf("Unix 时间戳: %d\n", xtime.NowUnix())

    // 时间解析
    t, _ := xtime.Parse(xtime.FormatDateTime, "2024-01-15 10:30:00")
    fmt.Printf("解析时间: %s\n", xtime.Format(t, xtime.FormatDate))

    // 时间区间
    now := xtime.Now()
    fmt.Printf("今天开始: %s\n", xtime.Format(xtime.StartOfDay(now), xtime.FormatDateTime))
    fmt.Printf("本周结束: %s\n", xtime.Format(xtime.EndOfWeek(now), xtime.FormatDateTime))

    // 时间判断
    fmt.Printf("是否闰年: %v\n", xtime.IsLeapYear(2024))
    fmt.Printf("是否今天: %v\n", xtime.IsToday(now))
}
```

---

## 总结

Jupiter 框架的 `pkg/util` 目录提供了丰富的工具包，涵盖了开发中常见的各种需求：

1. **基础工具**：`xattr`、`xcast`、`xstring` 等提供了基础数据类型操作
2. **数学和加密**：`xmath`、`xcrypto`、`xrand` 等提供了数学运算和加密解密能力
3. **数据处理**：`xslice`、`xmap`、`xstruct` 等提供了泛型数据处理能力
4. **并发控制**：`xgo`、`xcycle` 等提供了强大的并发管理能力
5. **错误处理**：`xerror`、`xretry` 等提供了完善的错误处理机制
6. **网络和文件**：`xnet`、`xfile` 等简化了网络和文件操作
7. **时间处理**：`xtime` 提供了丰富的时间处理功能
8. **调试辅助**：`xdebug` 提供了丰富的调试工具
9. **高级功能**：`xfieldmask`、`xp2c` 等提供了专业领域的解决方案

合理使用这些工具包可以大大提高开发效率，减少重复代码，提高代码质量和可维护性。在实际开发中，建议根据具体需求选择合适的工具包，并遵循最佳实践，以充分发挥这些工具包的优势。

## 下一步

- 了解 [Worker 命令执行](./worker.md)
- 学习 [日志组件](./xlog.md)
- 探索 [配置系统](./config.md)
- 查看 [MeiliSearch 存储](./meilisearch.md)