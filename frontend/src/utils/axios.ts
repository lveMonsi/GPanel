import axios from 'axios'

const axiosInstance = axios.create({
  baseURL: '',
  timeout: 10000,
})

const getReadableErrorMessage = (error: any): string | undefined => {
  if (error?.code === 'ECONNABORTED' || error?.message?.includes('timeout')) {
    return '请求超时，请稍后重试'
  }

  if (error?.message === 'Network Error') {
    return '网络连接失败，请检查网络后重试'
  }

  return undefined
}

const readBlobMessage = async (data: Blob): Promise<string | undefined> => {
  if (!data.type.includes('application/json')) {
    return undefined
  }

  try {
    const text = await data.text()
    const parsed = JSON.parse(text) as { message?: string }
    return parsed.message
  } catch {
    return undefined
  }
}

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
  async (error) => {
    if (error.response) {
      // 处理 401 错误
      if (error.response.status === 401) {
        localStorage.removeItem('token')
        window.location.href = '/login'
      }

      // 优先使用后端返回的错误信息
      let backendMessage = error.response.data?.message
      if (!backendMessage && error.response.data instanceof Blob) {
        backendMessage = await readBlobMessage(error.response.data)
      }
      if (backendMessage) {
        error.message = backendMessage
        return Promise.reject(error)
      }
    }

    const readableMessage = getReadableErrorMessage(error)
    if (readableMessage) {
      error.message = readableMessage
    }

    return Promise.reject(error)
  }
)

export default axiosInstance
