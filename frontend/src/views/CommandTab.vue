<template>
  <div class="command-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-button type="primary" @click="openCreateDialog">
            <el-icon><Plus /></el-icon>
            新建快速命令
          </el-button>
          <el-button type="primary" plain @click="batchDelete" :disabled="selectedIds.length === 0">
            <el-icon><Delete /></el-icon>
            删除
          </el-button>
          <el-button type="primary" plain @click="exportCommands">
            <el-icon><Download /></el-icon>
            导出
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="commands"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="command" label="命令" min-width="300" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEditDialog(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button link type="danger" @click="deleteCommand(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadCommands"
        @current-change="loadCommands"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="resetForm"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入命令名称" />
        </el-form-item>
        <el-form-item label="命令" prop="command">
          <el-input
            v-model="form.command"
            type="textarea"
            :rows="4"
            placeholder="请输入命令内容"
          />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" placeholder="请输入描述（可选）" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort" :min="0" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  Plus,
  Delete,
  Edit,
  Download
} from '@element-plus/icons-vue';
import {
  createQuickCommand,
  updateQuickCommand,
  deleteQuickCommand,
  searchQuickCommands,
  getAllQuickCommands
} from '@/api/modules/quick_command';
import type {
  QuickCommand,
  QuickCommandCreate,
  QuickCommandUpdate
} from '@/api/interface/quick_command';

const loading = ref(false);
const commands = ref<QuickCommand[]>([]);
const selectedIds = ref<number[]>([]);
const dialogVisible = ref(false);
const dialogTitle = ref('');
const formRef = ref();

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
});

const form = reactive<QuickCommandCreate & { id?: number }>({
  name: '',
  command: '',
  description: '',
  sort: 0
});

const rules = {
  name: [{ required: true, message: '请输入命令名称', trigger: 'blur' }],
  command: [{ required: true, message: '请输入命令内容', trigger: 'blur' }]
};

const loadCommands = async () => {
  loading.value = true;
  try {
    const res = await searchQuickCommands({
      page: pagination.page,
      pageSize: pagination.pageSize
    });
    commands.value = res.data.items;
    pagination.total = res.data.total;
  } catch (error) {
    ElMessage.error('加载快速命令失败');
  } finally {
    loading.value = false;
  }
};

const handleSelectionChange = (selection: QuickCommand[]) => {
  selectedIds.value = selection.map(item => item.id);
};

const openCreateDialog = () => {
  dialogTitle.value = '新建快速命令';
  dialogVisible.value = true;
};

const openEditDialog = (row: QuickCommand) => {
  dialogTitle.value = '编辑快速命令';
  Object.assign(form, {
    id: row.id,
    name: row.name,
    command: row.command,
    description: row.description,
    sort: row.sort
  });
  dialogVisible.value = true;
};

const resetForm = () => {
  formRef.value?.resetFields();
  Object.assign(form, {
    name: '',
    command: '',
    description: '',
    sort: 0
  });
};

const submitForm = async () => {
  await formRef.value?.validate();
  try {
    if (form.id) {
      await updateQuickCommand(form as QuickCommandUpdate);
      ElMessage.success('更新成功');
    } else {
      await createQuickCommand(form);
      ElMessage.success('创建成功');
    }
    dialogVisible.value = false;
    loadCommands();
  } catch (error) {
    ElMessage.error('操作失败');
  }
};

const deleteCommand = async (row: QuickCommand) => {
  try {
    await ElMessageBox.confirm('确定要删除该命令吗？', '提示', {
      type: 'warning'
    });
    await deleteQuickCommand({ ids: [row.id] });
    ElMessage.success('删除成功');
    loadCommands();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
    }
  }
};

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 个命令吗？`, '提示', {
      type: 'warning'
    });
    await deleteQuickCommand({ ids: selectedIds.value });
    ElMessage.success('删除成功');
    loadCommands();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
    }
  }
};

const exportCommands = async () => {
  try {
    const res = await getAllQuickCommands();
    const dataStr = JSON.stringify(res.data, null, 2);
    const blob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'quick-commands.json';
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success('导出成功');
  } catch (error) {
    ElMessage.error('导出失败');
  }
};

const acceptParams = () => {
  loadCommands();
};

defineExpose({
  acceptParams
});

onMounted(() => {
  loadCommands();
});
</script>

<style scoped>
.command-page {
  padding: 20px;
}

.card-header {
  display: flex;
  gap: 10px;
}
</style>