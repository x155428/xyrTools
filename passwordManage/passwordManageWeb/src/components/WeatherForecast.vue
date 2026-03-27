<!--顶部走马灯样式天气组件-->
<template>
  <div class="weather-forecast">
    <!-- 左侧时间 -->
    <div class="left">
      <span class="time">{{ currentTime }}</span>
    </div>

    <!-- 中间当前天气信息 -->
    <div class="center">
      <div class="current-city">{{ city }}</div>
      <div class="current-condition">
        <span class="weather">天气：{{ weather }}</span>
        <span class="temperature">| 温度：{{ temperature }}℃</span>
        <span class="wind">| 风向：{{ windDirection }} | 风力：{{ windPower }}级</span>
        <span class="humidity">| 湿度：{{ humidity }}%</span>
      </div>
    </div>

    <!-- 右侧未来三天天气横向滚动展示（精简为两行） -->
    <div class="right">
      <div class="forecast-container">
        <div class="forecast">
          <div v-for="(forecast, index) in weatherData" :key="index" class="forecast-item">
            <div class="date">{{ formatForecastDate(forecast.date).replace('月', '/').replace('日', '') }}({{
              forecast.week }})</div>
            <div class="weather-info">
              <span class="weather-desc">{{ forecast.dayweather }}</span>
              <span class="temp-range">{{ forecast.daytemp }}/{{ forecast.nighttemp }}℃</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
  import { ref, onMounted, onUnmounted } from 'vue';

  const city = ref('');
  const currentTime = ref('');
  const weather = ref('');
  const temperature = ref('');
  const windDirection = ref('');
  const windPower = ref('');
  const humidity = ref('');
  const weatherData = ref([]);
  //const apiKey = '3930d8582d268e14fa2a4ac576430c73';  // 替换高德地图API Key
  // 从本地存储读取天气API key
  const apiKey = ref(localStorage.getItem('weatherApiKey') || '');
  var t;

  // 获取当前城市的adcode
  const fetchCityAdcode = async () => {
    try {
      const response = await fetch(`https://restapi.amap.com/v3/ip?key=${apiKey.value}`);
      const data = await response.json();
      if (data.status === '1') {
        city.value = `${data.province} ${data.city}`;
        fetchWeatherData(data.adcode);
      } else {
        console.error('无法获取城市信息');
      }
    } catch (error) {
      console.error('获取城市失败', error);
    }
  };

  // 获取天气信息
  const fetchWeatherData = async (adcode) => {

    try {
      const response2 = await fetch(`https://restapi.amap.com/v3/weather/weatherInfo?key=${apiKey.value}&city=${adcode}&extensions=base`);
      const data2 = await response2.json();
      if (data2.status) {
        const currentWeather2 = data2.lives[0];
        humidity.value = currentWeather2.humidity;  //湿度
        weather.value = currentWeather2.weather;  // 白天气象
        temperature.value = currentWeather2.temperature;  // 白天温度
        windDirection.value = currentWeather2.winddirection;  // 白天风向
        windPower.value = currentWeather2.windpower;  // 白天风力
      }
    } catch (error) {
      console.error('获取实时天气基础信息失败', error);
    }

    try {
      const response = await fetch(`https://restapi.amap.com/v3/weather/weatherInfo?key=${apiKey.value}&city=${adcode}&extensions=all`);
      const data = await response.json();
      if (data.forecasts) {
        // 当前天气解析
        const currentWeather = data.forecasts[0];
        const casts = currentWeather.casts[0];  // 当前天气的第一天数据
        weather.value = casts.dayweather;  // 白天气象
        temperature.value = casts.daytemp;  // 白天温度
        windDirection.value = casts.daywind;  // 白天风向
        windPower.value = casts.daypower;  // 白天风力

        // 未来三天天气
        weatherData.value = data.forecasts[0].casts.slice(1, 4);  // 切片，未来三天，忽略当天
        updateTime();
      } else {
        console.error('无法获取未来天气数据');
      }
    } catch (error) {
      console.error('获取未来天气失败', error);
    }
  };

  // 实时更新时间
  // 格式化日期，只显示月日
  const formatForecastDate = (dateStr) => {
    const date = new Date(dateStr);
    return `${date.getMonth() + 1}月${date.getDate()}日`;
  };

  const updateTime = () => {
    const now = new Date();

    // 获取详细的日期和时间
    const options = {
      weekday: 'long',  // 星期几
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    };

    currentTime.value = now.toLocaleString('zh-CN', options);
  };

  onMounted(() => {
    fetchCityAdcode();
    t = setInterval(updateTime, 1000); // 每秒更新一次时间

    // 监听localStorage中weatherApiKey的变化（跨窗口）
    window.addEventListener('storage', handleStorageChange);

    // 监听自定义事件（同窗口）
    window.addEventListener('weatherApiKeyChanged', handleWeatherApiKeyChanged);
  });

  // 处理localStorage变化的函数（跨窗口）
  const handleStorageChange = (event) => {
    if (event.key === 'weatherApiKey') {
      // 更新API key并重新获取天气数据
      apiKey.value = event.newValue || '';
      fetchCityAdcode();
    }
  };

  // 处理自定义事件的函数（同窗口）
  const handleWeatherApiKeyChanged = (event) => {
    // 更新API key并重新获取天气数据
    apiKey.value = event.detail.newKey || '';
    fetchCityAdcode();
  };

  onUnmounted(() => {
    clearInterval(t);
    // 移除事件监听器
    window.removeEventListener('storage', handleStorageChange);
    window.removeEventListener('weatherApiKeyChanged', handleWeatherApiKeyChanged);
  });
