<template>
  <el-dialog
    v-model="dialogVisible"
    title="文件编辑器"
    width="95%"
    :close-on-click-modal="false"
    class="file-editor-dialog"
    :fullscreen="isFullscreen"
    @close="handleClose"
    @opened="onOpened"
  >
    <template #header>
      <div class="dialog-header">
        <span class="dialog-title">{{ currentFile?.name || '文件编辑器' }}</span>
        <div class="header-actions">
          <el-button
            v-if="!isMobile"
            :icon="isFullscreen ? 'FullScreen' : 'FullScreen'"
            text
            @click="toggleFullscreen"
          >
            {{ isFullscreen ? '退出全屏' : '全屏' }}
          </el-button>
          <el-button :icon="'Close'" text @click="handleClose"></el-button>
        </div>
      </div>
    </template>

    <!-- 工具栏 -->
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <el-button text @click="handleReset" :disabled="!hasChanges">
          <el-icon><RefreshRight /></el-icon>
          重置
        </el-button>
        <el-divider direction="vertical" />
        <el-button text @click="saveCurrentFile" :loading="saving">
          <el-icon><DocumentChecked /></el-icon>
          保存
        </el-button>
        <el-divider direction="vertical" />
        <el-dropdown trigger="click" @command="changeTheme">
          <el-button text>
            主题: {{ currentThemeLabel }}
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
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
        <el-divider direction="vertical" />
        <el-dropdown trigger="click" @command="changeLanguage">
          <el-button text>
            语言: {{ config.language }}
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
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
        <el-divider direction="vertical" />
        <el-dropdown trigger="click" @command="changeEOL">
          <el-button text>
            换行符: {{ config.eol === 1 ? 'LF' : 'CRLF' }}
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
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
        <el-divider direction="vertical" />
        <el-dropdown trigger="click">
          <el-button text>
            设置
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
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
          <el-button text size="small" @click="goToParentDir" :disabled="!canGoUp">
            <el-icon><Top /></el-icon>
            上级
          </el-button>
          <el-divider direction="vertical" />
          <el-button text size="small" @click="refreshTree">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
          <el-divider direction="vertical" />
          <el-dropdown trigger="click" @command="handleCreateInTree">
            <el-button text size="small">
              新建
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
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
                <el-icon v-if="data.isDir" color="#409eff"><Folder /></el-icon>
                <el-icon v-else color="#909399"><Document /></el-icon>
                <span class="node-label">{{ node.label }}</span>
              </span>
            </template>
          </el-tree>
        </div>
      </div>

      <!-- 拖拽分隔条 -->
      <div 
        v-show="showFileTree" 
        class="resize-handle"
        @mousedown="startResize"
      ></div>

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
          <el-button
            class="toggle-tree-btn"
            text
            size="small"
            @click="showFileTree = !showFileTree"
          >
            <el-icon v-if="showFileTree"><DArrowLeft /></el-icon>
            <el-icon v-else><DArrowRight /></el-icon>
          </el-button>
        </div>

        <!-- 编辑器容器 -->
        <div class="editor-wrapper">
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
        <span>{{ currentFile?.path || '' }}</span>
        <el-tag v-if="currentFile?.truncated" type="warning" size="small">
          已截断
        </el-tag>
      </div>
      <div class="footer-right">
        <span>行 {{ cursorPosition.line }}, 列 {{ cursorPosition.column }}</span>
        <span>{{ config.eol === 1 ? 'LF' : 'CRLF' }}</span>
        <span>UTF-8</span>
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
  RefreshRight,
  DocumentChecked,
  Top,
  Refresh,
  FolderAdd,
  DocumentAdd,
  Folder,
  Document,
  StarFilled,
  DArrowLeft,
  DArrowRight,
} from '@element-plus/icons-vue';
import { fileApi } from '@/api/modules/file';
import * as monaco from 'monaco-editor';

// 配置 Monaco Editor Web Worker
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';

