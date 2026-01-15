# OpenShrimp - TODO

## 当前目标

开发一个跨平台 CLI + 桌面 UI 工具，统一管理散落在系统各处的 AI 工具配置、临时文件和运行环境。

## 已完成事项

### 项目初始化
- [x] 创建项目目录结构
- [x] 初始化 Go module
- [x] 实现配置管理系统 (YAML)
- [x] 实现工具发现模块 (scan)
- [x] 实现清理模块 (cleanup)
- [x] 实现健康检查模块 (check)
- [x] 实现统计模块 (stats)
- [x] 实现基础 CLI 命令框架

### 核心命令
- [x] **模型切换命令 (`switch`)**
- [x] **符号链接管理命令 (`link`)**
- [x] 配置备份命令 (`backup`)
- [x] 配置恢复命令 (`restore`)
- [x] **守护进程命令 (`daemon`)**

### 质量保障
- [x] 单元测试 (tests/ 目录)
- [x] 集成测试

### Web UI (新增)
- [x] **Tauri + Vue 3 桌面应用架构**
- [x] HTTP + WebSocket 后端服务
- [x] Dashboard 概览组件
- [x] ToolList 工具列表组件
- [x] ModelSwitcher 模型切换组件
- [x] BackupPanel 备份管理组件
- [x] LinkManager 符号链接管理组件
- [x] SchedulerPanel 定时任务组件
- [x] TailwindCSS 样式

### 待完成

#### 质量保障
- [x] CI/CD 流水线

#### 高级功能
- [x] 定时任务支持 (cron/scheduled)
- [x] API 凭据安全管理 (Keychain/env vars)

#### 分发
- [ ] Homebrew formula (自动化)
- [ ] DEB/RPM 包
- [x] Tauri 桌面应用打包
- [x] Docker 镜像发布

---

## 开发阶段

### Phase 1-4: 核心功能 (已完成) ✓

- CLI 命令 (scan, cleanup, switch, link, backup, restore, context, check, stats)
- 单元测试 + 集成测试

### Phase 5: Web UI (已完成) ✓

**技术栈:**
- **Tauri 2.x** - 桌面应用框架 (Rust)
- **Vue 3 + TypeScript** - 前端框架
- **Vite 5** - 构建工具
- **TailwindCSS 3** - 样式框架
- **gorilla/mux** - Go HTTP 路由
- **gorilla/websocket** - 实时通信

**新增文件:**
```
cmd/daemon/server.go      # HTTP + WebSocket 服务器
ui/                       # Vue 3 前端项目
  src/components/         # 5 个 UI 组件
  vite.config.ts
  package.json
tauri/                    # Tauri 配置
```

**API 端点:**
| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/tools` | GET | 工具列表 |
| `/api/tools/{name}/cleanup` | POST | 清理工具 |
| `/api/models` | GET | 模型列表 |
| `/api/switch` | POST | 切换模型 |
| `/api/backups` | GET/POST | 备份管理 |
| `/api/links` | GET/POST/DELETE | 链接管理 |
| `/api/scheduler` | GET/POST | 定时任务管理 |
| `/api/scheduler/{id}` | PUT/DELETE | 更新/删除任务 |
| `/api/scheduler/{id}/run` | POST | 手动运行任务 |
| `/api/stats` | GET | 统计信息 |
| `/ws` | WebSocket | 实时更新 |

### Phase 6: 定时任务 (已完成) ✓

- cron 调度器 (robfig/cron)
- 自动清理任务
- 自动备份任务
- CLI 命令: `ai-mgr scheduler`
- Web UI 组件

### Phase 7: 高级功能 (已完成) ✓

8. **API 凭据管理**
   - Keychain/Keyring 集成 (macOS Keychain, Linux Secret Service)
   - 加密文件存储回退 (AES-256-GCM)
   - 环境变量安全管理
   - CLI 命令: `ai-mgr credentials`
   - Web UI 组件

### Phase 8: CI/CD 与分发 (已完成) ✓

10. **CI/CD 流水线**
    - GitHub Actions CI 工作流 (测试、构建、lint、安全扫描)
    - GitHub Actions Release 工作流 (多平台发布、Homebrew、Docker)
    - SBOM 生成

11. **分发支持**
    - Docker 多阶段构建
    - Homebrew formula 模板
    - Bash/Zsh 自动补全

### Phase 9: Tauri 桌面应用 (已完成) ✓

12. **Tauri 桌面应用**
    - Tauri 2.x 集成
    - 自动启动后台服务
    - 原生 macOS 应用 (.app + .dmg)
    - Rust 命令: get_app_version, run_cli_command, check_system_health, open_url

---

## API 端点 (更新)

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/tools` | GET | 工具列表 |
| `/api/tools/{name}/cleanup` | POST | 清理工具 |
| `/api/models` | GET | 模型列表 |
| `/api/switch` | POST | 切换模型 |
| `/api/backups` | GET/POST | 备份管理 |
| `/api/backups/{id}/restore` | POST | 恢复备份 |
| `/api/backups/{id}` | DELETE | 删除备份 |
| `/api/links` | GET/POST/DELETE | 链接管理 |
| `/api/scheduler` | GET/POST | 定时任务管理 |
| `/api/scheduler/{id}` | PUT/DELETE | 更新/删除任务 |
| `/api/scheduler/{id}/run` | POST | 手动运行任务 |
| `/api/credentials` | GET/POST | 凭据管理 |
| `/api/credentials/{model}/{key}` | DELETE | 删除凭据 |
| `/api/credentials/{model}` | GET | 模型凭据 |
| `/api/stats` | GET | 统计信息 |
| `/ws` | WebSocket | 实时更新 |

---

## 关键约束

### 技术约束
- **语言**: Go + Rust (Tauri) + TypeScript
- **平台**: macOS + Linux (Windows 待定)
- **配置格式**: YAML

### 功能约束
- **非侵入式**: 不修改原始工具的配置
- **隐私优先**: API Key 只存储在本地

---

*最后更新: 2026-01-15*

## 2026-01-15 修复记录

### Tauri 桌面应用 API 代理修复
- 修复前端 API 请求无法连接到 daemon 的问题
- 实现 HTTP 服务器 API 代理功能（端口 3456 → 19999）
- 添加超时处理和连接池管理
- 修复前端 ToolList 组件类型定义不匹配问题

### 文件变更
- `tauri/src-tauri/src/lib.rs` - 添加 API 代理和静态文件服务
- `ui/src/components/ToolList.vue` - 修复 ScanResult 接口字段名
- `tauri/src-tauri/tauri.conf.json` - 更新 Tauri 2.x 配置
- `tauri/src-tauri/capabilities/default.json` - 添加权限配置
