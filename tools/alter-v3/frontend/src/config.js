// API配置
export const API_CONFIG = {
  // 生产环境使用相对路径，开发环境使用本地地址
  BASE_URL: process.env.NODE_ENV === 'production' 
    ? '' // 生产环境使用相对路径，通过nginx代理
    : (process.env.REACT_APP_API_URL || 'http://127.0.0.1:8081'), // 开发环境
  ENDPOINTS: {
    CONFIG: '/api/config'
  }
};

// 获取完整的API URL
export const getApiUrl = (endpoint) => {
  return `${API_CONFIG.BASE_URL}${endpoint}`;
};
