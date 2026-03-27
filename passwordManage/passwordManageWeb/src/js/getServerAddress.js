export function getServerAddress() {
    let serverAddress = localStorage.getItem('savedServerAddress');
    let serverPort = localStorage.getItem('savedServerPort');

    // 如果本地存储没有服务器地址，则使用当前页面的 URL
    if (!serverAddress) {
        const currentUrl = window.location.href; // 获取当前页面的 URL
        const url = new URL(currentUrl); // 创建 URL 对象

        // 使用从 localStorage 获取的端口，如果未设置则使用默认端口 8443
        url.port = serverPort || 8443; // 设置端口
        serverAddress = `${url.protocol}//${url.hostname}:${url.port}`; // 组合 URL
    } else {
        // 如果已存在 serverAddress，则使用它，并确保使用端口
        serverAddress = `https://${serverAddress}:${serverPort || 8443}`;
    }

    return serverAddress;
}