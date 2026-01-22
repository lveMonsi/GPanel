<template>
  <div class="file-manager">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <!-- 左侧：导航按钮 -->
      <div class="toolbar-left">
        <el-tooltip content="返回上级" placement="top">
          <el-button :icon="ArrowLeft" circle @click="goBack" :disabled="currentPath === '/'" />
        </el-tooltip>
        <el-tooltip content="刷新" placement="top">
          <el-button :icon="Refresh" circle @click="refresh" />
        </el-tooltip>
        <el-tooltip :content="showHidden ? '隐藏隐藏文件' : '显示隐藏文件'" placement="top">
          <el-button circle :type="showHidden ? 'primary' : ''" :icon="showHidden ? View : Hide" @click="toggleHidden" />
        </el-tooltip>
      </div>

      <!-- 中间：面包屑导航 -->
      <div class="breadcrumb">
        <el-link @click.stop="jump('/')" :underline="false">
          <el-icon :size="18" color="#606266"><HomeFilled /></el-icon>
        </el-link>
        <span v-for="(path, index) in pathSegments" :key="index" class="breadcrumb-segment">
          <span class="breadcrumb-separator">/</span>
          <el-tooltip :content="path.name" placement="bottom" :disabled="path.name.length <= 20">
            <el-link
              :underline="false"
              class="path-segment"
              @click.stop="jump(path.url)"
            >
              {{ path.name.length > 20 ? path.name.substring(0, 18) + '...' : path.name }}
            </el-link>
          </el-tooltip>
        </span>
      </div>

      <!-- 右侧：搜索框 -->
      <div class="toolbar-right">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索文件"
          :prefix-icon="Search"
          @input="handleSearch"
          @clear="handleSearch"
          clearable
          class="search-input"
        />
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <!-- 左侧操作按钮 -->
      <div class="action-bar-left">
        <el-dropdown @command="handleCreate">
          <el-button type="primary">
            新建
            <el-icon class="ml-1"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="dir" :icon="FolderAdd">新建目录</el-dropdown-item>
              <el-dropdown-item command="file" :icon="DocumentAdd">新建文件</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <el-button :icon="Upload" @click="showUploadDialog" :disabled="currentPath === '/'">
          上传文件
        </el-button>

        <el-button-group v-if="selectedFiles.length > 0">
          <el-button :icon="Operation" @click="showCompressDialog">压缩</el-button>
          <el-button :icon="Operation" @click="showChmodDialog">权限</el-button>
          <el-button :icon="Delete" type="danger" @click="batchDelete">删除</el-button>
        </el-button-group>
      </div>

      <!-- 右侧信息 -->
      <div class="action-bar-right">
        共 {{ fileList.length }} 项
      </div>
    </div>

    <!-- 文件列表 -->
    <div class="file-list">
      <el-table
        :data="fileList"
        @row-click="handleRowClick"
        @row-dblclick="handleRowDblClick"
        @selection-change="handleSelectionChange"
        height="100%"
        stripe
        class="file-table"
      >
        <el-table-column type="selection" width="45" />
        <el-table-column prop="name" label="名称" min-width="250" sortable>
          <template #default="{ row }">
            <div class="file-name-cell">
              <el-icon v-if="row.isDir" :size="20" color="#409eff"><Folder /></el-icon>
              <el-icon v-else :size="20" color="#909399"><Document /></el-icon>
              <span class="file-name">{{ row.name }}</span>
              <el-tag v-if="row.isSymlink" size="small" type="info" effect="plain">链接</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="大小" width="120" sortable>
          <template #default="{ row }">
            <span class="file-meta">{{ row.isDir ? '-' : formatSize(row.size) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="mode" label="权限" width="100" />
        <el-table-column prop="user" label="用户" width="100" />
        <el-table-column prop="group" label="组" width="100" />
        <el-table-column prop="modTime" label="修改时间" width="180" sortable>
          <template #default="{ row }">
            <span class="file-meta">{{ formatDate(row.modTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button-group>
              <el-tooltip content="编辑" placement="top">
                <el-button size="small" :icon="Edit" @click.stop="editFile(row)" :disabled="row.isDir" />
              </el-tooltip>
              <el-tooltip content="下载" placement="top">
                <el-button size="small" :icon="Download" @click.stop="downloadFile(row)" :disabled="row.isDir" />
              </el-tooltip>
              <el-tooltip content="解压" placement="top">
                <el-button size="small" :icon="FolderOpened" @click.stop="decompressFile(row)" :disabled="!isCompressedFile(row)" />
              </el-tooltip>
              <el-tooltip content="删除" placement="top">
                <el-button size="small" :icon="Delete" type="danger" @click.stop="deleteFile(row)" />
              </el-tooltip>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 创建目录对话框 -->
    <el-dialog v-model="createDirDialogVisible" title="新建目录" width="450px" :close-on-click-modal="false">
      <el-form :model="createDirForm" label-width="80px">
        <el-form-item label="目录名" required>
          <el-input v-model="createDirForm.name" placeholder="请输入目录名" clearable @keyup.enter="createDir" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDirDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createDir" :loading="creating">确定</el-button>
      </template>
    </el-dialog>

    <!-- 创建文件对话框 -->
    <el-dialog v-model="createFileDialogVisible" title="新建文件" width="450px" :close-on-click-modal="false">
      <el-form :model="createFileForm" label-width="80px">
        <el-form-item label="文件名" required>
          <el-input v-model="createFileForm.name" placeholder="请输入文件名" clearable @keyup.enter="createFile" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createFileDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createFile" :loading="creating">确定</el-button>
      </template>
    </el-dialog>

    <!-- 上传文件对话框 -->
    <el-dialog v-model="uploadDialogVisible" title="上传文件" width="550px" :close-on-click-modal="false">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :on-change="handleFileChange"
        :file-list="uploadFileList"
        drag
        multiple
        class="upload-area"
      >
        <el-icon class="el-icon--upload" :size="50" color="#409eff"><UploadFilled /></el-icon>
        <div class="el-upload__text mt-2">
          将文件拖到此处，或<em class="text-blue-600">点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip text-gray-500 mt-2">
            支持上传多个文件
          </div>
        </template>
      </el-upload>
      <el-checkbox v-model="overwrite" class="mt-4">覆盖已存在的文件</el-checkbox>
      <template #footer>
        <el-button @click="uploadDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="uploadFiles" :loading="uploading">上传</el-button>
      </template>
    </el-dialog>

    <!-- 文件编辑器 -->
    <FileEditor
      v-if="editorVisible"
      :visible="editorVisible"
      :file-path="editingFilePath"
      @close="editorVisible = false"
      @save="handleEditorSave"
    />

    <!-- 压缩对话框 -->
    <el-dialog v-model="compressDialogVisible" title="压缩文件" width="450px" :close-on-click-modal="false">
      <el-form :model="compressForm" label-width="80px">
        <el-form-item label="压缩名称" required>
          <el-input v-model="compressForm.name" placeholder="请输入压缩文件名" clearable />
        </el-form-item>
        <el-form-item label="压缩格式" required>
          <el-select v-model="compressForm.type" placeholder="选择压缩格式" style="width: 100%">
            <el-option label="tar.gz" value="tar.gz" />
            <el-option label="zip" value="zip" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="compressDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="compressFiles" :loading="compressing">压缩</el-button>
      </template>
    </el-dialog>

    <!-- 权限修改对话框 -->
    <el-dialog v-model="chmodDialogVisible" title="修改权限" width="450px" :close-on-click-modal="false">
      <el-form :model="chmodForm" label-width="80px">
        <el-form-item label="权限模式" required>
          <el-input v-model="chmodForm.mode" placeholder="例如: 755" maxlength="3" clearable />
          <div class="text-xs text-gray-500 mt-1">请输入 3 位数字权限，如 755、644</div>
        </el-form-item>
        <el-form-item label="递归修改">
          <el-switch v-model="chmodForm.sub" />
          <div class="text-xs text-gray-500 mt-1">是否递归修改子目录和文件</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="chmodDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="chmodFiles">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
// 版本：2025-01-21-optimized-ui
import { ref, onMounted, watch, onBeforeMount, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  ArrowLeft,
  Refresh,
  FolderAdd,
  DocumentAdd,
  Upload,
  Search,
  FolderOpened,
  Folder,
  Document,
  Edit,
  Download,
  Delete,
  UploadFilled,
  Operation,
  HomeFilled,
  View,
  Hide,
  ArrowDown,
} from '@element-plus/icons-vue';
import { fileApi } from '@/api/modules/file';
import type { FileInfo } from '@/api/interface/file';
import FileEditor from '@/components/FileEditor.vue';

const currentPath = ref('/');
const fileList = ref<FileInfo[]>([]);
const showHidden = ref(false);
const searchKeyword = ref('');
const selectedFiles = ref<FileInfo[]>([]);
const drives = ref<string[]>([]);
const isWindows = ref(false);

const createDirDialogVisible = ref(false);
const createDirForm = ref({ name: '' });

const createFileDialogVisible = ref(false);
const createFileForm = ref({ name: '' });

const uploadDialogVisible = ref(false);
const uploadFileList = ref<File[]>([]);
const uploadRef = ref();
const uploading = ref(false);
const overwrite = ref(false);

const compressDialogVisible = ref(false);
const compressForm = ref({ name: '', type: 'tar.gz' });
const compressing = ref(false);

const chmodDialogVisible = ref(false);
const chmodForm = ref({ mode: '755', sub: false });

const editorVisible = ref(false);
const editingFilePath = ref('');
const creating = ref(false);

// 面包屑导航路径段
const pathSegments = computed(() => {
  if (currentPath.value === '/') {
    return [];
  }
  const path = currentPath.value.replace(/\\/g, '/');
  const parts = path.split('/').filter(p => p);
  const segments: Array<{ name: string; url: string }> = [];
  let currentUrl = '';
  parts.forEach((part, index) => {
    currentUrl += '/' + part;
    segments.push({
      name: part,
      url: currentUrl,
    });
  });
  return segments;
});

// 将显示路径转换为后端路径
const toBackendPath = (displayPath: string): string => {
  if (!displayPath) return '';
  console.log('toBackendPath input:', displayPath);

  // Windows: /C/xxx -> C:\xxx, /C -> C:\
  // Linux: /xxx -> /xxx
  if (displayPath.match(/^\/[A-Z]\/.+/i)) {
    const result = displayPath.replace(/^\/([A-Z])\//, '$1:\\').replace(/\//g, '\\');
    console.log('toBackendPath output (with path):', result);
    return result;
  }
  if (displayPath.match(/^\/[A-Z]\/?$/i)) {
    // /C 或 /C/ -> C:\
    const drive = displayPath.match(/^\/([A-Z])/i)?.[1];
    if (drive) {
      const result = drive + ':\\';
      console.log('toBackendPath output (drive only):', result);
      return result;
    }
  }
  console.log('toBackendPath output (linux):', displayPath);
  return displayPath;
};

// 将后端路径转换为显示路径
const toDisplayPath = (backendPath: string): string => {
  if (!backendPath) return '/';
  // Windows: C:\xxx -> /C/xxx
  // Linux: /xxx -> /xxx
  if (backendPath.match(/^[A-Z]:\\/i)) {
    return '/' + backendPath.charAt(0) + backendPath.substring(2).replace(/\\/g, '/');
  }
  return backendPath;
};

const loadDrives = async () => {
  try {
    const response = await fileApi.getDrives();
    if (response.data.code === 200) {
      drives.value = response.data.data;
      // 判断是否为 Windows 系统
      isWindows.value = drives.value.length > 1 || (drives.value.length === 1 && drives.value[0].match(/^[A-Z]:\\/i));
      // 设置默认路径
      currentPath.value = '/';
      loadFileList();
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载盘符列表失败');
  }
};

const loadFileList = async () => {
  try {
    console.log('loadFileList called, currentPath:', currentPath.value, 'isWindows:', isWindows.value);

    // 如果当前路径为 /，显示盘符列表（Windows）或根目录内容（Linux）
    if (currentPath.value === '/') {
      if (isWindows.value) {
        // Windows: 显示盘符列表
        console.log('Showing drive list for Windows');
        const response = await fileApi.getDrives();
        if (response.data.code === 200) {
          // 将盘符转换为文件列表格式
          fileList.value = response.data.data.map(drive => ({
            path: toDisplayPath(drive),
            name: drive,
            size: 0,
            isDir: true,
            isSymlink: false,
            isHidden: false,
            linkPath: '',
            extension: '',
            mode: '',
            mimeType: '',
            modTime: new Date().toISOString(),
            user: '',
            group: '',
            uid: '',
            gid: '',
          }));
          console.log('Drive list loaded:', fileList.value);
        } else {
          ElMessage.error(response.data.message);
        }
      } else {
        // Linux: 显示根目录内容
        console.log('Showing root directory for Linux');
        const response = await fileApi.getFileList({
          path: '/',
          showHidden: showHidden.value,
          search: searchKeyword.value,
          page: 1,
          pageSize: 1000,
          sortBy: 'name',
          sortOrder: 'ascending',
        });
        if (response.data.code === 200) {
          fileList.value = response.data.data.items.map(item => ({
            ...item,
            path: toDisplayPath(item.path),
          }));
        } else {
          ElMessage.error(response.data.message);
        }
      }
      return;
    }

    // 否则加载文件列表
    const backendPath = toBackendPath(currentPath.value);
    console.log('Loading file list for path:', currentPath.value, '-> backendPath:', backendPath);
    const response = await fileApi.getFileList({
      path: backendPath,
      showHidden: showHidden.value,
      search: searchKeyword.value,
      page: 1,
      pageSize: 1000,
      sortBy: 'name',
      sortOrder: 'ascending',
    });
    if (response.data.code === 200) {
      fileList.value = response.data.data.items.map(item => ({
        ...item,
        path: toDisplayPath(item.path),
      }));
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    console.error('Error loading file list:', error);
    ElMessage.error(error.message || '加载文件列表失败');
  }
};

const refresh = () => {
  loadFileList();
};

const jump = (path: string) => {
  currentPath.value = path;
  loadFileList();
};

const toggleHidden = () => {
  showHidden.value = !showHidden.value;
};

const goBack = () => {
  if (currentPath.value === '/') {
    return;
  }

  const path = currentPath.value.replace(/\\/g, '/');
  const parts = path.split('/').filter(p => p);

  // 如果只有一个部分（如 C），返回到根目录
  if (parts.length === 1) {
    currentPath.value = '/';
  } else {
    parts.pop();
    currentPath.value = '/' + parts.join('/') + '/';
  }

  loadFileList();
};

const navigateToPath = () => {
  let path = currentPath.value.trim();
  if (!path) {
    currentPath.value = '/';
    loadFileList();
    return;
  }

  // 如果用户输入的是 Windows 原生路径，转换为显示路径
  if (path.match(/^[A-Z]:\\/i)) {
    path = toDisplayPath(path);
    currentPath.value = path;
  } else if (!path.startsWith('/')) {
    // 如果路径不以 / 开头，添加 /
    path = '/' + path;
    currentPath.value = path;
  }

  loadFileList();
};

const handleRowClick = (row: FileInfo) => {
  if (row.isDir) {
    // 确保路径以 / 结尾
    let path = row.path;
    if (!path.endsWith('/')) {
      path = path + '/';
    }
    currentPath.value = path;
    loadFileList();
  }
};

const handleRowDblClick = (row: FileInfo) => {
  if (row.isDir) {
    // 确保路径以 / 结尾
    let path = row.path;
    if (!path.endsWith('/')) {
      path = path + '/';
    }
    currentPath.value = path;
    loadFileList();
  } else {
    editFile(row);
  }
};

const handleSelectionChange = (selection: FileInfo[]) => {
  selectedFiles.value = selection;
};

const handleSearch = () => {
  loadFileList();
};

const showCreateDirDialog = () => {
  createDirForm.value.name = '';
  createDirDialogVisible.value = true;
};

const createDir = async () => {
  if (!createDirForm.value.name) {
    ElMessage.warning('请输入目录名');
    return;
  }
  try {
    const backendPath = toBackendPath(currentPath.value + '/' + createDirForm.value.name);
    const response = await fileApi.createFile({
      path: backendPath,
      isDir: true,
    });
    if (response.data.code === 200) {
      ElMessage.success('创建目录成功');
      createDirDialogVisible.value = false;
      loadFileList();
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '创建目录失败');
  }
};

const showCreateFileDialog = () => {
  createFileForm.value.name = '';
  createFileDialogVisible.value = true;
};

const createFile = async () => {
  if (!createFileForm.value.name) {
    ElMessage.warning('请输入文件名');
    return;
  }
  try {
    const backendPath = toBackendPath(currentPath.value + '/' + createFileForm.value.name);
    const response = await fileApi.createFile({
      path: backendPath,
      isDir: false,
    });
    if (response.data.code === 200) {
      ElMessage.success('创建文件成功');
      createFileDialogVisible.value = false;
      loadFileList();
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '创建文件失败');
  }
};

const showUploadDialog = () => {
  uploadFileList.value = [];
  uploadDialogVisible.value = true;
};

const handleFileChange = (file: any) => {
  uploadFileList.value = file.raw ? [file.raw] : [];
};

const uploadFiles = async () => {
  if (uploadFileList.value.length === 0) {
    ElMessage.warning('请选择要上传的文件');
    return;
  }
  uploading.value = true;
  try {
    const backendPath = toBackendPath(currentPath.value);
    for (const file of uploadFileList.value) {
      await fileApi.uploadFile(backendPath, file, overwrite.value);
    }
    ElMessage.success('上传成功');
    uploadDialogVisible.value = false;
    loadFileList();
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败');
  } finally {
    uploading.value = false;
  }
};

const editFile = (row: FileInfo) => {
  if (row.isDir) {
    return;
  }
  editingFilePath.value = row.path;
  editorVisible.value = true;
};

const handleEditorSave = async (content: string) => {
  try {
    const backendPath = toBackendPath(editingFilePath.value);
    const response = await fileApi.saveFileContent({
      path: backendPath,
      content,
    });
    if (response.data.code === 200) {
      ElMessage.success('保存成功');
      editorVisible.value = false;
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败');
  }
};

const downloadFile = async (row: FileInfo) => {
  try {
    const backendPath = toBackendPath(row.path);
    const response = await fileApi.downloadFile(backendPath);
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', row.name);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (error: any) {
    ElMessage.error(error.message || '下载失败');
  }
};

const deleteFile = async (row: FileInfo) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ${row.isDir ? '目录' : '文件'} "${row.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    });
    const backendPath = toBackendPath(row.path);
    const response = await fileApi.deleteFile({
      path: backendPath,
      force: true,
    });
    if (response.data.code === 200) {
      ElMessage.success('删除成功');
      loadFileList();
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败');
    }
  }
};

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};

const formatDate = (dateStr: string): string => {
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN');
};

const isCompressedFile = (row: FileInfo): boolean => {
  const ext = row.extension.toLowerCase();
  return ext === '.zip' || ext === '.gz' || ext === '.tar' || ext === '.tgz';
};

const showCompressDialog = () => {
  compressForm.value.name = 'archive';
  compressDialogVisible.value = true;
};

const compressFiles = async () => {
  if (!compressForm.value.name) {
    ElMessage.warning('请输入压缩文件名');
    return;
  }

  const files = selectedFiles.value.map(f => toBackendPath(f.path));
  const dstPath = toBackendPath(currentPath.value);
  compressing.value = true;

  try {
    const response = await fileApi.compressFiles({
      files,
      dst: dstPath,
      name: compressForm.value.name,
      type: compressForm.value.type as 'zip' | 'tar.gz',
    });
    if (response.data.code === 200) {
      ElMessage.success('压缩成功');
      compressDialogVisible.value = false;
      loadFileList();
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '压缩失败');
  } finally {
    compressing.value = false;
  }
};

const decompressFile = async (row: FileInfo) => {
  try {
    await ElMessageBox.confirm(`确定要解压文件 "${row.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    });

    const ext = row.extension.toLowerCase();
    let type: 'zip' | 'tar.gz' = 'tar.gz';
    if (ext === '.zip') {
      type = 'zip';
    }

    const backendPath = toBackendPath(row.path);
    const dstPath = toBackendPath(currentPath.value);
    const response = await fileApi.decompressFile({
      path: backendPath,
      dst: dstPath,
      type,
    });

    if (response.data.code === 200) {
      ElMessage.success('解压成功');
      loadFileList();
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '解压失败');
    }
  }
};

const showChmodDialog = () => {
  chmodForm.value.mode = '755';
  chmodForm.value.sub = false;
  chmodDialogVisible.value = true;
};

const chmodFiles = async () => {
  if (!chmodForm.value.mode || chmodForm.value.mode.length !== 3) {
    ElMessage.warning('请输入有效的权限模式（3位数字）');
    return;
  }

  const mode = parseInt('0' + chmodForm.value.mode, 8);

  try {
    for (const file of selectedFiles.value) {
      const backendPath = toBackendPath(file.path);
      const response = await fileApi.chmodFile({
        path: backendPath,
        mode,
        sub: chmodForm.value.sub,
      });
      if (response.data.code !== 200) {
        ElMessage.error(`修改 ${file.name} 权限失败: ${response.data.message}`);
        return;
      }
    }
    ElMessage.success('权限修改成功');
    chmodDialogVisible.value = false;
    loadFileList();
  } catch (error: any) {
    ElMessage.error(error.message || '权限修改失败');
  }
};

const batchDelete = async () => {
  if (selectedFiles.value.length === 0) {
    return;
  }
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedFiles.value.length} 个项目吗？`,
      '批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    );
    for (const file of selectedFiles.value) {
      const backendPath = toBackendPath(file.path);
      const response = await fileApi.deleteFile({
        path: backendPath,
        force: true,
      });
      if (response.data.code !== 200) {
        ElMessage.error(`删除 ${file.name} 失败: ${response.data.message}`);
        return;
      }
    }
    ElMessage.success('批量删除成功');
    loadFileList();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '批量删除失败');
    }
  }
};

const handleCreate = (command: string) => {
  if (command === 'dir') {
    showCreateDirDialog();
  } else if (command === 'file') {
    showCreateFileDialog();
  }
};

watch(showHidden, () => {
  loadFileList();
});

onBeforeMount(() => {
  // 在组件挂载前设置默认路径
  currentPath.value = '/';
});

onMounted(() => {
  loadDrives();
});
</script>

<style scoped>
.file-manager {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f5f7fa;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  flex-wrap: wrap;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.breadcrumb {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  overflow: hidden;
}

.breadcrumb-segment {
  display: inline-flex;
  align-items: center;
}

.breadcrumb-separator {
  margin: 0 8px;
  color: #909399;
}

.path-segment {
  font-size: 14px;
  color: #606266;
  transition: color 0.2s;
  margin-right: 8px;
}

.path-segment:hover {
  color: #409eff;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 16px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  flex-wrap: wrap;
}

.action-bar-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.action-bar-right {
  font-size: 14px;
  color: #909399;
}

.search-input {
  width: 280px;
}

.file-list {
  flex: 1;
  overflow: auto;
  padding: 16px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin: 16px;
}

.file-table {
  border-radius: 8px;
  overflow: hidden;
}

.file-table :deep(.el-table__header-wrapper) {
  background: #f5f7fa;
}

.file-table :deep(.el-table__header th) {
  background: #f5f7fa;
  font-weight: 600;
  color: #303133;
}

.file-table :deep(.el-table__row) {
  transition: background-color 0.2s;
}

.file-table :deep(.el-table__row:hover) {
  background-color: #f5f7fa !important;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-name {
  font-weight: 500;
  color: #303133;
}

.file-meta {
  color: #606266;
}

.upload-area {
  padding: 20px;
}

.upload-area :deep(.el-upload-dragger) {
  padding: 40px;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  transition: all 0.3s;
}

.upload-area :deep(.el-upload-dragger:hover) {
  border-color: #409eff;
  background-color: #f0f9ff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left {
    justify-content: center;
  }

  .breadcrumb {
    order: 3;
    width: 100%;
  }

  .toolbar-right {
    width: 100%;
  }

  .search-input {
    width: 100%;
  }

  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .action-bar-left {
    justify-content: center;
  }

  .action-bar-right {
    text-align: center;
  }

  .file-table :deep(.el-table__body-wrapper) {
    overflow-x: auto;
  }
}

@media (max-width: 576px) {
  .breadcrumb {
    display: none;
  }

  .toolbar-right {
    order: 2;
  }
}

/* 动画效果 */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.toolbar,
.action-bar {
  animation: fadeIn 0.3s ease-out;
}
</style>