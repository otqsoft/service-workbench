<template>
  <div class="app">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <div class="logo-dot"></div>
        <span class="app-title">服务工作台</span>
        <span class="version-tag">v0.1</span>
      </div>
      <div class="toolbar-right">
        <button class="btn btn-start" @click="startAll">
          <span>▶</span> 启动全部
        </button>
        <button class="btn btn-stop" @click="stopAll">
          <span>■</span> 停止全部
        </button>
        <div class="divider"></div>
        <button class="btn btn-add" @click="openAddDialog">+ 添加服务</button>
      </div>
    </div>

    <div class="main">
      <!-- 左侧服务列表 -->
      <div class="sidebar">
        <div class="sidebar-header">服务列表 ({{ services.length }})</div>
        <div
          v-for="svc in services"
          :key="svc.config.id"
          class="service-item"
          :class="{ active: activeTab === svc.config.id, running: svc.status === 'running' }"
          @click="switchTab(svc.config.id)"
        >
          <div class="service-item-main">
            <span class="status-dot" :class="svc.status"></span>
            <span class="service-name">{{ svc.config.name }}</span>
          </div>
          <div class="service-actions" @click.stop>
            <button v-if="svc.status !== 'running'" class="mini-btn" @click="startService(svc.config.id)" title="启动">▶</button>
            <button v-else class="mini-btn" @click="stopService(svc.config.id)" title="停止">■</button>
            <button class="mini-btn" @click="editService(svc.config.id)" title="编辑">✎</button>
            <button class="mini-btn danger" @click="deleteService(svc.config.id)" title="删除">✕</button>
          </div>
        </div>
        <div v-if="services.length === 0" class="empty-tip">
          <div class="empty-icon">📦</div>
          <div>还没有服务</div>
          <div class="empty-sub">点击右上角"+ 添加服务"开始配置</div>
        </div>
      </div>

      <!-- 右侧 Tab 日志区 -->
      <div class="content">
        <div v-if="openTabs.length === 0" class="no-tab">
          <div class="no-tab-icon">🖥️</div>
          <div>从左侧选择一个服务查看日志</div>
        </div>
        <template v-else>
          <div class="tab-bar">
            <div
              v-for="tabId in openTabs"
              :key="tabId"
              class="tab"
              :class="{ active: activeTab === tabId }"
              @click="activeTab = tabId"
            >
              <span class="tab-dot" :class="getSvc(tabId)?.status"></span>
              <span class="tab-name">{{ getSvc(tabId)?.config.name || tabId }}</span>
              <span class="tab-close" @click.stop="closeTab(tabId)">×</span>
            </div>
          </div>
          <div class="log-area">
            <div v-for="tabId in openTabs" v-show="activeTab === tabId" :key="tabId" class="log-panel">
              <div class="log-toolbar">
                <button class="btn small" @click="clearLog(tabId)">清屏</button>
                <div class="search-box">
                  <span class="search-icon">🔍</span>
                  <input v-model="searchKeyword[tabId]" placeholder="过滤日志..." />
                </div>
                <span class="log-count">{{ filteredLogs(tabId).length }} 行</span>
              </div>
              <div class="log-view" ref="logViewRef">
                <div
                  v-for="(line, i) in filteredLogs(tabId)"
                  :key="i"
                  class="log-line"
                  :class="{ err: line.isErr }"
                  v-html="highlight(line.line, tabId)"
                ></div>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 退出确认弹窗 -->
    <div v-if="quitDialogVisible" class="modal-mask">
      <div class="modal quit-modal">
        <div class="quit-icon">
          <span class="quit-icon-dot"></span>
          ⏻
        </div>
        <div class="quit-title">确认退出</div>
        <div class="quit-text">退出将停止所有正在运行的服务，确定要退出吗？</div>
        <div class="modal-footer quit-footer">
          <button class="btn" @click="quitDialogVisible = false">取消</button>
          <button class="btn btn-quit" @click="confirmQuit">退出程序</button>
        </div>
      </div>
    </div>

    <!-- 配置弹窗 -->
    <div v-if="dialogVisible" class="modal-mask" @click.self="dialogVisible = false">
      <div class="modal">
        <div class="modal-title">{{ editingConfig.id ? '编辑服务' : '添加服务' }}</div>
        <div class="modal-body">
          <div class="form-row">
            <label>服务名称</label>
            <input v-model="editingConfig.name" placeholder="如：用户认证服务" />
          </div>
          <div class="form-row">
            <label>可执行程序</label>
            <div class="input-group">
              <input v-model="editingConfig.exePath" placeholder="选择 exe / py / bat 文件" />
              <button class="btn" @click="pickExe">浏览...</button>
            </div>
          </div>
          <div class="form-row">
            <label>工作目录</label>
            <div class="input-group">
              <input v-model="editingConfig.workDir" placeholder="留空则使用程序所在目录" />
              <button class="btn" @click="pickWorkDir">浏览...</button>
            </div>
          </div>
          <div class="form-row">
            <label>启动参数</label>
            <input v-model="argsStr" placeholder="如：--config prod.yaml --port 8081" />
          </div>
          <div class="form-row">
            <label>环境变量</label>
            <textarea v-model="envStr" rows="3" placeholder="每行一个 KEY=VALUE"></textarea>
          </div>
          <div class="form-row-2">
            <div class="form-row">
              <label>日志编码</label>
              <select v-model="editingConfig.logEncoding">
                <option value="utf-8">UTF-8</option>
                <option value="gbk">GBK</option>
              </select>
            </div>
            <div class="form-row">
              <label>启动顺序</label>
              <input v-model.number="editingConfig.startOrder" type="number" />
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn" @click="dialogVisible = false">取消</button>
          <button class="btn btn-primary" @click="saveConfig">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { ListServices, AddService, DeleteService, StartService, StopService, StartAll, StopAll, GetLogs, OpenFileDialog, OpenDirectoryDialog, ConfirmQuit } from '../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

