<template>
    <div>
        <div>
            <button @click="scrollToNext">下一处</button>
            <button @click="scrollToPrev">上一处</button>
            <button @click="clearLeft">清空左侧</button>
            <button @click="clearRight">清空右侧</button>
            <button @click="mergeLeft">合并到左侧</button>
            <button @click="mergeRight">合并到右侧</button>
            <!-- 配置按钮 -->
            <button @click="showConfig = true">设置编辑器样式</button>
        </div>

        <!-- 配置弹窗 -->
        <el-dialog title="编辑器样式设置" v-model="showConfig" width="400px" :close-on-click-modal="false">
            <div>
                <el-checkbox v-model="showLineNumbers">显示行号</el-checkbox>
            </div>
            <div style="margin-top: 10px;">
                <el-checkbox v-model="wrapLines">自动换行</el-checkbox>
            </div>
            <template #footer>
                <el-button @click="showConfig = false">关闭</el-button>
                <el-button type="primary" @click="applySettings">应用</el-button>
            </template>
        </el-dialog>

        <!-- Mergely 对比区域 -->
        <div ref="mergelyContainer" id="compare" style="margin: 0;padding: 0;width: 100%;height: 100%;">
            <!-- Mergely 内容自动填充 -->
        </div>
    </div>
</template>

<script setup>
    import { ref, onMounted, nextTick } from 'vue';
    import Mergely from 'mergely';
    import 'mergely/lib/mergely.css';
    import 'codemirror/lib/codemirror.css';
    import 'codemirror/mode/javascript/javascript.js';

    // 定义 refs
    const mergelyContainer = ref(null);
    let mergelyInstance = null;
    const showConfig = ref(false); // 控制弹窗显示
    const showLineNumbers = ref(true);
    const wrapLines = ref(true);

    // 示例文本内容
    const leftText = ``;

    const rightText = ``;

    // 在组件挂载后初始化 Mergely
    onMounted(async () => {
        await nextTick(); // 确保 DOM 已经渲染

        if (mergelyContainer.value) {
            // 初始化 Mergely
            mergelyInstance = new Mergely(mergelyContainer.value, {
                lhs: leftText, // 左侧文本
                rhs: rightText, // 右侧文本
                sidebar: true, // 显示侧边栏
                ignorews: false, // 忽略空白字符
                license: "lgpl-separate-notice", // 设置许可证以移除提示框
                wrap_lines: true, // 自动换行
                viewport: true, // 启用视口模式

            });

        } else {
            console.error('Mergely container is null!');
        }
    });


    // 应用配置
    const applySettings = () => {
        updateEditorStyles();
        showConfig.value = false; // 关闭弹窗
    };

    // 按钮事件处理
    const scrollToNext = () => {
        mergelyInstance.scrollToDiff('next');
    };

    const scrollToPrev = () => {
        mergelyInstance.scrollToDiff('prev');
    };

    const clearLeft = () => {
        mergelyInstance.clear('lhs');
    };

    const clearRight = () => {
        mergelyInstance.clear('rhs');
    };

    const mergeLeft = () => {
        mergelyInstance.merge('lhs');
    };

    const mergeRight = () => {
        mergelyInstance.merge('rhs');
    };

    // 更新编辑器样式
    const updateEditorStyles = () => {
        if (mergelyInstance) {
            mergelyInstance.options({
                line_numbers: showLineNumbers.value,
                wrap_lines: wrapLines.value,
            });
        }
    };
</script>

<style scoped>
    .compare {
        display: flex;
        width: 100%;
        height: 100%;
        border: 1px solid var(--border-color-light);
        overflow: hidden;
        max-height: 500px;
    }

    :deep(.mergely-editor) {
        width: 100%;
        height: 100%;
        overflow: hidden;
        max-height: 92vh;
    }

    :deep(.CodeMirror) {
        width: 100%;
        height: auto;
        overflow: hidden;
        line-height: 18px;
    }

    :deep(.CodeMirror-vscrollbar) {
        right: 0;
        top: 0;
        overflow-x: hidden;
        overflow-y: scroll;
    }


    :deep(.CodeMirror-scroll) {
        overflow: scroll !important;
        margin-bottom: -50px;
        margin-right: -50px;
        padding-bottom: 50px;
        height: 100%;
        outline: none;
        position: relative;
        z-index: 0;
    }

    /* 自定义差异样式 */
    :deep(.CodeMirror-linewidget.added) {
        background-color: #d6f5d6;
        /* 添加部分的背景色 */
    }

    :deep(.CodeMirror-linewidget.removed) {
        background-color: #ffdddd;
        /* 删除部分的背景色 */
        text-decoration: line-through;
        /* 删除线 */
    }

    :deep(.CodeMirror-linewidget.changed) {
        background-color: #ffffcc;
        /* 修改部分的背景色 */
    }
</style>