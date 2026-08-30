# util 工具包

本目录包含常用的 Go 工具包，提供各种实用的工具函数。

## 新增工具包

### xslice - 切片操作工具

提供切片相关的通用操作函数：

- `Contains` / `ContainsFunc` - 检查元素是否存在
- `Filter` - 过滤切片
- `Map` / `FlatMap` - 转换切片元素
- `Reduce` - 归约操作
- `Unique` / `UniqueFunc` - 去重
- `GroupBy` - 分组
- `Partition` - 分区
- `First` / `Last` - 查找第一个/最后一个满足条件的元素
- `Index` / `IndexFunc` - 查找索引
- `Count` - 统计满足条件的元素数量
- `All` / `Any` - 检查所有/任意元素满足条件
- `Sum` / `Min` / `Max` - 求和/最小值/最大值
- `Reverse` / `Copy` - 反转/复制
- `Chunk` / `Flatten` - 分块/展平
- `Compact` - 去除零值

### xtime - 时间工具

提供时间相关的工具函数：

- `Now` / `NowUnix` / `NowUnixMilli` - 获取当前时间
- `Today` / `Tomorrow` / `Yesterday` - 获取日期
- `Format` / `Parse` - 格式化和解析
- `UnixToDate` / `DateToUnix` - Unix 时间戳转换
- `Age` - 计算年龄
- `IsLeapYear` / `DaysInMonth` - 闰年和每月天数
- `StartOfDay` / `EndOfDay` - 一天的开始/结束
- `StartOfWeek` / `EndOfWeek` - 一周的开始/结束
- `StartOfMonth` / `EndOfMonth` - 一月的开始/结束
- `StartOfYear` / `EndOfYear` - 一年的开始/结束
- `IsToday` / `IsTomorrow` / `IsYesterday` - 判断是否是今天/明天/昨天
- `Between` - 判断时间是否在区间内
- `Elapsed` - 获取人类可读的时间间隔

### xerror - 错误处理工具

提供增强的错误处理功能：

- `New` / `Newf` - 创建错误
- `Wrap` / `Wrapf` - 包装错误
- `Cause` - 获取根因
- `WithStack` - 添加堆栈信息
- `WithMessage` / `WithMessagef` - 添加消息
- 常见错误类型：`NotFound`, `InvalidInput`, `Internal`, `Unauthorized`, `Forbidden`, `Conflict`, `Timeout`
- 错误判断函数：`IsNotFound`, `IsInvalidInput`, `IsInternal`, `IsUnauthorized`, `IsForbidden`, `IsConflict`, `IsTimeout`, `IsClosed`

### xcrypto - 加密工具

提供加密、哈希和编码工具：

- `MD5` / `SHA1` / `SHA256` / `SHA512` - 哈希计算
- `HmacMD5` / `HmacSHA1` / `HmacSHA256` - HMAC 计算
- `Base64Encode` / `Base64Decode` - Base64 编解码
- `Base64URLEncode` / `Base64URLDecode` - Base64 URL 编解码
- `AESEncrypt` / `AESDecrypt` - AES-CBC 加密解密
- `AESGCMEncrypt` / `AESGCMDecrypt` - AES-GCM 加密解密
- `RSAEncrypt` / `RSADecrypt` - RSA 加密解密
- `GenerateRSAKeyPair` / `RSAKeyToPEM` / `PEMToRSAKey` - RSA 密钥管理
- `GenerateRandomString` / `GenerateRandomBytes` - 生成随机数
- `XOR` - XOR 加密
- `Caesar` / `ROT13` - 经典加密算法

### xmath - 数学工具

提供数学相关的工具函数：

