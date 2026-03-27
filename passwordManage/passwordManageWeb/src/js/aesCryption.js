import tools from './tools.js';

export async function aesGcmEncrypt(data, key) {
    const iv = tools.generateRandomIV();
    // 判断密钥传入类型，增加处理
    let secretKeyBuffer;
    if (typeof key === 'string') {
        secretKeyBuffer = tools.hexToArrayBuffer(key);  // hex字符串密钥转ArrayBuffer
    } else if (key instanceof ArrayBuffer) {
        secretKeyBuffer = key;
    } else if (key instanceof Uint8Array) {
        secretKeyBuffer = key.buffer;
    } else {
        throw new TypeError('不支持加密的密钥类型。支持传入arrayBuffer和string类型。');
    }
    // 导入密钥
    const aesKey = await crypto.subtle.importKey(
        'raw',
        secretKeyBuffer,
        { name: 'AES-GCM' },
        false,
        ['encrypt']
    );

    var dataBuffer;
    // 将输入数据转换为 ArrayBuffer
    if (typeof data === 'string') {
        // 如果是字符串，使用 TextEncoder 进行转换
        dataBuffer = new TextEncoder().encode(data);
    } else if (data instanceof ArrayBuffer) {
        // 如果是 ArrayBuffer，直接赋值
        dataBuffer = data;
    } else {
        throw new TypeError('不支持加密的数据类型。支持传入arrayBuffer和string类型。');
    }

    // 使用 AES-GCM 加密
    const encryptedData = await crypto.subtle.encrypt(
        { name: 'AES-GCM', iv: iv },
        aesKey,
        dataBuffer
    );

    // 打印加密数据，调试用
    // const tagLength = 16; // GCM 标签长度通常是 16 字节
    // const ciphertextLength = encryptedData.byteLength - tagLength;
    // const ciphertext = encryptedData.slice(0, ciphertextLength); // 提取密文
    // const tag = encryptedData.slice(ciphertextLength); // 提取标签
    // console.log("Ciphertext:", tools.arrayBufferToBase64(ciphertext));
    // console.log("Tag:", tools.arrayBufferToBase64(tag));
    // console.log("IV:", tools.arrayBufferToBase64(iv));

    // 返回加密后的数据和 IV
    return {
        iv: tools.arrayBufferToBase64(iv),
        data: tools.arrayBufferToBase64(encryptedData)
    };
}


// AES-GCM 解密函数
export async function aesGcmDecrypt(encryptedDataJson, key) {
    try {
        // 解析json数据，解析出iv和密文data
        const encryptedJson = JSON.parse(encryptedDataJson);
        // iv
        const iv = tools.base64ToArrayBuffer(encryptedJson.iv);
        // 密文data
        const data = tools.base64ToArrayBuffer(encryptedJson.data);

        // 导入密钥
        const secretKeyBuffer = tools.hexToArrayBuffer(key);
        const aesKey = await crypto.subtle.importKey(
            "raw",
            secretKeyBuffer,
            { name: "AES-GCM" },
            false,
            ["decrypt"]
        );

        const decryptedData = await crypto.subtle.decrypt(
            {
                name: "AES-GCM",
                iv: iv
            },
            aesKey,
            data
        );
        return decryptedData;
    } catch (error) {
        console.error("Data integrity check failed or incorrect key/iv!");
        throw error;
    }
}