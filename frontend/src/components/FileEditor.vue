<template>
  <el-dialog
    v-model="dialogVisible"
    width="70%"
    :close-on-click-modal="false"
    class="file-editor-dialog"
    :fullscreen="isFullscreen"
    :show-close="false"
    @close="handleClose"
    @opened="onOpened"
  >
    <template #header>
      <div class="dialog-header">
        <span class="dialog-title" :title="currentFile?.path || ''">{{ currentFile?.path || '文件编辑器' }}</span>
        <div class="header-actions">
          <el-button
            v-if="!isMobile"
            class="header-btn"
            text
            @click="toggleFullscreen"
          >
            <el-icon><FullScreen /></el-icon>
          </el-button>
          <el-button class="header-btn close-btn" text @click="handleClose">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
    </template>

    <!-- 工具栏 -->
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <el-text class="toolbar-text" @click="handleReset" :class="{ 'is-disabled': !hasChanges }">重置</el-text>
        <el-divider direction="vertical" class="toolbar-divider" />
        <el-text class="toolbar-text" @click="saveCurrentFile">
          <span v-if="saving">保存中...</span>
          <span v-else>保存</span>
        </el-text>
        <el-divider direction="vertical" class="toolbar-divider" />
        <el-dropdown trigger="click" placement="bottom-start" @command="changeTheme">
          <span class="el-dropdown-link">主题</span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item 
                v-for="item in themes" 
                :key="item.value" 
                :command="item.value"
              >
                <div class="dropdown-item">
                  <span>{{ item.label }}</span>
                  <el-icon v-if="config.theme === item.value"><Check /></el-icon>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-divider direction="vertical" class="toolbar-divider" />
        <el-dropdown trigger="click" placement="bottom-start" @command="changeLanguage">
          <span class="el-dropdown-link">语言</span>
          <template #dropdown>
            <el-dropdown-menu class="language-dropdown">
              <el-dropdown-item 
                v-for="item in languages" 
                :key="item.label" 
                :command="item.label"
              >
                <div class="dropdown-item">
                  <span>{{ item.label }}</span>
                  <el-icon v-if="config.language === item.label"><Check /></el-icon>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-divider direction="vertical" class="toolbar-divider" />
        <el-dropdown trigger="click" placement="bottom-start" @command="changeEOL">
          <span class="el-dropdown-link">换行符</span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item 
                v-for="item in eols" 
                :key="item.value" 
                :command="item.value"
              >
                <div class="dropdown-item">
                  <span>{{ item.label }}</span>
                  <el-icon v-if="config.eol === item.value"><Check /></el-icon>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-divider direction="vertical" class="toolbar-divider" />
        <el-dropdown trigger="click" placement="bottom-start">
          <span class="el-dropdown-link">设置</span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="toggleMinimap">
                <div class="dropdown-item">
                  <span>小地图</span>
                  <el-icon v-if="config.minimap"><Check /></el-icon>
                </div>
              </el-dropdown-item>
              <el-dropdown-item @click="toggleWordWrap">
                <div class="dropdown-item">
                  <span>自动换行</span>
                  <el-icon v-if="config.wordWrap === 'on'"><Check /></el-icon>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <!-- 主内容区 -->
    <div v-show="loading" class="loading-container">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <span>加载中...</span>
    </div>
    <div v-show="!loading" class="editor-main">
      <!-- 左侧文件树 -->
      <div v-show="showFileTree" class="file-tree-panel">
        <div class="tree-header">
          <el-text size="small" class="tree-header-btn" @click="goToParentDir" :class="{ 'is-disabled': !canGoUp }">
            <el-icon><Top /></el-icon>
            <span class="tree-header-text">上级</span>
          </el-text>
          <el-divider direction="vertical" class="tree-divider-v" />
          <el-text size="small" class="tree-header-btn" @click="refreshTree">
            <el-icon><Refresh /></el-icon>
            <span class="tree-header-text">刷新</span>
          </el-text>
          <el-divider direction="vertical" class="tree-divider-v" />
          <el-dropdown trigger="click" @command="handleCreateInTree">
            <el-text size="small" class="tree-header-btn">
              新建
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-text>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="dir">
                  <el-icon><FolderAdd /></el-icon>
                  新建目录
                </el-dropdown-item>
                <el-dropdown-item command="file">
                  <el-icon><DocumentAdd /></el-icon>
                  新建文件
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
        <el-divider class="tree-divider" />
        <div class="tree-content">
          <el-tree
            ref="treeRef"
            :data="treeData"
            :props="treeProps"
            :highlight-current="true"
            :expand-on-click-node="false"
            node-key="path"
            @node-click="handleTreeNodeClick"
          >
            <template #default="{ node, data }">
              <span class="tree-node">
                <el-icon v-if="data.isDir" class="tree-icon folder-icon"><Folder /></el-icon>
                <el-icon v-else class="tree-icon file-icon"><Document /></el-icon>
                <span class="node-label" :title="node.label">{{ node.label }}</span>
              </span>
            </template>
          </el-tree>
        </div>
      </div>

      <!-- 分隔线 -->
      <div v-if="showFileTree" class="tree-separator">
        <el-icon
          class="toggle-tree-icon"
          @click="showFileTree = false"
        >
          <DArrowLeft />
        </el-icon>
      </div>

      <!-- 右侧编辑区 -->
      <div class="editor-panel">
        <!-- 标签页 -->
        <div class="tabs-container" v-if="fileTabs.length > 0">
          <el-tabs
            v-model="activeTabPath"
            type="card"
            :closable="fileTabs.length > 1"
            @tab-remove="removeTab"
            @tab-change="changeTab"
          >
            <el-tab-pane
              v-for="tab in fileTabs"
              :key="tab.path"
              :name="tab.path"
            >
              <template #label>
                <el-tooltip :content="tab.path" placement="bottom">
                  <span class="tab-label">
                    {{ tab.name }}
                    <el-icon v-if="tab.modified" class="modified-icon"><StarFilled /></el-icon>
                  </span>
                </el-tooltip>
              </template>
            </el-tab-pane>
          </el-tabs>
        </div>

        <!-- 编辑器容器 -->
        <div class="editor-wrapper">
          <el-icon
            v-if="!showFileTree"
            class="show-tree-icon"
            @click="showFileTree = true"
          >
            <DArrowRight />
          </el-icon>
          <div v-if="fileTabs.length === 0" class="empty-editor">
            <el-empty description="请从左侧选择文件或关闭编辑器">
              <el-button type="primary" @click="showFileTree = true">显示文件树</el-button>
            </el-empty>
          </div>
          <div ref="editorContainer" class="monaco-container"></div>
        </div>
      </div>
    </div>

    <!-- 底部状态栏 -->
    <div class="editor-footer">
      <div class="footer-left">
        <span class="footer-text">{{ currentFile?.path || '' }}</span>
        <el-tag v-if="currentFile?.truncated" type="warning" size="small">
          已截断
        </el-tag>
      </div>
      <div class="footer-right">
        <el-divider direction="vertical" class="footer-divider" />
        <el-dropdown trigger="click" placement="top" @command="changeTheme">
          <span class="el-dropdown-link footer-link">
            {{ currentThemeLabel }}
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="item in themes" :key="item.value" :command="item.value">
                <div class="dropdown-item">
                  <span>{{ item.label }}</span>
                  <el-icon v-if="config.theme === item.value"><Check /></el-icon>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-divider direction="vertical" class="footer-divider" />
        <el-dropdown trigger="click" placement="top" @command="changeEOL">
          <span class="el-dropdown-link footer-link">
            {{ config.eol === 0 ? 'CRLF' : 'LF' }}
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item v-for="item in eols" :key="item.value" :command="item.value">
                <div class="dropdown-item">
                  <span>{{ item.label }}</span>
                  <el-icon v-if="config.eol === item.value"><Check /></el-icon>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-divider direction="vertical" class="footer-divider" />
        <el-dropdown trigger="click" placement="top" @command="changeLanguage">
          <span class="el-dropdown-link footer-link">
            {{ config.language }}
          </span>
          <template #dropdown>
            <el-dropdown-menu class="language-dropdown">
              <el-dropdown-item v-for="item in languages" :key="item.label" :command="item.label">
                <div class="dropdown-item">
                  <span>{{ item.label }}</span>
                  <el-icon v-if="config.language === item.label"><Check /></el-icon>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-divider direction="vertical" class="footer-divider" />
        <span class="footer-text">行 {{ cursorPosition.line }}, 列 {{ cursorPosition.column }}</span>
        <el-divider direction="vertical" class="footer-divider" />
        <span class="footer-text">UTF-8</span>
        <el-tag v-if="currentFile?.size" type="info" size="small">
          {{ formatSize(currentFile.size) }}
        </el-tag>
      </div>
    </div>

    <!-- 新建文件/目录对话框 -->
    <el-dialog
      v-model="createDialogVisible"
      :title="createType === 'dir' ? '新建目录' : '新建文件'"
      width="400px"
      append-to-body
    >
      <el-form @submit.prevent>
        <el-form-item :label="createType === 'dir' ? '目录名' : '文件名'">
          <el-input
            v-model="createName"
            :placeholder="createType === 'dir' ? '请输入目录名' : '请输入文件名'"
            @keyup.enter="confirmCreate"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmCreate" :loading="creating">确定</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed, onBeforeUnmount, shallowRef, nextTick } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  ArrowDown,
  Check,
  Loading,
  Top,
  Refresh,
  FolderAdd,
  DocumentAdd,
  Folder,
  Document,
  StarFilled,
  DArrowLeft,
  DArrowRight,
  FullScreen,
  Close,
} from '@element-plus/icons-vue';
import { fileApi } from '@/api/modules/file';
import type * as Monaco from 'monaco-editor';

