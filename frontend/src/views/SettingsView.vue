<template>
  <div class="page">
    <header class="navbar">
      <div class="navbar-left" @click="goHome">
        <div class="logo">
          <RiRadarFill size="22" color="#fff" />
        </div>
        <span class="brand-cn">文制星</span>
        <span class="brand-en">Boomerang</span>
      </div>
      <div class="navbar-right">
        <button class="back-btn" @click="goBack">
          <RiArrowLeftLine size="16" color="#8B7355" />
          <span>返回</span>
        </button>
      </div>
    </header>

    <div class="body">
      <aside class="sidebar">
        <div class="sidebar-header">
          <span class="sidebar-title">系统设置</span>
          <div class="sidebar-divider" />
          <span class="sidebar-subtitle">Settings</span>
        </div>
        <nav class="nav-list">
          <div class="nav-indicator" :style="indicatorStyle" />
          <button
            v-for="item in navItems"
            :key="item.id"
            class="nav-item"
            :class="{ 'nav-item-active': activeNav === item.id }"
            @click="activeNav = item.id"
          >
            <component :is="item.icon" :size="'22'" />
            <div class="nav-texts">
              <span class="nav-label">{{ item.label }}</span>
              <span class="nav-sublabel">{{ item.sublabel }}</span>
            </div>
          </button>
        </nav>
      </aside>

      <main class="content">
        <div class="content-header">
          <div class="content-header-left">
            <div class="content-header-icon" :style="{ background: currentNav.color }">
              <component :is="currentNav.icon" :size="'20'" color="#fff" />
            </div>
            <div class="content-header-texts">
              <h3 class="content-title">{{ currentNav.title }}</h3>
              <span class="content-subtitle">{{ currentNav.desc }}</span>
            </div>
          </div>
          <div class="content-header-actions">
            <button class="action-btn-cancel" @click="goBack">取消</button>
            <button class="action-btn-save" :style="{ background: currentNav.color }" @click="saveSettings">保存</button>
          </div>
        </div>

        <div class="content-scroll">
          <!-- 主题设置 -->
          <div v-if="activeNav === 'theme'" class="panel">
            <div class="theme-page-title">主题设置</div>
            <div class="theme-page-desc">选择你喜欢的界面风格</div>
            <div class="theme-cards-row">
              <div
                v-for="t in themes"
                :key="t.id"
                class="theme-card-new"
                :class="{ 'theme-card-selected': settingsStore.theme === t.id }"
                @click="settingsStore.setTheme(t.id)"
              >
                <div class="theme-preview-area" :style="{ background: t.previewBg }">
                  <div class="theme-preview-nav" :style="{ background: t.navBg }">
                    <div class="preview-logo" :style="{ background: t.logoBg }" />
                    <span class="preview-brand" :style="{ color: t.brandColor }">文制星</span>
                  </div>
                  <div class="theme-preview-body">
                    <div class="preview-sidebar" :style="{ background: t.sidebarBg }" />
                    <div class="preview-content">
                      <div class="preview-card" :style="{ background: t.card1Bg }" />
                      <div class="preview-card-sm" :style="{ background: t.card2Bg }" />
              </div>
            </div>
          </div>

          <!-- 技能预览模态框 -->
          <div v-if="showSkillPreview" class="skill-preview-overlay" @click.self="closeSkillPreview">
            <div class="skill-preview-modal">
              <div class="skill-preview-header">
                <span class="skill-preview-title">{{ currentSkillName }}</span>
                <button class="skill-preview-close" @click="closeSkillPreview">
                  <RiCloseLine size="18" color="#8B7355" />
                </button>
              </div>
              <pre class="skill-preview-content">{{ currentSkillContent }}</pre>
            </div>
          </div>
                <div class="theme-info-area">
                  <div class="theme-info-texts">
                    <span class="theme-info-name" :style="{ color: t.textColor }">{{ t.name }}</span>
                    <span class="theme-info-desc">{{ t.desc }}</span>
                  </div>
                  <div class="theme-check-circle" :class="{ 'theme-checked-circle': settingsStore.theme === t.id }">
                    <RiCheckLine v-if="settingsStore.theme === t.id" size="14" color="#fff" />
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 模板设置 -->
          <div v-if="activeNav === 'template'" class="panel">
            <div class="settings-card">
              <div class="tpl-shelf">
                <div
                  v-for="tpl in settingsStore.templates"
                  :key="tpl.id"
                  class="tpl-card"
                  :class="{ 'tpl-card-selected': settingsStore.selectedTemplateId === tpl.id }"
                  @click="selectTemplate(tpl.id)"
                >
                  <div class="tpl-card-cover">
                    <div class="tpl-card-icon" style="background: #C23B22">
                      <RiFileTextLine size="24" color="#fff" />
                    </div>
                    <span class="tpl-card-cat">{{ tpl.category }}</span>
                    <span class="tpl-card-label">投标模板</span>
                    <button class="tpl-card-delete-btn" @click.stop="settingsStore.removeTemplate(tpl.id)" title="删除">
                      <RiDeleteBinLine size="16" color="#C43A31" />
                    </button>
                  </div>
                  <div class="tpl-card-info">
                    <span class="tpl-card-name">{{ tpl.name }}</span>
                    <span class="tpl-card-desc">{{ tpl.description || '暂无描述' }}</span>
                  </div>
                  <div v-if="settingsStore.selectedTemplateId === tpl.id" class="tpl-card-check">
                    <RiCheckLine size="12" color="#fff" />
                  </div>
                </div>
                <div class="tpl-card tpl-card-add" @click="addTemplate">
                  <div class="tpl-add-icon">
                    <RiAddLine size="22" color="#8B7355" />
                  </div>
                  <span class="tpl-add-text">添加模板</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 技能管理 -->
          <div v-if="activeNav === 'skills'" class="panel">
            <div class="settings-card">
              <div class="tpl-shelf">
                <div
                  v-for="(skill, idx) in displaySkills"
                  :key="skill.id + '-' + idx"
                  class="tpl-card"
                  @click="selectedSkillCard = idx"
                >
                  <div class="tpl-card-cover">
                    <div class="tpl-card-icon" :style="{ background: skill.iconBg }">
                      <component :is="skill.iconComp" :size="'24'" color="#ffffff" />
                    </div>
                    <span class="tpl-card-cat">{{ skill.category }}</span>
                    <label class="toggle-switch-compact" @click.stop="handleToggleSkill(skill)">
                      <input type="checkbox" :checked="handleIsEnabled(skill)" />
                      <span class="toggle-slider"></span>
                    </label>
                    <button
                      v-if="skill.path || skill.id.startsWith('custom_')"
                      class="tpl-card-skill-delete-btn"
                      @click.stop="handleDeleteCustomSkill(skill)"
                      title="从设置中删除"
                    >
                      <RiDeleteBinLine size="16" color="#C43A31" />
                    </button>
                  </div>
                  <div class="tpl-card-info">
                    <span class="tpl-card-name">{{ skill.name }}</span>
                    <span class="tpl-card-desc" :title="skill.desc">{{ skill.desc }}</span>
                  </div>
                </div>
                <div v-if="hasMoreSkills" class="tpl-card tpl-card-more" @click="showMoreSkills = true">
                  <div class="tpl-add-icon">
                    <RiArrowDownSLine size="22" color="#8B7355" />
                  </div>
                  <span class="tpl-add-text">更多技能 ({{ visibleSkillsCount }})</span>
                </div>
                <input type="file" accept=".md" ref="skillFileInput" style="display:none" @change="handleSkillFileSelect" />
                <div class="tpl-card tpl-card-add" @click="openAddSkillPicker">
                  <div class="tpl-add-icon">
                    <RiAddLine size="22" color="#8B7355" />
                  </div>
                  <span class="tpl-add-text">添加技能</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 导出设置 -->
          <div v-if="activeNav === 'export'" class="panel">
            <div class="export-page-title">导出设置</div>
            <div class="export-page-desc">配置标书导出的默认格式</div>
            <div class="export-cards-row">
              <div
                class="export-card-item"
                :class="{ 'export-card-selected': settingsStore.exportFormat === 'word' }"
                @click="settingsStore.setExportFormat('word')"
              >
                <div class="export-card-header">
                  <div class="export-card-icon" style="background: #E8F0F8">
                    <RiFileWord2Line size="28" color="#2D6A9F" />
                  </div>
                  <div class="export-card-titles">
                    <span class="export-card-name">Word 格式</span>
                    <span class="export-card-ext">.docx</span>
                  </div>
                  <div class="export-check" :class="{ 'export-checked': settingsStore.exportFormat === 'word' }">
                    <RiCheckLine v-if="settingsStore.exportFormat === 'word'" size="18" color="#fff" />
                  </div>
                </div>
                <div class="export-features">
                  <div v-for="f in wordFeatures" :key="f" class="export-feature">
                    <RiCheckLine size="16" color="#2D8A4E" />
                    <span>{{ f }}</span>
                  </div>
                </div>
              </div>
              <div
                class="export-card-item"
                :class="{ 'export-card-selected': settingsStore.exportFormat === 'md' }"
                @click="settingsStore.setExportFormat('md')"
              >
                <div class="export-card-header">
                  <div class="export-card-icon" style="background: #F0E8D8">
                    <RiMarkdownLine size="28" color="#8B7355" />
                  </div>
                  <div class="export-card-titles">
                    <span class="export-card-name">Markdown 格式</span>
                    <span class="export-card-ext">.md</span>
                  </div>
                  <div class="export-check" :class="{ 'export-checked': settingsStore.exportFormat === 'md' }">
                    <RiCheckLine v-if="settingsStore.exportFormat === 'md'" size="18" color="#fff" />
                  </div>
                </div>
                <div class="export-features">
                  <div v-for="f in mdFeatures" :key="f" class="export-feature">
                    <RiCheckLine size="16" color="#2D8A4E" />
                    <span>{{ f }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- API Key -->
          <div v-if="activeNav === 'apikey'" class="panel api-full-panel">
            <div class="config-panel">
              <div class="config-tabs">
                <button class="config-tab" :class="{ 'config-tab-selected': configTab === 'custom' }" @click="configTab = 'custom'">自定义配置</button>
                <button class="config-tab" :class="{ 'config-tab-selected': configTab === 'provider' }" @click="configTab = 'provider'">模型制造商</button>
              </div>

              <div class="config-form">
                <template v-if="configTab === 'provider'">
                  <div class="form-field">
                    <label class="form-label">服务商</label>
                    <select v-model="selectedProvider" class="form-select" @change="onProviderChange">
                      <option value="" disabled>请选择服务商</option>
                      <optgroup label="国内">
                        <option value="阿里云">阿里云 (通义千问)</option>
                        <option value="百度">百度 (文心一言)</option>
                        <option value="智谱">智谱 (GLM)</option>
                        <option value="DeepSeek">DeepSeek</option>
                        <option value="Moonshot">Moonshot (月之暗面)</option>
                        <option value="零一万物">零一万物 (Yi)</option>
                      </optgroup>
                      <optgroup label="国外">
                        <option value="OpenAI">OpenAI (GPT)</option>
                        <option value="Anthropic">Anthropic (Claude)</option>
                        <option value="Google">Google (Gemini)</option>
                        <option value="Mistral">Mistral AI</option>
                        <option value="Groq">Groq</option>
                      </optgroup>
                    </select>
                  </div>
                  <div class="form-field">
                    <label class="form-label">模型</label>
                    <select v-model="selectedModelName" class="form-select">
                      <option value="" disabled>请选择模型</option>
                      <option v-for="m in providerModels" :key="m" :value="m">{{ m }}</option>
                    </select>
                  </div>
                </template>
                <template v-else>
                  <div class="form-field">
                    <label class="form-label">API格式</label>
                    <select v-model="customApiFormat" class="form-select">
                      <option value="openai">OpenAI Chat Completions 格式</option>
                      <option value="anthropic">Anthropic Messages 格式</option>
                    </select>
                  </div>
                  <div class="form-field">
                    <label class="form-label">自定义地址</label>
                    <input v-model="customEndpoint" class="form-input" placeholder="https://api.example.com/v1" />
                  </div>
                  <div class="form-field">
                    <label class="form-label">模型ID</label>
                    <input v-model="customModelId" class="form-input" placeholder="例如: gpt-4" />
                  </div>
                </template>
                <div class="form-field">
                  <label class="form-label">API Key</label>
                  <div class="form-input-row">
                    <input
                      v-model="settingsStore.apiKeyForm.key"
                      :type="settingsStore.apiKeyForm.keyVisible ? 'text' : 'password'"
                      class="form-input"
                      placeholder="sk-..."
                    />
                    <button class="form-key-toggle" @click="settingsStore.toggleKeyVisibility()">
                      <RiEyeOffLine v-if="!settingsStore.apiKeyForm.keyVisible" size="16" color="#8B7355" />
                      <RiEyeLine v-else size="16" color="#8B7355" />
                    </button>
                  </div>
                </div>
              </div>

              <div v-if="settingsStore.apiKeys.length > 0" class="saved-keys">
                <span class="saved-keys-title">已保存的密钥</span>
                <div v-for="key in settingsStore.apiKeys" :key="key.id" class="saved-key-item">
                  <div class="saved-key-info">
                    <span class="saved-key-provider">{{ key.provider }}</span>
                    <span class="saved-key-model">{{ key.modelName || key.model }}</span>
                  </div>
                  <button class="saved-key-test" @click="handleTestKey(key)" :disabled="testingKey === key.id">
                    <RiLoaderLine v-if="testingKey === key.id" size="14" color="#8B7355" class="spin" />
                    <RiCheckLine v-else-if="testSuccess[key.id]" size="14" color="#2D8A4E" />
                    <RiCloseLine v-else-if="testFailed[key.id]" size="14" color="#C43A31" />
                    <span>{{ testSuccess[key.id] ? '可用' : testFailed[key.id] ? '失败' : '测试' }}</span>
                  </button>
                  <button class="saved-key-delete" @click="settingsStore.removeApiKey(key.id)">
                    <RiDeleteBinLine size="14" color="#C43A31" />
                  </button>
                  <label class="toggle-switch" @click.stop="settingsStore.toggleApiKey(key.id)">
                    <input type="checkbox" :checked="key.enabled" />
                    <span class="toggle-slider"></span>
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- Template Modal -->
    <div v-if="showTemplateModal" class="modal-overlay" @click.self="showTemplateModal = false">
      <div class="modal-content">
        <h3>添加模板</h3>
        <div class="form-field">
          <label>模板名称</label>
          <input v-model="newTemplateName" class="form-input" placeholder="输入模板名称" />
        </div>
        <div class="form-field">
          <label>上传文件</label>
          <input type="file" accept=".docx,.doc" @change="handleTemplateFileSelect" class="form-input" />
        </div>
        <div class="modal-actions">
          <button class="modal-btn-cancel" @click="showTemplateModal = false">取消</button>
          <button class="modal-btn-save" @click="saveNewTemplate">保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from '../stores/settingsStore'
import { testApiKey } from '../api/client'
import { RiRadarFill, RiArrowLeftLine, RiPaletteLine, RiBookmarkLine, RiFileListLine, RiFileDownloadLine, RiKeyLine, RiCheckLine, RiAddLine, RiDeleteBinLine, RiFileTextLine, RiBuildingLine, RiServerLine, RiCustomerServiceLine, RiFileWord2Line, RiMarkdownLine, RiEyeOffLine, RiEyeLine, RiLoaderLine, RiCloseLine, RiArrowDownSLine } from '@remixicon/vue'

const router = useRouter()

const goHome = () => router.push('/')
function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/')
  }
}