const services = ref([])
const openTabs = ref([])
const activeTab = ref('')
const logsMap = ref({})
const searchKeyword = ref({})
const logViewRef = ref(null)

const dialogVisible = ref(false)
const quitDialogVisible = ref(false)
const editingConfig = ref({
  id: '', name: '', exePath: '', workDir: '', args: [], env: {},
  logFilePath: '', logEncoding: 'utf-8', autoRestart: false, startOrder: 0, startDelay: 0
})
const argsStr = ref('')
const envStr = ref('')

async function refreshList() {
  services.value = await ListServices()
}

function getSvc(id) {
  return services.value.find(s => s.config.id === id)
}

function switchTab(id) {
  if (!openTabs.value.includes(id)) {
    openTabs.value.push(id)
    GetLogs(id).then(lines => {
      logsMap.value[id] = lines.map(l => ({ line: l, isErr: l.toLowerCase().includes('error') || l.toLowerCase().includes('panic'), ts: 0 }))
    })
  }
  activeTab.value = id
}

function closeTab(id) {
  openTabs.value = openTabs.value.filter(t => t !== id)
  if (activeTab.value === id) {
    activeTab.value = openTabs.value[openTabs.value.length - 1] || ''
  }
}

async function startService(id) {
  await StartService(id)
  setTimeout(refreshList, 300)
}

async function stopService(id) {
  await StopService(id)
  setTimeout(refreshList, 300)
}

async function startAll() {
  await StartAll()
  setTimeout(refreshList, 500)
}

async function stopAll() {
  await StopAll()
  setTimeout(refreshList, 500)
}

function openAddDialog() {
  editingConfig.value = {
    id: '', name: '', exePath: '', workDir: '', args: [], env: {},
    logFilePath: '', logEncoding: 'utf-8', autoRestart: false, startOrder: 0, startDelay: 0
  }
  argsStr.value = ''
  envStr.value = ''
  dialogVisible.value = true
}

function editService(id) {
  const svc = getSvc(id)
  if (!svc) return
  editingConfig.value = { ...svc.config }
  argsStr.value = (svc.config.args || []).join(' ')
  envStr.value = Object.entries(svc.config.env || {}).map(([k, v]) => `${k}=${v}`).join('\n')
  dialogVisible.value = true
}

