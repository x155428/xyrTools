<template>
    <div class="file-integrity-checker">
        <h2>文件完整性验证</h2>

        <!-- 校验方法选择 -->
        <div class="section">
            <h3>1. 选择校验方法</h3>
            <el-select v-model="selectedAlgorithm" placeholder="请选择校验算法" @change="onAlgorithmChange">
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
            <!-- outputLength 动态输入框 -->
            <div v-if="showOutputLength" style="margin-top: 16px;">
                <el-form-item label="输出长度 (字节)">
                    <el-select v-model="outputLengthStr" placeholder="选择或输入长度" filterable allow-create
                        style="width: 200px">
                        <el-option v-for="len in currentCommonLengths" :key="len" :label="len + ' 字节'"
                            :value="String(len)" />
                    </el-select>
                    <span style="margin-left: 8px; color: #888;">
                        范围: {{ outputLengthRange.min }} - {{ outputLengthRange.max }} 字节
                    </span>
                </el-form-item>
            </div>
        </div>

        <!-- 文件选择部分 -->
        <div class="section">
            <h3>2. 选择或上传文件</h3>
            <el-upload drag action="" :auto-upload="false" :on-change="handleFileChange" :file-list="userFileList"
                :on-remove="handleFileRemove" :limit="1" :on-exceed="handleExceed">
                <i class="el-icon-upload"></i>
                <div class="el-upload__text">拖拽文件到此处，或点击上传</div>
            </el-upload>
            <div v-if="fileName" class="file-info">选中文件: <strong>{{ fileName }}</strong></div>
        </div>

        <!-- 校验值输入或上传部分 -->
        <div class="section">
            <h3>3. 输入或上传预期校验值</h3>
            <el-upload drag action="" :auto-upload="false" :on-change="handleHashFileChange" :file-list="hashFileList"
                :on-remove="handleHashFileRemove" :limit="1" :on-exceed="handleExceed">
                <i class="el-icon-upload"></i>
                <div class="el-upload__text">拖拽哈希文件到此处，或点击上传</div>
            </el-upload>
            <p class="or">或</p>
            <el-input v-model="expectedHash" placeholder="手动输入预期哈希值" clearable />
            <div v-if="hashFileName" class="file-info">选中的哈希文件: <strong>{{ hashFileName }}</strong></div>
        </div>

        <!-- 验证按钮 -->
        <div class="section">
            <h3>4. 验证文件完整性</h3>
            <el-button type="primary" @click="checkIntegrity"
                :disabled="!userFile || !expectedHash || !userFileList.length">
                验证完整性
            </el-button>
        </div>

        <!-- 结果弹窗 -->
        <el-dialog v-model="showResult" title="验证结果" width="400px" @close="closeModal">
            <template #default>
                <p v-if="verificationResult" class="success">文件验证通过！</p>
                <p v-else class="error">文件验证失败！</p>
                <p><strong>计算出的哈希值:</strong> {{ actualHash }}</p>
                <p><strong>预期的哈希值:</strong> {{ expectedHash }}</p>
            </template>
            <template #footer>
                <el-button type="primary" @click="closeModal">关闭</el-button>
            </template>
        </el-dialog>
    </div>
</template>