interface NavItem {
  id: string
  label: string
  sublabel: string
  title: string
  desc: string
  icon: any
  color: string
}

const navItems: NavItem[] = [
  { id: 'theme', label: '主题设置', sublabel: 'Theme', title: '主题设置', desc: '选择你喜欢的界面风格', icon: RiPaletteLine, color: '#C23B22' },
  { id: 'template', label: '模板设置', sublabel: 'Template', title: '模板设置', desc: '管理标书模板', icon: RiBookmarkLine, color: '#C8A45C' },
  { id: 'skills', label: '技能管理', sublabel: 'Skills', title: '技能管理', desc: '管理 AI 写作技能', icon: RiFileListLine, color: '#5B8C5A' },
  { id: 'export', label: '导出设置', sublabel: 'Export', title: '导出设置', desc: '配置标书导出的默认格式', icon: RiFileDownloadLine, color: '#2D6A9F' },
  { id: 'apikey', label: 'API Key', sublabel: '', title: 'API Key', desc: '管理 AI 模型密钥', icon: RiKeyLine, color: '#6366F1' },
]

const activeNav = ref('theme')
const settingsStore = useSettingsStore()

const themes = [
  {
    id: 'light' as const, name: '羊皮纸', desc: '温润雅致 · 默认主题',
    previewBg: '#FBF7F0', navBg: '#F5EFE3', logoBg: '#C23B22',
    brandColor: '#3D2B1F', sidebarBg: '#F5EFE3', card1Bg: '#F0E8D8', card2Bg: '#F5EFE3',
    textColor: '#C23B22',
  },
  {
    id: 'dark' as const, name: '深色', desc: '深邃护眼 · 夜间模式',
    previewBg: '#2C2416', navBg: '#3D3224', logoBg: '#C23B22',
    brandColor: '#E8DCC8', sidebarBg: '#3D3224', card1Bg: '#4A3D2C', card2Bg: '#3D3224',
    textColor: '#E8DCC8',
  },
  {
    id: 'paper' as const, name: '白纸', desc: '清爽干净 · 极简模式',
    previewBg: '#FFFFFF', navBg: '#F5F5F5', logoBg: '#C23B22',
    brandColor: '#3D2B1F', sidebarBg: '#F5F5F5', card1Bg: '#F0F0F0', card2Bg: '#F5F5F5',
    textColor: '#3D2B1F',
  },
]