- `Abs` / `Max` / `Min` / `Clamp` - 绝对值/最大/最小/限制
- `Sum` / `Average` / `Median` / `Mode` - 统计函数
- `Pow` / `Sqrt` / `Cbrt` - 幂/平方根/立方根
- `Log` / `Log10` / `Log2` / `Exp` - 对数/指数
- `Ceil` / `Floor` / `Round` / `Trunc` - 取整
- `Sign` / `IsNaN` / `IsInf` / `IsFinite` - 符号/特殊值判断
- `Lerp` / `Map` - 插值/映射
- `GCD` / `LCM` - 最大公约数/最小公倍数
- `Factorial` / `Fibonacci` - 阶乘/斐波那契
- `IsPrime` / `Factors` - 素数判断/因数分解
- `IsEven` / `IsOdd` - 偶数/奇数判断
- `Digits` / `ReverseDigits` / `IsPalindrome` - 数字操作

## 老工具包补充

### xstring - 字符串工具补充

新增函数：
- `Split` / `SplitN` - 字符串分割
- `Join` - 字符串连接
- `Repeat` - 字符串重复
- `Contains` / `ContainsAny` / `ContainsRune` - 包含检查
- `HasPrefix` / `HasSuffix` - 前缀/后缀检查
- `Index` / `IndexAny` / `IndexRune` / `LastIndex` / `LastIndexAny` - 索引查找
- `Count` - 计数
- `ToUpper` / `ToLower` / `ToTitle` - 大小写转换
- `Trim` / `TrimLeft` / `TrimRight` / `TrimSpace` - 去除空白
- `Replace` / `ReplaceAll` - 替换
- `EqualFold` - 忽略大小写比较
- `IsBlank` / `IsNotBlank` / `DefaultIfBlank` - 空白检查
- `Truncate` - 截断
- `Wrap` / `WrapIfNotEmpty` / `Unwrap` - 包装/解包装
- `Reverse` - 反转
- `PadLeft` / `PadRight` / `Center` - 填充

### xmap - Map 工具补充

新增函数：
- `Keys` / `Values` / `Entries` - 获取键/值/条目
- `ContainsKey` / `ContainsValue` - 包含检查
- `GetOrDefault` / `GetOrElse` - 获取值或默认值
- `Filter` - 过滤
- `MapValues` / `MapKeys` - 映射转换
- `ForEach` - 遍历
- `Size` / `IsEmpty` / `IsNotEmpty` - 大小检查
- `Clear` / `Copy` - 清空/复制
- `Merge` / `MergeInPlace` - 合并
- `KeysSorted` / `KeysSortedFunc` - 排序的键
- `Equal` - 相等比较
- `Pick` / `Omit` - 选择/排除字段
- `GroupBy` - 分组
- `Reduce` - 归约
- `Any` / `All` / `Count` - 条件检查
- `FromSlice` / `FromSliceWithValues` - 从切片创建

### xcast - 类型转换工具补充

新增函数：
- `ToString` / `ToStringE` - 转换为字符串
- `ToInt` / `ToIntE` / `ToInt64` / `ToInt64E` - 转换为整数
- `ToFloat64` / `ToFloat64E` - 转换为浮点数
- `ToBool` / `ToBoolE` - 转换为布尔值
- `ToStringSlice` / `ToIntSlice` / `ToFloat64Slice` - 转换为切片
- `ToStringMap` / `ToStringMapString` / `ToMap` - 转换为 Map
- `ToSlice` - 转换为 []any
- `ToUint` / `ToUintE` / `ToUint64` / `ToUint64E` - 转换为无符号整数
- `IsEmpty` / `IsNotEmpty` / `IsZero` / `IsNotZero` - 空值检查
- `ToPtr` / `ToVal` - 指针转换
- `ToStringPtr` / `ToIntPtr` / `ToInt64Ptr` / `ToFloat64Ptr` / `ToBoolPtr` - 创建指针
- `IntToString` / `Int64ToString` / `Float64ToString` - 数值转字符串
- `StringToInt` / `StringToInt64` / `StringToFloat64` / `StringToBool` - 字符串转数值

### xnet - 网络工具补充

