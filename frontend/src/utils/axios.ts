import axios from 'axios'

const axiosInstance = axios.create({
  baseURL: '',
  timeout: 10000,
})

// 请求拦截器 - 自动添加 token
axiosInstance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理错误
axiosInstance.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response) {
      // 处理 401 错误
      if (error.response.status === 401) {
        localStorage.removeItem('token')
        window.location.href = '/login'
      }
      
      // 优先使用后端返回的错误信息
      const backendMessage = error.response.data?.message
      if (backendMessage) {
        error.message = backendMessage
      }
    }
    return Promise.reject(error)
  }
)

export default axiosInstance