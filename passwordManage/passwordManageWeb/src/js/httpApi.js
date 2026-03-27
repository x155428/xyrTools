// 封装一个测试api地址的httpTest对象

import axios from 'axios';
const httpTest = axios.create({
     baseURL: 'https://xiaoyudemo.com:9091',
     timeout: 10000,
     withCredentials: true,
     headers: {
          'X-Requested-With': 'XMLHttpRequest',
     },
});

export default httpTest