async function saveConfig() {
  editingConfig.value.args = argsStr.value ? argsStr.value.split(/\s+/) : []
  const env = {}
  envStr.value.split('\n').forEach(line => {
    line = line.trim()
    if (!line) return
    const idx = line.indexOf('=')
    if (idx > 0) env[line.slice(0, idx)] = line.slice(idx + 1)
  })
  editingConfig.value.env = env
  await AddService(editingConfig.value)
  dialogVisible.value = false
  refreshList()
}

async function deleteService(id) {
  if (!confirm('确定删除该服务配置？')) return
  await DeleteService(id)
  closeTab(id)
  refreshList()
}

async function pickExe() {
  const path = await OpenFileDialog('选择可执行程序')
  if (path) editingConfig.value.exePath = path
}

async function pickWorkDir() {
  const path = await OpenDirectoryDialog('选择工作目录')
  if (path) editingConfig.value.workDir = path
}

function confirmQuit() {
  quitDialogVisible.value = false
  ConfirmQuit()
}

function clearLog(id) {
  logsMap.value[id] = []
}

function filteredLogs(id) {
  const all = logsMap.value[id] || []
  const kw = searchKeyword.value[id]
  if (!kw) return all
  return all.filter(l => l.line.toLowerCase().includes(kw.toLowerCase()))
}

function highlight(line, id) {
  const kw = searchKeyword.value[id]
  if (!kw) return escapeHtml(line)
  const escaped = escapeHtml(line)
  const re = new RegExp(`(${escapeRegex(kw)})`, 'gi')
  return escaped.replace(re, '<span class="hl">$1</span>')
}

function escapeHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
}
function escapeRegex(s) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

let logHandler, statusHandler
onMounted(async () => {
  await refreshList()
  logHandler = (entry) => {
    if (!logsMap.value[entry.serviceId]) logsMap.value[entry.serviceId] = []
    logsMap.value[entry.serviceId].push({ line: entry.line, isErr: entry.isErr, ts: entry.timestamp })
    if (logsMap.value[entry.serviceId].length > 2000) {
      logsMap.value[entry.serviceId] = logsMap.value[entry.serviceId].slice(-2000)
    }
    nextTick(() => {
      if (logViewRef.value) logViewRef.value.scrollTop = logViewRef.value.scrollHeight
    })
  }
  statusHandler = (info) => {
    const idx = services.value.findIndex(s => s.config.id === info.config.id)
    if (idx >= 0) services.value[idx] = info
    else services.value.push(info)
  }
  EventsOn('log', logHandler)
  EventsOn('status', statusHandler)
  // 关闭窗口时后端会拦截并发出此事件，弹出退出确认框
  EventsOn('before-close', () => { quitDialogVisible.value = true })
})

onUnmounted(() => {
  EventsOff('log')
  EventsOff('status')
})
</script>

<style>
:root {
  --bg-primary: #0a1628;
  --bg-secondary: #0d1d33;
  --bg-tertiary: #12253f;
  --bg-hover: rgba(0, 212, 255, 0.06);
  --border: rgba(59, 130, 246, 0.16);
  --border-strong: rgba(0, 212, 255, 0.35);
  --text-primary: #e2e8f0;
  --text-secondary: #94a3b8;
  --text-muted: #64748b;
  --accent: #00d4ff;
  --accent-soft: rgba(0, 212, 255, 0.12);
  --accent-2: #3b82f6;
  --success: #34d399;
  --danger: #f87171;
  --warning: #fbbf24;
}

* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: "Microsoft YaHei", "PingFang SC", "Segoe UI", sans-serif;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-size: 13px;
  overflow: hidden;
}
.app {
  display: flex; flex-direction: column; height: 100vh;
  background:
    radial-gradient(ellipse 60% 40% at 15% 0%, rgba(59, 130, 246, 0.08), transparent),
    radial-gradient(ellipse 50% 45% at 85% 100%, rgba(0, 212, 255, 0.05), transparent),
    var(--bg-primary);
}