const selectedSkillCard = ref(-1)
const showMoreSkills = ref(false)
const skillFileInput = ref<HTMLInputElement | null>(null)
const showTemplateModal = ref(false)
const newTemplateName = ref('')
const newTemplateFile = ref<File | null>(null)

const showSkillPreview = ref(false)
const currentSkillContent = ref('')
const currentSkillName = ref('')

const hiddenSkills = ref(new Set<string>())

const toggleSkillDisplay = (skillId: string) => {
  if (hiddenSkills.value.has(skillId)) {
    hiddenSkills.value.delete(skillId)
  } else {
    hiddenSkills.value.add(skillId)
  }
  try { localStorage.setItem('hidden_skills', JSON.stringify([...hiddenSkills.value])) } catch {}
}

const isSkillVisible = (skill: any) => !hiddenSkills.value.has(skill.id)

// Handle toggle for skill management — respects both enabled and hidden
const handleToggleSkill = (skill: { id: string; path?: string }) => {
  if (skill.id.startsWith('custom_')) {
    settingsStore.toggleSkillEnabled(skill.id)
  } else if (skill.path) {
    const localSkill = settingsStore.localSkills.find(s => s.id === skill.id)
    if (localSkill) {
      localSkill.enabled = !localSkill.enabled
    }
  }
}

