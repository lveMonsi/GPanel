<template>
  <div class="port-rules">
    <div class="rules-toolbar">
      <div class="toolbar-left">
        <el-button type="primary" @click="handleAdd">
          <el-icon class="mr-1"><Plus /></el-icon>
          添加规则
        </el-button>
        <el-button type="danger" :disabled="selectedRows.length === 0" @click="handleBatchDelete">
          <el-icon class="mr-1"><Delete /></el-icon>
          批量删除
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchInfo"
          placeholder="搜索端口或IP"
          clearable
          style="width: 280px"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <div class="rules-content">
      <el-table
        :data="currentPageData"
        v-loading="loading"
        @selection-change="handleSelectionChange"
        stripe
        class="rules-table"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column type="index" label="#" width="60" />
        <el-table-column prop="protocol" label="协议" width="100" />
        <el-table-column prop="port" label="端口" min-width="120" />
        <el-table-column prop="address" label="允许IP" min-width="180">
          <template #default="{ row }">
            {{ formatAddress(row.address) }}
          </template>
        </el-table-column>
        <el-table-column prop="strategy" label="策略" width="100">
          <template #default="{ row }">
            <el-tag :type="row.strategy === 'accept' ? 'success' : 'danger'" size="small">
              {{ row.strategy === 'accept' ? '允许' : '拒绝' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" size="small" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="rules-footer">
        <div class="footer-left">共 {{ total }} 条规则</div>
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
          small
        />
      </div>
    </div>

    <!-- 添加/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑规则' : '添加规则'"
      width="500px"
    >
      <el-form :model="form" :rules="rulesValidator" label-width="80px" ref="formRef">
        <el-form-item label="协议" prop="protocol">
          <el-select v-model="form.protocol" placeholder="选择协议" style="width: 100%">
            <el-option label="TCP" value="tcp" />
            <el-option label="UDP" value="udp" />
            <el-option label="TCP/UDP" value="tcp/udp" />
          </el-select>
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input v-model="form.port" placeholder="例如: 80 或 80-100" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="允许IP" prop="address">
          <el-input v-model="form.address" placeholder="留空表示所有IP，或输入IP地址" />
          <div class="form-tip">
            <div>支持输入 IP 或 IP 段：172.16.10.11 或 172.16.0.0/24</div>
            <div>多个 IP 或 IP 段请用 "," 隔开：172.16.10.11,172.16.0.0/24</div>
          </div>
        </el-form-item>
        <el-form-item label="策略" prop="strategy">
          <el-radio-group v-model="form.strategy">
            <el-radio label="accept">允许</el-radio>
            <el-radio label="drop">拒绝</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Delete, Search } from '@element-plus/icons-vue';
import { searchRules, operatePortRule, updatePortRule } from '@/api/modules/firewall';
import type { FireInfo, PortRuleOperate } from '@/api/interface/firewall';

const emit = defineEmits(['refresh']);

const loading = ref(false);
const saving = ref(false);
const allRules = ref<FireInfo[]>([]);
const selectedRows = ref<FireInfo[]>([]);
const searchInfo = ref('');
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();

// 分页相关
const currentPage = ref(1);
const pageSize = ref(20);
const total = ref(0);

const currentPageData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return allRules.value.slice(start, end);
});

const form = ref<PortRuleOperate>({
  operation: 'add',
  protocol: 'tcp',
  port: '',
  strategy: 'accept',
  address: '',
});

const editRow = ref<FireInfo | null>(null);