// Monaco Editor 动态导入
let monaco: typeof Monaco | null = null;
let monacoLoading: Promise<typeof Monaco> | null = null;

const loadMonaco = async (): Promise<typeof Monaco> => {
  if (monaco) return monaco;
  if (monacoLoading) return monacoLoading;

  monacoLoading = (async () => {
    const [monacoModule, editorWorker, jsonWorker, cssWorker, htmlWorker, tsWorker] = await Promise.all([
      import('monaco-editor'),
      import('monaco-editor/esm/vs/editor/editor.worker?worker'),
      import('monaco-editor/esm/vs/language/json/json.worker?worker'),
      import('monaco-editor/esm/vs/language/css/css.worker?worker'),
      import('monaco-editor/esm/vs/language/html/html.worker?worker'),
      import('monaco-editor/esm/vs/language/typescript/ts.worker?worker'),
    ]);

    // 注册 Worker
    (self as any).MonacoEnvironment = {
      getWorker(_: any, label: string) {
        if (label === 'json') {
          return new jsonWorker.default();
        }
        if (label === 'css' || label === 'scss' || label === 'less') {
          return new cssWorker.default();
        }
        if (label === 'html' || label === 'handlebars' || label === 'razor') {
          return new htmlWorker.default();
        }
        if (label === 'typescript' || label === 'javascript') {
          return new tsWorker.default();
        }
        return new editorWorker.default();
      },
    };

    monaco = monacoModule;
    return monacoModule;
  })();

  return monacoLoading;
};

