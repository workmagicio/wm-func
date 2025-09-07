import { useState, useCallback } from 'react'

interface TagResponse {
  success: boolean
  message: string
  data?: any
}

export const useTagManagement = () => {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // 添加标签
  const addTag = useCallback(async (tenantId: number, platform: string, tagName: string): Promise<TagResponse> => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/tags', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          tenant_id: tenantId,
          platform,
          tag_name: tagName
        })
      })

      const result: TagResponse = await response.json()

      if (!result.success) {
        setError(result.message || '添加标签失败')
      }

      return result
    } catch (err) {
      const errorMessage = '网络请求失败，请检查服务器连接'
      setError(errorMessage)
      console.error('添加标签失败:', err)
      return {
        success: false,
        message: errorMessage
      }
    } finally {
      setLoading(false)
    }
  }, [])

  // 删除标签
  const removeTag = useCallback(async (tenantId: number, platform: string, tagName: string): Promise<TagResponse> => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch('/api/tags', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          tenant_id: tenantId,
          platform,
          tag_name: tagName
        })
      })

      const result: TagResponse = await response.json()

      if (!result.success) {
        setError(result.message || '删除标签失败')
      }

      return result
    } catch (err) {
      const errorMessage = '网络请求失败，请检查服务器连接'
      setError(errorMessage)
      console.error('删除标签失败:', err)
      return {
        success: false,
        message: errorMessage
      }
    } finally {
      setLoading(false)
    }
  }, [])

  // 获取标签
  const getTags = useCallback(async (tenantId: number, platform: string): Promise<TagResponse> => {
    setLoading(true)
    setError(null)

    try {
      const response = await fetch(`/api/tags/${tenantId}/${platform}`)
      const result: TagResponse = await response.json()

      if (!result.success) {
        setError(result.message || '获取标签失败')
      }

      return result
    } catch (err) {
      const errorMessage = '网络请求失败，请检查服务器连接'
      setError(errorMessage)
      console.error('获取标签失败:', err)
      return {
        success: false,
        message: errorMessage
      }
    } finally {
      setLoading(false)
    }
  }, [])

  // 清除错误
  const clearError = useCallback(() => {
    setError(null)
  }, [])

  return {
    loading,
    error,
    addTag,
    removeTag,
    getTags,
    clearError
  }
}
