<template>
    <el-container style="max-width: 100%; margin: auto; padding: 2em;display: block;">
        <!-- 顶部按钮区 -->
        <div class="header-actions">
            <el-row justify="end" :gutter="10">
                <el-button type="primary" @click="handleQuery">查询</el-button>
                <el-button type="danger" @click="handleDelete">删除</el-button>
                <el-button type="success" @click="handleInsert">入库</el-button>
            </el-row>
        </div>

        <!-- 可滚动数据区 -->
        <div class="scrollable-data">
            <el-form label-position="top" style="width: 100%">
                <!-- 模块名 -->
                <el-form-item label="模块名">
                    <el-select v-model="current.module" clearable filterable allow-create default-first-option
                        placeholder="模块名，例如：Kernel32.dll" style="width: 100%" @visible-change="handleChange">
                        <el-option v-for="mod in moduleOptions" :key="mod" :label="mod" :value="mod" />
                    </el-select>
                </el-form-item>

                <!-- 函数名 -->
                <el-form-item label="函数名">
                    <el-select v-model="current.functionName" clearable filterable allow-create default-first-option
                        placeholder="函数名，例如：EnumProcesses" style="width: 100%" @visible-change="handleChange">
                        <el-option v-for="fn in functionOptions" :key="fn" :label="fn" :value="fn" />
                    </el-select>
                </el-form-item>

                <!-- 原型 -->
                <el-form-item label="原型">
                    <el-input v-model="current.funcDefine" type="textarea" autosize placeholder="函数原型定义，例如 
            BOOL EnumProcesses(
                [out] DWORD *lpidProcess, // 用于接收进程ID的数组
                [in] DWORD cb, // 数组的字节大小
                [out] LPDWORD lpcbNeeded // 实际返回的字节
            );" />
                </el-form-item>

                <!-- 特征值 -->
                <el-form-item label="特征">
                    <el-input type="textarea" v-model="current.signature" autosize placeholder="特征值" />
                </el-form-item>

                <!-- 功能描述 -->
                <el-form-item label="功能描述">
                    <el-input type="textarea" v-model="current.funcDesc" autosize
                        placeholder="例如：EnumProcesses函数是一个Windows API函数，它可以获取系统中每个进程对象的进程标识符（PID）。PID是一个唯一的数字，用于标识一个进程。通过该函数，攻击者可以枚举系统进程以注入和规避沙箱。" />
                </el-form-item>

                <!-- 其他 -->
                <el-form-item label="其他">
                    <el-input type="textarea" v-model="current.other" autosize placeholder="调用注意事项、用例、返回值含义等补充信息
            例如：如果函数成功，返回值为非零；如果失败，返回值为零，可以调用
GetLastError函数获取错误代码。要使用这个函数，需要包含psapi.h头
文件，并链接psapi.lib库文件
用例：
DWORD aProcesses[1024], cbNeeded, cProcesses; // 进程ID数组，需要的字节数，进程数量
unsigned int i;
if (!EnumProcesses(aProcesses, sizeof(aProcesses),&cbNeeded)) // 调用EnumProcesses函数获取进程列表
    {
        return 1; // 如果失败，返回1
    }" />
                </el-form-item>
            </el-form>
        </div>
    </el-container>
</template>

