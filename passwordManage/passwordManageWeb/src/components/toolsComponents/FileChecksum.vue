<template>
    <div class="file-integrity-checker">
        <h2>文件与文件夹校验值计算</h2>
        <div>
            <el-select v-model="selectedAlgorithms" multiple clearable placeholder="请选择校验算法" style="width: 100%;">
                <el-option-group label="经典密码学算法">
                    <el-option label="SHA-1" value="SHA-1" />
                    <el-option label="SHA-224" value="SHA-224" />
                    <el-option label="SHA-256" value="SHA-256" />
                    <el-option label="SHA-384" value="SHA-384" />
                    <el-option label="SHA-512" value="SHA-512" />
                    <el-option label="RIPEMD160" value="RIPEMD160" />
                    <el-option label="KECCAK-256" value="KECCAK-256" />
                    <el-option label="SM3" value="SM3" />
                </el-option-group>

                <el-option-group label="非安全或兼容性算法">
                    <el-option label="MD5" value="MD5" />
                </el-option-group>
                <el-option-group label="现代推荐算法（安全且性能优异）">
                    <el-option label="BLAKE3" value="BLAKE3" />
                    <el-option label="BLAKE2b" value="BLAKE2B" />
                    <el-option label="BLAKE2s" value="BLAKE2S" />
                    <el-option label="SHA3-224" value="SHA3-224" />
                    <el-option label="SHA3-256" value="SHA3-256" />
                    <el-option label="SHA3-384" value="SHA3-384" />
                    <el-option label="SHA3-512" value="SHA3-512" />
                </el-option-group>

                <el-option-group label="高性能非密码学哈希">
                    <el-option label="XXH32" value="XXH32" />
                    <el-option label="XXH64" value="XXH64" />
                    <el-option label="XXH3" value="XXH3" />
                </el-option-group>

                <el-option-group label="校验和">
                    <el-option label="CRC32" value="CRC32" />
                    <el-option label="CRC64" value="CRC64" />
                </el-option-group>
            </el-select>

            <!-- 参数输入区 -->
            <div v-for="algo in selectedAlgorithms" :key="algo" class="config-item">
                <template v-if="supportsOutputLength(algo)">
                    <div class="param-row">
                        <span>{{ algo }} 输出长度：</span>
                        <el-select v-model="outputLengthMap[algo]" filterable allow-create default-first-option
                            placeholder="请选择或输入" style="width: 140px">
                            <el-option v-for="len in getRecommendedLengths(algo)" :key="len" :label="len + ' 字节'"
                                :value="len" />
                        </el-select>
                        <span v-if="!isValidLength(algo, outputLengthMap[algo])" style="color: red; margin-left: 5px;">
                            非法长度
                        </span>
                    </div>
                </template>
            </div>
        </div>
        <div class="upload-sections">
            <!-- 左侧：文件上传 -->
            <div class="upload-section">
                <div class="section-header">

                    <h3>文件计算</h3>
                    <div class="actions">
                        <el-button size="small" @click="clearResults('file')">清除结果</el-button>
                        <el-button size="small" @click="exportResults('file')">导出结果</el-button>
                    </div>
                </div>



                <el-upload drag action="" :auto-upload="false" :file-list="fileList" :on-change="handleFiles" multiple>
                    <i class="el-icon-upload"></i>
                    <div class="el-upload__text">
                        拖拽文件到此处，或点击上传
                    </div>
                </el-upload>

                <div class="results">
                    <h3>计算结果</h3>
                    <ul class="scrollable-list">
                        <li v-for="(file, index) in results" :key="index">
                            <p><strong>文件名:</strong> {{ file.name }}</p>
                            <p><strong>路径:</strong> {{ file.path }}</p>
                            <p><strong>大小:</strong> {{ formatSize(file.size) }}</p>
                            <div v-for="(hashValue, hashName) in file.hashes" :key="hashName">
                                <p><strong>{{ hashName }}:</strong> {{ hashValue }}</p>
                            </div>
                            <hr />
                        </li>
                    </ul>
                </div>
            </div>

            <!-- 右侧：文件夹上传 -->
            <div class="upload-section">
                <div class="section-header">
                    <h3>文件夹批量计算</h3>
                    <div class="actions">
                        <el-button size="small" @click="clearResults('folder')">清除结果</el-button>
                        <el-button size="small" @click="exportResults('folder')">导出结果</el-button>
                    </div>
                </div>

                <file-upload ref="folderUpload" :directory="true" multiple drop @input-file="handleFolderFiles">
                    <div class="dropzone">
                        <i class="el-icon-folder"></i>
                        <p>拖拽文件夹到此处（chrome只能点击）</p>
                    </div>
                </file-upload>

                <div class="results">
                    <h3>计算结果</h3>
                    <ul class="scrollable-list">
                        <li v-for="(file, index) in folderResults" :key="index">
                            <p><strong>文件名:</strong> {{ file.name }}</p>
                            <p><strong>路径:</strong> {{ file.path }}</p>
                            <p><strong>大小:</strong> {{ formatSize(file.size) }}</p>
                            <div v-for="(hashValue, hashName) in file.hashes" :key="hashName">
                                <p><strong>{{ hashName }}:</strong> {{ hashValue }}</p>
                            </div>
                            <hr />
                        </li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
    import { ref, computed } from 'vue'
    import FileUpload from "vue-upload-component";
    import { ElCheckbox, ElCheckboxGroup, ElButton } from "element-plus";
    import {
        createMD5,
        createSHA1,
        createSHA224,
        createSHA256,
        createSHA384,
        createSHA512,
        createSHA3,
        createBLAKE2b,
        createBLAKE2s,
        createBLAKE3,
        createRIPEMD160,
        createKeccak,
        createSM3,
        createCRC32,
        createCRC64,
        createXXHash32,
        createXXHash64,
        createXXHash3
    } from 'hash-wasm';

    const MAX_SIZE = 10 * 1024 * 1024; // 10MB
    const CHUNK_SIZE = 2 * 1024 * 1024; // 超过10M，分块大小，2MB

    // 选择的算法
    const selectedAlgorithms = ref(["SHA-1", "SHA-256", "MD5", "CRC32"])
    const outputLengthMap = ref({})

    // 左侧文件上传数据
    const results = ref([]);
    const fileList = ref([]);

    // 右侧文件夹上传数据
    const folderResults = ref([]);

    const outputLengthCapable = {
        BLAKE3: { min: 1, max: 32, step: 1, recommend: [16, 32] },
        BLAKE2B: { min: 8, max: 64, step: 8, recommend: [32, 64] },
        BLAKE2S: { min: 8, max: 32, step: 8, recommend: [16, 32] },
        XXH3: { min: 8, max: 16, step: 8, recommend: [8, 16] },
    }

    // 格式化文件大小
    const formatSize = (size) => {
        const units = ["B", "KB", "MB", "GB"];
        let unitIndex = 0;
        while (size >= 1024 && unitIndex < units.length - 1) {
            size /= 1024;
            unitIndex++;
        }
        return `${size.toFixed(2)} ${units[unitIndex]}`;
    };

    // 处理左侧文件上传
    const handleFiles = async (file, fileList) => {
        //results.value = []; // 清空之前的结果
        for (const fileItem of fileList) {
            if (fileItem.raw) {
                await processFile(fileItem.raw, results);
            }
        }
        fileList.splice(0, fileList.length); // 清空文件列表
    };

    // 处理右侧文件夹上传
    const handleFolderFiles = async (newFile, oldFile) => {
        if (!newFile || !newFile.file) return;
        if (newFile.file.webkitRelativePath || newFile.file.path) {
            await processFile(newFile.file, folderResults);
        }
        // 清空当前上传组件的文件
        newFile.remove();
    };

    // 处理单个文件
    const processFile = async (file, targetResults) => {

        const hashes = await calculateHashes(file, selectedAlgorithms.value);

        const fileInfo = {
            name: file.name,
            path: file.webkitRelativePath || file.name,
            size: file.size,
            hashes
        };

        targetResults.value.push(fileInfo);
    };

    // 清除结果
    const clearResults = (type) => {
        if (type === "file") {
            results.value = [];
        } else if (type === "folder") {
            folderResults.value = [];
        }
    };

    // 导出结果
    const exportResults = (type) => {
        const targetResults = type === "file" ? results.value : folderResults.value;
        const blob = new Blob([JSON.stringify(targetResults, null, 2)], { type: "application/json" });
        const url = URL.createObjectURL(blob);

        const link = document.createElement("a");
        link.href = url;
        link.download = `${type}_results.json`;
        link.click();
        URL.revokeObjectURL(url);
    };

    // 校验选择器
    const createHasher = async (algorithm, options = {}) => {
        const outputLengthtmp = outputLengthMap.value[algorithm];  // 可变输出长度

        switch (algorithm.toUpperCase()) {
            case 'MD5': return await createMD5();
            case 'SHA-1': return await createSHA1();
            case 'SHA-224': return await createSHA224();
            case 'SHA-256': return await createSHA256();
            case 'SHA-384': return await createSHA384();
            case 'SHA-512': return await createSHA512();

            case 'SHA3-224': return await createSHA3(224);
            case 'SHA3-256': return await createSHA3(256);
            case 'SHA3-384': return await createSHA3(384);
            case 'SHA3-512': return await createSHA3(512);

            case 'BLAKE2B': return await createBLAKE2b(outputLengthtmp || 64);
            case 'BLAKE2S': return await createBLAKE2s(outputLengthtmp || 32);
            case 'BLAKE3': return await createBLAKE3(outputLengthtmp || 32);

            case 'RIPEMD160': return await createRIPEMD160();
            case 'KECCAK-256': return await createKeccak(256);

            case 'SM3': return await createSM3();
            case 'CRC32': return await createCRC32();
            case 'CRC64': return await createCRC64();

            case 'XXH32': return await createXXHash32();
            case 'XXH64': return await createXXHash64();
            case 'XXH3': return await createXXHash3(outputLengthtmp || 16);


            default:
                throw new Error(`不支持的哈希算法: ${algorithm}。\n可选算法包括: MD5, SHA-1, SHA-224, SHA-256, SHA-384, SHA-512, SHA3-224/256/384/512, BLAKE2b/s, BLAKE3, RIPEMD160, KECCAK-256, SM3, CRC32/64, XXH32/64/3。`);
        }
    };

    // 计算哈希值
    const calculateHashes = async (file, algorithms) => {
        const isSmallFile = file.size <= MAX_SIZE;
        const result = {};
        const hashers = {};

        // 创建哈希器
        for (const algo of algorithms) {
            const hasher = await createHasher(algo);
            hasher.init();
            hashers[algo] = hasher;
        }

        if (isSmallFile) {
            // 小文件：一次性读取
            const buffer = await file.arrayBuffer();
            const uint8Array = new Uint8Array(buffer);
            for (const hasher of Object.values(hashers)) {
                hasher.update(uint8Array);
            }
        } else {
            // 大文件：分块读取
            let offset = 0;
            while (offset < file.size) {
                const chunk = file.slice(offset, offset + CHUNK_SIZE);
                const buffer = await chunk.arrayBuffer();
                const uint8Array = new Uint8Array(buffer);

                for (const hasher of Object.values(hashers)) {
                    hasher.update(uint8Array);
                }

                offset += CHUNK_SIZE;
            }
        }

        // 获取哈希结果
        for (const algo of algorithms) {
            result[algo] = hashers[algo].digest();
        }

        return result;
    };

    // 支持输出长度的算法列表
    const supportsOutputLength = (algo) => {
        return Object.hasOwn(outputLengthCapable, algo)
    }

    // 获取推荐长度列表
    const getRecommendedLengths = (algo) => {
        return outputLengthCapable[algo]?.recommend || []
    }

    // 校验输出长度是否合法
    const isValidLength = (algo, len) => {
        const cap = outputLengthCapable[algo]
        if (!cap || len === undefined) return true
        return (
            Number.isInteger(len) &&
            len >= cap.min &&
            len <= cap.max &&
            len % cap.step === 0
        )
    }

</script>

<style scoped>
    .file-integrity-checker {
        margin: 20px;
    }

    .upload-sections {
        display: flex;
        gap: 20px;
    }

    .upload-section {
        flex: 1;
        border: 1px solid var(--border-color-light);
        border-radius: 5px;
        padding: 20px;
    }

    .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;
    }

    .dropzone {
        border: 2px dashed var(--border-color-light);
        padding: 20px;
        text-align: center;
        cursor: pointer;
    }

    .results {
        margin-top: 20px;
    }

    .scrollable-list {
        max-height: 300px;
        overflow-y: auto;
    }

    ul {
        list-style: none;
        padding: 0;
    }

    li {
        margin-bottom: 20px;
    }

    hr {
        margin-top: 10px;
        border: 0;
        border-top: 1px solid var(--border-color-light);
    }
</style>