// Props
const props = defineProps<{
  visible: boolean;
  filePath: string;
}>();

// Emits
const emit = defineEmits<{
  close: [];
  save: [content: string];
}>();

// ==================== 常量定义 ====================
const STORAGE_KEYS = {
  THEME: 'gpanel-editor-theme',
  WORD_WRAP: 'gpanel-editor-wordwrap',
  MINIMAP: 'gpanel-editor-minimap',
};

// 主题选项
const themes = [
  { label: 'Visual Studio Dark', value: 'vs-dark' },
  { label: 'Visual Studio Light', value: 'vs' },
  { label: 'High Contrast', value: 'hc-black' },
];

// 语言选项
const languages = [
  { label: 'plaintext', value: ['txt'] },
  { label: 'json', value: ['json'] },
  { label: 'vue', value: ['vue'] },
  { label: 'typescript', value: ['ts'] },
  { label: 'javascript', value: ['js'] },
  { label: 'html', value: ['html', 'htm'] },
  { label: 'css', value: ['css'] },
  { label: 'scss', value: ['scss', 'sass'] },
  { label: 'less', value: ['less'] },
  { label: 'markdown', value: ['md', 'markdown'] },
  { label: 'yaml', value: ['yml', 'yaml'] },
  { label: 'xml', value: ['xml'] },
  { label: 'php', value: ['php'] },
  { label: 'sql', value: ['sql'] },
  { label: 'go', value: ['go'] },
  { label: 'python', value: ['py'] },
  { label: 'java', value: ['java'] },
  { label: 'kotlin', value: ['kt'] },
  { label: 'shell', value: ['sh', 'bash'] },
  { label: 'ini', value: ['ini', 'conf', 'cfg'] },
  { label: 'dockerfile', value: ['dockerfile'] },
  { label: 'lua', value: ['lua'] },
  { label: 'rust', value: ['rs'] },
  { label: 'c', value: ['c', 'h'] },
  { label: 'cpp', value: ['cpp', 'hpp', 'cc'] },
];

// 换行符选项
const eols = [
  { label: 'LF (Linux/macOS)', value: 1 }, // EndOfLineSequence.LF
  { label: 'CRLF (Windows)', value: 0 }, // EndOfLineSequence.CRLF
];

// ==================== 状态定义 ====================
const dialogVisible = ref(props.visible);
const loading = ref(false);
const saving = ref(false);
const creating = ref(false);
const isFullscreen = ref(false);
const isMobile = ref(false);

// 编辑器相关
const editorContainer = ref<HTMLElement | null>(null);
const editor = shallowRef<Monaco.editor.IStandaloneCodeEditor | null>(null);
const cursorPosition = ref({ line: 1, column: 1 });

// 编辑器配置
const config = ref({
  theme: localStorage.getItem(STORAGE_KEYS.THEME) || 'vs-dark',
  language: 'plaintext',
  eol: 1, // LF
  wordWrap: (localStorage.getItem(STORAGE_KEYS.WORD_WRAP) as 'on' | 'off') || 'on',
  minimap: localStorage.getItem(STORAGE_KEYS.MINIMAP) !== 'false',
});