const handleIsEnabled = (skill: { id: string; enabled?: boolean }) => {
  return skill.enabled !== false
}

const handleDeleteCustomSkill = (skill: { id: string }) => {
  if (skill.id.startsWith('custom_')) {
    settingsStore.removeCustomSkill(skill.id)
  } else if (skill.path) {
    // Remove from localSkills and hidden set, but don't delete local file
    settingsStore.localSkills = settingsStore.localSkills.filter(s => s.id !== skill.id)
    hiddenSkills.value.delete(skill.id)
    try { localStorage.setItem('hidden_skills', JSON.stringify([...hiddenSkills.value])) } catch {}
  }
}

const allManageableSkills = computed(() => {
  const local = settingsStore.localSkills
  const custom = settingsStore.customSkills
  return [...local, ...custom]
})

onMounted(() => {
  settingsStore.fetchModels()
  settingsStore.fetchTemplates()
  settingsStore.fetchLocalSkills()
  settingsStore.loadCustomSkillsFromLocalStorage()
  loadSavedConfig()
  try {
    const saved = localStorage.getItem('hidden_skills')
    if (saved) hiddenSkills.value = new Set(JSON.parse(saved))
  } catch {}
})

const displaySkills = computed(() => {
  const icons = [
    { icon: RiFileTextLine, bg: '#C23B22' },
    { icon: RiBuildingLine, bg: '#2D6A9F' },
    { icon: RiServerLine, bg: '#2D8A4E' },
    { icon: RiCustomerServiceLine, bg: '#D4A017' },
  ]
  return allManageableSkills.value.map((skill, idx) => ({
    id: skill.id,
    name: skill.name,
    desc: skill.description || '',
    category: skill.category || '自定义',
    iconComp: icons[idx % icons.length].icon,
    iconBg: icons[idx % icons.length].bg,
    path: (skill as any).path || '',
    enabled: (skill as any).enabled !== false,
  }))
})

const hasMoreSkills = computed(() => displaySkills.value.length > 10 && !showMoreSkills.value)

const visibleSkillsCount = computed(() => {
  return allManageableSkills.value.filter(s => !hiddenSkills.value.has(s.id)).length
})

const addTemplate = () => {
  showTemplateModal.value = true
  newTemplateName.value = ''
  newTemplateFile.value = null
}

const openAddSkillPicker = () => {
  skillFileInput.value?.click()
}

const openSkill = async (skill: { id: string; name: string; path: string }) => {
  if (!skill.path) return
  
  currentSkillName.value = skill.name
  showSkillPreview.value = true
  
  try {
    const res = await fetch('/api/skills/content?path=' + encodeURIComponent(skill.path))
    if (res.ok) {
      currentSkillContent.value = await res.text()
    } else {
      currentSkillContent.value = '# ' + skill.name + '\n\n(无法加载技能文件内容)'
    }
  } catch {
    currentSkillContent.value = '# ' + skill.name + '\n\n(加载失败，请检查技能文件路径)'
  }
}

const closeSkillPreview = () => {
  showSkillPreview.value = false
  currentSkillContent.value = ''
  currentSkillName.value = ''
}

const handleSkillFileSelect = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return
  try {
    const text = await file.text()
    const parsed = parseFrontmatter(text, file.name)
    settingsStore.addCustomSkill(parsed)
    alert('技能添加成功')
  } catch {
    alert('读取文件失败')
  }
  ;(event.target as HTMLInputElement).value = ''
}

const handleTemplateFileSelect = (e: Event) => {
  newTemplateFile.value = (e.target as HTMLInputElement).files?.[0] || null
}

const saveNewTemplate = () => {
  if (!newTemplateName.value.trim()) {
    alert('请输入模板名称')
    return
  }
  if (!newTemplateFile.value) {
    alert('请选择 .docx 文件')
    return
  }
  // 保存到本地 localStorage，不调用后端 API
  settingsStore.addTemplate(newTemplateName.value.trim(), newTemplateFile.value)
  showTemplateModal.value = false
  newTemplateName.value = ''
  newTemplateFile.value = null
}

