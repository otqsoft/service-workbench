<template>
  <div class="app">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <span class="app-title">Service Workbench</span>
      </div>
      <div class="toolbar-right">
        <button class="btn btn-primary" @click="startAll">启动全部</button>
        <button class="btn btn-danger" @click="stopAll">停止全部</button>
        <button class="btn" @click="openAddDialog">+ 添加服务</button>
      </div>
    </div>

    <div class="main">
      <!-- 左侧服务列表 -->
      <div class="sidebar">
        <div
          v-for="svc in services"
          :key="svc.config.id"
          class="service-item"
          :class="{ active: activeTab === svc.config.id }"
          @click="switchTab(svc.config.id)"
        >
          <span class="status-dot" :class="svc.status"></span>
          <span class="service-name">{{ svc.config.name }}</span>
          <div class="service-actions" @click.stop>
            <button v-if="svc.status !== 'running'" class="mini-btn" @click="startService(svc.config.id)" title="启动">▶</button>
            <button v-else class="mini-btn" @click="stopService(svc.config.id)" title="停止">■</button>
            <button class="mini-btn" @click="editService(svc.config.id)" title="编辑">✎</button>
            <button class="mini-btn" @click="deleteService(svc.config.id)" title="删除">✕</button>
          </div>
        </div>
        <div v-if="services.length === 0" class="empty-tip">
          还没有服务，点击右上角添加
        </div>
      </div>

      <!-- 右侧 Tab 日志区 -->
      <div class="content">
        <div v-if="openTabs.length === 0" class="no-tab">
          从左侧选择一个服务查看日志
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
              {{ getSvc(tabId)?.config.name || tabId }}
              <span class="tab-close" @click.stop="closeTab(tabId)">×</span>
            </div>
          </div>
          <div class="log-area">
            <div v-for="tabId in openTabs" v-show="activeTab === tabId" :key="tabId" class="log-panel">
              <div class="log-toolbar">
                <button class="mini-btn" @click="clearLog(tabId)">清屏</button>
                <label class="search-box">
                  <input v-model="searchKeyword[tabId]" placeholder="搜索..." />
                </label>
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
import { ListServices, AddService, DeleteService, StartService, StopService, StartAll, StopAll, GetLogs, OpenFileDialog, OpenDirectoryDialog } from '../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

const services = ref([])
const openTabs = ref([])
const activeTab = ref('')
const logsMap = ref({})
const searchKeyword = ref({})
const logViewRef = ref(null)

const dialogVisible = ref(false)
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
})

onUnmounted(() => {
  EventsOff('log')
  EventsOff('status')
})
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: "Microsoft YaHei", sans-serif; background: #1e1e1e; color: #ddd; }
.app { display: flex; flex-direction: column; height: 100vh; }

.toolbar {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8px 16px; background: #2d2d2d; border-bottom: 1px solid #444;
}
.app-title { font-weight: bold; font-size: 16px; }
.toolbar-right { display: flex; gap: 8px; }

.btn {
  padding: 6px 14px; border: 1px solid #555; background: #3a3a3a; color: #ddd;
  border-radius: 4px; cursor: pointer; font-size: 13px;
}
.btn:hover { background: #4a4a4a; }
.btn-primary { background: #2563eb; border-color: #2563eb; }
.btn-primary:hover { background: #3b82f6; }
.btn-danger { background: #dc2626; border-color: #dc2626; }
.btn-danger:hover { background: #ef4444; }

.main { display: flex; flex: 1; overflow: hidden; }

.sidebar {
  width: 240px; background: #252526; border-right: 1px solid #444;
  overflow-y: auto; padding: 8px 0;
}
.service-item {
  display: flex; align-items: center; padding: 8px 12px; cursor: pointer;
  gap: 8px; font-size: 13px;
}
.service-item:hover { background: #2d2d2d; }
.service-item.active { background: #37373d; border-left: 3px solid #2563eb; padding-left: 9px; }
.service-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.status-dot, .tab-dot {
  width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0;
  background: #666;
}
.status-dot.running, .tab-dot.running { background: #22c55e; }
.status-dot.stopped, .tab-dot.stopped { background: #666; }
.status-dot.error, .tab-dot.error { background: #ef4444; }
.status-dot.starting, .tab-dot.starting { background: #eab308; }

.service-actions { display: none; gap: 2px; }
.service-item:hover .service-actions { display: flex; }
.mini-btn {
  background: none; border: none; color: #aaa; cursor: pointer; font-size: 12px; padding: 2px 4px;
}
.mini-btn:hover { color: #fff; }

.empty-tip { padding: 20px; text-align: center; color: #666; font-size: 13px; }

.content { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.no-tab { flex: 1; display: flex; align-items: center; justify-content: center; color: #555; }

.tab-bar { display: flex; background: #2d2d2d; border-bottom: 1px solid #444; overflow-x: auto; }
.tab {
  padding: 8px 16px; cursor: pointer; font-size: 13px; display: flex; align-items: center; gap: 6px;
  border-right: 1px solid #444; white-space: nowrap;
}
.tab.active { background: #1e1e1e; }
.tab-close { margin-left: 4px; color: #888; cursor: pointer; }
.tab-close:hover { color: #fff; }

.log-area { flex: 1; overflow: hidden; }
.log-panel { height: 100%; display: flex; flex-direction: column; }
.log-toolbar {
  display: flex; align-items: center; gap: 12px; padding: 6px 12px;
  background: #252526; border-bottom: 1px solid #444;
}
.search-box input {
  padding: 4px 8px; background: #333; border: 1px solid #555; color: #ddd;
  border-radius: 3px; font-size: 12px; width: 200px;
}
.log-count { color: #666; font-size: 12px; margin-left: auto; }

.log-view {
  flex: 1; overflow-y: auto; padding: 8px 12px; font-family: Consolas, monospace;
  font-size: 12px; line-height: 1.5; text-align: left;
}
.log-line {
  white-space: pre-wrap; word-break: break-all; text-align: left;
  display: block; padding: 0; margin: 0;
}
.log-line.err { color: #f87171; }
.log-line .hl { background: #854d0e; color: #fef08a; }

.modal-mask {
  position: fixed; inset: 0; background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
  background: #2d2d2d; border-radius: 8px; width: 560px; max-height: 80vh;
  overflow-y: auto; border: 1px solid #555;
}
.modal-title { padding: 16px; font-size: 16px; font-weight: bold; border-bottom: 1px solid #444; }
.modal-body { padding: 16px; }
.form-row { margin-bottom: 12px; }
.form-row label { display: block; font-size: 12px; color: #999; margin-bottom: 4px; }
.form-row input, .form-row select, .form-row textarea {
  width: 100%; padding: 6px 10px; background: #1e1e1e; border: 1px solid #555;
  color: #ddd; border-radius: 4px; font-size: 13px;
}
.input-group { display: flex; gap: 8px; }
.input-group input { flex: 1; }
.modal-footer { padding: 12px 16px; border-top: 1px solid #444; display: flex; justify-content: flex-end; gap: 8px; }
</style>
