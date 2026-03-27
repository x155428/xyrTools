<template>
  <div class="container">
    <!-- 顶部按钮区 -->
    <div class="header">
      <div class="header-left">
        <el-button type="primary" @click="handleAdd">添加</el-button>
        <el-button type="danger" @click="handleDelete">删除</el-button>
        <el-button type="success" @click="refresh">刷新</el-button>
        <el-button type="primary" @click="handleImport">导入</el-button>
        <el-button type="primary" @click="handleExportAll">全量导出</el-button>
      </div>
      <div class="header-right">
        <el-input v-model="queryParams.userName" placeholder="搜索用户名" style="max-width: 200px; margin-right: 10px;"
          @keyup.enter="userNameSearch()">
          <template #append>
            <el-button :icon="Search" @click="userNameSearch()"></el-button>
          </template>
        </el-input>
        <el-button @click="isAdvQuery=!isAdvQuery"><span>高级搜索</span></el-button>
      </div>
    </div>

    <!-- 表格展示区 -->
    <div class="content table-container">
      <el-table :data="tableData" style="width: 100%" :height="tableHeight" header-fixed border
        @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55"></el-table-column>
        <el-table-column prop="appName" label="应用名" width="180" />
        <el-table-column prop="username" label="用户名" width="180" />
        <el-table-column prop="password" label="密码" width="120">
          <template #default="scope">
            <template v-if="scope.row.inputType === 'password'">
              <span>{{ scope.row.password}}</span>
            </template>
            <template v-else-if="scope.row.inputType === 'keyFile'">
              <el-button size="small" @click="handleDownloadKeyFile(scope.row.id)">下载密钥文件</el-button>
            </template>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="URL" />
        <el-table-column prop="notes" label="备注" />
        <el-table-column prop="tags" label="标签" />
        <el-table-column prop="strength" label="密码强度" />
        <el-table-column label="操作" width="120" align="center">
          <template #default="scope">
            <div style="display: flex; justify-content: center; align-items: center;">
              <el-button size="small" @click="handleEditItem(scope.row)">编辑</el-button>
              <el-button v-if="scope.row.inputType === 'password'" size="small"
                @click="copyToClipboard(scope.row.password)">复制</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 底部分页区 -->
    <div class="footer">
      <el-pagination :current-page="queryParams.page" :page-size="queryParams.pageSize" :total="total"
        layout="total,sizes,prev, pager, next,jumper" @size-change="handleSizeChange"
        @current-change="handlePageChange" />
    </div>

    <!-- 弹出框 -->
    <el-dialog v-model="isModalVisible" @close="resetForm" width="50%"
      style="min-width: 520px;max-width: 520px;border-radius: 20px;background: linear-gradient(to bottom right, #e3f2fd, #bbdefb);">
      <template #header>
        <span class="dialog-title">添加密码记录</span>
        <el-button size="small" @click="toggleAllEncryption">{{ allEncrypted ? '取消全加密' : '一键全加密' }}</el-button>
      </template>
      <el-form :model="form" ref="addform" label-width="120px" label-position="left">
        <el-form-item label="应用名">
          <div style="display: flex; align-items: center;">
            <el-input v-model="form.appName" />
            <el-checkbox v-model="form.isAppNameEncrypted" style="margin-left: 10px;">加密</el-checkbox>
          </div>
        </el-form-item>
        <el-form-item label="用户名">
          <div style="display: flex; align-items: center;">
            <el-input v-model="form.username" />
            <el-checkbox v-model="form.isUsernameEncrypted" style="margin-left: 10px;">加密</el-checkbox>
          </div>
        </el-form-item>

        <!-- 切换密码和密钥文件 -->
        <el-form-item label="密码或密钥文件">
          <div class="toggle-buttons">
            <el-segmented v-model="form.inputType"
              :options="[{ label: '密码', value: 'password' }, { label: '密钥文件', value: 'keyFile' }]" />
            <el-tooltip content="密钥或密码字段作为私密数据必须加密，不可选择！" placement="top">
              <span style="margin-left: 30px; color: red;">
                说明*
              </span>
            </el-tooltip>
          </div>
        </el-form-item>

        <!-- 密码输入框 -->
        <el-form-item v-if="form.inputType === 'password'" label="密码">
          <div style="display: flex; align-items: center; gap: 10px; width: 100%;">
            <el-input v-model="form.password" type="password" placeholder="输入密码" show-password style="flex: 1;" />
            <el-button type="primary" size="small" @click="showGeneratePasswordDialog">生成密码</el-button>
          </div>
        </el-form-item>

        <!-- 密码生成配置弹窗 -->
        <el-dialog v-model="isGeneratePasswordDialogVisible" title="生成密码" width="400px">
          <div style="padding: 20px;">
            <div style="margin-bottom: 20px;">
              <label style="display: block; margin-bottom: 10px;">密码长度: {{ passwordLength }}</label>
              <el-slider v-model="passwordLength" :min="6" :max="32" show-stops />
            </div>
            <div style="margin-bottom: 20px;">
              <label style="display: block; margin-bottom: 10px;">字符集选择:</label>
              <div style="display: flex; flex-direction: column; gap: 10px;">
                <el-checkbox v-model="includeUppercase">包含大写字母 (A-Z)</el-checkbox>
                <el-checkbox v-model="includeLowercase">包含小写字母 (a-z)</el-checkbox>
                <el-checkbox v-model="includeNumbers">包含数字 (0-9)</el-checkbox>
                <el-checkbox v-model="includeSpecialChars">包含特殊字符 (!@#$%^&*)</el-checkbox>
              </div>
            </div>
            <div style="display: flex; justify-content: space-between; margin-top: 30px;">
              <el-button @click="isGeneratePasswordDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="generatePassword">生成密码</el-button>
            </div>
          </div>
        </el-dialog>

        <!-- 密钥文件上传框 -->
        <el-form-item v-if="form.inputType === 'keyFile'" label="浏览">
          <el-upload action="''" :limit="1" :before-upload="handleBeforeUpload" v-model="form.keyFile"
            :show-file-list="false">
            <el-button type="primary">上传密钥文件</el-button>
          </el-upload>
        </el-form-item>

        <el-form-item label="URL">
          <div style="display: flex; align-items: center;">
            <el-input v-model="form.url" />
            <el-checkbox v-model="form.isUrlEncrypted" style="margin-left: 10px;">加密</el-checkbox>
          </div>
        </el-form-item>

        <el-form-item label="备注">
          <div style="display: flex; align-items: center;">
            <el-mention v-model="form.notes" type="textarea" />
            <el-checkbox v-model="form.isNotesEncrypted" style="margin-left: 10px;">加密</el-checkbox>
          </div>
        </el-form-item>

        <el-form-item label="标签">
          <div style="display: flex; align-items: center;">
            <el-select v-model="form.tags" placeholder="选择标签" filterable allow-create style="width:150px" clearable>
              <el-option label="浏览器" value="浏览器" />
              <el-option label="银行卡" value="银行卡" />
              <el-option label="QQ" value="QQ" />
              <el-option label="微信" value="微信" />
              <el-option label="手机号" value="手机号" />
              <el-option label="身份证" value="身份证" />
              <el-option label="生日" value="生日" />
              <el-option label="密钥" value="密钥" />
              <el-option label="证书" value="证书" />
            </el-select>
            <el-checkbox v-model="form.isTagsEncrypted" style="margin-left: 10px;">加密</el-checkbox>
          </div>
        </el-form-item>

        <span class="dialog-title">本条记录加密配置</span>
        <!-- 选择加密算法 -->
        <el-form-item label="选择加密算法">
          <el-select v-model="form.choseCrypto" @change="handleCryptoChange" placeholder="选择加密算法" style="width:380px;"
            clearable>
            <el-option label="AES-GCM" value="AES-GCM" />
            <el-option label="ChaCha20-Poly1305（待实现）" value="ChaCha20-Poly1305" />
          </el-select>
        </el-form-item>
        <el-checkbox v-if="isSymmetricEncryption&&form.choseCrypto=='AES-GCM'" @change="handleCustomKeyChange"
          v-model="keySet" style="margin-left: 10px;">使用自己的加密密钥</el-checkbox>
        <!-- 密钥输入框或文件上传 -->
        <el-form-item label="输入/上传密钥">
          <div style="display: flex; gap: 10px; align-items: center;">
            <!-- 如果选择对称加密（如 AES、ChaCha20），显示输入框 -->
            <el-input v-if="isSymmetricEncryption" v-model="form.key" :disabled="keySet"
              placeholder="输入密钥HEX，例如：4c52f7e17bc804786377f7f27690deacc6e7453b45e11792" style="width: 370px;" />
            <!-- 如果选择非对称加密（如 RSA、ECC），显示上传组件 -->
            <!-- <el-upload v-else action="" :limit="1" :before-upload="handleKeyFileUpload" :show-file-list="false"
              style="flex: 1;">
              <el-button>上传密钥文件</el-button>
            </el-upload> -->
          </div>
        </el-form-item>

        <!--提交-->
        <el-form-item>
          <el-button type="primary" @click="handleSave">保存</el-button>
          <el-button @click="handleCancel">取消</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <!--高级查询表单-->
    <el-dialog class="query-form" v-model="isAdvQuery" title="高级搜索" width="600px">
      <el-form :model="queryParams" label-width="80px" @submit.native.prevent="fetchData">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="应用名">
              <el-input v-model="queryParams.appName" placeholder="输入应用名" clearable />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="用户名">
              <el-input v-model="queryParams.userName" placeholder="输入用户名" clearable />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="URL">
              <el-input v-model="queryParams.url" placeholder="输入URL" clearable />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="标签">
              <el-input v-model="queryParams.tags" placeholder="多个用逗号分隔" clearable />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item style="text-align: center; margin-top: 20px;">
          <el-button type="primary" @click="fetchData" style="width: 120px;" native-type="submit">查询</el-button>
          <el-button @click="resetAdvQuery" style="width: 120px; margin-left: 10px;">重置</el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <!--导入弹出框-->
    <el-dialog class="import-form" v-model="isImport" title="导入记录" width="80%" :max-width="400"
      :style="{borderRadius: '12px', boxShadow: '0 4px 20px rgba(0, 0, 0, 0.15)'}">

      <el-form :model="importForm" label-width="80px">
        <el-form-item label="Excel文件">
          <el-upload class="importUpload" v-model:file-list="importFileList" accept=".xlsx,.xls,.csv" :limit="1"
            :before-upload="importBeforeUpload" :on-exceed="handleImportExceed" :http-request="()=>{}">
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">
                仅支持上传 .xlsx 或 .xls 或 .csv 格式的文件
              </div>
            </template>
          </el-upload>
        </el-form-item>

        <el-form-item label="加密算法">
          <el-select v-model="importForm.choseCrypto" placeholder="请选择加密算法" style="width: 100%;"
            @change="handleImportCryptoChange">
            <el-option label="AES-GCM" value="AES-GCM" />
            <el-option label="ChaCha20-Poly1305（待实现）" value="ChaCha20-Poly1305" />
          </el-select>
        </el-form-item>

        <el-form-item label="密钥设置">
          <el-radio-group v-model="importForm.keySource" @change="handleImportKeySourceChange" style="width: 100%;">
            <el-radio label="custom">设置密钥</el-radio>
            <el-radio label="file">使用文件中配置密钥</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="密钥" v-if="importForm.keySource === 'custom'">
          <div style="display: flex; align-items: center;">
            <el-input v-model="importForm.key" type="text" placeholder="请输入加密密钥" style="flex: 1; margin-right: 10px;" />
            <el-checkbox v-model="importForm.autoFillMasterKey"
              @change="handleImportAutoFillKeyChange">自动填充为主密钥</el-checkbox>
          </div>
        </el-form-item>

        <el-form-item label="加密设置">
          <el-checkbox v-model="isAllImportEncrypted">全选</el-checkbox>
        </el-form-item>

        <el-form-item label="应用名">
          <el-checkbox v-model="importForm.isAppNameEncrypted">加密</el-checkbox>
        </el-form-item>

        <el-form-item label="用户名">
          <el-checkbox v-model="importForm.isUsernameEncrypted">加密</el-checkbox>
        </el-form-item>

        <el-form-item label="URL">
          <el-checkbox v-model="importForm.isUrlEncrypted">加密</el-checkbox>
        </el-form-item>

        <el-form-item label="备注">
          <el-checkbox v-model="importForm.isNotesEncrypted">加密</el-checkbox>
        </el-form-item>

        <el-form-item label="标签">
          <el-checkbox v-model="importForm.isTagsEncrypted">加密</el-checkbox>
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="isImport = false">取消</el-button>
          <el-button type="primary" @click="handleImportFileUpload()">导入</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
  import { ref, reactive, onBeforeMount, onMounted, onUnmounted, computed, watch } from "vue";
  import {
    ElButton,
    ElTable,
    ElTableColumn,
    ElPagination,
    ElDialog,
    ElForm,
    ElInput,
    ElSelect,
    ElOption,
    ElUpload,
    ElRow,
    ElCol,
    ElFormItem,
    ElMessage,
    ElMessageBox,
    ElSlider,
    ElCheckbox
  } from "element-plus";
  import { Search } from '@element-plus/icons-vue'
  import { aesGcmEncrypt, aesGcmDecrypt } from '@/js/aesCryption.js';
  import { checkSession } from '@/js/checkSession.js';
  import { useRouter } from 'vue-router';
  import http from '@/js/http.js'
  import tools from '@/js/tools.js';
  import keyExchange from '@/js/keyExchange.js';
  import isAxiosError from 'axios';
  import * as XLSX from 'xlsx';
  import { getServerAddress } from '@/js/getServerAddress.js';

  // 定义响应式数据，展示页设置和表格数据
  const tableData = ref([]);
  const total = ref(0); // 总数据量
  // 查询条件状态
  const queryParams = reactive({
    appName: '',
    userName: '',
    url: '',
    tags: '', // 支持逗号分隔的多个标签
    page: 1,
    pageSize: 20
  })

  // 窗口大小变化时重新计算表格高度的函数
  const handleResize = () => {
    // 调整空间分配，确保内容完全显示且布局紧凑
    const footerHeight = 80; // 底部空间
    const headerHeight = 180; // 头部空间
    const additionalSpace = 20; // 少量额外安全空间
    tableHeight.value = `calc(100vh - ${footerHeight + headerHeight + additionalSpace}px)`;
  };

  // 定义加密密钥
  const userMasterAesKey = ref('');
  // 存储密钥文件的二进制内容和元信息
  const fileInfo = ref(null);

  // 存储选中的行
  const selectedRows = ref([]);

  // 是否为对称加密（AES 或 ChaCha20）
  const isSymmetricEncryption = computed(() => {
    return form.choseCrypto === "AES-GCM" || form.choseCrypto === "ChaCha20";
  });

  // 是否全加密
  const allEncrypted = ref(false);

  // 高级查询组件显示
  const isAdvQuery = ref(false);
  // 导入对话框显示
  const isImport = ref(false);

  const router = useRouter();
  // 部分组件控制，表格高度 
  const tableHeight = ref("0");
  const isModalVisible = ref(false); // 控制弹出框的显示
  const keySet = ref(false);

  // 密码生成相关配置
  const isGeneratePasswordDialogVisible = ref(false);
  const passwordLength = ref(16);
  const includeUppercase = ref(true);
  const includeLowercase = ref(true);
  const includeNumbers = ref(true);
  const includeSpecialChars = ref(true);

  // 导入相关配置
  const importForm = reactive({
    file: null,
    choseCrypto: '',       // 选择加密算法
    keySource: 'custom',   // 密钥来源: custom(自定义), file(文件)
    autoFillMasterKey: false, // 是否自动填充为主密钥
    key: '',               // 密钥
    isAppNameEncrypted: false,
    isUsernameEncrypted: false,
    isUrlEncrypted: false,
    isNotesEncrypted: false,
    isTagsEncrypted: false
  });
  const importFileList = ref([]);

  // 计算属性：检查是否所有加密字段都已选中
  const isAllImportEncrypted = computed({
    get: () => {
      return importForm.isAppNameEncrypted &&
        importForm.isUsernameEncrypted &&
        importForm.isUrlEncrypted &&
        importForm.isNotesEncrypted &&
        importForm.isTagsEncrypted;
    },
    set: (value) => {
      importForm.isAppNameEncrypted = value;
      importForm.isUsernameEncrypted = value;
      importForm.isUrlEncrypted = value;
      importForm.isNotesEncrypted = value;
      importForm.isTagsEncrypted = value;
    }
  });

  // 处理Excel导入
  const handleImport = async () => {

    // 检查AES密钥是否配置
    if (!userMasterAesKey.value) {
      ElMessage({
        message: '密钥未配置，功能不可用！',
        type: 'error',
        grouping: true,
      });
      return;
    }
    // 打开导入框
    isImport.value = true;
  };

  // 初始化模板
  const initialForm = {
    appName: '',               // 应用名
    isAppNameEncrypted: false, // 应用名是否加密
    username: '',              // 用户名
    isUsernameEncrypted: false, // 用户名是否加密
    inputType: 'password',     // 密码或密钥文件类型（'password' 或 'keyFile'）
    password: '',              // 密码
    keyFile: null,               // 密钥文件
    url: '',                   // URL
    isUrlEncrypted: false,     // URL是否加密
    notes: '',                 // 备注
    isNotesEncrypted: false,   // 备注是否加密
    tags: '',                  // 标签
    isTagsEncrypted: false,    // 标签是否加密
    choseCrypto: '',       // 选择加密算法
    key: '',                  // 密钥
  }
  // 定义表单数据
  const form = reactive({ ...initialForm })

  // 一键全加密逻辑
  const toggleAllEncryption = () => {
    allEncrypted.value = !allEncrypted.value;
    form.isAppNameEncrypted = allEncrypted.value;
    form.isUsernameEncrypted = allEncrypted.value;
    form.isUrlEncrypted = allEncrypted.value;
    form.isNotesEncrypted = allEncrypted.value;
    form.isTagsEncrypted = allEncrypted.value;
  };



  const handleCancel = () => {
    isModalVisible.value = false;
    allEncrypted.value = false;
    resetForm();
  };

  // 点击添加按钮,显示弹出框
  const handleAdd = () => {
    resetForm();
    isModalVisible.value = true;
  };

  //点击删除按钮
  const handleDelete = async () => {
    if (selectedRows.value.length === 0) {
      ElMessage({
        message: '请选择要删除的记录',
        type: 'warning',
        grouping: true,
      })
      return;
    }
    // 弹窗二次确认
    const confirmResult = await ElMessageBox.confirm('确认删除选中的记录吗？', '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    if (!confirmResult) {
      return;
    }

    const ids = selectedRows.value.map(row => row.id);

    const response = await http.post("/deleteRecords", { ids });
    if (response.status >= 200 && response.status < 300) {
      ElMessage({
        message: '删除成功',
        type: 'success',
        grouping: true,
      })
      fetchData(queryParams.page, queryParams.pageSize);
    } else {
      ElMessage({
        message: `删除失败: ${response.data.message}`,
        type: 'error',
        grouping: true,
      })
    }
  };

  // 刷新
  const refresh = () => {
    // 刷新查询条件
    queryParams.appName = '';
    queryParams.userName = '';
    queryParams.url = '';
    queryParams.tags = '';
    // 刷新表格数据
    fetchData(queryParams.page, queryParams.pageSize);
    ElMessage({
      message: '刷新成功',
      type: 'success',
      grouping: true,
    })

  };

  //操作编辑项
  const handleEditItem = (row) => {
    console.log("编辑项:", row);
    // 打开编辑弹窗
    isModalVisible.value = true;
    // 填充表单数据
    Object.assign(form, row);

  };

  // 显示密码生成弹窗
  const showGeneratePasswordDialog = () => {
    isGeneratePasswordDialogVisible.value = true;
  };

  // 生成密码函数
  const generatePassword = () => {
    // 验证至少选择了一个字符集
    if (!includeUppercase.value && !includeLowercase.value && !includeNumbers.value && !includeSpecialChars.value) {
      ElMessage({
        message: '请至少选择一个字符集',
        type: 'warning',
        grouping: true,
      })
      return;
    }

    // 定义字符集
    let charset = '';
    const uppercaseChars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    const lowercaseChars = 'abcdefghijklmnopqrstuvwxyz';
    const numberChars = '0123456789';
    const specialChars = '!@#$%^&*()_+-=[]{}|;:,.<>?';

    // 根据选择添加字符集
    if (includeUppercase.value) charset += uppercaseChars;
    if (includeLowercase.value) charset += lowercaseChars;
    if (includeNumbers.value) charset += numberChars;
    if (includeSpecialChars.value) charset += specialChars;

    // 生成密码
    let password = '';
    for (let i = 0; i < passwordLength.value; i++) {
      const randomIndex = Math.floor(Math.random() * charset.length);
      password += charset[randomIndex];
    }

    // 检查密码是否包含所选字符集的至少一个字符
    let isValid = true;
    if (includeUppercase.value && !/[A-Z]/.test(password)) isValid = false;
    if (includeLowercase.value && !/[a-z]/.test(password)) isValid = false;
    if (includeNumbers.value && !/[0-9]/.test(password)) isValid = false;
    if (includeSpecialChars.value && !/[^A-Za-z0-9]/.test(password)) isValid = false;

    // 如果不满足条件，重新生成
    if (!isValid) {
      generatePassword();
      return;
    }

    // 设置密码并关闭弹窗
    form.password = password;
    isGeneratePasswordDialogVisible.value = false;
    ElMessage({
      message: '密码生成成功',
      type: 'success',
      grouping: true,
    })
  };

  //操作复制项
  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      ElMessage({
        message: '已复制到剪贴板',
        type: 'success',
        grouping: true,
      })
    }
    catch (error) {
      ElMessage({
        message: '复制失败',
        type: 'error',
        grouping: true,
      })
    }
  };

  //分页页码点击
  const handlePageChange = (page) => {
    queryParams.page = page;
    fetchData(page, queryParams.pageSize); // 根据页码请求数据
  };

  // 分页大小变化
  const handleSizeChange = (size) => {
    queryParams.pageSize = size; // 更新分页大小
    fetchData(queryParams.page, queryParams.pageSize); // 重新加载数据
  };

  // 密钥状态变化时的逻辑
  const handleCustomKeyChange = (checked) => {
    if (checked) {
      // 选中时设置默认密钥
      form.key = userMasterAesKey.value;
    } else {
      // 取消选中时清空密钥
      form.key = '';
    }
  };

  // 加密算法选择变化时的逻辑
  const handleCryptoChange = (value) => {
    // 如果选择了算法，且算法不是AES-GCM或ChaCha20，弹出提示
    if (value && value !== 'AES-GCM' && value !== 'ChaCha20-Poly1305') {
      alert("该加密算法暂未实现！"); // 设置默认密钥
      value = '';
      // 取消选择的算法
      form.choseCrypto = '';
    }
  };

  // 导入时加密算法选择变化时的逻辑
  const handleImportCryptoChange = (value) => {
    // 如果选择了算法，且算法不是AES-GCM或ChaCha20，弹出提示
    if (value && value !== 'AES-GCM' && value !== 'ChaCha20-Poly1305') {
      alert("该加密算法暂未实现！"); // 设置默认密钥
      value = '';
      // 取消选择的算法
      importForm.choseCrypto = '';
    }
  };

  // 导入时密钥来源选择变化时的逻辑
  const handleImportKeySourceChange = () => {
    // 重置自动填充选项
    importForm.autoFillMasterKey = false;

    if (importForm.keySource === 'custom') {
      // 使用自定义密钥时，保留当前密钥
      // 可以选择是否自动填充，由用户决定
    } else if (importForm.keySource === 'file') {
      // 使用文件密钥时，清空输入框和自动填充选项
      importForm.key = '';
    }
  };

  // 导入时自动填充主密钥选择变化时的逻辑
  const handleImportAutoFillKeyChange = (checked) => {
    if (checked) {
      // 选中时填充用户主密钥
      importForm.key = userMasterAesKey.value;
      // 确保密钥来源为自定义
      importForm.keySource = 'custom';
    } else {
      // 取消选中时清空密钥
      importForm.key = '';
    }
  };

  const resetForm = () => {
    // 清除所有已有属性
    Object.keys(form).forEach(key => delete form[key])
    // 赋值初始模板，深拷贝避免引用污染
    Object.assign(form, JSON.parse(JSON.stringify(initialForm)))
  };

  // 初始化表格的高度
  onMounted(() => {
    handleResize(); // 初始化时计算一次

    window.addEventListener('resize', handleResize);

    fetchData(queryParams.page, queryParams.pageSize);

    //aes密钥，测试使用
    //const userMasterAesKey = "c81ca881b2340b509b23258066e67056f0e07ff54ecfb4dedf5b20c0ae5a70a1"
    // 从本地加载密钥
    userMasterAesKey.value = localStorage.getItem('aesKey');
    if (userMasterAesKey.value === null) {
      console.log("本地未找到 AES 密钥。");
    } else {
      console.log("从本地加载的 AES 密钥:", userMasterAesKey.value);
      const aesKeyArrayBuffer = tools.hexToArrayBuffer(userMasterAesKey.value);
      const aesKeyUint8Array = new Uint8Array(aesKeyArrayBuffer);
      const aesKeyBase64 = btoa(String.fromCharCode(...aesKeyUint8Array));
      console.log("Base64 编码后的 AES 密钥:", aesKeyBase64);
    }
  });

  // 处理表格选中行变化
  const handleSelectionChange = (selection) => {
    selectedRows.value = selection;
  };


  // 数据请求函数
  const fetchData = async () => {
    const params = {
      page: queryParams.page,
      pageSize: queryParams.pageSize,
    };
    // 按需添加条件
    if (queryParams.appName) params.appName = queryParams.appName;
    if (queryParams.userName) params.userName = queryParams.userName;
    if (queryParams.url) params.url = queryParams.url;
    if (queryParams.tags) {
      // 将 tags 数组转换为逗号分隔的字符串
      params.tags = queryParams.tags.split(',').filter(t => t.trim()).join(',');
    }
    try {
      // 从后端 API 获取分页数据
      const res = await http.get("/queryData", { params: params });
      const data = res.data.data;
      // 检查 data 是否为空
      if (!data || !data.data || !Array.isArray(data.data)) {
        ElMessage({
          message: '未获取到有效数据',
          type: 'warning',
          grouping: true,
        })
        isAdvQuery.value = false;
        tableData.value = [];
        total.value = 0;
        return;
      }

      // 遍历所有数据记录
      for (const [key, record] of Object.entries(data.data)) {
        // 处理 strength 字段
        if (record.strength) {
          record.strength = record.strength.valid ? '' : record.strength.String || '';
        }

        if (userMasterAesKey.value) {
          try {
            let algo = null;
            let datakey = null;
            if (record.choseCrypto) {
              // 解密获取加密算法
              const decryptedAlgoBuf = await decryptPassword(record.choseCrypto, userMasterAesKey.value);
              algo = new TextDecoder().decode(decryptedAlgoBuf);
              // 将解密后的值赋值回record对象
              record.choseCrypto = algo;
            }
            if (record.key) {
              // 解密获取数据密钥
              const decryptedDataKeyBuf = await decryptPassword(record.key, userMasterAesKey.value);
              datakey = new TextDecoder().decode(decryptedDataKeyBuf);
              // 将解密后的值赋值回record对象
              record.key = datakey;
            }

            // 若提供了 AES 密钥，则进行解密
            if (algo && datakey) {
              for (const [fieldKey, fieldValue] of Object.entries(record)) {
                if (fieldKey === 'choseCrypto' || fieldKey === 'key') {
                  continue;
                }
                if (typeof fieldValue === 'string' && fieldValue.startsWith('{') && fieldValue.endsWith('}')) {
                  const passwdArray = await decryptPassword(fieldValue, datakey);
                  // 解析数据
                  record[fieldKey] = new TextDecoder().decode(passwdArray);
                }
              }
            }
          } catch (error) {
            console.log(`解析出错:`, record.dataKey);
          }
        }


      }
      isAdvQuery.value = false;
      // 赋值给表格绑定数据
      tableData.value = data.data;
      total.value = data.total;
    } catch (error) {
      ElMessage({
        message: '请求异常',
        type: 'error',
        grouping: true,
      })
      return;
    }
  };


  //需要存储的密钥文件上传前处理文件，读取文件信息并记录
  const handleBeforeUpload = async (file) => {
    // 获取文件元信息
    const { name, size, type } = file;
    // 读取文件内容并计算哈希
    const reader = new FileReader();
    const fileHash = await calculateFileHash(file);
    console.log(size, type, name, fileHash, file.type);
    // 读取文件内容为二进制
    reader.onload = (event) => {
      const binaryData = event.target.result; // ArrayBuffer 类型的二进制数据
      console.log("密钥文件读取成功，二进制内容：", binaryData);
      // 将文件的元信息和二进制内容保存到变量中
      fileInfo.value = {
        name: file.name,         // 文件名
        size: file.size,         // 文件大小
        type: file.type || 'unknown',  // 文件类型（如果为空，则标记为 'unknown'）
        content: binaryData,     // 文件的二进制数据（ArrayBuffer）
        fileHash: fileHash,  //源文件hash
      };
      console.log('文件信息:', fileInfo.value);
    }
    // 读取失败时触发的回调
    reader.onerror = (event) => {
      console.error("文件读取失败:", event.target.error); // 错误信息
      alert("文件读取失败，请检查文件格式或文件内容！");
    };

    // 读取文件进度时触发的回调，调试用
    reader.onprogress = (event) => {
      if (event.lengthComputable) {
        const percent = (event.loaded / event.total) * 100;
        console.log(`文件读取进度: ${percent.toFixed(2)}%`);
      }
    };
    await reader.readAsArrayBuffer(file);
    return false;
  };

  const beforeFileUpload = (file) => {
    const isValid = ['application/x-pem-file', 'application/x-zip-compressed'].includes(file.type);
    if (!isValid) {
      this.$message.error('请上传有效的密钥文件（.key或.zip）。');
    }
    return isValid;
  };

  // 计算文件的哈希值（SHA-256）
  const calculateFileHash = (file) => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = async (event) => {
        const buffer = event.target.result;
        const hashBuffer = await crypto.subtle.digest('SHA-256', buffer);
        const hashArray = Array.from(new Uint8Array(hashBuffer));
        const hashHex = hashArray.map(byte => byte.toString(16).padStart(2, '0')).join('');
        resolve(hashHex);
      };
      reader.onerror = reject;
      reader.readAsArrayBuffer(file);
    });
  };


  // 保存新增数据
  const handleSave = async () => {
    // 检查AES密钥是否为空，没密钥不给用
    if (!userMasterAesKey.value) {
      ElMessage({
        message: "密钥未配置，功能不可用！",
        type: 'error',
        grouping: true,
      })
      return;
    }

    // 创建变量保存数据
    var saveData = { ...form };
    console.log('表单数据:', saveData);
    if (!saveData.appName && !saveData.username && !saveData.password && !saveData.keyFile && !saveData.url && !saveData.notes && (!saveData.tags || saveData.tags.length === 0)) {
      alert("数据为空！请填写数据！");
      return;
    }
    // 该记录加密密钥是否配置
    if (!saveData.key) {
      ElMessage({
        message: "该记录加密密钥未配置，功能不可用！",
        type: 'error',
        grouping: true,
      })
      return;
    }
    //获取密码类型，如果为文件类型，合并数据
    if (saveData.inputType === "keyFile") {
      saveData.password = null;
      if (fileInfo.value && fileInfo.value.content) {
        saveData.keyFile = {
          name: fileInfo.value.name,
          size: fileInfo.value.size,
          type: fileInfo.value.type,
          content: fileInfo.value.content,
          fileHash: fileInfo.value.fileHash,
        };
      }
    }

    // 处理最终的表单数据，发送请求、更新状态等
    //console.log('合并后未加密表单数据:', saveData);
    //1、遍历数据，查找加密字段
    for (const key of Object.keys(saveData)) {
      if (key.startsWith('is') && key.endsWith('Encrypted')) {
        const fieldName = key.replace(/^is(.*)Encrypted$/, '$1'); // 提取字段名
        const fieldNameWithLowerFirstLetter = fieldName.charAt(0).toLowerCase() + fieldName.slice(1); // 首字母小写
        const isEncrypted = saveData[key]; // 获取加密标识（true 或 false）
        const fieldValue = saveData[fieldNameWithLowerFirstLetter]; // 获取实际字段值

        // 如果加密标识为 true
        if (isEncrypted) {
          // 如果字段值不为空，执行加密，目前用的AES-GCM，后期可调整为对应加密算法
          if (fieldValue && fieldValue.trim() !== '') {
            const dataEncryptedResult = await aesGcmEncrypt(fieldValue, saveData.key);
            saveData[fieldNameWithLowerFirstLetter] = {
              iv: dataEncryptedResult.iv,
              data: dataEncryptedResult.data,
            };

          }
        }
      }

      // 处理嵌套的 keyFile 对象（可以根据需求加密某些属性）
      if (key === 'keyFile' && saveData[key] && saveData[key].content) {
        const file = saveData[key];
        if (file.content) {
          const fileContentEncryptedResult = await aesGcmEncrypt(file.content, saveData.key);
          file.content = {
            iv: fileContentEncryptedResult.iv,
            data: fileContentEncryptedResult.data,
          };
        }
      }
    }
    if (saveData.password != null) {
      // 密码加密
      // 如果有选择加密方式，则使用选择的加密方式和密钥加密密码
      // 否则使用默认的 AES-GCM 加密方式和用户主密钥加密
      const pwdEncrypted = await aesGcmEncrypt(saveData.password, saveData.key);
      saveData.password = {
        iv: pwdEncrypted.iv,
        data: pwdEncrypted.data,
      }
    }

    //主密钥加密加密算法和加密密钥
    const choseEncrypted = await aesGcmEncrypt(saveData.choseCrypto, userMasterAesKey.value);
    saveData.choseCrypto = {
      iv: choseEncrypted.iv,
      data: choseEncrypted.data,
    }
    const keyEncrypted = await aesGcmEncrypt(saveData.key, userMasterAesKey.value);
    saveData.key = {
      iv: keyEncrypted.iv,
      data: keyEncrypted.data,
    }

    // 2. JSON 序列化
    const jsonData = JSON.stringify(saveData);
    const tmpAesKey = await keyExchange.getAesKey();
    // 3. AES 加密 JSON 数据
    const encryptedData = await aesGcmEncrypt(jsonData, tmpAesKey);
    // 4. 发送 AES 加密后的数据到后端 API
    try {
      const response = await http.post("/saveSecret", {
        iv: encryptedData.iv,        // Base64 编码的 IV
        encryptedData: encryptedData.data  // Base64 编码的加密数据
      }, {
        headers: {
          'Content-Type': 'application/json'
        }
      });
      resetForm();
      isModalVisible.value = false;
      allEncrypted.value = false;
      keySet.value = false;
      fetchData(queryParams.page, queryParams.pageSize);
      ElMessage({
        message: "数据保存成功",
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      isModalVisible.value = false;
      allEncrypted.value = false;
      keySet.value = false;
      resetForm();
      fetchData(queryParams.page, queryParams.pageSize);
    }
  };


  //检查session是否过期 
  onBeforeMount(async () => {
    const isLogin = await checkSession();
    if (!isLogin) {
      router.push('/');
    }
  })

  // 组件卸载时移除事件监听器
  onUnmounted(() => {
    window.removeEventListener('resize', handleResize);
  })

  // 解密数据记录函数
  // json字符串(iv  data) 返回明文
  // key参数可选，如果不提供则使用默认的aesKeyTest.value
  const decryptPassword = async (encryptedPassword, key = null) => {
    // 确定使用的密钥
    const usedKey = key || userMasterAesKey.value;
    // 如果密钥为空，则返回
    if (!usedKey) {
      ElMessage({
        message: '无主密钥无法解密',
        type: 'error',
        grouping: true,
      })
      return;
    }
    const decryptedPasswd = await aesGcmDecrypt(encryptedPassword, usedKey);
    return decryptedPasswd;
  }


  // 用户名搜索
  const userNameSearch = async () => {
    try {
      // 从后端 API 获取分页数据
      await fetchData();
    } catch (error) {
      // 弹窗显示错误详细信息
      ElMessage({
        message: error,
        type: 'error',
        grouping: true,
      })
    }
  };

  const resetAdvQuery = () => {
    queryParams.appName = '';
    queryParams.userName = '';
    queryParams.url = '';
    queryParams.tags = '';
    fetchData();
  };


  const handleDownloadKeyFile = async (id) => {
    // 获取本地密钥
    const aesKey = localStorage.getItem('aesKey');
    if (userMasterAesKey.value == '' || userMasterAesKey.value == null || aesKey == undefined || aesKey == '') {
      ElMessage({
        message: '无主密钥不可用',
        type: 'error',
        grouping: true,
      })
      return;
    }
    try {
      const response = await http.get("/downloadKeyFile", {
        params: {
          id: id
        },
      });

      // 检查响应状态
      if (response.data.code !== 200) {
        ElMessage({
          message: `下载失败: ${response.data.message || '未知错误'}`,
          type: 'error',
          grouping: true,
        })
        return;
      }

      // 从返回包中获取数据
      const keyFileJson = response.data.data.key_file;
      if (!keyFileJson) {
        ElMessage({
          message: '未找到密钥文件数据',
          type: 'error',
          grouping: true,
        })
        return;
      }
      const crypto = response.data.data.chose_encrypt;
      const cryptoKey = response.data.data.key;

      // 解析key_file为JSON对象
      const parsedKeyFile = JSON.parse(keyFileJson);

      // 获取文件名和内容
      const keyFileName = parsedKeyFile.name;
      const keyFileContent = parsedKeyFile.content;

      // 处理文件内容 - 如果是对象，转换为JSON字符串
      let keyFileContentJson;
      if (typeof keyFileContent === 'object' && keyFileContent !== null) {
        keyFileContentJson = JSON.stringify(keyFileContent);
      } else {
        keyFileContentJson = keyFileContent;
      }

      // 获取解密算法和密钥
      const choseCrypto = await decryptPassword(crypto, userMasterAesKey.value);
      const decryptAlgo = new TextDecoder().decode(choseCrypto);
      const key = await decryptPassword(cryptoKey, userMasterAesKey.value);
      const decryptKey = new TextDecoder().decode(key);

      // 解密文件内容
      const keyFileContentDecrypted = await decryptPassword(keyFileContentJson, decryptKey);

      // 创建Blob并下载
      const blob = new Blob([keyFileContentDecrypted], { type: 'application/octet-stream' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = keyFileName;
      document.body.appendChild(a);
      a.click();
      ElMessage({
        message: "文件下载成功",
        type: 'success',
        grouping: true,
      })
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('下载密钥文件失败:', error);
      ElMessage({
        message: '下载密钥文件失败，请稍后重试',
        type: 'error',
        grouping: true,
      })
    }
  }


  // 解析excel文件为 JSON
  const parseExcelFile = async (file) => {
    const data = await file.arrayBuffer();
    const uint8Array = new Uint8Array(data);
    const fileExtension = file.name.split('.').pop().toLowerCase();

    let workbook;
    switch (fileExtension) {
      case 'xlsx':
        workbook = XLSX.read(uint8Array, { type: 'array' });
      case 'xls':
        if (!isLikelyUTF8(uint8Array)) {
          try {
            workbook = XLSX.read(uint8Array, {
              type: 'array',
              codepage: 936 // GBK代码页
            });
          } catch (e) {
            console.warn('GBK代码页读取失败，尝试默认方式');
            workbook = XLSX.read(uint8Array, { type: 'array' });
          }
        }
        break;
      case 'csv':
        const decodedText = await decodeCSV(uint8Array);
        workbook = XLSX.read(decodedText, { type: 'string' });
        break;
    }
    const sheetName = workbook.SheetNames[0];
    const worksheet = workbook.Sheets[sheetName];
    const rows = XLSX.utils.sheet_to_json(worksheet);

    // 暴力点，全字段强制 string，可能会导致数据丢失
    const normalized = rows.map(row => {
      const newRow = {};
      for (const key in row) {
        newRow[key] = row[key] != null ? String(row[key]) : "";
      }
      return newRow;
    });

    return normalized;
  };

  const importBeforeUpload = (file) => {
    importForm.file = file;
    return file;
  }
  // 解析CSV文件为字符串，根据是否为UTF-8编码选择不同的解码器
  const decodeCSV = async (uint8Array) => {
    const gbkDecoder = new TextDecoder('gbk');
    const utf8Decoder = new TextDecoder('utf-8');
    const isUtf8 = isLikelyUTF8(uint8Array);
    if (isUtf8) {
      return utf8Decoder.decode(uint8Array);
    } else {
      return gbkDecoder.decode(uint8Array);
    }
  }

  // 检查是否为UTF-8编码，通过尝试解码为UTF-8来判断，不一定精确
  const isLikelyUTF8 = (uint8Array, chunkSize = 4096) => {
    // 检查是否为空数据，空数据视为UTF-8
    if (uint8Array.length === 0) return true; // 空数据视为UTF-8

    // 创建UTF-8的编解码器，设置fatal:true以确保遇到无效序列时抛出异常
    const utf8Decoder = new TextDecoder('utf-8', { fatal: true });

    // 以4096字节为分块进行检测，使用stream选项处理跨分块的多字节字符
    let offset = 0;
    try {
      // 循环检测每个分块是否为有效的UTF-8
      while (offset < uint8Array.length) {
        // 计算当前分块的结束位置
        const end = Math.min(offset + chunkSize, uint8Array.length);
        // 获取当前分块
        const chunk = uint8Array.slice(offset, end);
        // 判断是否是最后一个分块
        const isFinalChunk = (end === uint8Array.length);
        // 解码当前分块，如果遇到无效序列会抛出异常
        // 对于非最后一个分块，设置stream: true以正确处理跨分块的多字节字符
        utf8Decoder.decode(chunk, { stream: !isFinalChunk });
        // 更新偏移量，准备检测下一个分块
        offset = end;
      }
      // 如果所有分块都解码成功，说明是有效的UTF-8
      return true;
    } catch (error) {
      // 任何一个分块解码失败，说明不是有效的UTF-8
      return false;
    }
  }

  const handleImportExceed = (files) => {
    const file = files[0];
    importForm.file = file;
    importFileList.value = [
      { name: file.name, raw: file }
    ];
  }

  const handleImportFileUpload = async () => {
    try {
      const file = importForm.file;
      // 检查文件类型
      const allowedTypes = ['.xlsx', '.xls', '.csv'];
      const fileExtension = file.name.split('.').pop().toLowerCase();
      if (!allowedTypes.includes('.' + fileExtension)) {
        ElMessage({
          message: '请上传Excel、CSV文件',
          type: 'error',
          grouping: true,
        })
        return false;
      }
      const jsonDatas = await parseExcelFile(file);

      // Chrome导出Excel字段映射到应用字段
      const chromeFieldMapping = {
        'name': 'appName',
        'url': 'url',
        'username': 'username',
        'password': 'password',
        'note': 'notes',
        'key': 'key' // 保留key字段用于特殊处理
      };

      // 处理Chrome导出的字段映射
      const mappedDatas = jsonDatas.map(item => {
        const mappedItem = {};
        for (const [chromeField, appField] of Object.entries(chromeFieldMapping)) {
          if (item.hasOwnProperty(chromeField)) {
            mappedItem[appField] = item[chromeField];
          }
        }
        // 保留其他可能存在的字段
        for (const [field, value] of Object.entries(item)) {
          if (!chromeFieldMapping.hasOwnProperty(field)) {
            mappedItem[field] = value;
          }
        }
        return mappedItem;
      });

      const encryptMapping = {
        isAppNameEncrypted: "appName",
        isUsernameEncrypted: "username",
        isUrlEncrypted: "url",
        isNotesEncrypted: "notes",
        isTagsEncrypted: "tags"
      };
      // 分析importForm中数据并提取需要加密的字段
      const encryptFieldSet = new Set(
        Object.entries(importForm)
          .filter(([key, value]) => value === true && encryptMapping[key])
          .map(([key]) => encryptMapping[key])
      );

      // 处理json数据的异步加密
      for (const jsonData of mappedDatas) {
        jsonData.choseCrypto = importForm.choseCrypto;
        // 确定当前记录使用的加密密钥
        let encryptionKey;
        if (importForm.keySource === 'file' && jsonData.key) {
          // 如果密钥来源于文件，且当前记录有自己的key，则使用自己的key
          encryptionKey = jsonData.key;
        } else {
          // 否则使用设置的统一密钥
          encryptionKey = importForm.key;
        }
        jsonData.key = encryptionKey;

        // 遍历对象的属性，而不是尝试迭代对象本身
        for (const field in jsonData) {
          if (encryptFieldSet.has(field) || field === "password" || field === "key" || field === "choseCrypto") {
            // 使用已导入的aesGcmEncrypt异步函数进行加密
            const cryptoTmp = await aesGcmEncrypt(jsonData[field], encryptionKey);
            // 保持加密结果为对象格式，与后端EncryptedString结构匹配
            jsonData[field] = cryptoTmp;
          }
        }
        // 规范字段，增加缺失字段
        jsonData.inputType = 'password';
        jsonData.keyFile = jsonData.keyFile || null;

        // 添加后端需要的加密状态布尔字段
        jsonData.isAppNameEncrypted = encryptFieldSet.has('appName');
        jsonData.isUsernameEncrypted = encryptFieldSet.has('username');
        jsonData.isUrlEncrypted = encryptFieldSet.has('url');
        jsonData.isNotesEncrypted = encryptFieldSet.has('notes');// notes和tag导出文件里没有，为了适配数据结构
        jsonData.isTagsEncrypted = encryptFieldSet.has('tags');
      }
      //console.log(mappedDatas);
      // 发送到接口importByFile
      const response = await importByFile(mappedDatas);
      if (response.code === 200) {
        ElMessage({
          message: "导入成功",
          type: 'success',
          grouping: true,
        })
        //关闭弹窗
        isImport.value = false;
        // 清空文件列表
        importFileList.value = [];
        // 清空表单数据
        importForm.file = null;
        importForm.keySource = 'file';
        importForm.choseCrypto = 'AES-GCM';
        importForm.key = '';
        importForm.file = null;
        importForm.isAppNameEncrypted = false;
        importForm.isUsernameEncrypted = false;
        importForm.isUrlEncrypted = false;
        importForm.isNotesEncrypted = false;
        importForm.isTagsEncrypted = false;
        // 刷新
        refresh();
      } else {
        ElMessage({
          message: response.msg || "导入失败",
          type: 'error',
          grouping: true,
        })
      }
      return true;
    } catch (error) {
      console.error('导入文件解析失败:', error);
      ElMessage({
        message: '导入文件解析失败，请稍后重试',
        type: 'error',
        grouping: true,
      })
      return false;
    }
  }

  const importByFile = async (jsonDatas) => {
    try {
      const jsonData = JSON.stringify(jsonDatas);
      // 交换密钥
      const tmpAesKey = await keyExchange.getAesKey();
      // AES 加密 JSON 数据
      const encryptedData = await aesGcmEncrypt(jsonData, tmpAesKey);
      const response = await http.post('/importByFile', {
        iv: encryptedData.iv,        // Base64 编码的 IV
        encryptedData: encryptedData.data  // Base64 编码的加密数据
      }, {
        headers: {
          'Content-Type': 'application/json'
        }
      });
      return response.data;
    } catch (error) {
      console.error('导入文件失败:', error);
      throw error;
    }
  }

  const handleExportAll = async () => {
    exportAll();
    // try {
    //   const response = await exportAll();
    //   if (response.code === 200) {
    //     ElMessage({
    //       message: "导出成功",
    //       type: 'success',
    //       grouping: true,
    //     })
    //     // 下载文件
    //     downloadFile(response.data, 'passwords.json');
    //   } else {
    //     ElMessage({
    //       message: response.msg || "导出失败",
    //       type: 'error',
    //       grouping: true,
    //     })
    //   }
    // } catch (error) {
    //   console.error('导出文件失败:', error);
    //   ElMessage({
    //     message: '导出文件失败，请稍后重试',
    //     type: 'error',
    //     grouping: true,
    //   })
    // }
  }

  const exportAll = () => {
    const serverAddress = getServerAddress();
    window.location.href = serverAddress + '/exportAll';
  }

</script>

<style scoped>
  /* 容器样式 */
  .container {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: var(--space-sm);
    background-color: var(--bg-secondary);
  }

  /* 头部样式 */
  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 10px;
    padding: var(--space-sm);
    margin: var(--space-sm);
    background-color: var(--bg-card);
    border-radius: 12px;
    box-shadow: var(--shadow-light);
    border: 1px solid var(--border-light);
  }

  .header-left {
    display: flex;
    gap: 10px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  @media (max-width: 768px) {
    .header {
      flex-direction: column;
      align-items: stretch;
    }

    .header-left,
    .header-right {
      justify-content: center;
    }

    .header-right {
      flex-direction: column;
      width: 100%;
    }

    .header-right .el-input {
      max-width: none;
      width: 100%;
      margin-right: 0;
      margin-bottom: 10px;
    }
  }

  /* 内容区域样式 */
  .content {
    flex: 1;
    margin-bottom: 10px;
    overflow: hidden;
  }

  /* 表格容器样式 */
  .table-container {
    background-color: var(--bg-card);
    border-radius: 12px;
    box-shadow: var(--shadow-light);
    border: 1px solid var(--border-light);
    overflow: hidden;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  /* 底部样式 */
  .footer {
    background-color: var(--bg-card);
    padding: 12px 15px;
    border-top: 1px solid var(--border-light);
    text-align: center;
    box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.05);
    margin-top: 10px;
  }

  /* 对话框标题样式 */
  .dialog-title {
    display: flex;
    justify-content: center;
    font-weight: 700;
    font-size: 1.4rem;
    color: var(--text-primary);
    margin-bottom: 10px;
  }



  /* Element Plus 表格样式覆盖 */
  .el-table {
    border: none !important;
    border-radius: 12px;
    overflow: hidden;
  }

  .el-table th {
    background-color: var(--bg-tertiary) !important;
    color: var(--text-primary) !important;
    font-weight: 600;
    border-bottom: 1px solid var(--border-light) !important;
  }

  .el-table td {
    border-bottom: 1px solid var(--border-light) !important;
    color: var(--text-secondary);
  }

  .el-table__body tr:hover>td {
    background-color: rgba(22, 93, 255, 0.04) !important;
  }

  /* Element Plus 按钮样式覆盖
  .el-button {
    transition: all 0.3s ease !important;
    border-radius: 8px !important;
  }

  .el-button--primary {
    background-color: var(--primary-color) !important;
    border-color: var(--primary-color) !important;
  }

  .el-button--primary:hover {
    background-color: var(--primary-hover) !important;
    border-color: var(--primary-hover) !important;
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(22, 93, 255, 0.3);
  }

  .el-button--danger {
    background-color: var(--error-color) !important;
    border-color: var(--error-color) !important;
  }

  .el-button--danger:hover {
    background-color: var(--error-hover) !important;
    border-color: var(--error-hover) !important;
  }

  .el-button--success {
    background-color: var(--success-color) !important;
    border-color: var(--success-color) !important;
  }

  .el-button--success:hover {
    background-color: var(--success-hover) !important;
    border-color: var(--success-hover) !important;
  }*/

  /* Element Plus 输入框样式覆盖 */
  .el-input__wrapper {
    border-radius: 8px !important;
    background-color: var(--bg-input) !important;
  }

  .el-input__wrapper:focus-within {
    box-shadow: 0 0 0 2px rgba(22, 93, 255, 0.2) !important;
    border-color: var(--primary-color) !important;
  }

  /* Element Plus 对话框样式覆盖 */
  .el-dialog {
    border-radius: 16px !important;
    overflow: hidden;
  }

  .el-dialog__header {
    background-color: var(--bg-card) !important;
    border-bottom: 1px solid var(--border-light) !important;
  }

  .el-dialog__body {
    background-color: var(--bg-secondary) !important;
    padding: 24px !important;
  }

  /* Element Plus 表单样式覆盖 */
  .el-form-item {
    margin-bottom: 20px !important;
  }

  .el-form-item__label {
    color: var(--text-primary) !important;
    font-weight: 500 !important;
  }

  /* Element Plus 选择框样式覆盖 */
  .el-select .el-input__wrapper {
    border-radius: 8px !important;
  }

  /* 高级查询表单样式 */
  .query-form .el-dialog__body {
    background-color: var(--bg-card) !important;
    padding: 24px !important;
  }

  .query-form .el-form {
    margin-bottom: 0;
  }

  .query-form .el-input__wrapper {
    border-radius: 8px !important;
    background-color: var(--bg-input) !important;
  }

  .query-form .el-form-item__label {
    color: var(--text-primary) !important;
    font-weight: 500 !important;
  }

  /* 切换按钮组样式 */
  .toggle-buttons {
    margin-bottom: 10px;
  }

  /* 加密提示样式 */
  .toggle-buttons .el-tooltip__content {
    background-color: var(--bg-secondary) !important;
    color: var(--text-primary) !important;
    border: 1px solid var(--border-light) !important;
    box-shadow: var(--shadow-medium);
  }

  /* 说明文字样式 */
  .toggle-buttons span[style*="color: red"] {
    color: var(--error-color) !important;
    font-size: 14px;
  }

  /* 响应式设计 */
  @media (max-width: 1024px) {
    .header {
      flex-direction: column;
      align-items: stretch;
    }

    .header>div {
      width: 100%;
      display: flex;
      justify-content: space-between;
      margin-bottom: 10px;
    }

    .header .el-input {
      max-width: none;
    }
  }

  @media (max-width: 768px) {
    .container {
      padding: 12px;
    }

    .el-dialog {
      width: 95% !important;
      min-width: unset !important;
      max-width: unset !important;
    }

    .el-table {
      font-size: 13px;
    }

    .el-table th,
    .el-table td {
      padding: 8px 5px;
    }
  }

  .import-form :deep(.el-dialog__header) {
    border-radius: 12px 12px 0 0;
  }

  .import-form :deep(.el-dialog__footer) {
    border-radius: 0 0 12px 12px;
  }

  /* 确保表单元素在小屏幕上也能正常显示 */
  @media (max-width: 480px) {
    .import-form :deep(.el-form-item__label) {
      width: 70px;
    }

    .import-form :deep(.el-form-item__content) {
      margin-left: 80px;
    }
  }
</style>