function parseFrontmatter(content: string, filename: string) {
  const result: any = { name: '', description: '', prompt: '', category: '自定义', enabled: true }
  if (!content.startsWith('---')) {
    result.name = filename.replace('.md', '')
    const firstLine = content.split('\n')[0]
    result.description = firstLine.substring(0, 100)
    result.prompt = content
    return result
  }
  const match = content.match(/^---\n([\s\S]*?)\n---/)
  if (!match) {
    result.name = filename.replace('.md', '')
    return result
  }
  const fm: Record<string, string> = {}
  match[1].split('\n').forEach(line => {
    const colonIdx = line.indexOf(':')
    if (colonIdx > -1) {
      const key = line.substring(0, colonIdx).trim()
      let val = line.substring(colonIdx + 1).trim()
      val = val.replace(/^["']|["']$/g, '')
      if (key && val) fm[key] = val
    }
  })
  result.name = fm.name || filename.replace('.md', '')
  result.description = fm.description || ''
  result.prompt = fm.prompt || content
  result.category = fm.category || '自定义'
  return result
}

const wordFeatures = [
  '保留完整格式与排版样式',
  '支持表格、图片、页眉页脚',
  '兼容 Microsoft Word / WPS',
  '支持目录自动生成',
]

const mdFeatures = [
  '纯文本格式，轻量易读',
  '适合版本管理与协作',
  '可快速转换为 HTML/PDF',
  '兼容各类 Markdown 编辑器',
]

const currentNav = computed(() => navItems.find(i => i.id === activeNav.value)!)

const configTab = ref<'provider' | 'custom'>(
  (() => { try { const v = localStorage.getItem('cfg_tab'); return (v === 'provider' || v === 'custom') ? v : 'custom' } catch { return 'custom' } })()
)
const customApiFormat = ref(
  (() => { try { return localStorage.getItem('cfg_format') || 'openai' } catch { return 'openai' } })()
)
const customEndpoint = ref(localStorage.getItem('cfg_endpoint') || '')
const customModelId = ref(localStorage.getItem('cfg_model_id') || '')

const providerModelsMap: Record<string, string[]> = {
  '阿里云': ['qwen3.7-max', 'qwen3.7-plus', 'qwen3.6-plus', 'qwen3.5-plus', 'qwen3-max', 'qwen-plus', 'qwen-flash', 'qwen3-coder-plus'],
  '百度': ['ernie-4.0', 'ernie-3.5', 'ernie-speed'],
  '智谱': ['glm-5.1', 'glm-4.6', 'glm-4'],
  'DeepSeek': ['deepseek-v4-pro', 'deepseek-r1', 'deepseek-chat'],
  'Moonshot': ['kimi-k2.6', 'moonshot-v1-128k', 'moonshot-v1-32k'],
  '零一万物': ['yi-large', 'yi-medium', 'yi-34b-chat'],
  'OpenAI': ['gpt-5.6-sol', 'gpt-5.5', 'gpt-5.4', 'gpt-5.4-pro', 'gpt-5.3-codex', 'o4-mini', 'o3', 'gpt-4.1'],
  'Anthropic': ['claude-fable-5', 'claude-opus-4.8', 'claude-opus-4.7', 'claude-sonnet-4.6', 'claude-haiku-4.5'],
  'Google': ['gemini-2.5-pro', 'gemini-2.5-flash', 'gemini-2.0-pro'],
  'Mistral': ['mistral-large-3', 'mistral-medium-3', 'mistral-small-3'],
  'Groq': ['llama-3.3-70b', 'llama-3.1-8b', 'mixtral-8x7b-32768'],
}

const selectedProvider = ref('')
const selectedModelName = ref('')
const providerModels = computed(() => providerModelsMap[selectedProvider.value] || [])

const onProviderChange = () => {
  selectedModelName.value = ''
}

const selectTemplate = (id: string) => {
  settingsStore.setSelectedTemplate(
    settingsStore.selectedTemplateId === id ? '' : id
  )
}

const testingKey = ref('')
const testSuccess = ref<Record<string, boolean>>({})
const testFailed = ref<Record<string, boolean>>({})

const handleTestKey = async (key: any) => {
  testingKey.value = key.id
  testSuccess.value[key.id] = false
  testFailed.value[key.id] = false
  try {
    const res = await testApiKey({
      provider: key.provider,
      model: key.model,
      key: key.key,
      endpoint: key.endpoint || '',
      format: key.format || 'openai',
    })
    testingKey.value = ''
    if (res.data.available) {
      testSuccess.value[key.id] = true
    } else {
      testFailed.value[key.id] = true
    }
  } catch {
    testingKey.value = ''
    testFailed.value[key.id] = true
  }
}

const saveSettings = () => {
  if (activeNav.value === 'apikey') {
    persistConfig()
    addApiKeyEntry()
    return
  }
  persistConfig()
  alert('设置已保存')
  router.push('/')
}

const resetForm = () => {
  settingsStore.apiKeyForm = { provider: '', model: '', key: '', keyVisible: false }
  customApiFormat.value = ''
  customEndpoint.value = ''
  customModelId.value = ''
  selectedProvider.value = ''
  selectedModelName.value = ''
}

const addApiKeyEntry = () => {
  const key = settingsStore.apiKeyForm.key.trim()
  if (!key) {
    alert('请输入 API Key')
    return
  }
  if (configTab.value === 'custom') {
    if (!customEndpoint.value.trim() || !customModelId.value.trim()) {
      alert('请填写自定义地址和模型 ID')
      return
    }
    settingsStore.addApiKey({
      id: Date.now().toString(),
      provider: customApiFormat.value === 'anthropic' ? 'Anthropic (自定义)' : 'OpenAI (自定义)',
      model: customModelId.value.trim(),
      modelName: customModelId.value.trim(),
      key: key,
      endpoint: customEndpoint.value.trim(),
      format: customApiFormat.value,
      enabled: true,
    })
  } else {
    if (!selectedProvider.value || !selectedModelName.value) {
      alert('请选择服务商和模型')
      return
    }
    settingsStore.addApiKey({
      id: Date.now().toString(),
      provider: selectedProvider.value,
      model: selectedModelName.value,
      modelName: selectedModelName.value,
      key: key,
      enabled: true,
    })
  }
  resetForm()
}

function loadSavedConfig() {
  try {
    const saved = localStorage.getItem('api_config')
    if (saved) {
      const cfg = JSON.parse(saved)
      configTab.value = cfg.configTab || 'custom'
      customApiFormat.value = cfg.customApiFormat || 'openai'
      customEndpoint.value = cfg.customEndpoint || ''
      customModelId.value = cfg.customModelId || ''
      selectedProvider.value = cfg.selectedProvider || ''
      selectedModelName.value = cfg.selectedModelName || ''
    }
  } catch {}
}

function persistConfig() {
  localStorage.setItem('api_config', JSON.stringify({
    configTab: configTab.value,
    customApiFormat: customApiFormat.value,
    customEndpoint: customEndpoint.value,
    customModelId: customModelId.value,
    selectedProvider: selectedProvider.value,
    selectedModelName: selectedModelName.value,
  }))
}

const indicatorStyle = computed(() => {
  const idx = navItems.findIndex(i => i.id === activeNav.value)
  const itemHeight = 56
  const gap = 2
  return {
    top: `${idx * (itemHeight + gap)}px`,
    height: `${itemHeight}px`,
    background: currentNav.value.color,
  }
})
</script>

<style scoped>
.page {
  min-height: 100vh;
  background: #FDF6E3;
  display: flex;
  flex-direction: column;
  font-family: 'Noto Sans SC', -apple-system, BlinkMacSystemFont, sans-serif;
}

/* ── Navbar ── */
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
  padding: 0 32px;
  background: #FDF6E3;
  flex-shrink: 0;
}

.navbar-left {
  display: flex;
  align-items: center;
  gap: 0;
  cursor: pointer;
}

.logo {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  background: #C23B22;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.brand-cn {
  font-size: 22px;
  font-weight: 700;
  font-family: 'Noto Serif SC', serif;
  color: #3D2B1F;
  margin-left: 12px;
}

.brand-en {
  font-size: 12px;
  color: #8B7355;
  margin-left: 6px;
  line-height: 1;
  align-self: flex-end;
  padding-bottom: 3px;
}

.navbar-right {
  display: flex;
  align-items: center;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 16px;
  background: #F0E8D5;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: #8B7355;
  transition: background 0.2s;
}

.back-btn:hover {
  background: #E0D5C0;
}

/* ── Body ── */
.body {
  display: flex;
  flex: 1;
  min-height: 0;
}

/* ── Sidebar ── */
.sidebar {
  width: 280px;
  flex-shrink: 0;
  background: #FBF7EF;
  border-right: 1px solid #E0D5C0;
  display: flex;
  flex-direction: column;
  position: relative;
}

.sidebar-header {
  padding: 20px 24px 20px;
}

.sidebar-title {
  font-size: 16px;
  font-weight: 600;
  color: #3D2B1F;
}

.sidebar-divider {
  width: 100%;
  height: 1px;
  background: #E0D5C0;
  margin-top: 2px;
}

.sidebar-subtitle {
  font-size: 12px;
  color: #8B7355;
  margin-top: 8px;
  display: block;
}

.nav-list {
  flex: 1;
  padding: 0 16px 16px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  position: relative;
}

.nav-indicator {
  position: absolute;
  left: 16px;
  right: 16px;
  z-index: 0;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  transition: all 0.3s ease-out;
  pointer-events: none;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: none;
  border-radius: 12px;
  background: transparent;
  cursor: pointer;
  text-align: left;
  width: 100%;
  position: relative;
  z-index: 1;
  transition: all 0.2s;
}

.nav-texts {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.nav-label {
  font-size: 15px;
  font-weight: 600;
  color: #3D2B1F;
  transition: color 0.2s;
}

.nav-item-active {
  color: #fff;
}

.nav-item-active .nav-label {
  color: #fff;
}

.nav-sublabel {
  font-size: 11px;
  color: #8B7355;
  transition: color 0.2s;
}

.nav-item-active .nav-sublabel {
  color: rgba(255,255,255,0.7);
}

/* ── Content ── */
.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: #EAE5D9;
}

.content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 32px;
  background: #FDF6E3;
  flex-shrink: 0;
}