// 文件树相关
const showFileTree = ref(true);
const treeRef = ref();
const treeData = ref<any[]>([]);
const treeProps = {
  children: 'children',
  label: 'name',
};
const currentTreePath = ref('');

// 文件标签页
interface FileTab {
  path: string;
  name: string;
  content: string;
  originalContent: string;
  modified: boolean;
  language: string;
  size: number;
  mode: string;
  truncated: boolean;  // 文件是否被截断
}

const fileTabs = ref<FileTab[]>([]);
const activeTabPath = ref('');

// 创建对话框
const createDialogVisible = ref(false);
const createType = ref<'file' | 'dir'>('file');
const createName = ref('');

// ==================== 计算属性 ====================
const currentThemeLabel = computed(() => {
  const theme = themes.find(t => t.value === config.value.theme);
  return theme ? theme.label : 'Visual Studio Dark';
});

const currentFile = computed(() => {
  return fileTabs.value.find(t => t.path === activeTabPath.value);
});

const hasChanges = computed(() => {
  return currentFile.value?.modified || false;
});

const canGoUp = computed(() => {
  if (!currentTreePath.value) return false;
  return currentTreePath.value !== '/';
});

// ==================== 工具函数 ====================
const getLanguageFromExtension = (filePath: string): string => {
  if (!filePath) return 'plaintext';
  const ext = filePath.split('.').pop()?.toLowerCase() || '';
  
  for (const lang of languages) {
    if (lang.value.includes(ext)) {
      return lang.label;
    }
  }
  
  // 特殊处理 Dockerfile
  if (filePath.toLowerCase().endsWith('dockerfile')) {
    return 'dockerfile';
  }
  
  return 'plaintext';
};

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};

const getDirectoryPath = (filePath: string): string => {
  if (!filePath) return '/';
  const lastSlash = Math.max(filePath.lastIndexOf('/'), filePath.lastIndexOf('\\'));
  if (lastSlash <= 0) return '/';
  return filePath.substring(0, lastSlash);
};

const getFileName = (filePath: string): string => {
  if (!filePath) return '';
  const lastSlash = Math.max(filePath.lastIndexOf('/'), filePath.lastIndexOf('\\'));
  return filePath.substring(lastSlash + 1);
};

// ==================== 编辑器操作 ====================
const initEditor = async (): Promise<void> => {
  if (!editorContainer.value) return;

  // 动态加载 Monaco Editor
  const monacoInstance = await loadMonaco();

  // 销毁旧编辑器
  if (editor.value) {
    editor.value.dispose();
    editor.value = null;
  }

  await nextTick();

  if (!editorContainer.value) return;

  editor.value = monacoInstance.editor.create(editorContainer.value, {
    value: currentFile.value?.content || '',
    language: config.value.language,
    theme: config.value.theme,
    automaticLayout: true,
    fontSize: 14,
    lineNumbers: 'on',
    minimap: { enabled: config.value.minimap },
    wordWrap: config.value.wordWrap,
    scrollBeyondLastLine: false,
    renderWhitespace: 'selection',
    tabSize: 4,
    insertSpaces: true,
    folding: true,
    foldingStrategy: 'auto',
    showFoldingControls: 'mouseover',
    matchBrackets: 'always',
    autoClosingBrackets: 'always',
    autoClosingQuotes: 'always',
    formatOnPaste: true,
    formatOnType: true,
  });

  // 设置换行符
  if (currentFile.value) {
    editor.value.getModel()?.setEOL(config.value.eol);
  }

  // 监听内容变化
  editor.value.onDidChangeModelContent(() => {
    if (editor.value && currentFile.value) {
      const newContent = editor.value.getValue();
      currentFile.value.content = newContent;
      currentFile.value.modified = newContent !== currentFile.value.originalContent;
    }
  });

  // 监听光标位置
  editor.value.onDidChangeCursorPosition((e) => {
    cursorPosition.value = {
      line: e.position.lineNumber,
      column: e.position.column,
    };
  });

  // 快捷键保存
  editor.value.addCommand(monacoInstance.KeyMod.CtrlCmd | monacoInstance.KeyCode.KeyS, () => {
    saveCurrentFile();
  });
};

const updateEditorContent = async () => {
  if (!editor.value || !currentFile.value) return;

  const monacoInstance = await loadMonaco();
  const model = editor.value.getModel();
  if (model) {
    model.setValue(currentFile.value.content);
    monacoInstance.editor.setModelLanguage(model, currentFile.value.language);

    // 检测换行符
    if (currentFile.value.content.includes('\r\n')) {
      config.value.eol = 0; // CRLF
    } else {
      config.value.eol = 1; // LF
    }
    model.setEOL(config.value.eol);
  }
};

