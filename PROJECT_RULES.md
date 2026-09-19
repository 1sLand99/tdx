### 8.1 项目概况与技术栈

- **项目**：通达信(tdx)协议 Go 实现（标准行情 7709 + 盘后数据 report file + 扩展行情 exhq 7727）。
- **技术栈**：Go 1.23+，依赖 `github.com/injoyai/conv`（类型转换）、`github.com/injoyai/logs`（日志）。
- **目录**：
  - `protocol/`：协议帧编解码与数据结构（`model_*.go`、`unit.go`、`frame.go`、`types.go`）。
  - `client.go` / `client_exhq.go`：连接与业务 API；`gbbq.go`、`manage.go` 等根级扩展。
  - `extend/`：离线数据读写（`local.go` 读 / `local_write.go` 写）、行情拉取、爬虫、HTTP server、指标计算（`model_kline.go`）。
  - `lib/`：bse（北交所官网爬虫，已弃用）、gbbq、xorms、zip 工具。
  - `example/`：大量独立可运行示例，每个示例一个目录，`go run <dir>` 直接运行。
- **外部依赖**：真实连接通达信服务器（`Dial*`/`DialDefault`）；示例与测试大多需要联网，测试可能因行情服务器/数据源而异，失败时先排查网络与数据可用性。

### 8.2 输出数据统一放 `./output/`（约定）

- 所有示例、脚本、爬虫、测试等**生成/落盘的数据文件**统一写入项目根目录 `./output/`，并按场景建子目录（如 `./output/dump/`、`./output/lc1/`）。
- **禁止**：把输出数据写入项目根目录其他位置、`example/` 内部、或散落任意路径。
- 目录若不存在则自动创建（如 `os.MkdirAll("./output/<sub>/", 0o755)`）。
- `./output/` 已加入 `.gitignore`，不提交到版本库。
- 已有历史产物（如 `dump/`、`.testdata_lc1/`）如继续使用，应迁移至 `./output/` 下的对应子目录，迁移完成后删除旧目录。

### 8.3 关键约定与踩坑（摘要）

- 代码/目录带交易所前缀，如 `sz000001`、`sh000001`、`bj899050`；北交所市场编码 `ExchangeBJ=2`。
- 内部 `Price` 单位=厘（元×1000，int64），`.Float64()` 得元；`Volume` 单位=手（股÷100）。
- 指数与股票成交量单位不同：本地文件 `.day/.lc1/.lc5` 中**指数存"手"、股票存"股"（×100）**，读写均按 `protocol.IsIndex(c)` 区分。
- 分钟线 float32 精度限制：往返会差 ±1厘，测试断言需容差；`.lc1` 每个交易日首条有"量=0额=0"占位记录（`.lc5` 无），读写逻辑须对称处理。
- 完整细节见根目录 `MEMORY.md`。