.content-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.content-header-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.content-header-texts {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.content-title {
  font-size: 18px;
  font-weight: 700;
  color: #3D2B1F;
  margin: 0;
}

.content-subtitle {
  font-size: 12px;
  color: #8B7355;
}

.content-header-actions {
  display: flex;
  gap: 12px;
}

.action-btn-cancel {
  padding: 12px 24px;
  border: 0.7px solid #E0D5C0;
  border-radius: 12px;
  background: #F5EFE0;
  color: #8B7355;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.action-btn-save {
  padding: 12px 24px;
  border: none;
  border-radius: 12px;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.content-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 0 32px 32px;
  min-height: 0;
}

.panel {
  background: transparent;
}

/* ── Theme Settings ── */
.theme-page-title {
  font-size: 24px;
  font-weight: 700;
  color: #3D2B1F;
  margin-bottom: 4px;
}

.theme-page-desc {
  font-size: 14px;
  color: #9B8C7C;
  margin-bottom: 32px;
}

.theme-cards-row {
  display: flex;
  gap: 24px;
}

.theme-card-new {
  flex: 1;
  background: #fff;
  border-radius: 16px;
  overflow: hidden;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.theme-card-new:hover {
  border-color: #E0D5C0;
}

.theme-card-selected {
  border-color: #C23B22;
  box-shadow: 0 4px 20px rgba(196, 61, 61, 0.15);
}

.theme-preview-area {
  height: 200px;
  display: flex;
  flex-direction: column;
}

.theme-preview-nav {
  height: 40px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
  flex-shrink: 0;
}

.preview-logo {
  width: 21px;
  height: 20px;
  border-radius: 9999px;
  flex-shrink: 0;
}

.preview-brand {
  font-size: 10px;
  font-weight: 600;
}

.theme-preview-body {
  flex: 1;
  display: flex;
}

.preview-sidebar {
  width: 51px;
  flex-shrink: 0;
}

.preview-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
}

.preview-card {
  height: 60px;
  border-radius: 8px;
}

.preview-card-sm {
  height: 40px;
  border-radius: 8px;
}

.theme-info-area {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: #fff;
}

.theme-info-texts {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.theme-info-name {
  font-size: 15px;
  font-weight: 600;
}

.theme-info-desc {
  font-size: 12px;
  color: #9B8C7C;
}

.theme-check-circle {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #D4C4A8;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  flex-shrink: 0;
}

.theme-checked-circle {
  background: #C23B22;
}

.settings-card {
  background: #F5EFE0;
  border: 0.7px solid #E0D5C0;
  border-radius: 16px;
  padding: 32px;
}

/* ── Template Shelf ── */
.tpl-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  gap: 16px;
}

.tpl-empty-text {
  font-size: 14px;
  color: #8B7355;
}

.tpl-card-selected {
  box-shadow: 0 0 0 2px #C23B22;
}

.tpl-shelf {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
}

.tpl-card {
  width: 200px;
  border-radius: 12px;
  background: #fff;
  position: relative;
  cursor: pointer;
  overflow: hidden;
  transition: box-shadow 0.2s;
}

.tpl-card:hover {
  box-shadow: 0 4px 16px rgba(0,0,0,0.1);
}

.tpl-card-cover {
  height: 240px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  position: relative;
}

.tpl-card-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
}