// 注册 Worker
(self as any).MonacoEnvironment = {
  getWorker(_: any, label: string) {
    if (label === 'json') {
      return new jsonWorker();
    }
    if (label === 'css' || label === 'scss' || label === 'less') {
      return new cssWorker();
    }
    if (label === 'html' || label === 'handlebars' || label === 'razor') {
      return new htmlWorker();
    }
    if (label === 'typescript' || label === 'javascript') {
      return new tsWorker();
    }
    return new editorWorker();
  },
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
  { label: 'LF (Linux/macOS)', value: monaco.editor.EndOfLineSequence.LF },
  { label: 'CRLF (Windows)', value: monaco.editor.EndOfLineSequence.CRLF },
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
const editor = shallowRef<monaco.editor.IStandaloneCodeEditor | null>(null);
const cursorPosition = ref({ line: 1, column: 1 });

// 编辑器配置
const config = ref({
  theme: localStorage.getItem(STORAGE_KEYS.THEME) || 'vs-dark',
  language: 'plaintext',
  eol: monaco.editor.EndOfLineSequence.LF,
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
const initEditor = (): Promise<void> => {
  return new Promise((resolve) => {
    if (!editorContainer.value) {
      resolve();
      return;
    }
    
    // 销毁旧编辑器
    if (editor.value) {
      editor.value.dispose();
      editor.value = null;
    }

    nextTick(() => {
      if (!editorContainer.value) {
        resolve();
        return;
      }
      
      editor.value = monaco.editor.create(editorContainer.value, {
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
      editor.value.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
        saveCurrentFile();
      });

      resolve();
    });
  });
};

const updateEditorContent = () => {
  if (!editor.value || !currentFile.value) return;
  
  const model = editor.value.getModel();
  if (model) {
    model.setValue(currentFile.value.content);
    monaco.editor.setModelLanguage(model, currentFile.value.language);
    
    // 检测换行符
    if (currentFile.value.content.includes('\r\n')) {
      config.value.eol = monaco.editor.EndOfLineSequence.CRLF;
    } else {
      config.value.eol = monaco.editor.EndOfLineSequence.LF;
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
const changeTheme = (theme: string) => {
  config.value.theme = theme;
  monaco.editor.setTheme(theme);
  localStorage.setItem(STORAGE_KEYS.THEME, theme);
};

const changeLanguage = (language: string) => {
  config.value.language = language;
  if (editor.value) {
    const model = editor.value.getModel();
    if (model) {
      monaco.editor.setModelLanguage(model, language);
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

<style scoped>
.file-editor-dialog :deep(.el-dialog__header) {
  padding: 0;
  margin: 0;
}

.file-editor-dialog :deep(.el-dialog__body) {
  padding: 0;
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.dialog-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 4px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 12px;
  color: #909399;
}

.editor-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.file-tree-panel {
  width: 250px;
  min-width: 150px;
  max-width: 500px;
  display: flex;
  flex-direction: column;
  background: #fafafa;
  border-right: 1px solid #e4e7ed;
}

.tree-header {
  display: flex;
  align-items: center;
  padding: 8px;
  gap: 4px;
}

.tree-divider {
  margin: 0;
}

.tree-content {
  flex: 1;
  overflow: auto;
  padding: 8px;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 6px;
}

.node-label {
  font-size: 13px;
}

.resize-handle {
  width: 4px;
  cursor: col-resize;
  background: transparent;
  transition: background-color 0.2s;
}

.resize-handle:hover {
  background: #409eff;
}

.editor-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.tabs-container {
  display: flex;
  align-items: center;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
}

.tabs-container :deep(.el-tabs) {
  flex: 1;
  margin-bottom: 0;
}

.tabs-container :deep(.el-tabs__header) {
  margin: 0;
  border-bottom: none;
}

.tabs-container :deep(.el-tabs__nav-wrap) {
  padding: 0 8px;
}

.toggle-tree-btn {
  margin-right: 8px;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.modified-icon {
  color: #e6a23c;
  font-size: 12px;
}

.editor-wrapper {
  flex: 1;
  display: flex;
  overflow: hidden;
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
}

.editor-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 16px;
  background: #f5f7fa;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  color: #606266;
}

.footer-left {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-width: 120px;
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
</style>