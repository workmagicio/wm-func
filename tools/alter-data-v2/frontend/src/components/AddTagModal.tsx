import React, { useState, useEffect, useMemo, useCallback, useRef } from 'react'
import './AddTagModal.css'

interface AddTagModalProps {
  tenantId: number
  platform: string
  isOpen: boolean
  onClose: () => void
  onSuccess: (tenantId: number, tagName: string, updatedTags?: string[]) => void
  existingTags?: string[]
}

const AddTagModal: React.FC<AddTagModalProps> = ({ 
  tenantId, 
  platform, 
  isOpen, 
  onClose, 
  onSuccess,
  existingTags = []
}) => {
  const [tagName, setTagName] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  
  // 使用ref来避免useCallback依赖问题
  const onSuccessRef = useRef(onSuccess)
  const onCloseRef = useRef(onClose)
  
  // 更新ref值
  useEffect(() => {
    onSuccessRef.current = onSuccess
    onCloseRef.current = onClose
  }, [onSuccess, onClose])

  // 重置状态当模态框打开时
  useEffect(() => {
    console.log('🔄 AddTagModal isOpen changed:', isOpen, 'tenantId:', tenantId)
    if (isOpen) {
      console.log('✅ Modal opening - resetting state, existingTags count:', existingTags.length)
      setTagName('')
      setError(null)
    } else {
      console.log('❌ Modal closing')
    }
  }, [isOpen, tenantId]) // 移除existingTags.length依赖，避免无限循环

  // 处理键盘事件
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!isOpen) return
      
      if (event.key === 'Escape') {
        onCloseRef.current()
      } else if (event.key === 'Enter' && !loading && tagName.trim()) {
        handleSubmit()
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, loading, tagName]) // 移除onClose依赖

  // 处理现有标签点击 - 直接提交
  const handleExistingTagClick = useCallback(async (tag: string) => {
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
          platform: platform,
          tag_name: tag
        })
      })

      const result = await response.json()

      if (result.success) {
        onSuccessRef.current(tenantId, tag, result.data?.tags) // 使用ref避免依赖
        onCloseRef.current()   // 使用ref避免依赖
      } else {
        setError(result.message || '添加标签失败')
      }
    } catch (err) {
      console.error('添加标签失败:', err)
      setError('网络请求失败，请检查服务器连接')
    } finally {
      setLoading(false)
    }
  }, [tenantId, platform]) // 移除onSuccess和onClose依赖

  // 处理自定义标签提交
  const handleSubmit = async () => {
    if (!tagName.trim()) {
      setError('标签名称不能为空')
      return
    }

    if (tagName.trim().length > 20) {
      setError('标签名称不能超过20个字符')
      return
    }

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
          platform: platform,
          tag_name: tagName.trim()
        })
      })

      const result = await response.json()

      if (result.success) {
        onSuccessRef.current(tenantId, tagName.trim(), result.data?.tags) // 使用ref避免依赖
        onCloseRef.current()   // 使用ref避免依赖
        setTagName('') // 清空输入
      } else {
        setError(result.message || '添加标签失败')
      }
    } catch (err) {
      console.error('添加标签失败:', err)
      setError('网络请求失败，请检查服务器连接')
    } finally {
      setLoading(false)
    }
  }

  const handleBackdropClick = (e: React.MouseEvent) => {
    console.log('🖱️ Backdrop clicked')
    if (e.target === e.currentTarget) {
      console.log('🚪 Closing modal via backdrop')
      onCloseRef.current()
    }
  }

  // 缓存标签按钮渲染，避免每次都重新创建
  const tagButtons = useMemo(() => {
    console.log('🏷️ Rendering tag buttons, count:', existingTags.length, 'loading:', loading)
    const startTime = performance.now()
    
    if (!existingTags.length) return null
    
    const buttons = existingTags.map((tag) => (
      <button
        key={tag}
        type="button"
        className={`tag-button ${tag.startsWith('err_') ? 'error' : 'normal'}`}
        onClick={() => handleExistingTagClick(tag)}
        disabled={loading}
      >
        {tag}
      </button>
    ))
    
    const endTime = performance.now()
    console.log('⏱️ Tag buttons rendered in:', (endTime - startTime).toFixed(2), 'ms')
    return buttons
  }, [existingTags, loading]) // 移除handleExistingTagClick依赖

  if (!isOpen) return null

  return (
    <div className="modal-backdrop" onClick={handleBackdropClick}>
      <div className="add-tag-modal">
        <div className="modal-header">
          <h3>添加标签</h3>
          <button className="modal-close" onClick={() => {
            console.log('❌ Close button clicked')
            onCloseRef.current()
          }}>
            ×
          </button>
        </div>
        
        <div className="modal-body">
          <div className="tenant-info">
            <span className="tenant-label">租户ID:</span>
            <span className="tenant-value">{tenantId}</span>
          </div>
          
          {/* 现有标签按钮 */}
          {tagButtons && (
            <div className="existing-tags-section">
              <label className="section-label">选择现有标签</label>
              <div className="tag-buttons">
                {tagButtons}
              </div>
            </div>
          )}

          {/* 自定义标签输入 */}
          <div className="input-group">
            <label htmlFor="tag-input">或输入新标签</label>
            <input
              id="tag-input"
              type="text"
              placeholder="输入标签名称（最多20个字符）"
              value={tagName}
              onChange={(e) => {
                setTagName(e.target.value)
                setError(null)
              }}
              maxLength={20}
              autoFocus
              disabled={loading}
              className="tag-input"
            />
            <div className="input-hint">
              {tagName.length}/20 字符 • 30天后自动过期
            </div>
          </div>

          {error && (
            <div className="error-message">
              <span className="error-icon">⚠️</span>
              {error}
            </div>
          )}
        </div>

        <div className="modal-actions">
          <button 
            className="btn btn-secondary" 
            onClick={() => {
              console.log('🚫 Cancel button clicked')
              onCloseRef.current()
            }}
            disabled={loading}
          >
            取消
          </button>
          <button 
            className="btn btn-primary" 
            onClick={handleSubmit}
            disabled={loading || !tagName.trim()}
          >
            {loading ? (
              <>
                <span className="loading-spinner"></span>
                添加中...
              </>
            ) : (
              '确认添加'
            )}
          </button>
        </div>
      </div>
    </div>
  )
}

export default AddTagModal