const rulesValidator = {
  protocol: [{ required: true, message: '请选择协议', trigger: 'change' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  strategy: [{ required: true, message: '请选择策略', trigger: 'change' }],
};

const formatAddress = (address: string) => {
  if (!address || address === '0.0.0.0/0' || address === '::/0' || address === 'Anywhere') {
    return '所有 IP';
  }
  return address;
};

const loadRules = async () => {
  try {
    loading.value = true;
    const res = await searchRules({
      page: currentPage.value,
      pageSize: pageSize.value,
      type: 'port',
      info: searchInfo.value,
    });
    // 检查响应格式
    if (res && res.data) {
      if (Array.isArray(res.data)) {
        allRules.value = res.data;
        total.value = res.data.length;
      } else if (res.data.items && Array.isArray(res.data.items)) {
        allRules.value = res.data.items;
        total.value = res.data.total || res.data.items.length;
      } else {
        console.warn('Unexpected response format:', res);
        allRules.value = [];
        total.value = 0;
      }
    } else {
      allRules.value = [];
      total.value = 0;
    }
  } catch (error) {
    console.error('加载规则失败:', error);
    ElMessage.error('加载规则失败');
    allRules.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page: number) => {
  currentPage.value = page;
  loadRules();
};

const handleSizeChange = (size: number) => {
  pageSize.value = size;
  currentPage.value = 1;
  loadRules();
};

const handleSelectionChange = (rows: FireInfo[]) => {
  selectedRows.value = rows;
};

const handleSearch = () => {
  currentPage.value = 1;
  loadRules();
};

const handleAdd = () => {
  isEdit.value = false;
  form.value = {
    operation: 'add',
    protocol: 'tcp',
    port: '',
    strategy: 'accept',
    address: '',
  };
  editRow.value = null;
  dialogVisible.value = true;
};

const handleEdit = (row: FireInfo) => {
  isEdit.value = true;
  form.value = {
    operation: 'add',
    protocol: row.protocol,
    port: row.port,
    strategy: row.strategy,
    address: row.address === '0.0.0.0/0' || row.address === 'Anywhere' ? '' : row.address,
  };
  editRow.value = row;
  dialogVisible.value = true;
};

const handleSave = async () => {
  if (!formRef.value) return;
  await formRef.value.validate();

  try {
    saving.value = true;

    if (isEdit.value && editRow.value) {
      // 编辑：端口不能修改，IP可以修改
      const oldRule: PortRuleOperate = {
        operation: 'remove',
        protocol: editRow.value.protocol,
        port: editRow.value.port,
        strategy: editRow.value.strategy,
        address: editRow.value.address === '0.0.0.0/0' || editRow.value.address === 'Anywhere' ? '' : editRow.value.address,
      };

      const newRule: PortRuleOperate = {
        operation: 'add',
        protocol: editRow.value.protocol,
        port: editRow.value.port,
        strategy: form.value.strategy,
        address: form.value.address,
      };

      await updatePortRule({ oldRule, newRule });
    } else {
      // 添加
      await operatePortRule(form.value);
    }

    ElMessage.success('保存成功');
    dialogVisible.value = false;
    await loadRules();
    emit('refresh');
  } catch (error: any) {
    console.error('保存失败:', error);
    ElMessage.error(error.message || '保存失败');
  } finally {
    saving.value = false;
  }
};

const handleDelete = async (row: FireInfo) => {
  try {
    await ElMessageBox.confirm('确认删除此规则吗?', '确认删除', {
      type: 'warning',
    });

    await operatePortRule({
      operation: 'remove',
      protocol: row.protocol,
      port: row.port,
      strategy: row.strategy,
      address: row.address,
    });

    ElMessage.success('删除成功');
    await loadRules();
    emit('refresh');
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除失败:', error);
      ElMessage.error(error.message || '删除失败');
    }
  }
};

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确认删除选中的 ${selectedRows.value.length} 条规则吗?`, '确认删除', {
      type: 'warning',
    });

    for (const row of selectedRows.value) {
      await operatePortRule({
        operation: 'remove',
        protocol: row.protocol,
        port: row.port,
        strategy: row.strategy,
        address: row.address,
      });
    }

    ElMessage.success('批量删除成功');
    await loadRules();
    emit('refresh');
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error);
      ElMessage.error(error.message || '批量删除失败');
    }
  }
};

onMounted(() => {
  loadRules();
});
</script>

<style scoped>
.port-rules {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fafafa;
}

.rules-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.rules-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px;
  overflow: hidden;
}

.rules-table {
  flex: 1;
  border-radius: 8px;
  overflow: hidden;
}

.rules-table :deep(.el-table__header-wrapper) {
  background: #f5f7fa;
}

.rules-table :deep(.el-table__header th) {
  background: #f5f7fa;
  font-weight: 600;
  color: #303133;
  border-bottom: 1px solid #e4e7ed;
}

.rules-table :deep(.el-table__row) {
  transition: background-color 0.2s;
}

.rules-table :deep(.el-table__row:hover) {
  background-color: #f0f9ff !important;
}

.rules-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-top: 1px solid #f0f0f0;
  margin-top: 12px;
}

.footer-left {
  font-size: 13px;
  color: #909399;
}

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

.mr-1 {
  margin-right: 4px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .rules-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-left,
  .toolbar-right {
    width: 100%;
  }

  .toolbar-right :deep(.el-input) {
    width: 100% !important;
  }

  .rules-footer {
    flex-direction: column;
    gap: 12px;
  }
}
</style>