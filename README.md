# Service Workbench

Windows 平台下的多进程服务管理工作台，把命令行窗口里跑的 Go/Python 服务收拢到一个 GUI 窗口统一管理，交付客户时无需接触命令行。

## 功能特性

- **多服务 Tab 管理**：每个服务一个独立 Tab，状态指示灯实时显示运行/停止/异常
- **实时日志查看**：捕获子进程 stdout/stderr，支持搜索、高亮、清屏
- **服务配置管理**：可执行路径、工作目录、启动参数、环境变量、日志编码等
- **一键操作**：单个启动/停止/重启，一键启动全部、停止全部
- **配置持久化**：所有服务配置存在 `services.json`，随 exe 一起交付即可

## 环境要求

| 依赖 | 版本 | 说明 |
|---|---|---|
| Go | >= 1.25 | Wails v2.16 要求 Go 1.25+ |
| Node.js | >= 18 | 前端构建，推荐 LTS |
| Wails CLI | v2.x | `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| WebView2 | 预装或运行时自动下载 | Windows 10/11 一般已自带 |

## 项目结构

```
service-workbench/
├── main.go              # 入口
├── app.go               # Wails 绑定（前端可调用的方法）
├── internal/
│   ├── model/           # 数据结构
│   ├── manager/         # 服务进程管理、日志捕获
│   └── config/         # 配置 JSON 读写
├── frontend/            # Vue3 前端
│   ├── src/
│   │   └── App.vue      # 主界面
│   └── wailsjs/         # Wails 自动生成的绑定
├── build/
│   └── bin/
│       └── ServiceWorkbench.exe   # 编译产物
└── wails.json            # Wails 项目配置
```

## 开发启动

```bash
# 1. 安装前端依赖
cd frontend
npm install
cd ..

# 2. 开发模式运行（前端热更新）
wails dev
```

## 编译打包

```bash
# 编译生产版单文件 exe
wails build

# 产物位置
# build/bin/ServiceWorkbench.exe
```

## 使用说明

### 添加服务

1. 点击右上角 **+ 添加服务**
2. 填写服务名称
3. 选择可执行程序：
   - Go 服务：选编译好的 `.exe`
   - Python 服务：选 `.py` 文件（自动用系统 python 启动）；如果 Python 不在 PATH，直接选你的 `python.exe`，在"启动参数"里填脚本路径
   - 批处理：选 `.bat` / `.cmd`
4. 工作目录留空则自动用程序所在目录
5. 启动参数：如 `--config prod.yaml --port 8081`
6. 环境变量：每行一个 `KEY=VALUE`
7. 保存后从左侧列表点服务名进入日志 Tab

### 快捷操作

- 左侧服务项 hover 显示：启动 ▶ / 停止 ■ / 编辑 ✎ / 删除 ✕
- 顶部工具栏：启动全部 / 停止全部
- 日志窗口：右上角搜索框过滤关键词，清屏按钮清空当前显示

### 配置文件

所有服务配置保存在 exe 同目录下的 `services.json`，可直接编辑或随交付包一起拷贝。

## 已知限制（待优化）

- 停止方式目前是强制结束进程，暂不支持优雅信号
- Python GBK 日志暂未做编码转换，可能乱码
- 崩溃自动重启、系统托盘、磁盘日志 tail 等功能待开发