const destroyEditor = () => {
  if (editor.value) {
    editor.value.dispose();
    editor.value = null;
  }
};

// ==================== 文件操作 ====================
const loadFileContent = async (filePath: string): Promise<FileTab | null> => {
  try {
    const response = await fileApi.getFileContent({ path: filePath });
    if (response.data.code === 200) {
      const data = response.data.data;
      const tab: FileTab = {
        path: filePath,
        name: getFileName(filePath),
        content: data.content || '',
        originalContent: data.content || '',
        modified: false,
        language: getLanguageFromExtension(filePath),
        size: data.size || 0,
        mode: data.mode || '',
        truncated: data.truncated || false,
      };
      return tab;
    } else {
      ElMessage.error(response.data.message);
      return null;
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载文件内容失败');
    return null;
  }
};

const openFile = async (filePath: string) => {
  // 检查是否已打开
  const existingTab = fileTabs.value.find(t => t.path === filePath);
  if (existingTab) {
    activeTabPath.value = filePath;
    config.value.language = existingTab.language;
    updateEditorContent();
    return;
  }

  loading.value = true;
  const tab = await loadFileContent(filePath);
  loading.value = false;

  if (tab) {
    fileTabs.value.push(tab);
    activeTabPath.value = filePath;
    config.value.language = tab.language;
    updateEditorContent();
    
    // 大文件提示
    if (tab.truncated) {
      ElMessage.warning({
        message: `文件较大（${formatSize(tab.size)}），只显示最后 300 行内容`,
        duration: 5000,
      });
    }
  }
};

const saveCurrentFile = async () => {
  if (!currentFile.value || !currentFile.value.modified) {
    ElMessage.warning('文件未修改');
    return;
  }

  saving.value = true;
  try {
    const response = await fileApi.saveFileContent({
      path: currentFile.value.path,
      content: currentFile.value.content,
    });
    
    if (response.data.code === 200) {
      currentFile.value.originalContent = currentFile.value.content;
      currentFile.value.modified = false;
      ElMessage.success('保存成功');
      emit('save', currentFile.value.content);
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败');
  } finally {
    saving.value = false;
  }
};

const handleReset = () => {
  if (!currentFile.value || !currentFile.value.modified) {
    ElMessage.warning('文件未修改');
    return;
  }
  
  currentFile.value.content = currentFile.value.originalContent;
  currentFile.value.modified = false;
  updateEditorContent();
  ElMessage.success('已重置');
};

// ==================== 标签页操作 ====================
const removeTab = async (targetPath: string) => {
  const tab = fileTabs.value.find(t => t.path === targetPath);
  if (!tab) return;

  if (tab.modified) {
    try {
      await ElMessageBox.confirm('文件已修改但未保存，确定要关闭吗？', '提示', {
        confirmButtonText: '保存',
        cancelButtonText: '不保存',
        distinguishCancelAndClose: true,
        type: 'warning',
      });
      // 保存后关闭
      const response = await fileApi.saveFileContent({
        path: tab.path,
        content: tab.content,
      });
      if (response.data.code === 200) {
        doRemoveTab(targetPath);
      }
    } catch (action: any) {
      if (action === 'cancel') {
        // 不保存，直接关闭
        doRemoveTab(targetPath);
      }
      // 其他情况（如点击关闭按钮）不做操作
    }
  } else {
    doRemoveTab(targetPath);
  }
};

const doRemoveTab = (targetPath: string) => {
  const index = fileTabs.value.findIndex(t => t.path === targetPath);
  if (index === -1) return;

  fileTabs.value.splice(index, 1);

  // 如果关闭的是当前标签，切换到其他标签
  if (activeTabPath.value === targetPath) {
    if (fileTabs.value.length > 0) {
      const newIndex = Math.min(index, fileTabs.value.length - 1);
      activeTabPath.value = fileTabs.value[newIndex].path;
      config.value.language = fileTabs.value[newIndex].language;
      updateEditorContent();
    } else {
      activeTabPath.value = '';
      if (editor.value) {
        editor.value.setValue('');
      }
    }
  }
};

const changeTab = (path: string) => {
  const tab = fileTabs.value.find(t => t.path === path);
  if (tab) {
    activeTabPath.value = path;
    config.value.language = tab.language;
    updateEditorContent();
  }
};

// ==================== 文件树操作 ====================
const loadTreeData = async (path: string) => {
  try {
    const response = await fileApi.getFileList({
      path: path,
      showHidden: false,
      page: 1,
      pageSize: 1000,
      sortBy: 'name',
      sortOrder: 'ascending',
    });

    if (response.data.code === 200) {
      const items = response.data.data.items || [];
      // 排序：目录在前，文件在后
      const sorted = items.sort((a: any, b: any) => {
        if (a.isDir && !b.isDir) return -1;
        if (!a.isDir && b.isDir) return 1;
        return a.name.localeCompare(b.name);
      });
      
      treeData.value = sorted.map((item: any) => ({
        path: item.path,
        name: item.name,
        isDir: item.isDir,
        children: item.isDir ? [] : undefined,
      }));
      currentTreePath.value = path;
    }
  } catch (error: any) {
    console.error('加载文件树失败:', error);
  }
};

const handleTreeNodeClick = (data: any) => {
  if (data.isDir) {
    loadTreeData(data.path);
  } else {
    openFile(data.path);
  }
};

const goToParentDir = () => {
  const parentPath = getDirectoryPath(currentTreePath.value);
  if (parentPath && parentPath !== currentTreePath.value) {
    loadTreeData(parentPath);
  }
};

const refreshTree = () => {
  if (currentTreePath.value) {
    loadTreeData(currentTreePath.value);
    ElMessage.success('刷新成功');
  }
};

// ==================== 创建文件/目录 ====================
const handleCreateInTree = (command: string) => {
  createType.value = command as 'file' | 'dir';
  createName.value = '';
  createDialogVisible.value = true;
};

const confirmCreate = async () => {
  if (!createName.value) {
    ElMessage.warning(createType.value === 'dir' ? '请输入目录名' : '请输入文件名');
    return;
  }

  creating.value = true;
  try {
    const newPath = `${currentTreePath.value}/${createName.value}`.replace(/\/+/g, '/');
    const response = await fileApi.createFile({
      path: newPath,
      isDir: createType.value === 'dir',
    });

    if (response.data.code === 200) {
      ElMessage.success(createType.value === 'dir' ? '创建目录成功' : '创建文件成功');
      createDialogVisible.value = false;
      refreshTree();
      
      // 如果是文件，自动打开
      if (createType.value === 'file') {
        openFile(newPath);
      }
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '创建失败');
  } finally {
    creating.value = false;
  }
};

// ==================== 配置操作 ====================
const changeTheme = async (theme: string) => {
  config.value.theme = theme;
  const monacoInstance = await loadMonaco();
  monacoInstance.editor.setTheme(theme);
  localStorage.setItem(STORAGE_KEYS.THEME, theme);
};

const changeLanguage = async (language: string) => {
  config.value.language = language;
  if (editor.value) {
    const monacoInstance = await loadMonaco();
    const model = editor.value.getModel();
    if (model) {
      monacoInstance.editor.setModelLanguage(model, language);
    }
  }
};

const changeEOL = (eol: number) => {
  config.value.eol = eol;
  if (editor.value) {
    editor.value.getModel()?.pushEOL(eol);
  }
};

const toggleMinimap = () => {
  config.value.minimap = !config.value.minimap;
  localStorage.setItem(STORAGE_KEYS.MINIMAP, String(config.value.minimap));
  if (editor.value) {
    editor.value.updateOptions({ minimap: { enabled: config.value.minimap } });
  }
};

const toggleWordWrap = () => {
  config.value.wordWrap = config.value.wordWrap === 'on' ? 'off' : 'on';
  localStorage.setItem(STORAGE_KEYS.WORD_WRAP, config.value.wordWrap);
  if (editor.value) {
    editor.value.updateOptions({ wordWrap: config.value.wordWrap });
  }
};

const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value;
};

// ==================== 拖拽调整 ====================
const startResize = (e: MouseEvent) => {
  const startX = e.clientX;
  const treePanel = document.querySelector('.file-tree-panel') as HTMLElement;
  if (!treePanel) return;
  
  const startWidth = treePanel.offsetWidth;
  
  const onMouseMove = (e: MouseEvent) => {
    const newWidth = startWidth + (e.clientX - startX);
    if (newWidth >= 150 && newWidth <= 500) {
      treePanel.style.width = `${newWidth}px`;
    }
  };
  
  const onMouseUp = () => {
    document.removeEventListener('mousemove', onMouseMove);
    document.removeEventListener('mouseup', onMouseUp);
  };
  
  document.addEventListener('mousemove', onMouseMove);
  document.addEventListener('mouseup', onMouseUp);
};

// ==================== 对话框操作 ====================
const onOpened = async () => {
  await initEditor();
  if (props.filePath) {
    const dirPath = getDirectoryPath(props.filePath);
    loadTreeData(dirPath);
    openFile(props.filePath);
  }
};

const handleClose = async () => {
  // 检查是否有未保存的文件
  const modifiedTabs = fileTabs.value.filter(t => t.modified);
  
  if (modifiedTabs.length > 0) {
    try {
      await ElMessageBox.confirm(
        `有 ${modifiedTabs.length} 个文件未保存，确定要关闭吗？`,
        '提示',
        {
          confirmButtonText: '全部保存',
          cancelButtonText: '不保存',
          distinguishCancelAndClose: true,
          type: 'warning',
        }
      );
      // 保存所有
      for (const tab of modifiedTabs) {
        await fileApi.saveFileContent({
          path: tab.path,
          content: tab.content,
        });
      }
    } catch (action: any) {
      if (action !== 'cancel') {
        return; // 取消关闭
      }
    }
  }

  destroyEditor();
  fileTabs.value = [];
  activeTabPath.value = '';
  emit('close');
};

// ==================== 生命周期 ====================
watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal;
  if (!newVal) {
    destroyEditor();
    fileTabs.value = [];
    activeTabPath.value = '';
  }
});