.tpl-card-cat {
  font-size: 14px;
  font-weight: 600;
  color: #3D2B1F;
}

.tpl-card-label {
  font-size: 12px;
  color: #8B7355;
}

.tpl-card-info {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.tpl-card-name {
  font-size: 13px;
  font-weight: 500;
  color: #3D2B1F;
}

.tpl-card-desc {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: normal;
  font-size: 11px;
  color: #8B7355;
  cursor: help;
}

.tpl-card-check {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: #C23B22;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ── Add Template Card ── */
.tpl-card-add {
  background: #F5EFE3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 340px;
  gap: 8px;
  border: 1.5px dashed #D4C4A8;
  transition: border-color 0.2s, background 0.2s;
}

.tpl-card-add:hover {
  border-color: #C23B22;
  background: #F0E8D5;
}

/* ── Toggle Button on Skill Cards ── */
.tpl-card-toggle {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  border: none;
  background: rgba(255,255,255,0.9);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s;
  z-index: 2;
}

.tpl-card-toggle:hover {
  background: #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.tpl-card-toggle.disabled {
  opacity: 0.6;
}

.tpl-card-toggle.disabled:hover {
  opacity: 0.8;
}

/* ── More Button Card ── */
.tpl-card-more {
  background: #F5EFE3;
  border: 1.5px dashed #D4C4A8;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 340px;
  gap: 8px;
  transition: border-color 0.2s, background 0.2s;
  cursor: pointer;
}

.tpl-card-more:hover {
  border-color: #C23B22;
  background: #F0E8D5;
}

/* ── Template Modal ── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: #fff;
  border-radius: 16px;
  padding: 32px;
  width: 400px;
  max-height: 80vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.2);
  animation: slideUp 0.2s ease-out;
}

@keyframes slideUp {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
}

.modal-btn-cancel {
  padding: 10px 24px;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  background: #F5EFE0;
  color: #8B7355;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.modal-btn-cancel:hover {
  background: #E0D5C0;
}

.modal-btn-save {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: #C23B22;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.modal-btn-save:hover {
  opacity: 0.9;
}

.modal-btn-test {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: #2D8A4E;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.modal-btn-test:hover {
  opacity: 0.9;
}

.tpl-add-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #F0E8D5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.tpl-add-text {
  font-size: 14px;
  font-weight: 500;
  color: #8B7355;
}

/* ── Export Settings ── */
.export-page-title {
  font-size: 24px;
  font-weight: 700;
  color: #3D2B1F;
  margin-bottom: 4px;
}

.export-page-desc {
  font-size: 14px;
  color: #9B8C7C;
  margin-bottom: 32px;
}

.export-cards-row {
  display: flex;
  gap: 24px;
}

.export-card-item {
  flex: 1;
  background: #fff;
  border-radius: 16px;
  padding: 32px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.export-card-item:hover {
  border-color: #E0D5C0;
}

.export-card-selected {
  border-color: #C23B22;
  box-shadow: 0 4px 20px rgba(196, 61, 61, 0.1);
}

.export-card-header {
  display: flex;
  align-items: center;
  gap: 16px;
}

.export-card-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.export-card-titles {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
}

.export-card-name {
  font-size: 20px;
  font-weight: 700;
  color: #3D2B1F;
}

.export-card-ext {
  font-size: 13px;
  color: #8B7355;
}

.export-check {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #D4C4A8;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
  flex-shrink: 0;
}

.export-checked {
  background: #C23B22;
}

.export-features {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.export-feature {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #3D2B1F;
}

.export-feature svg {
  flex-shrink: 0;
}

/* ── API Key (Model List + Config Form) ── */
.api-full-panel {
  background: #F5EFE0;
  border: 0.7px solid #E0D5C0;
  border-radius: 16px;
  padding: 32px;
}

/* ── Config Panel ── */
.config-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 24px;
  min-width: 0;
}

.config-tabs {
  display: flex;
  gap: 4px;
  background: #F0E8D5;
  border-radius: 12px;
  padding: 4px;
  width: fit-content;
}

.config-tab-selected {
  padding: 8px 20px;
  border: none;
  border-radius: 8px;
  background: #C23B22;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.config-tab {
  padding: 8px 20px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #8B7355;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.config-form {
  background: #fff;
  border-radius: 16px;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-field {
  display: flex;
  align-items: center;
  gap: 16px;
}

.form-label {
  width: 80px;
  font-size: 12px;
  font-weight: 500;
  color: #5C4033;
  flex-shrink: 0;
}

.form-input {
  flex: 1;
  background: #FAFAF5;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 13px;
  color: #3D2B1F;
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus {
  border-color: #C23B22;
  box-shadow: 0 0 0 1px rgba(194, 59, 34, 0.15);
}

.form-select {
  flex: 1;
  background: #FAFAF5;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 13px;
  color: #3D2B1F;
  outline: none;
  transition: border-color 0.2s;
  cursor: pointer;
  appearance: none;
  -webkit-appearance: none;
  -moz-appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='%238B7355'%3E%3Cpath d='M12 16L6 10H18L12 16Z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
}

.form-select:hover {
  border-color: #D4C4A8;
}

.form-select:focus {
  border-color: #C23B22;
  box-shadow: 0 0 0 1px rgba(194, 59, 34, 0.15);
}

.form-select optgroup {
  font-size: 12px;
  font-weight: 600;
  color: #8B7355;
  background: #FBF7EF;
}

.form-select option {
  font-size: 13px;
  color: #3D2B1F;
  background: #fff;
  padding: 8px;
}

.form-input-row {
  display: flex;
  align-items: center;
  gap: 0;
  flex: 1;
  position: relative;
}

.form-input-row .form-input {
  padding-right: 36px;
}

.form-key-toggle {
  border: none;
  background: transparent;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  border-radius: 4px;
}

.form-key-toggle:hover {
  background: rgba(194, 59, 34, 0.05);
}

/* ── Saved Keys ── */
.saved-keys {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.saved-keys-title {
  font-size: 12px;
  font-weight: 600;
  color: #8B7355;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 12px 0 4px;
}

.saved-key-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  background: #FAFAF5;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
}

.saved-key-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.saved-key-provider {
  font-size: 13px;
  font-weight: 600;
  color: #3D2B1F;
}

.saved-key-model {
  font-size: 11px;
  color: #8B7355;
}

.saved-key-delete {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: background 0.2s;
  flex-shrink: 0;
}

.saved-key-delete:hover {
  background: rgba(196, 58, 49, 0.08);
}

/* ── Toggle Switch (inside skill card, top-left) ── */
.toggle-switch-compact {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 20px;
  border-radius: 999px;
  background: #D4C4A8;
  cursor: pointer;
  position: absolute;
  top: 12px;
  left: 12px;
  flex-shrink: 0;
  transition: background 0.2s;
}

.toggle-switch-compact input {
  display: none;
}

.toggle-switch-compact .toggle-slider {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.2s;
}

.toggle-switch-compact input:checked + .toggle-slider {
  transform: translateX(16px);
}

.toggle-switch-compact:has(input:checked) {
  background: #C23B22;
}
.toggle-switch {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 20px;
  border-radius: 999px;
  background: #D4C4A8;
  cursor: pointer;
  position: relative;
  flex-shrink: 0;
  transition: background 0.2s;
}

.toggle-switch input {
  display: none;
}

.toggle-switch .toggle-slider {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #fff;
  transition: transform 0.2s;
}

.toggle-switch input:checked + .toggle-slider {
  transform: translateX(16px);
}

.toggle-switch:has(input:checked) {
  background: #C23B22;
}

/* ── Test Button ── */
.saved-key-test {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border: 0.7px solid #E0D5C0;
  border-radius: 6px;
  background: transparent;
  font-size: 12px;
  color: #8B7355;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
  flex-shrink: 0;
}

.saved-key-test:hover:not(:disabled) {
  border-color: #C23B22;
  color: #C23B22;
  background: rgba(194, 59, 34, 0.04);
}

.saved-key-test:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-key-toggle:hover {
  background: rgba(194, 59, 34, 0.05);
  border-radius: 4px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.form-btn-cancel {
  padding: 10px 24px;
  border: 0.7px solid #E0D5C0;
  border-radius: 8px;
  background: #F5EFE0;
  color: #8B7355;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.form-btn-add {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  background: #C23B22;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.form-btn-add:hover {
  opacity: 0.9;
}

/* ── Template Card Delete Button (top-right) ── */
.tpl-card-delete-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  opacity: 0;
  z-index: 3;
}

.tpl-card:hover .tpl-card-delete-btn {
  opacity: 1;
}

.tpl-card-delete-btn:hover svg {
  color: #C23B22 !important;
}

/* ── Skill Preview Icon ── */
.tpl-card-preview {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: #F0E8D5;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.tpl-card-preview:hover {
  background: #C23B22;
  color: #fff;
}

.tpl-card-preview svg {
  color: inherit !important;
}

/* ── Skill Delete Button (top-right) ── */
.tpl-card-skill-delete-btn {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 28px;
  height: 28px;
  border: none;
  background: rgba(240, 232, 213, 0.9);
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  z-index: 2;
  opacity: 0;
}

.tpl-card:hover .tpl-card-skill-delete-btn {
  opacity: 1;
}

.tpl-card-skill-delete-btn:hover {
  background: #C23B22;
}

.tpl-card-skill-delete-btn:hover svg {
  color: #fff !important;
}

/* ── Skill Preview Modal ── */
.skill-preview-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.skill-preview-modal {
  background: #fff;
  border-radius: 16px;
  width: 90%;
  max-width: 700px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
}

.skill-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #E0D5C0;
}

.skill-preview-title {
  font-size: 16px;
  font-weight: 600;
  color: #3D2B1F;
}

.skill-preview-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.skill-preview-close:hover {
  background: #F0E8D5;
}

.skill-preview-content {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  margin: 0;
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #3D2B1F;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
