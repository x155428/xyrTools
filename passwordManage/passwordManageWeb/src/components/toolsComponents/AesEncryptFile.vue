<template>
  <div class="aes-encryptor">
    <h2>AES 文件加密/解密工具</h2>

    <!-- 密钥输入 -->
    <div class="key-input">
      <el-input v-model="customKey" placeholder="请输入十六进制密钥或留空使用默认密钥" clearable></el-input>
    </div>

    <!-- 文本输入 -->
    <div class="text-input">
      <el-input type="textarea" v-model="textContent" placeholder="请输入需要加密或解密的文本" rows="4"></el-input>
      <div class="text-actions">
        <el-button @click="encryptText" type="primary">加密文本</el-button>
        <el-button @click="decryptText" type="danger">解密文本</el-button>
      </div>
    </div>

    <!-- 文件上传 -->
    <el-upload class="upload-demo" :before-upload="handleBeforeUpload" :file-list="fileList" drag>
      <i class="el-icon-upload"></i>
      <div class="el-upload__text">将文件拖到此处，或点击上传</div>
      <div class="el-upload__tip">支持文件加密或解密</div>
    </el-upload>

    <!-- 文件列表 -->
    <div v-if="uploadedFiles.length > 0" class="file-list">
      <h3>文件列表</h3>
      <ul>
        <li v-for="(file, index) in uploadedFiles" :key="index">
          <div class="file-name">{{ file.name }}</div>
          <div class="file-actions">
            <el-button @click="encryptFile(file)" type="primary" size="small">加密</el-button>
            <el-button @click="decryptFile(file)" type="danger" size="small">解密</el-button>
            <el-button @click="saveFileToDirectory(file)" type="success" size="small">保存</el-button>
            <el-button @click="deleteFile(index)" type="warning" size="small">删除</el-button>
          </div>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup>
  import { ref } from "vue";
  import { ElMessage, ElButton } from "element-plus";

  const defaultKey = "0123456789abcdef0123456789abcdef"; // 默认密钥
  const customKey = ref(""); // 用户输入的密钥
  const uploadedFiles = ref([]); // 上传的文件列表
  const fileList = ref([]); // Element Plus 显示的文件列表
  const textContent = ref(""); // 输入的文本内容

  // 获取有效密钥
  function getEncryptionKey() {
    const key = customKey.value.trim();
    if (key) {
      if (key.length !== 32) {
        ElMessage({
          message: "密钥必须是 32 位十六进制字符串！",
          type: 'error',
          grouping: true,
        })
        return null;
      }
      return new TextEncoder().encode(key);
    }
    return new TextEncoder().encode(defaultKey);
  }

  // 处理文件上传
  async function handleBeforeUpload(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        uploadedFiles.value.push({
          name: file.name,
          content: new Uint8Array(reader.result), // 文件内容
          isEncrypted: false, // 初始状态未加密
        });
        ElMessage({
          message: `文件 ${file.name} 已上传！`,
          type: 'success',
          grouping: true,
        })
        resolve(false); // 阻止自动上传
      };
      reader.onerror = () => {
        ElMessage({
          message: `无法读取文件 ${file.name}`,
          type: 'error',
          grouping: true,
        })
        reject(false);
      };
      reader.readAsArrayBuffer(file); // 读取为 ArrayBuffer
    });
  }

  // 加密文件
  async function encryptFile(file) {
    const key = getEncryptionKey();
    if (!key) return;

    try {
      const iv = crypto.getRandomValues(new Uint8Array(12)); // 生成随机 IV
      const cryptoKey = await crypto.subtle.importKey(
        "raw",
        key,
        { name: "AES-GCM" },
        false,
        ["encrypt"]
      );
      const encryptedContent = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv },
        cryptoKey,
        file.content
      );

      // 更新文件状态
      file.content = new Uint8Array([...iv, ...new Uint8Array(encryptedContent)]);
      file.name = `${file.name}.enc`;
      file.isEncrypted = true;

      ElMessage({
        message: `文件 ${file.name} 加密成功！`,
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      console.error("加密失败：", error);
      ElMessage({
        message: `文件 ${file.name} 加密失败！`,
        type: 'error',
        grouping: true,
      })
    }
  }

  // 解密文件
  async function decryptFile(file) {
    const key = getEncryptionKey();
    if (!key) return;

    try {
      const iv = file.content.slice(0, 12); // 提取前 12 字节的 IV
      const encryptedContent = file.content.slice(12); // 剩余为加密内容

      const cryptoKey = await crypto.subtle.importKey(
        "raw",
        key,
        { name: "AES-GCM" },
        false,
        ["decrypt"]
      );
      const decryptedContent = await crypto.subtle.decrypt(
        { name: "AES-GCM", iv },
        cryptoKey,
        encryptedContent
      );

      // 更新文件状态
      file.content = new Uint8Array(decryptedContent);
      file.name = file.name.replace(/\.enc$/, ""); // 去掉 `.enc` 后缀
      file.isEncrypted = false;

      ElMessage({
        message: `文件 ${file.name} 解密成功！`,
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      console.error("解密失败：", error);
      ElMessage({
        message: `文件 ${file.name} 解密失败，请检查密钥！`,
        type: 'error',
        grouping: true,
      })
    }
  }

  // 保存文件
  async function saveFileToDirectory(file) {
    try {
      const blob = new Blob([file.content], {
        type: file.isEncrypted ? "application/octet-stream" : "text/plain",
      });
      const link = document.createElement("a");
      link.href = URL.createObjectURL(blob);
      link.download = file.name;
      link.click();
      URL.revokeObjectURL(link.href);

      ElMessage({
        message: `文件 ${file.name} 已保存！`,
        type: 'success',
        grouping: true,
      })
    } catch (error) {
      console.error("保存文件失败：", error.message);
      ElMessage({
        message: `保存文件失败，请检查或重试！`,
        type: 'error',
        grouping: true,
      })
    }
  }

  // 删除文件
  function deleteFile(index) {
    uploadedFiles.value.splice(index, 1); // 从文件列表中删除文件
    ElMessage({
      message: `文件已删除！`,
      type: 'success',
      grouping: true,
    })
  }

  // 加密文本
  function encryptText() {
    const key = getEncryptionKey();
    if (!key || textContent.value === "") {
      ElMessage({
        message: "文本为空或密钥错误！",
        type: 'error',
        grouping: true,
      })
      return;
    }

    try {
      const iv = crypto.getRandomValues(new Uint8Array(12));

      const cryptoKey = crypto.subtle.importKey(
        "raw",
        key,
        { name: "AES-GCM" },
        false,
        ["encrypt"]
      );

      cryptoKey.then(key => {
        return crypto.subtle.encrypt(
          { name: "AES-GCM", iv },
          key,
          new TextEncoder().encode(textContent.value)
        );
      }).then(encryptedContent => {
        textContent.value = btoa(String.fromCharCode(...new Uint8Array([...iv, ...new Uint8Array(encryptedContent)])));
        ElMessage({
          message: "文本加密成功！",
          type: 'success',
          grouping: true,
        })
      });
    } catch (error) {
      ElMessage({
        message: "文本加密失败，请检查密钥！",
        type: 'error',
        grouping: true,
      })
    }
  }

  // 解密文本
  function decryptText() {
    const key = getEncryptionKey();
    if (!key || textContent.value === "") {
      ElMessage({
        message: "文本为空或密钥错误！",
        type: 'error',
        grouping: true,
      })
      return;
    }

    try {
      const encryptedData = atob(textContent.value);
      const iv = new Uint8Array(encryptedData.slice(0, 12).split("").map(c => c.charCodeAt(0)));
      const encryptedContent = new Uint8Array(encryptedData.slice(12).split("").map(c => c.charCodeAt(0)));

      const cryptoKey = crypto.subtle.importKey(
        "raw",
        key,
        { name: "AES-GCM" },
        false,
        ["decrypt"]
      );

      cryptoKey.then(key => {
        return crypto.subtle.decrypt(
          { name: "AES-GCM", iv },
          key,
          encryptedContent
        );
      }).then(decryptedContent => {
        textContent.value = new TextDecoder().decode(decryptedContent);
        ElMessage({
          message: "文本解密成功！",
          type: 'success',
          grouping: true,
        })
      });
    } catch (error) {
      ElMessage({
        message: "文本解密失败，请检查密钥！",
        type: 'error',
        grouping: true,
      })
    }
  }
</script>

<style scoped>
  .aes-encryptor {
    font-family: Arial, sans-serif;
    max-width: 800px;
    margin: 20px auto;
    padding: 20px;
    border: 1px solid var(--border-color-light);
    border-radius: 8px;
    box-shadow: var(--shadow-sm);
  }

  .key-input,
  .text-input {
    margin-bottom: 20px;
  }

  .text-actions {
    margin-top: 10px;
    display: flex;
    justify-content: space-between;
  }

  .upload-demo {
    margin-top: 20px;
  }

  .file-list ul {
    list-style: none;
    padding: 0;
  }

  .file-list li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin: 10px 0;
  }

  .file-actions {
    display: flex;
    gap: 10px;
  }

  button,
  .el-button {
    background: #007bff;
    color: white;
    border: none;
    padding: 5px 10px;
    border-radius: 4px;
    cursor: pointer;
  }

  button:hover,
  .el-button:hover {
    background: #0056b3;
  }
</style>