/* 顶部工具栏 */
.toolbar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 0 16px; height: 52px;
  background: linear-gradient(180deg, rgba(18, 37, 63, 0.65), rgba(13, 29, 51, 0.35));
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.toolbar-left { display: flex; align-items: center; gap: 10px; }
.logo-dot {
  width: 10px; height: 10px; border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), var(--accent-2));
  box-shadow: 0 0 10px rgba(0, 212, 255, 0.7);
  animation: pulse 2.4s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { box-shadow: 0 0 6px rgba(0, 212, 255, 0.45); }
  50% { box-shadow: 0 0 14px rgba(0, 212, 255, 0.85); }
}
.app-title { font-weight: 600; font-size: 15px; letter-spacing: 1px; }
.version-tag {
  font-size: 11px; color: var(--accent);
  background: var(--accent-soft); padding: 2px 9px; border-radius: 10px;
  border: 1px solid var(--border-strong);
}
.toolbar-right { display: flex; align-items: center; gap: 8px; }
.divider { width: 1px; height: 20px; background: var(--border); margin: 0 4px; }

/* 按钮 */
.btn {
  padding: 6px 14px; border: 1px solid var(--border);
  background: var(--bg-tertiary); color: var(--text-primary);
  border-radius: 6px; cursor: pointer; font-size: 12px;
  transition: all 0.15s ease;
}
.btn:hover { background: var(--bg-hover); border-color: var(--border-strong); color: var(--accent); }
.btn:active { transform: scale(0.97); }
.btn.small { padding: 4px 10px; font-size: 11px; }
.btn-primary {
  background: linear-gradient(135deg, var(--accent), var(--accent-2));
  border-color: transparent; color: #04101f; font-weight: 600;
}
.btn-primary:hover { filter: brightness(1.12); color: #04101f; }

/* 工具栏操作按钮：语义化配色 */
.toolbar-right .btn { font-weight: 600; letter-spacing: 0.5px; padding: 7px 16px; }
.btn-start {
  background: rgba(52, 211, 153, 0.12); border-color: rgba(52, 211, 153, 0.45);
  color: var(--success);
}
.btn-start:hover {
  background: rgba(52, 211, 153, 0.22); border-color: var(--success);
  color: #6ee7b7; box-shadow: 0 0 14px rgba(52, 211, 153, 0.25);
  transform: translateY(-1px);
}
.btn-start:active { transform: translateY(0) scale(0.97); }
.btn-stop {
  background: rgba(248, 113, 113, 0.1); border-color: rgba(248, 113, 113, 0.45);
  color: var(--danger);
}
.btn-stop:hover {
  background: rgba(248, 113, 113, 0.2); border-color: var(--danger);
  color: #fca5a5; box-shadow: 0 0 14px rgba(248, 113, 113, 0.25);
  transform: translateY(-1px);
}
.btn-stop:active { transform: translateY(0) scale(0.97); }
.btn-add {
  background: transparent; border-color: var(--border); color: var(--text-secondary);
}
.btn-add:hover {
  background: var(--accent-soft); border-color: var(--border-strong);
  color: var(--accent); transform: translateY(-1px);
}
.btn-add:active { transform: translateY(0) scale(0.97); }

/* 退出确认按钮 */
.btn-quit {
  background: linear-gradient(135deg, #f87171, #ef4444);
  border-color: transparent; color: #fff; font-weight: 600;
}
.btn-quit:hover { filter: brightness(1.12); color: #fff; box-shadow: 0 0 14px rgba(248, 113, 113, 0.35); }

/* 退出确认弹窗 */
.quit-modal {
  width: 380px; text-align: center; padding: 28px 24px 20px;
}
.quit-icon {
  width: 56px; height: 56px; margin: 0 auto 14px;
  display: flex; align-items: center; justify-content: center;
  font-size: 26px; color: var(--danger);
  border: 1px solid rgba(248, 113, 113, 0.35); border-radius: 50%;
  background: rgba(248, 113, 113, 0.08);
  position: relative;
}
.quit-icon-dot {
  position: absolute; inset: -1px; border-radius: 50%;
  border: 1px solid rgba(248, 113, 113, 0.35);
  animation: quit-ring 2s ease-out infinite;
}
@keyframes quit-ring {
  0% { transform: scale(1); opacity: 0.8; }
  100% { transform: scale(1.45); opacity: 0; }
}
.quit-title { font-size: 15px; font-weight: 600; margin-bottom: 8px; }
.quit-text { font-size: 12px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 4px; }
.quit-footer { justify-content: center; border-top: none; padding-top: 10px; }

/* 主区域 */
.main { display: flex; flex: 1; overflow: hidden; }

/* 左侧边栏 */
.sidebar {
  width: 260px; background: rgba(13, 29, 51, 0.5);
  border-right: 1px solid var(--border);
  display: flex; flex-direction: column;
  flex-shrink: 0;
}
.sidebar-header {
  padding: 14px 16px 8px; font-size: 11px; color: var(--text-muted);
  letter-spacing: 1px;
}
.service-item {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px; margin: 2px 8px; cursor: pointer;
  transition: background 0.15s ease;
  border-radius: 8px;
  border-left: 3px solid transparent;
}
.service-item:hover { background: var(--bg-hover); }
.service-item.active { background: var(--accent-soft); border-left-color: var(--accent); }
.service-item.running.active { border-left-color: var(--success); }
.service-item-main { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; }
.service-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

/* 状态点 */
.status-dot, .tab-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
  background: var(--text-muted);
  transition: background 0.2s;
}
.status-dot.running, .tab-dot.running {
  background: var(--success);
  box-shadow: 0 0 8px rgba(52, 211, 153, 0.8);
  animation: pulse-green 2s ease-in-out infinite;
}
@keyframes pulse-green {
  0%, 100% { box-shadow: 0 0 4px rgba(52, 211, 153, 0.5); }
  50% { box-shadow: 0 0 10px rgba(52, 211, 153, 0.95); }
}
.status-dot.stopped, .tab-dot.stopped { background: var(--text-muted); }
.status-dot.error, .tab-dot.error { background: var(--danger); box-shadow: 0 0 8px rgba(248, 113, 113, 0.8); }
.status-dot.starting, .tab-dot.starting {
  background: var(--warning);
  box-shadow: 0 0 8px rgba(251, 191, 36, 0.8);
  animation: pulse-green 1s ease-in-out infinite;
}

/* 服务项操作按钮：常驻布局，用透明度切换，避免 hover 时高度跳动 */
.service-actions { display: flex; gap: 2px; opacity: 0; visibility: hidden; transition: opacity 0.15s ease; }
.service-item:hover .service-actions { opacity: 1; visibility: visible; }
.mini-btn {
  background: none; border: none; color: var(--text-secondary);
  cursor: pointer; font-size: 12px; padding: 3px 5px; border-radius: 4px;
  transition: all 0.12s ease;
}
.mini-btn:hover { color: var(--accent); background: rgba(0, 212, 255, 0.1); }
.mini-btn.danger:hover { color: var(--danger); background: rgba(248, 113, 113, 0.12); }

/* 空状态 */
.empty-tip {
  padding: 40px 20px; text-align: center; color: var(--text-muted); font-size: 13px;
}
.empty-icon { font-size: 36px; margin-bottom: 12px; opacity: 0.5; }
.empty-sub { font-size: 11px; margin-top: 6px; opacity: 0.7; }

/* 右侧内容区 */
.content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.no-tab {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center; color: var(--text-muted); gap: 12px;
}
.no-tab-icon { font-size: 48px; opacity: 0.3; }

/* Tab 栏 */
.tab-bar {
  display: flex; background: rgba(13, 29, 51, 0.45);
  border-bottom: 1px solid var(--border);
  overflow-x: auto; flex-shrink: 0;
}
.tab {
  padding: 0 16px; height: 38px; cursor: pointer;
  display: flex; align-items: center; gap: 8px;
  font-size: 12px; border-right: 1px solid var(--border);
  white-space: nowrap; color: var(--text-secondary);
  transition: background 0.15s ease, color 0.15s ease;
  border-bottom: 2px solid transparent;
}
.tab:hover { background: var(--bg-hover); color: var(--text-primary); }
.tab.active {
  background: rgba(0, 212, 255, 0.05); color: var(--accent);
  border-bottom-color: var(--accent);
}
.tab-name { max-width: 150px; overflow: hidden; text-overflow: ellipsis; }
.tab-close {
  color: var(--text-muted); cursor: pointer;
  width: 18px; height: 18px; display: flex; align-items: center; justify-content: center;
  border-radius: 4px; font-size: 14px;
  transition: all 0.12s ease;
}
.tab-close:hover { color: var(--text-primary); background: rgba(148, 163, 184, 0.15); }

/* 日志区 */
.log-area { flex: 1; overflow: hidden; }
.log-panel { height: 100%; display: flex; flex-direction: column; }
.log-toolbar {
  display: flex; align-items: center; gap: 12px; padding: 7px 12px;
  background: rgba(13, 29, 51, 0.45); border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.search-box {
  display: flex; align-items: center; gap: 6px;
  background: rgba(10, 22, 40, 0.8); border: 1px solid var(--border);
  border-radius: 6px; padding: 4px 10px;
  transition: border-color 0.15s ease;
}
.search-box:focus-within { border-color: var(--border-strong); }
.search-box input {
  background: none; border: none; outline: none;
  color: var(--text-primary); font-size: 12px; width: 180px;
}
.search-box input::placeholder { color: var(--text-muted); }
.search-icon { font-size: 11px; opacity: 0.5; }
.log-count { color: var(--text-muted); font-size: 11px; margin-left: auto; }

.log-view {
  flex: 1; overflow-y: auto; padding: 12px 14px;
  font-family: "Cascadia Code", Consolas, "Courier New", monospace;
  font-size: 12px; line-height: 1.65;
  text-align: left;
}
.log-line {
  white-space: pre-wrap; word-break: break-all; text-align: left;
  display: block; padding: 0 4px; margin: 0;
  color: var(--text-secondary);
  border-radius: 3px;
}
.log-line:hover { background: rgba(59, 130, 246, 0.05); }
.log-line.err { color: var(--danger); background: rgba(248, 113, 113, 0.06); }
.log-line .hl { background: var(--warning); color: #04101f; border-radius: 2px; padding: 0 2px; }

/* 弹窗 */
.modal-mask {
  position: fixed; inset: 0; background: rgba(4, 16, 31, 0.7);
  display: flex; align-items: center; justify-content: center; z-index: 100;
  backdrop-filter: blur(4px);
}
.modal {
  background: var(--bg-secondary); border-radius: 12px;
  width: 560px; max-height: 85vh; overflow-y: auto;
  border: 1px solid var(--border-strong);
  box-shadow: 0 0 0 1px rgba(0, 212, 255, 0.08), 0 24px 64px rgba(0, 0, 0, 0.55);
}
.modal-title {
  padding: 16px 20px; font-size: 15px; font-weight: 600;
  border-bottom: 1px solid var(--border);
  color: var(--accent);
}
.modal-body { padding: 20px; }
.form-row { margin-bottom: 14px; }
.form-row-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.form-row label {
  display: block; font-size: 11px; color: var(--text-muted); margin-bottom: 5px;
}
.form-row input, .form-row select, .form-row textarea {
  width: 100%; padding: 8px 12px;
  background: rgba(10, 22, 40, 0.8); border: 1px solid var(--border);
  color: var(--text-primary); border-radius: 6px; font-size: 13px;
  outline: none; transition: border-color 0.15s ease;
}
.form-row input::placeholder, .form-row textarea::placeholder { color: var(--text-muted); }
.form-row input:focus, .form-row select:focus, .form-row textarea:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.12);
}
.input-group { display: flex; gap: 8px; }
.input-group input { flex: 1; }
.modal-footer {
  padding: 14px 20px; border-top: 1px solid var(--border);
  display: flex; justify-content: flex-end; gap: 8px;
}

/* 滚动条美化 */
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(59, 130, 246, 0.25); border-radius: 4px; }
::-webkit-scrollbar-thumb:hover { background: rgba(0, 212, 255, 0.45); }

/* 无障碍：尊重系统减弱动态效果设置 */
@media (prefers-reduced-motion: reduce) {
  .logo-dot, .status-dot.running, .status-dot.starting, .tab-dot.running, .quit-icon-dot {
    animation: none;
  }
  .service-actions { transition: none; }
  * { transition-duration: 0.01ms !important; }
}
</style>