<script setup>
    import { ref, computed } from "vue";
    import { ElMessageBox } from "element-plus";
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

    const CHUNK_SIZE = 2 * 1024 * 1024; // 2MB 分块文件处理

    const userFile = ref(null); // 用户上传的文件
    const fileName = ref(""); // 文件名称
    const expectedHash = ref(""); // 用户输入或解析出的哈希值
    const hashFileName = ref(""); // 哈希文件名称
    const selectedAlgorithm = ref("SHA-1"); // 算法选择
    const outputLengthStr = ref('') // 可变输出长度
    const verificationResult = ref(null); // 验证结果
    const actualHash = ref(""); // 文件计算出的实际哈希值
    const showResult = ref(false); // 弹窗显示状态
    const userFileList = ref([]); // 存储上传的文件列表
    const hashFileList = ref([]); // 存储上传的哈希文件列表
    let timeoutId = null; // 自动关闭弹窗的计时器
    // 定义支持 outputLength 的算法及其合法范围和推荐值
    const outputLengthConfig = {
        'BLAKE3': { min: 1, max: 32, common: [32, 16] },
        'BLAKE2B': { min: 1, max: 64, common: [64, 48, 32] },
        'BLAKE2S': { min: 1, max: 32, common: [32, 24, 16] },
        'XXH3': { min: 1, max: 16, common: [16, 8] },
    }



    // 是否支持可变长度输出、控制输入框显示/隐藏
    const showOutputLength = computed(() => selectedAlgorithm.value in outputLengthConfig)

    // 计算 outputLength 范围
    const outputLengthRange = computed(() => {
        return outputLengthConfig[selectedAlgorithm.value] || { min: 1, max: 32 }
    })

    const currentCommonLengths = computed(() => {
        return outputLengthConfig[selectedAlgorithm.value]?.common || []
    })

    // 自动设置默认 outputLength
    const onAlgorithmChange = () => {
        if (showOutputLength.value) {
            const defaultVal = outputLengthConfig[selectedAlgorithm.value].common[0]
            outputLengthStr.value = String(defaultVal)
        } else {
            outputLengthStr.value = ''
        }
    }

    // 如果你需要传入给 createHasher 的值：
    const outputLength = computed(() => {
        const num = parseInt(outputLengthStr.value)
        const { min, max } = outputLengthRange.value
        return (num >= min && num <= max) ? num : undefined
    })
    const createHasher = async (algorithm, options = {}) => {
        const outputLengthtmp = outputLength.value;  // 可变输出长度

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

    // 处理文件选择
    const handleFileChange = (file, fileList) => {
        if (file && file.raw) {
            userFile.value = file.raw;  // 只保留 File 对象
            fileName.value = file.name;
            userFileList.value = fileList;
        }
    };

    // 处理哈希文件选择
    const handleHashFileChange = async (file, fileList) => {
        if (file && file.raw) {
            hashFileName.value = file.name;
            hashFileList.value = fileList;

            const reader = new FileReader();
            // 文件读取成功后处理
            reader.onload = () => {
                const hashText = reader.result?.toString(); // 获取文件内容
                if (hashText) {
                    expectedHash.value = hashText.trim().toUpperCase();  // 去掉空白字符并赋值给预期哈希
                    console.log("文件内容已读取并赋值给 expectedHash", expectedHash.value);
                }
            };
            // 读取文件内容为文本
            reader.readAsText(file.raw);
        }
    };



    // 移除文件时的处理
    const handleFileRemove = () => {
        userFile.value = null;
        fileName.value = "";
        userFileList.value = []; // 清空文件列表
    };

    // 移除哈希文件时的处理
    const handleHashFileRemove = () => {
        expectedHash.value = "";
        hashFileName.value = "";
        hashFileList.value = []; // 清空哈希文件列表
    };

    // 验证文件完整性
    const checkIntegrity = async () => {
        if (!userFile.value || !expectedHash.value) {
            ElMessageBox.alert("请确保已选择文件并提供预期哈希值！", "提示", {
                confirmButtonText: "确定",
            });
            return;
        }

        const actualHashValue = await calculateHash(userFile.value);
        actualHash.value = actualHashValue.toUpperCase();
        verificationResult.value = actualHashValue.toUpperCase() === expectedHash.value.toUpperCase();

        // 显示弹窗并设置自动关闭
        showResult.value = true;
        if (timeoutId) clearTimeout(timeoutId); // 清除之前的计时器
        timeoutId = setTimeout(() => {
            closeModal();
        }, 3000);
    };

    // 动态计算文件哈希值
    const calculateHash = async (fileBuffer) => {
        try {
            const algorithm = selectedAlgorithm.value;
            const hasher = await createHasher(algorithm);
            hasher.init();

            // 如果传入的是 ArrayBuffer，小型直接计算
            if (fileBuffer instanceof ArrayBuffer) {
                hasher.update(new Uint8Array(fileBuffer));
                return hasher.digest();
            }

            // 如果传入的是 File 对象，大文件按块读取
            if (fileBuffer instanceof File) {
                const totalChunks = Math.ceil(fileBuffer.size / CHUNK_SIZE);
                for (let i = 0; i < totalChunks; i++) {
                    const start = i * CHUNK_SIZE;
                    const end = Math.min(fileBuffer.size, start + CHUNK_SIZE);
                    const blob = fileBuffer.slice(start, end);

                    const chunkBuffer = await blob.arrayBuffer();
                    hasher.update(new Uint8Array(chunkBuffer));
                    // 延迟处理，避免阻塞主线程
                    await new Promise((resolve) => setTimeout(resolve, 0));
                }
                return hasher.digest();
            }
        } catch (error) {
            console.error("哈希计算出错:", error);
            ElMessageBox.alert(`哈希计算出错: ${error.message}`, "错误", {
                confirmButtonText: "确定",
            });
        }
    };



    // 手动关闭弹窗
    const closeModal = () => {
        showResult.value = false;
        if (timeoutId) clearTimeout(timeoutId);
    };

    const handleExceed = () => {
        ElMessageBox.alert("请确保只上传一个文件！", "提示", {
            confirmButtonText: "确定",
        });
    }

    // 校验可变输出长度值是否合法
    function validateOutputLength(algorithm, length) {
        const algo = algorithm.toUpperCase();

        switch (algo) {
            case 'XXH3-64':
                if (length === 8) return { valid: true, recommended: 8 };
                return { valid: false, recommended: 8 };

            case 'XXH3-128':
                if (length === 16) return { valid: true, recommended: 16 };
                return { valid: false, recommended: 16 };

            case 'BLAKE2B':
                if (Number.isInteger(length) && length >= 1 && length <= 64) {
                    return { valid: true, recommended: 64 };
                }
                return { valid: false, recommended: 64 };

            case 'BLAKE2S':
                if (Number.isInteger(length) && length >= 1 && length <= 32) {
                    return { valid: true, recommended: 32 };
                }
                return { valid: false, recommended: 32 };

            case 'BLAKE3':
                if (Number.isInteger(length) && length >= 1) {
                    return { valid: true, recommended: 32 };
                }
                return { valid: false, recommended: 32 };

            default:
                // 不支持 outputLength 参数的算法默认返回不校验
                return { valid: false, recommended: null };
        }
    }
</script>

<style scoped>
    .section {
        margin-bottom: 20px;
    }

    .file-info {
        margin-top: 10px;
    }

    .or {
        text-align: center;
        margin: 10px 0;
    }

    .success {
        color: green;
    }

    .error {
        color: red;
    }
</style>