watch(dialogVisible, (newVal) => {
  if (!newVal) {
    emit('close');
  }
});

watch(activeTabPath, () => {
  updateEditorContent();
});

onBeforeUnmount(() => {
  destroyEditor();
});
</script>

<style scoped lang="scss">
/* CSS变量定义 */
.file-editor-dialog {
  --editor-header-bg: var(--el-bg-color);
  --editor-toolbar-bg: var(--el-fill-color-lighter);
  --editor-footer-bg: var(--el-fill-color-lighter);
  --editor-border-color: var(--el-border-color-lighter);
  --editor-text-color: var(--text-primary, var(--el-text-color-primary));
  --editor-text-secondary: var(--text-secondary, var(--el-text-color-regular));
  --editor-primary: var(--el-color-primary);
}

/* 对话框基础样式 */
.file-editor-dialog :deep(.el-dialog) {
  background: var(--editor-header-bg);
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  border-radius: 8px;
  overflow: hidden;
}

.file-editor-dialog :deep(.el-dialog__header) {
  padding: 0;
  margin: 0;
  flex-shrink: 0;
}

.file-editor-dialog :deep(.el-dialog__body) {
  padding: 0;
  flex: 1;
  min-height: 400px;
  display: flex;
  flex-direction: column;
  background: var(--editor-header-bg);
  overflow: hidden;
}

