import React, { useState, useEffect, useCallback } from 'react'
import AttributionChart from '../components/AttributionChart'
import AttributionTagFilter from '../components/AttributionTagFilter'
import AddTagModal from '../components/AddTagModal'
import { useTagManagement } from '../hooks/useTagManagement'
import './AttributionAnalysis.css'

interface AttributionDateSequence {
  date: string
  platform_data?: { [key: string]: number }
  total_attribution: number
  is_recent_zero: boolean
}

interface PlatformTotal {
  platform: string
  total_attribution: number
  daily_average: number
}

interface AttributionTenantData {
  tenant_id: number
  date_sequence: AttributionDateSequence[]
  platform_totals: PlatformTotal[]
  total_attribution_avg: number
  tags: string[]
  recent_zero_days: number
  has_recent_zeros: boolean
  customer_type: string
  register_time: string
}

interface AllAttributionApiResponse {
  success: boolean
  data: AttributionTenantData[]
  message: string
}

const AttributionAnalysis: React.FC = () => {
  const [data, setData] = useState<AttributionTenantData[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdateTime, setLastUpdateTime] = useState<string>('')
  
  // 标签筛选相关状态
  const [filteredData, setFilteredData] = useState<AttributionTenantData[]>([])
  const [showingFiltered, setShowingFiltered] = useState(false)
  
  // 标签添加相关状态
  const [isAddTagModalOpen, setIsAddTagModalOpen] = useState(false)
  const [selectedTenantId, setSelectedTenantId] = useState<number | null>(null)
  const [removingTags, setRemovingTags] = useState<Set<string>>(new Set())
  
  // Tag管理hooks
  const { removeTag } = useTagManagement()

  // 标签筛选处理函数
  const handleFilterChange = useCallback((filtered: AttributionTenantData[]) => {
    setFilteredData(filtered)
    setShowingFiltered(Array.isArray(filtered) && Array.isArray(data) && filtered.length !== data.length)
  }, [data])

  // 标签添加处理
  const handleAddTag = (tenantId: number) => {
    setSelectedTenantId(tenantId)
    setIsAddTagModalOpen(true)
  }

  // 标签删除处理
  const handleRemoveTag = async (tenantId: number, tag: string, platform: string = 'attribution') => {
    const tagKey = `${tenantId}_${platform}_${tag}`
    if (removingTags.has(tagKey)) return

    setRemovingTags(prev => new Set(prev).add(tagKey))
    
    try {
      await removeTag(tenantId, tag, platform)
      // 刷新数据
      await fetchAllAttributionData(true)
    } catch (error) {
      console.error('删除标签失败:', error)
    } finally {
      setRemovingTags(prev => {
        const newSet = new Set(prev)
        newSet.delete(tagKey)
        return newSet
      })
    }
  }

  // 获取所有租户的归因数据
  const fetchAllAttributionData = async (needRefresh: boolean = false) => {
    setLoading(true)
    setError(null)
    
    try {
      const response = await fetch(`/api/attribution/all?needRefresh=${needRefresh}`)
      const result: AllAttributionApiResponse = await response.json()
      
      if (result.success) {
        const responseData = Array.isArray(result.data) ? result.data : []
        setData(responseData)
        setLastUpdateTime(new Date().toLocaleString('zh-CN'))
      } else {
        setError(result.message || '获取归因数据失败')
      }
    } catch (err) {
      setError('网络请求失败: ' + (err as Error).message)
    } finally {
      setLoading(false)
    }
  }

  // 页面加载时获取数据
  useEffect(() => {
    fetchAllAttributionData()
  }, [])

  // 刷新数据
  const handleRefresh = () => {
    fetchAllAttributionData(true)
  }

  // 格式化差异值
  const formatDiff = (diff: number): string => {
    if (diff === 0) return '0'
    const sign = diff > 0 ? '+' : ''
    return `${sign}${diff.toLocaleString()}`
  }

  // 获取差异颜色
  const getDiffColor = (diff: number): string => {
    if (diff > 0) return '#28a745'
    if (diff < 0) return '#dc3545'
    return '#6c757d'
  }

  // 计算租户的总体差异（归因总数 vs 总归因平均）
  const calculateTenantDiff = (tenant: AttributionTenantData): number => {
    const totalAttribution = tenant?.platform_totals?.reduce((sum, platform) => sum + (platform?.total_attribution || 0), 0) || 0
    const avgValue = tenant?.total_attribution_avg || 0
    return totalAttribution - (avgValue * 30) // 假设30天的差异
  }

  // 获取标签样式类名
  const getTagClassName = (tag: string): string => {
    if (tag.startsWith('err_')) {
      return 'tag error-tag'
    }
    return 'tag normal-tag'
  }

  // 加载状态
  if (loading) {
    return (
      <div className="attribution-analysis">
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <p>正在加载归因数据...</p>
        </div>
      </div>
    )
  }

  // 错误状态
  if (error) {
    return (
      <div className="attribution-analysis">
        <div className="error-message">
          <h3>❌ 数据加载失败</h3>
          <p>{error}</p>
          <button onClick={handleRefresh} className="retry-button">
            重新加载
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="attribution-analysis">
      <div className="attribution-header">
        <div className="header-left">
          <h1>📊 归因数据分析</h1>
        </div>
        <div className="header-controls">
          {lastUpdateTime && (
            <div className="last-update">
              最后更新: {lastUpdateTime}
            </div>
          )}
          <button onClick={handleRefresh} className="refresh-button">
            🔄 刷新数据
          </button>
        </div>
      </div>

      {/* 标签筛选 */}
      <div className="filter-section">
        <AttributionTagFilter
          data={data}
          onFilterChange={handleFilterChange}
        />
      </div>

      {/* 数据概览 */}
      <div className="overview-cards">
        <div className="overview-card">
          <div className="overview-title">总租户数</div>
          <div className="overview-value">{Array.isArray(data) ? data.length : 0}</div>
          <div className="overview-subtitle">有归因数据的租户</div>
        </div>
        <div className="overview-card">
          <div className="overview-title">异常租户</div>
          <div className="overview-value error">
            {Array.isArray(data) ? data.filter(tenant => tenant?.tags?.some(tag => tag?.startsWith('err_'))).length : 0}
          </div>
          <div className="overview-subtitle">有异常标签的租户</div>
        </div>
        <div className="overview-card">
          <div className="overview-title">零值异常</div>
          <div className="overview-value warning">
            {Array.isArray(data) ? data.filter(tenant => tenant?.has_recent_zeros).length : 0}
          </div>
          <div className="overview-subtitle">最近3天有零值</div>
        </div>
      </div>

      {/* 租户归因数据列表 */}
      <div className="attribution-content">
        {((showingFiltered ? filteredData : data)?.length || 0) === 0 ? (
          <div className="no-data">
            <p>{showingFiltered ? '没有符合筛选条件的租户' : '暂无归因数据'}</p>
          </div>
        ) : (
          <div className="tenant-grid">
            {Array.isArray(showingFiltered ? filteredData : data) ? (showingFiltered ? filteredData : data).map((tenant) => (
              <div key={tenant?.tenant_id} className="attribution-tenant-card">
                <div className="tenant-card-header">
                  <div className="tenant-basic-info">
                    <div className="tenant-title-row">
                      <h3 className="tenant-id">租户 ID: {tenant?.tenant_id}</h3>
                      <span className={`customer-type-badge ${tenant?.customer_type === 'new' ? 'new-customer' : 'old-customer'}`}>
                        {tenant?.customer_type === 'new' ? '🆕 新客户' : '👤 老客户'}
                      </span>
                    </div>
                    {tenant?.register_time && (
                      <div className="register-time">
                        📅 注册时间: {new Date(tenant.register_time).toLocaleDateString('zh-CN')}
                      </div>
                    )}
                    <div className="tenant-stats">
                      <div className="stat-item">
                        <span className="stat-label">总归因平均:</span>
                        <span className="stat-value">{tenant?.total_attribution_avg?.toFixed(1) || '0'}</span>
                      </div>
                      <div className="stat-item">
                        <span className="stat-label">零值天数:</span>
                        <span className={`stat-value ${tenant?.has_recent_zeros ? 'warning' : ''}`}>
                          {tenant?.recent_zero_days || 0}天
                        </span>
                      </div>
                    </div>
                  </div>
                  
                  <div className="tenant-diff">
                    <div 
                      className="diff-value" 
                      style={{ color: getDiffColor(calculateTenantDiff(tenant)) }}
                    >
                      {formatDiff(calculateTenantDiff(tenant))}
                    </div>
                    <div className="diff-label">归因差异</div>
                  </div>
                </div>

                {/* 平台汇总 */}
                {tenant?.platform_totals && tenant?.platform_totals?.length > 0 && (
                  <div className="platform-summary">
                    <h4>平台汇总</h4>
                    <div className="platform-totals-grid">
                      {tenant?.platform_totals?.map(platform => (
                        <div key={platform?.platform} className="platform-item">
                          <span className="platform-name">{platform?.platform}:</span>
                          <span className="platform-value">
                            {platform?.total_attribution || 0} (日均 {platform?.daily_average?.toFixed(1) || '0'})
                          </span>
                        </div>
                      )) || []}
                    </div>
                  </div>
                )}

                {/* 标签 */}
                <div className="tenant-tags-section">
                  <div className="tags-header">
                    <h5>标签</h5>
                    <button 
                      className="add-tag-btn"
                      onClick={() => handleAddTag(tenant?.tenant_id || 0)}
                    >
                      + 添加标签
                    </button>
                  </div>
                  
                  {tenant?.tags && tenant?.tags?.length > 0 ? (
                    <div className="tenant-tags">
                      {tenant?.tags?.map((tag, index) => (
                        <span key={index} className={getTagClassName(tag || '')}>
                          {tag}
                          <button
                            className="remove-tag-btn"
                            onClick={() => handleRemoveTag(tenant?.tenant_id || 0, tag || '')}
                            disabled={removingTags.has(`${tenant?.tenant_id}_attribution_${tag}`)}
                          >
                            ×
                          </button>
                        </span>
                      )) || []}
                    </div>
                  ) : (
                    <div className="no-tags">
                      <p>暂无标签</p>
                    </div>
                  )}
                </div>

                {/* 归因数据图表 */}
                {tenant?.date_sequence && tenant?.date_sequence?.length > 0 ? (
                  <AttributionChart
                    title="归因数据趋势"
                    data={tenant}
                  />
                ) : (
                  <div className="no-chart-data">
                    <p>暂无图表数据</p>
                  </div>
                )}
              </div>
            )) : []}
          </div>
        )}
      </div>
      
      {/* 添加标签模态框 */}
      <AddTagModal
        isOpen={isAddTagModalOpen}
        onClose={() => {
          setIsAddTagModalOpen(false)
          setSelectedTenantId(null)
        }}
        tenantId={selectedTenantId || 0}
        platform="attribution"
        onSuccess={(tenantId, tagName, updatedTags) => {
          console.log('✅ Tag added successfully:', { tenantId, tagName, updatedTags })
          fetchAllAttributionData(true)
          setIsAddTagModalOpen(false)
          setSelectedTenantId(null)
        }}
      />
    </div>
  )
}

export default AttributionAnalysis
