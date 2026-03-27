import { fromByteArray, toByteArray } from 'base64-js';
export default {
    // 将十六进制字符串转换为 ArrayBuffer
    hexToArrayBuffer(hex) {
        // 检查输入是否为字符串
        if (typeof hex !== 'string') {
            throw new Error(`输入必须是字符串类型，当前输入类型是 ${typeof hex}`);
        }

        // 去除首尾空格
        hex = hex.trim();

        // 检查输入是否为空
        if (hex.length === 0) {
            throw new Error('输入的十六进制字符串不能为空');
        }

        // 检查输入是否为有效的十六进制字符串
        if (!/^[0-9a-fA-F]+$/.test(hex)) {
            throw new Error('输入的字符串不是有效的十六进制字符串');
        }

        // 检查输入长度是否为偶数
        if (hex.length % 2 !== 0) {
            throw new Error('输入的十六进制字符串长度必须为偶数');
        }

        const buffer = new ArrayBuffer(hex.length / 2);
        const view = new Uint8Array(buffer);

        try {
            for (let i = 0; i < hex.length; i += 2) {
                const byte = parseInt(hex.slice(i, i + 2), 16);
                if (isNaN(byte)) {
                    throw new Error(`无法解析十六进制字符: ${hex.slice(i, i + 2)}`);
                }
                view[i / 2] = byte;
            }
        } catch (error) {
            console.error('解析十六进制字符串时出错:', error);
            throw new Error('解析十六进制字符串时发生错误，请检查输入');
        }

        return buffer;
    },

    // 将 ArrayBuffer 转换为 Base64 字符串
    arrayBufferToBase64(buffer) {
        return fromByteArray(new Uint8Array(buffer));
    },

    // 将 Base64 字符串转换为 ArrayBuffer
    base64ToArrayBuffer(base64) {
        return toByteArray(base64).buffer;
    },

    // 生成随机 IV (12 字节) 用于 AES-GCM
    generateRandomIV() {
        const iv = new Uint8Array(12); // GCM 模式要求 12 字节的 IV
        window.crypto.getRandomValues(iv);
        // 检查，如果为0再重新生成
        if (iv.every(byte => byte === 0)) {
            return this.generateRandomIV();
        }
        return iv;
    },

    // 生成一个随机 AES 密钥（256 位）
    generateAESKey() {
        // 使用 CryptoJS 随机生成 256 位密钥
        const key = CryptoJS.lib.WordArray.random(256 / 8);  // 256 位 / 8 = 32 字节
        return key.toString(CryptoJS.enc.Hex); // 转换为十六进制字符串
    },

    printHex(aesKey) {
        // 将字节数组转换为十六进制字符串
        let hexString = Array.from(aesKey)
            .map(byte => byte.toString(16).padStart(2, '0'))
            .join('');

        // 打印十六进制字符串
        console.log("生成的 AES 密钥:", hexString);
    }
}