.file-editor-dialog :deep(.el-dialog.is-fullscreen) {
  max-height: 100vh;
  border-radius: 0;
}

.file-editor-dialog :deep(.el-dialog.is-fullscreen .el-dialog__body) {
  min-height: unset;
}

/* 头部样式 */
.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--editor-header-bg);
  border-bottom: 1px solid var(--editor-border-color);
}

.dialog-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--editor-text-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 70%;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.header-btn {
  --el-button-text-color: var(--editor-text-color);
  --el-button-hover-text-color: var(--editor-primary);
  --el-button-bg-color: transparent;
  --el-button-hover-bg-color: var(--el-fill-color);
  --el-button-border-color: transparent;
  --el-button-hover-border-color: transparent;
  min-width: 32px !important;
  min-height: 32px !important;
  padding: 6px !important;
  font-size: 18px !important;
  color: var(--editor-text-color) !important;
  background-color: transparent !important;
  border: 1px solid transparent !important;

  :deep(.el-icon),
  :deep(i),
  :deep(svg) {
    color: inherit !important;
    fill: currentColor;
    stroke: currentColor;
  }
  
  &:hover {
    color: var(--editor-primary) !important;
    background-color: var(--el-fill-color) !important;
  }
}

.close-btn:hover {
  --el-button-hover-text-color: var(--el-color-danger);
  color: var(--el-color-danger) !important;
}

/* 工具栏样式 */
.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--editor-toolbar-bg);
  border-bottom: 1px solid var(--editor-border-color);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 0;
}

.toolbar-text {
  cursor: pointer;
  color: var(--editor-text-color);
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
  
  &:hover {
    color: var(--editor-primary);
    background-color: var(--el-fill-color-light);
  }
  
  &.is-disabled {
    cursor: not-allowed;
    color: var(--el-text-color-placeholder);
    &:hover {
      background-color: transparent;
    }
  }
}

.toolbar-divider {
  margin: 0 4px;
  height: 14px;
}

.el-dropdown-link {
  cursor: pointer;
  color: var(--editor-text-color);
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  
  &:hover {
    color: var(--editor-primary);
    background-color: var(--el-fill-color-light);
  }
}

/* 加载状态 */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 12px;
  color: var(--editor-text-secondary);
}

/* 主内容区 */
.editor-main {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 300px;
}

/* 文件树面板 */
.file-tree-panel {
  width: 220px;
  min-width: 180px;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  background: var(--el-bg-color);
  border-right: 1px solid var(--editor-border-color);
  flex-shrink: 0;
  overflow: hidden;
}

.tree-header {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 8px;
  gap: 0;
}

