import http from '@/js/http.js'
export default {
    //获取aeskey
    async getAesKey() {
        //密钥交换
        // 创建密钥对
        const clientKeyPair = await this.generateEccKeyPair();
        //公钥导出为base64，spki格式
        const clientPublicKeyBase64 = await this.exportPublicEccKey(clientKeyPair.publicKey);
            // 发送公钥到服务器
            const response = await http.post("/getAesKey", {
                clientPublicKeyBase64
            }, {
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            // 解析服务器的公钥
            const serverPublicKeyBase64 = response.data.data;
            // 导入服务器的公钥，创建公钥对象
            const serverPublicKey = await this.importPublicKey(serverPublicKeyBase64);

            // 计算共享秘密
            const sharedSecret = await this.deriveSharedSecret(clientKeyPair.privateKey, serverPublicKey);
            //console.log("Shared Secret (Hex):", this.toHexString(sharedSecret));
            const sharedSecretString = String.fromCharCode.apply(null, sharedSecret);

            // 生成 AES 密钥
            const aesKey = await this.deriveAesKey(sharedSecret);
            //计算key的hash，调试用
            //const hash1 = await crypto.subtle.digest("SHA-256", aesKey);
            //console.log("Hash of AES Key:", this.toHexString(new Uint8Array(hash1)));
            return aesKey;
    },

    //创建ecc密钥对
    async generateEccKeyPair() {
        const curve = "P-521";
        const keyPair = await crypto.subtle.generateKey(
            { name: "ECDH", namedCurve: curve },
            true,
            ["deriveBits"]
        );
        return keyPair;
    },

    // 导出公钥
    async exportPublicEccKey(publicKey) {
        const publicKeyBuffer = await crypto.subtle.exportKey("spki", publicKey);
        return btoa(String.fromCharCode.apply(null, new Uint8Array(publicKeyBuffer)));
    },
    // 导入公钥
    async importPublicKey(publicKeyBase64) {
        const publicKeyBuffer = Uint8Array.from(atob(publicKeyBase64), c => c.charCodeAt(0));
        return crypto.subtle.importKey(
            "spki",
            publicKeyBuffer,
            { name: "ECDH", namedCurve: "P-521" },
            true,
            []
        );
    },

    // 计算共享秘密
    async deriveSharedSecret(privateKey, serverPublicKey) {
        const sharedBits = await window.crypto.subtle.deriveBits(
            {
                name: "ECDH",
                public: serverPublicKey
            },
            privateKey,
            521 // 使用 521 位的共享秘密
        );
        return new Uint8Array(sharedBits); // 返回字节数组
    },

    // 生成AES密钥
    async deriveAesKey(sharedSecret) {
        const hash = await crypto.subtle.digest("SHA-256", sharedSecret);
        const hashArray = new Uint8Array(hash);
        const aesKey = hashArray.slice(0, 32); // 取前 32 字节作为 AES-256 密钥，切片保证是32字节
        return aesKey;
    },

    //辅助函数，转成16进制打印
    toHexString(byteArray) {
        return Array.from(byteArray, byte => byte.toString(16).padStart(2, '0')).join('');
    },

    // 解密数据
    async decryptAES(encryptedData, key, iv) {
        try {
            // 导入 AES 密钥
            const aesKey = await crypto.subtle.importKey(
                'raw', // 原始格式
                key, // 密钥数据
                { name: 'AES-GCM' }, // 使用 AES-GCM 模式
                false, // 不需要导出
                ['decrypt'] // 只需要解密功能
            );

            // GCM 模式认证标签的长度是 16 字节
            const authTagLength = 16;

            // 提取密文和认证标签
            const cipherText = encryptedData.slice(0, encryptedData.length - authTagLength);
            const authTag = encryptedData.slice(encryptedData.length - authTagLength);

            // 认证标签和密文合并进行解密
            const fullData = new Uint8Array([...cipherText, ...authTag]);

            // 解密数据
            const decryptedBuffer = await crypto.subtle.decrypt(
                {
                    name: 'AES-GCM',
                    iv: iv,  // 使用给定的 IV
                    tagLength: authTagLength * 8, // 认证标签长度（单位是位）
                },
                aesKey,
                fullData // 密文加认证标签
            );

            // 将解密后的 ArrayBuffer 转换为字符串
            const decryptedText = new TextDecoder().decode(decryptedBuffer);
            return decryptedText;
        } catch (error) {
            console.error('解密失败:', error);
            throw new Error('解密失败，请检查密钥、数据或认证标签是否正确');
        }

    },

    // 加密数据
    async encryptData(jsonData, tmpAesKey) {
        // 生成随机 IV 向量（长度为 12 字节，AES-GCM模式加密）
        const iv = window.crypto.getRandomValues(new Uint8Array(12));

        // 将 jsonData 转换为字节数组
        const encoder = new TextEncoder();
        //console.log(jsonData)
        const dataToEncrypt = encoder.encode(jsonData);

        // 将 Uint8Array密钥导入为CryptoKey
        const aesCryptoKey = await window.crypto.subtle.importKey(
            "raw",                     // 密钥格式为原始字节（Uint8Array）
            tmpAesKey,                 // Uint8Array 格式的密钥
            { name: "AES-GCM" },       // AES-GCM 加密模式
            true,                      // 可导出（根据需求设置）
            ["encrypt", "decrypt"]     // 允许加密和解密操作
        );

        // 使用 tmpAesKey 加密 jsonData
        const encryptedData = await crypto.subtle.encrypt(
            {
                name: "AES-GCM",
                iv: iv, // 使用生成的 IV 向量
                tagLength: 128 // 默认 128 位（16字节）验证标签长度
            },
            aesCryptoKey, // AES 密钥
            dataToEncrypt // 数据
        );

        // 转换密文格式为Uint8Array，便于传输
        const encryptedDataArray = new Uint8Array(encryptedData);
        const ivArray = Array.from(iv);

        // 返回IV和密文数据
        return {
            iv: ivArray, // IV 向量，AES-GCM模式加密    
            encryptedData: Array.from(encryptedDataArray) // 转换为数组便于 JSON 序列化
        };
    }
}