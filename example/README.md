# license-scanner 样例（example/）

本目录是一个可被扫描的迷你源码树，用于演示 `license-scanner` 的扫描能力。

## 目录结构

- `LICENSE` —— MIT 许可证文件（license 文件名命中，内容签名识别为 MIT）
- `src/lib/lib.go` —— 带头注释 `// Apache License ... Apache Software Foundation` 的源文件（按源码头识别为 Apache-2.0）
- `src/lib/helper.go` —— 任何许可证声明都没有的源文件（被跳过，不计入 findings）

## 运行

```bash
go run . -path example -target Apache-2.0
```

## 期望输出

```
LICENSE                                 [MIT]        OK           MIT obligations do not exceed Apache-2.0
src/lib/lib.go                          [Apache-2.0] OK           Apache-2.0 obligations do not exceed Apache-2.0

2 file(s), 0 incompatible with Apache-2.0
```

（不同 Go 版本下文件名对齐宽度可能略有差异，但 `MIT`/`Apache-2.0` 与 `OK`、以及末尾 `2 file(s), 0 incompatible` 应一致。`src/lib/helper.go` 因无许可证声明不出现在 findings 中。）

指定不兼容目标时（如 `-target MIT`）会判 `src/lib/lib.go` 为 `INCOMPATIBLE` 并以退出码 1 结束。