<script setup>
    import { ref } from 'vue'
    import http from '@/js/http.js'
    import { ElMessage, ElMessageBox } from 'element-plus'
    import keyExchange from '@/js/keyExchange.js';

    const moduleOptions = ref([])  // 模块名选项
    const functionOptions = ref([]) // 函数名选项
    const current = ref({
        module: '',
        functionName: '',
        funcDefine: '',
        signature: '',
        funcDesc: '',
        other: '',
    })
    let cachedResults = []// 缓存的查询结果

    //下拉框切换时
    function handleChange(visible) {
        if (!visible) {
            // 下拉框隐藏时，获取选中的值
            const selectedModule = (current.value?.module ?? '').trim() // 获取当前选中的模块名
            const selectedFunction = (current.value?.functionName ?? '').trim() // 获取当前选中的函数名
            // 如果值非空，则从缓存中获取该值的函数名选项
            if (selectedModule && selectedFunction) {
                const result = cachedResults.filter(api => {
                    // 匹配模块名和函数名相同的项目
                    if (selectedModule && selectedFunction) return api.module === selectedModule && api.functionName === selectedFunction
                })
                // 更新current.value
                if (result.length > 0) {
                    Object.assign(current.value, result[0])
                } else {
                    Object.assign(current.value, { funcDefine: '', signature: '', funcDesc: '', other: '' })
                }
            } else {
                functionOptions.value = [] // 如果值为空，则清空函数名选项
            }
        }

    }

    // 同步模块名和函数名选项
    function syncOptions(result) {
        moduleOptions.value = Array.from(new Set(result.map(r => r.module)))
        functionOptions.value = Array.from(new Set(result.map(r => r.functionName)))
    }


    //  按钮处理事件
    // 查询按钮点击事件
    const handleQuery = async () => {
        moduleOptions.value = [] // 清空模块名选项
        functionOptions.value = [] // 清空函数名选项
        const result = await queryData()
        if (result) {
            Object.assign(cachedResults, result)

            if (cachedResults.length === 0) {
                current.value = { funcDefine: '', signature: '', funcDesc: '', other: '' }
                ElMessage({
                    message: '未找到匹配项',
                    type: 'warning',
                    grouping: true,
                })
                return
            }
            Object.assign(current.value, cachedResults[0])
            syncOptions(cachedResults)
        } else {
            ElMessage({
                message: '无记录！',
                type: 'error',
                grouping: true,
            })
        }

    }

    const handleInsert = async () => {
        const m = (current.value?.module ?? '').trim()
        const f = (current.value?.functionName ?? '').trim()
        // 两个数据确定唯一api
        if (m && f) {
            const queryResult = await queryData()
            if (queryResult) {
                ElMessageBox.confirm('API 已存在，是否覆盖？', '提示', {
                    type: 'warning'
                }).then(async () => {
                    //构造数据发送
                    try {
                        const jsonData = current.value;
                        const response = await http.post("/insertWinApi", jsonData);
                        ElMessage({
                            message: '覆盖成功',
                            type: 'success',
                            grouping: true,
                        })
                    } catch (err) {
                        // 统一处理错误（包含网络错误和 HTTP 4xx/5xx 错误）
                        const errorMsg = err.response?.data?.message || err.message;
                        ElMessage({
                            message: `插入失败: ${errorMsg}`,
                            type: 'error',
                            grouping: true,
                        })
                    }
                }).catch(() => { })
            } else {
                //构造数据发送
                try {
                    const jsonData = current.value;
                    const response = await http.post("/insertWinApi", jsonData,);
                    ElMessage({
                        message: '插入成功',
                        type: 'success',
                        grouping: true,
                    })
                } catch (err) {
                    // 错误处理 
                    const errorMsg = err.response?.data?.message || err.message;
                    ElMessage({
                        message: `插入失败: ${errorMsg}`,
                        type: 'error',
                        grouping: true,
                    })
                }
                ElMessage({
                    message: '已添加',
                    type: 'success',
                    grouping: true,
                })
            }
        } else {
            ElMessage({
                message: '数据缺失，不可入库！',
                type: 'warning',
                grouping: true,
            })
        }
    }

    const handleDelete = async () => {
        const m = (current.value?.module ?? '').trim()
        const f = (current.value?.functionName ?? '').trim()
        // 两个数据确定唯一api
        if (m && f) {
            const queryResult = await queryData()
            if (queryResult) {
                ElMessageBox.confirm('确认删除该 API？', '警告', {
                    type: 'warning'
                }).then(async () => {
                    try {
                        await http.get("/deleteWinApi", { params: { module: m, functionName: f } });
                        // 判断响应状态
                        if (err.response?.status === 200) {
                            ElMessage({
                                message: '删除成功',
                                type: 'success',
                                grouping: true,
                            })
                            // 更新缓存数据
                            cachedResults = cachedResults.filter(api => api.module !== m && api.functionName !== f);
                            syncOptions(cachedResults);
                        } else {
                            ElMessage({
                                message: '删除失败',
                                type: 'error',
                                grouping: true,
                            })
                        }
                    } catch (err) {
                        // 错误处理 
                        const errorMsg = err.response?.data?.message || err.message;
                        ElMessage({
                            message: `删除失败: ${errorMsg}`,
                            type: 'error',
                            grouping: true,
                        })
                    }
                })
            }
        }
    }

    // 查询数据
    async function queryData() {
        const m = (current.value?.module ?? '').trim() // 模块名，去除首尾空格
        const f = (current.value?.functionName ?? '').trim()// 函数名，去除首尾空格
        if (!m && !f) {
            ElMessage({
                message: '空查询！',
                type: 'warning',
                grouping: true,
            })
            return
        }
        //从服务器查询
        const response = await http.get("/queryWinApi", {
            params: { module: m, functionName: f }
        });

        // 响应数据直接在 response.data 中
        return response.data;
    }
</script>

<style scoped>
    .header-actions {
        width: 100%;
        height: 3em;
        display: flex;
    }

    .scrollable-data {
        width: 100%;
        max-height: 680px;
        overflow-y: auto;
        padding-right: 10px;
        border-top: 1px solid var(--border-color-light);
        padding-top: 1em;
    }

    .el-form-item {
        margin-bottom: 20px;
    }
</style>