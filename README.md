# FetalMovementCount
一款运行在PC端的胎动计数软件，常见的数胎动的主要是手机APP，现在大部分人使用PC办公，所以使用PC端软件数胎动可能更方便易用，固开发该软件。

## 编译测试
使用以下命令进行编译测试：

```rsrc -manifest FetalMovementCount.exe.manifest -o rsrc.syso```

```go build -ldflags="-H windowsgui"```


## 待完善功能
- 日志输出到文件
- 输出胎动报告（csv文件简单一些，或者excel，其他）
- 界面优化
- 老板键
- 软件信息
- 其他待考虑...