</script>

<style scoped>
  /* 使用base.css中定义的变量 */
  .weather-forecast {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    height: 100%;
    padding: 0 var(--space-sm);
    border-radius: var(--radius-md);
    overflow: hidden;
    font-family: 'Microsoft YaHei', Arial, sans-serif;
    font-size: var(--font-size-sm);
    line-height: 24px;
    color: var(--text-main);
  }

  .left {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 150px;
    margin-right: var(--space-sm);
    background: transparent;
  }

  .time {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    color: var(--text-secondary);
    padding: var(--space-xs) var(--space-sm);
    border-radius: var(--radius-sm);
    background: transparent;
  }

  .center {
    flex-grow: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    background: transparent;
  }

  .current-city {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    color: var(--primary);
    margin-right: var(--space-md);
    background: transparent;
  }

  .current-condition {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-normal);
    color: var(--text-secondary);
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    justify-content: center;
    background: transparent;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .current-condition span {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .right {
    display: flex;
    align-items: center;
    min-width: 180px;
    margin-left: var(--space-sm);
    background: transparent;
    overflow: hidden;
  }

  .forecast-container {
    width: 100%;
    overflow-x: auto;
    overflow-y: hidden;
    scrollbar-width: none;
    /* Firefox隐藏滚动条 */
    -ms-overflow-style: none;
    /* IE隐藏滚动条 */
  }

  .forecast-container::-webkit-scrollbar {
    display: none;
    /* Chrome/Safari隐藏滚动条 */
  }

  .forecast {
    display: flex;
    gap: var(--space-sm);
    padding: var(--space-xs) 0;
    width: max-content;
  }

  .forecast-item {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    background: transparent;
    padding: var(--space-xs) var(--space-xs);
    border-radius: var(--radius-sm);
    box-shadow: none;
    color: var(--text-secondary);
    min-width: 50px;
    height: 30px;
    text-align: center;
    transition: all 0.3s ease;
  }

  .forecast-item:hover {
    background: var(--bg-primary);
    transform: translateY(0);
    box-shadow: none;
  }

  .forecast-item .date {
    font-weight: var(--font-weight-medium);
    font-size: calc(var(--font-size-sm) - 2px);
    margin-bottom: 0;
    white-space: nowrap;
    text-align: center;
    line-height: 14px;
  }

  .weather-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0;
  }

  .weather-desc {
    font-size: calc(var(--font-size-sm) - 2px);
    text-align: center;
    line-height: 12px;
  }

  .temp-range {
    font-size: calc(var(--font-size-sm) - 2px);
    font-weight: var(--font-weight-medium);
    line-height: 12px;
  }

  .temp-range::before {
    display: none;
  }

  /* 横向滚动条美化（隐藏但功能保留） */
  .forecast-container {
    scroll-behavior: smooth;
  }
</style>