新增函数：
- `IsPortAvailable` - 检查端口是否可用
- `IsTCPPortOpen` / `IsUDPPortOpen` - 检查端口是否开放
- `GetAvailablePort` / `GetAvailablePorts` - 获取可用端口
- `IsIPv4` / `IsIPv6` - IP 版本检查
- `IsPrivateIP` / `IsLoopbackIP` / `IsLinkLocalIP` - IP 类型检查
- `NormalizeIPv6` - IPv6 标准化
- `IsCIDR` / `IsMACAddress` - 地址格式检查
- `GetInterfaceByName` / `GetInterfaces` - 网络接口
- `GetHostByName` / `GetAddrInfo` - DNS 解析

### xfile - 文件工具补充

新增函数：
- `ReadFile` / `ReadFileString` - 读取文件
- `WriteFile` / `WriteFileString` - 写入文件
- `AppendFile` / `AppendFileString` - 追加内容
- `CopyFile` / `MoveFile` - 复制/移动文件
- `DeleteFile` / `DeleteDir` - 删除文件/目录
- `GetFileSize` / `GetFileMode` - 获取文件信息
- `IsFile` - 检查是否为文件
- `GetTempDir` / `CreateTempFile` / `CreateTempDir` - 临时文件
- `GetExt` / `GetBase` / `GetDir` - 路径解析
- `Abs` / `Rel` / `Join` / `Split` - 路径操作
- `Walk` / `Glob` - 遍历/匹配
- `Chmod` / `Chown` - 权限修改
- `GetHomeDir` / `GetWorkDir` / `GetExecDir` - 目录获取

## 使用示例

```go
import (
    "github.com/weijinwen/jupiter/pkg/util/xslice"
    "github.com/weijinwen/jupiter/pkg/util/xtime"
    "github.com/weijinwen/jupiter/pkg/util/xerror"
    "github.com/weijinwen/jupiter/pkg/util/xcrypto"
    "github.com/weijinwen/jupiter/pkg/util/xmath"
    "github.com/weijinwen/jupiter/pkg/util/xstring"
    "github.com/weijinwen/jupiter/pkg/util/xmap"
    "github.com/weijinwen/jupiter/pkg/util/xcast"
    "github.com/weijinwen/jupiter/pkg/util/xnet"
    "github.com/weijinwen/jupiter/pkg/util/xfile"
)

// xslice 示例
nums := []int{1, 2, 3, 4, 5, 6}
evens := xslice.Filter(nums, func(i int) bool { return i%2 == 0 })
// evens = [2, 4, 6]

// xtime 示例
now := xtime.Now()
formatted := xtime.Format(now, xtime.FormatDateTime)
age := xtime.Age(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))

// xerror 示例
err := xerror.NotFound("user 123")
if xerror.IsNotFound(err) {
    // handle not found
}

// xcrypto 示例
hash := xcrypto.SHA256("hello world")
encrypted, _ := xcrypto.AESEncrypt(key, plaintext)

// xmath 示例
result := xmath.Max(10, 20)
avg := xmath.Average(1, 2, 3, 4, 5)

// xstring 示例
s := xstring.Split("a,b,c", ",")
joined := xstring.Join([]string{"a", "b", "c"}, "-")

// xmap 示例
m := map[string]int{"a": 1, "b": 2, "c": 3}
keys := xmap.Keys(m)
filtered := xmap.Filter(m, func(k string, v int) bool { return v > 1 })

// xcast 示例
s := xcast.ToString(123)
i := xcast.ToInt("456")

// xnet 示例
isAvailable := xnet.IsPortAvailable("localhost", 8080)
isIPv4 := xnet.IsIPv4("192.168.1.1")

// xfile 示例
data, _ := xfile.ReadFile("test.txt")
xfile.WriteFile("output.txt", []byte("hello"))
```

## 运行测试

```bash
# 运行所有测试
go test ./pkg/util/...

# 运行特定包测试
go test ./pkg/util/xslice/...
go test ./pkg/util/xtime/...
go test ./pkg/util/xerror/...
go test ./pkg/util/xcrypto/...
go test ./pkg/util/xmath/...
go test ./pkg/util/xstring/...
go test ./pkg/util/xmap/...
go test ./pkg/util/xcast/...
go test ./pkg/util/xnet/...
go test ./pkg/util/xfile/...
```