.tree-header-btn {
  cursor: pointer;
  color: var(--editor-text-color);
  padding: 2px 6px;
  border-radius: 4px;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  
  &:hover {
    color: var(--editor-primary);
    background-color: var(--el-fill-color-light);
  }
  
  &.is-disabled {
    cursor: not-allowed;
    color: var(--el-text-color-placeholder);
    &:hover {
      background-color: transparent;
    }
  }
}

.tree-header-text {
  margin-left: 2px;
}

.tree-divider-v {
  margin: 0 4px;
  height: 12px;
}

.tree-divider {
  margin: 0;
}

.tree-content {
  flex: 1;
  overflow: auto;
  padding: 4px 8px;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
}

.tree-icon {
  font-size: 16px;
  flex-shrink: 0;
}

.folder-icon {
  color: var(--editor-primary);
}

.file-icon {
  color: var(--editor-text-secondary);
}

.node-label {
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 分隔线 */
.tree-separator {
  position: relative;
  width: 1px;
  background-color: var(--editor-border-color);
  flex-shrink: 0;
}

.toggle-tree-icon {
  position: absolute;
  top: 50%;
  left: -9px;
  transform: translateY(-50%);
  cursor: pointer;
  padding: 8px 2px;
  background-color: var(--el-fill-color-lighter);
  border-radius: 4px 0 0 4px;
  color: var(--editor-text-color);
  transition: all 0.2s;
  
  &:hover {
    background-color: var(--el-fill-color-dark);
    color: var(--editor-primary);
  }
}

/* 编辑器面板 */
.editor-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}

/* 标签页 */
.tabs-container {
  display: flex;
  align-items: center;
  background: var(--editor-toolbar-bg);
  border-bottom: 1px solid var(--editor-border-color);
}

.tabs-container :deep(.el-tabs) {
  flex: 1;
  --el-tabs-header-height: 29px;
}

.tabs-container :deep(.el-tabs__header) {
  margin: 0;
  border-bottom: none;
  height: 29px;
}

.tabs-container :deep(.el-tabs__nav-wrap) {
  padding: 0 4px;
  height: 28px;
  line-height: 28px;
}

.tabs-container :deep(.el-tabs__nav) {
  border: none !important;
  border-radius: 0 !important;
}

.tabs-container :deep(.el-tabs__item) {
  height: 28px;
  line-height: 28px;
  padding: 0 12px !important;
  color: var(--editor-text-color) !important;
  
  &:hover {
    color: var(--editor-primary) !important;
  }
  
  &.is-active {
    color: var(--editor-primary) !important;
    background-color: var(--el-bg-color);
    border-bottom: 2px solid var(--editor-primary);
  }
}

.tabs-container :deep(.el-tabs__item .is-icon-close) {
  color: var(--editor-text-secondary) !important;
}

.tabs-container :deep(.el-tabs__item .is-icon-close:hover) {
  color: var(--editor-primary) !important;
  background-color: var(--el-fill-color) !important;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.modified-icon {
  color: var(--el-color-warning);
  font-size: 10px;
}

/* 编辑器容器 */
.editor-wrapper {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 200px;
  position: relative;
}

.show-tree-icon {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  cursor: pointer;
  padding: 16px 4px;
  background-color: var(--el-fill-color-lighter);
  border-radius: 0 4px 4px 0;
  color: var(--editor-text-color);
  z-index: 10;
  transition: all 0.2s;
  
  &:hover {
    background-color: var(--el-fill-color-dark);
    color: var(--editor-primary);
  }
}

.empty-editor {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.monaco-container {
  width: 100%;
  height: 100%;
  min-height: 200px;
}

/* 底部状态栏 */
.editor-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  height: 24px;
  background: var(--editor-footer-bg);
  border-top: 1px solid var(--editor-border-color);
  font-size: 12px;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
  flex: 1;
  min-width: 0;
}

.footer-text {
  color: var(--editor-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 0;
}

.footer-link {
  color: var(--editor-text-color);
  font-size: 12px;
  
  &:hover {
    color: var(--editor-primary);
  }
}

.footer-divider {
  margin: 0 8px;
  height: 12px;
}

/* 下拉菜单项 */
.dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
  min-width: 140px;
}

.language-dropdown :deep(.el-dropdown-menu) {
  max-height: 300px;
  overflow-y: auto;
}

/* 深色主题适配 */
.file-editor-dialog :deep(.monaco-editor) {
  background: inherit;
}

.file-editor-dialog :deep(.monaco-editor .margin) {
  background: inherit;
}

/* 暗色主题下的样式调整 */
.dark .file-editor-dialog {
  --editor-header-bg: #1e1e1e;
  --editor-toolbar-bg: #252526;
  --editor-footer-bg: #252526;
  --editor-border-color: #3c3c3c;
  --editor-text-color: #e5eaf3;
  --editor-text-secondary: #cfd3dc;
}
</style>
