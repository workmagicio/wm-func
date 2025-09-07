import React, { useState, useEffect, useCallback } from 'react'
import TrendChart from '../components/TrendChart'
import AddTagModal from '../components/AddTagModal'
import TagFilter from '../components/TagFilter'
import PlatformSelector from '../components/PlatformSelector'
import { useTagManagement } from '../hooks/useTagManagement'
import { getEnabledPlatforms, getPlatformDisplayName } from '../config/platforms'
import './Dashboard.css'

interface DateSequence {
  date: string
  api_data: number
  data: number
  remove_data: number
}

interface Tenant {
  tenant_id: number
  last_30_day_diff: number
  date_sequence: DateSequence[]
  tags: string[]
  register_time?: string
}

interface ApiResponse {
  success: boolean
  data: {
    new_tenants: Tenant[]
    old_tenants: Tenant[]
    data_last_load_time: string
    data_type?: string
    last_data_date?: string
  }
  message: string
  global_tags?: string[]
}

const Dashboard: React.FC<{}> = () => {
  const platforms = getEnabledPlatforms()
  
  // 从URL参数或localStorage恢复平台选择
  const getInitialPlatform = () => {
    // 1. 优先从URL参数获取
    const urlParams = new URLSearchParams(window.location.search)
    const urlPlatform = urlParams.get('platform')
    if (urlPlatform && platforms.some(p => p.id === urlPlatform)) {
      return urlPlatform
    }
    
    // 2. 从localStorage获取
    const savedPlatform = localStorage.getItem('selectedPlatform')
    if (savedPlatform && platforms.some(p => p.id === savedPlatform)) {
      return savedPlatform
    }
    
    // 3. 默认值
    return platforms[0]?.id || 'googleAds'
  }
  
  const [platform, setPlatform] = useState<string>(getInitialPlatform())
  const [data, setData] = useState<ApiResponse['data'] | null>(null)
  const [globalTags, setGlobalTags] = useState<string[]>([])
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)
  const [updatingTenants, setUpdatingTenants] = useState<Set<number>>(new Set())
  
  // Tag管理相关状态
  const [isAddTagModalOpen, setIsAddTagModalOpen] = useState(false)
  const [selectedTenantId, setSelectedTenantId] = useState<number | null>(null)
  const [removingTags, setRemovingTags] = useState<Set<string>>(new Set())
  
  // 筛选相关状态
  const [filteredNewTenants, setFilteredNewTenants] = useState<Tenant[]>([])
  const [filteredOldTenants, setFilteredOldTenants] = useState<Tenant[]>([])
  const [showingFiltered, setShowingFiltered] = useState(false)
  
  // Tag管理hooks
  const { removeTag } = useTagManagement()

  // 筛选处理函数
  const handleFilterChange = useCallback((filteredNew: Tenant[], filteredOld: Tenant[]) => {
    setFilteredNewTenants(filteredNew)
    setFilteredOldTenants(filteredOld)
    
    // 判断是否在使用筛选
    const isFiltering = filteredNew.length !== (data?.new_tenants?.length || 0) || 
                       filteredOld.length !== (data?.old_tenants?.length || 0)
    setShowingFiltered(isFiltering)
  }, [data])

  // 保存平台选择到localStorage和URL
  useEffect(() => {
    // 保存到localStorage
    localStorage.setItem('selectedPlatform', platform)
    
    // 更新URL参数
    const url = new URL(window.location.href)
    url.searchParams.set('platform', platform)
    window.history.replaceState({}, '', url.toString())
  }, [platform])

  // 使用后端返回的全局标签列表（已经过滤和排序）

  // 获取数据
  const fetchData = async (needRefresh = false) => {
    try {
      setLoading(true)
      setError(null)
      
      const params = new URLSearchParams({ platform })
      if (needRefresh) {
        params.append('needRefresh', 'true')
      }
      
      const response = await fetch(`/api/alter-data?${params}`)
      const result: ApiResponse = await response.json()
      
      if (result.success) {
        // 确保数据结构完整
        const safeData = {
          ...result.data,
          new_tenants: result.data?.new_tenants || [],
          old_tenants: result.data?.old_tenants || [],
          data_type: result.data?.data_type || 'dual_source',
          last_data_date: result.data?.last_data_date || ''
        }
        console.log('📊 Data received:', {
          newTenants: safeData.new_tenants.length,
          oldTenants: safeData.old_tenants.length,
          dataType: safeData.data_type,
          lastDataDate: safeData.last_data_date
        })
        setData(safeData)
        const tags = result.global_tags || []
        console.log('🌐 Global tags received from API:', tags.length, 'tags:', tags)
        setGlobalTags(tags)
      } else {
        setError(result.message || '获取数据失败')
      }
    } catch (err) {
      setError('网络请求失败，请检查服务器连接')
      console.error('API Error:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [platform])

  const handleRefresh = () => {
    fetchData(true)
  }

  // Tag管理方法
  const handleAddTag = (tenantId: number) => {
    console.log('🏷️ Opening add tag modal for tenant:', tenantId)
    const startTime = performance.now()
    setSelectedTenantId(tenantId)
    setIsAddTagModalOpen(true)
    const endTime = performance.now()
    console.log('⏱️ Add tag modal state updated in:', (endTime - startTime).toFixed(2), 'ms')
  }

  const handleCloseTagModal = useCallback(() => {
    console.log('🚪 Closing add tag modal')
    const startTime = performance.now()
    setIsAddTagModalOpen(false)
    setSelectedTenantId(null)
    const endTime = performance.now()
    console.log('⏱️ Close tag modal state updated in:', (endTime - startTime).toFixed(2), 'ms')
  }, [])

  const handleTagAdded = useCallback(async (tenantId: number, tagName: string, updatedTags?: string[]) => {
    // 添加tag成功后只更新当前tenant的标签，不刷新整个页面
    if (!data) return

    // 更新全局标签列表（如果后端返回了）
    if (updatedTags) {
      setGlobalTags(updatedTags)
    }

    // 更新数据状态
    const updateTenantTags = (tenants: Tenant[]) => 
      tenants.map(tenant => 
        tenant.tenant_id === tenantId 
          ? { ...tenant, tags: [...(tenant.tags || []), tagName] }
          : tenant
      )

    const updatedData = {
      ...data,
      new_tenants: updateTenantTags(data.new_tenants),
      old_tenants: updateTenantTags(data.old_tenants)
    }

    setData(updatedData)

    // 如果正在使用筛选，也更新筛选结果
    if (showingFiltered) {
      setFilteredNewTenants(updateTenantTags(filteredNewTenants))
      setFilteredOldTenants(updateTenantTags(filteredOldTenants))
    }
  }, [data, showingFiltered, filteredNewTenants, filteredOldTenants])

  const handleRemoveTag = async (tenantId: number, tagName: string) => {
    // 防止删除默认标签
    if (tagName === 'code_filter_region') {
      alert('无法删除系统默认标签')
      return
    }

    const tagKey = `${tenantId}-${tagName}`
    
    if (window.confirm(`确定要删除标签"${tagName}"吗？`)) {
      try {
        setRemovingTags(prev => new Set(prev).add(tagKey))
        
        const result = await removeTag(tenantId, platform, tagName)
        
        if (result.success) {
          // 更新全局标签列表（如果后端返回了）
          if (result.data?.tags) {
            setGlobalTags(result.data.tags)
          }

          // 删除成功，只更新当前tenant的标签，不刷新整个页面
          if (data) {
            const updateTenantTags = (tenants: Tenant[]) => 
              tenants.map(tenant => 
                tenant.tenant_id === tenantId 
                  ? { ...tenant, tags: (tenant.tags || []).filter(tag => tag !== tagName) }
                  : tenant
              )

            const updatedData = {
              ...data,
              new_tenants: updateTenantTags(data.new_tenants),
              old_tenants: updateTenantTags(data.old_tenants)
            }

            setData(updatedData)

            // 如果正在使用筛选，也更新筛选结果
            if (showingFiltered) {
              setFilteredNewTenants(updateTenantTags(filteredNewTenants))
              setFilteredOldTenants(updateTenantTags(filteredOldTenants))
            }
          }
        } else {
          alert(`删除标签失败: ${result.message}`)
        }
      } catch (err) {
        console.error('删除标签出错:', err)
        alert('删除标签失败，请检查网络连接')
      } finally {
        setRemovingTags(prev => {
          const newSet = new Set(prev)
          newSet.delete(tagKey)
          return newSet
        })
      }
    }
  }

  // 补齐数据
  const handleUpdateTenantData = async (tenantId: number) => {
    try {
      setUpdatingTenants(prev => new Set(prev).add(tenantId))
      
      const params = new URLSearchParams({ 
        platform,
        tenantId: tenantId.toString()
      })
      
      const response = await fetch(`/api/alter-data?${params}`)
      const result: ApiResponse = await response.json()
      
      if (result.success && result.data) {
        // 只更新当前租户的数据，而不是刷新整个页面
        setData(prevData => {
          if (!prevData) return prevData
          
          // 查找并更新对应的租户数据
          const updatedNewTenants = prevData.new_tenants.map(tenant => {
            if (tenant.tenant_id === tenantId) {
              // 从返回的数据中找到更新后的租户数据
              const updatedTenant = [...result.data.new_tenants, ...result.data.old_tenants]
                .find(t => t.tenant_id === tenantId)
              return updatedTenant || tenant
            }
            return tenant
          })
          
          const updatedOldTenants = prevData.old_tenants.map(tenant => {
            if (tenant.tenant_id === tenantId) {
              // 从返回的数据中找到更新后的租户数据
              const updatedTenant = [...result.data.new_tenants, ...result.data.old_tenants]
                .find(t => t.tenant_id === tenantId)
              return updatedTenant || tenant
            }
            return tenant
          })
          
          return {
            ...prevData,
            new_tenants: updatedNewTenants,
            old_tenants: updatedOldTenants,
            data_last_load_time: result.data.data_last_load_time
          }
        })
        
        // 同时更新筛选后的数据
        setFilteredNewTenants(prevFiltered => 
          prevFiltered.map(tenant => {
            if (tenant.tenant_id === tenantId) {
              const updatedTenant = [...result.data.new_tenants, ...result.data.old_tenants]
                .find(t => t.tenant_id === tenantId)
              return updatedTenant || tenant
            }
            return tenant
          })
        )
        
        setFilteredOldTenants(prevFiltered => 
          prevFiltered.map(tenant => {
            if (tenant.tenant_id === tenantId) {
              const updatedTenant = [...result.data.new_tenants, ...result.data.old_tenants]
                .find(t => t.tenant_id === tenantId)
              return updatedTenant || tenant
            }
            return tenant
          })
        )
        
        console.log(`租户 ${tenantId} 的数据补齐成功`)
      } else {
        console.error(`租户 ${tenantId} 数据补齐失败:`, result.message)
        alert(`数据补齐失败: ${result.message}`)
      }
    } catch (err) {
      console.error(`租户 ${tenantId} 数据补齐出错:`, err)
      alert('数据补齐失败，请检查网络连接')
    } finally {
      setUpdatingTenants(prev => {
        const newSet = new Set(prev)
        newSet.delete(tenantId)
        return newSet
      })
    }
  }

  const getPlatformName = (platformId: string) => {
    return getPlatformDisplayName(platformId)
  }

  const formatDiff = (diff: number) => {
    return diff.toLocaleString('en-US', { 
      maximumFractionDigits: 0,
      useGrouping: true 
    })
  }

  const getDiffColor = (diff: number) => {
    if (diff > 0) return '#4caf50'
    if (diff < 0) return '#f44336'
    return '#666'
  }

    const formatRegisterTime = (registerTime?: string) => {
    if (!registerTime) return ''
    try {
      const date = new Date(registerTime)
      return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
    } catch (error) {
      return registerTime
    }
  }

  // 获取tag的CSS类名
  const getTagClassName = (tag: string, tenantType: 'new' | 'old') => {
    if (tag.startsWith('err_')) {
      return 'tag tag-error'
    }
    return `tag tag-${tenantType}`
  }

  // 加载状态
  if (loading) {
    return (
      <div className="dashboard">
        <div className="loading">
          <div className="loading-spinner"></div>
          <p>正在加载数据...</p>
        </div>
      </div>
    )
  }

  // 错误状态
  if (error) {
    return (
      <div className="dashboard">
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

  // 没有数据
  if (!data) {
    return (
      <div className="dashboard">
        <div className="no-data">
          <p>暂无数据</p>
        </div>
      </div>
    )
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <div className="header-left">
          <h2>{getPlatformName(platform)} 数据差异分析</h2>
          <PlatformSelector
            platforms={platforms}
            selectedPlatform={platform}
            onPlatformChange={setPlatform}
          />
        </div>
        <div className="header-controls">
          <div className="last-update">
            最后更新: {new Date(data.data_last_load_time).toLocaleString('zh-CN')}
          </div>
          <button onClick={handleRefresh} className="refresh-button">
            🔄 刷新数据
          </button>
        </div>
      </div>

      <div className="dashboard-content">
        {/* Tag筛选器 */}
        {data && data.new_tenants && data.old_tenants && (
          <TagFilter
            newTenants={data.new_tenants.map(tenant => ({
              ...tenant,
              register_time: tenant.register_time || ''
            }))}
            oldTenants={data.old_tenants.map(tenant => ({
              ...tenant,
              register_time: tenant.register_time || ''
            }))}
            onFilterChange={handleFilterChange}
          />
        )}

        {/* 数据概览 */}
        <div className="overview-cards">
          <div className="overview-card">
            <div className="overview-title">最近15天注册</div>
            <div className="overview-value">
              {showingFiltered ? filteredNewTenants.length : data?.new_tenants?.length || 0}
              {showingFiltered && (
                <span className="filter-indicator">/{data?.new_tenants?.length || 0}</span>
              )}
            </div>
            <div className="overview-subtitle">新客户数量</div>
          </div>
          <div className="overview-card">
            <div className="overview-title">老客户</div>
            <div className="overview-value">
              {showingFiltered ? filteredOldTenants.length : data?.old_tenants?.length || 0}
              {showingFiltered && (
                <span className="filter-indicator">/{data?.old_tenants?.length || 0}</span>
              )}
            </div>
            <div className="overview-subtitle">注册 ≥ 15天</div>
          </div>
          <div className="overview-card">
            <div className="overview-title">总差异</div>
            <div className="overview-value" style={{
              color: getDiffColor(
                [...(showingFiltered ? filteredNewTenants : data.new_tenants), 
                 ...(showingFiltered ? filteredOldTenants : data.old_tenants)]
                  .reduce((sum, tenant) => sum + tenant.last_30_day_diff, 0)
              )
            }}>
              {formatDiff(
                [...(showingFiltered ? filteredNewTenants : data.new_tenants), 
                 ...(showingFiltered ? filteredOldTenants : data.old_tenants)]
                  .reduce((sum, tenant) => sum + tenant.last_30_day_diff, 0)
              )}
            </div>
            <div className="overview-subtitle">最近30天累计</div>
          </div>
        </div>

        {/* 最近15天注册的客户 */}
        <div className="section">
          <div className="section-header">
            <h2 className="section-title">🌟 最近15天注册的客户</h2>
            <div className="section-subtitle">
              共 {showingFiltered ? filteredNewTenants.length : data?.new_tenants?.length || 0} 个新客户
              {showingFiltered && (
                <span className="filter-indicator">（筛选结果，总共{data?.new_tenants?.length || 0}个）</span>
              )}
              （按30天差异绝对值降序排列）
            </div>
          </div>
          
          <div className="tenant-grid">
            {(showingFiltered ? filteredNewTenants : data?.new_tenants || []).map((tenant) => (
            <div key={tenant.tenant_id} className="tenant-card">
              <div className="tenant-card-header">
                <div className="tenant-header-row">
                  <div className="tenant-basic-info">
                    <h4 className="tenant-id">租户 ID: {tenant.tenant_id}</h4>
                    {tenant.register_time && (
                      <div className="register-time">注册时间: {formatRegisterTime(tenant.register_time)}</div>
                    )}
                  </div>
                  <div className="tenant-diff">
                    <div 
                      className="diff-value" 
                      style={{ color: getDiffColor(tenant.last_30_day_diff) }}
                    >
                      {formatDiff(tenant.last_30_day_diff)}
                    </div>
                    <div className="diff-label">30天差异</div>
                    <button 
                      className={`update-data-button ${updatingTenants.has(tenant.tenant_id) ? 'updating' : ''}`}
                      onClick={() => handleUpdateTenantData(tenant.tenant_id)}
                      disabled={updatingTenants.has(tenant.tenant_id)}
                    >
                      {updatingTenants.has(tenant.tenant_id) ? '补齐中...' : '补齐数据'}
                    </button>
                  </div>
                </div>
                <div className="tenant-tags-section">
                  <div className="tenant-tags">
                    {(tenant.tags || []).filter(tag => tag && tag.trim() !== '').map((tag, index) => (
                      <span key={index} className={getTagClassName(tag, 'new')}>
                        {tag}
                        {!tag.startsWith('err_') && tag !== 'code_filter_region' && (
                          <button
                            className="tag-remove"
                            onClick={() => handleRemoveTag(tenant.tenant_id, tag)}
                            disabled={removingTags.has(`${tenant.tenant_id}-${tag}`)}
                            title="删除标签"
                          >
                            {removingTags.has(`${tenant.tenant_id}-${tag}`) ? '⏳' : '×'}
                          </button>
                        )}
                      </span>
                    ))}
                  </div>
                  <button 
                    className="add-tag-btn"
                    onClick={() => handleAddTag(tenant.tenant_id)}
                    title="添加标签"
                  >
                    <span>+</span>
                    <span>标签</span>
                  </button>
                </div>
              </div>
              
              {tenant.date_sequence && tenant.date_sequence.length > 0 && (
                <>
                  <TrendChart 
                    title="数据趋势对比"
                    data={tenant.date_sequence}
                    dataType={data?.data_type}
                    lastDataDate={data?.last_data_date}
                  />
                </>
              )}
            </div>
            ))}
          </div>
        </div>

        {/* 老客户 */}
        <div className="section">
          <div className="section-header">
            <h2 className="section-title">👥 老客户</h2>
            <div className="section-subtitle">
              共 {showingFiltered ? filteredOldTenants.length : data?.old_tenants?.length || 0} 个老客户
              {showingFiltered && (
                <span className="filter-indicator">（筛选结果，总共{data?.old_tenants?.length || 0}个）</span>
              )}
              （按30天差异绝对值降序排列）
            </div>
          </div>
          
          <div className="tenant-grid">
            {(showingFiltered ? filteredOldTenants : data?.old_tenants || []).map((tenant) => (
            <div key={tenant.tenant_id} className="tenant-card">
              <div className="tenant-card-header">
                <div className="tenant-header-row">
                  <div className="tenant-basic-info">
                    <h4 className="tenant-id">租户 ID: {tenant.tenant_id}</h4>
                    {tenant.register_time && (
                      <div className="register-time">注册时间: {formatRegisterTime(tenant.register_time)}</div>
                    )}
                  </div>
                  <div className="tenant-diff">
                    <div 
                      className="diff-value" 
                      style={{ color: getDiffColor(tenant.last_30_day_diff) }}
                    >
                      {formatDiff(tenant.last_30_day_diff)}
                    </div>
                    <div className="diff-label">30天差异</div>
                    <button 
                      className={`update-data-button ${updatingTenants.has(tenant.tenant_id) ? 'updating' : ''}`}
                      onClick={() => handleUpdateTenantData(tenant.tenant_id)}
                      disabled={updatingTenants.has(tenant.tenant_id)}
                    >
                      {updatingTenants.has(tenant.tenant_id) ? '补齐中...' : '补齐数据'}
                    </button>
                  </div>
                </div>
                <div className="tenant-tags-section">
                  <div className="tenant-tags">
                    {(tenant.tags || []).filter(tag => tag && tag.trim() !== '').map((tag, index) => (
                      <span key={index} className={getTagClassName(tag, 'old')}>
                        {tag}
                        {!tag.startsWith('err_') && tag !== 'code_filter_region' && (
                          <button
                            className="tag-remove"
                            onClick={() => handleRemoveTag(tenant.tenant_id, tag)}
                            disabled={removingTags.has(`${tenant.tenant_id}-${tag}`)}
                            title="删除标签"
                          >
                            {removingTags.has(`${tenant.tenant_id}-${tag}`) ? '⏳' : '×'}
                          </button>
                        )}
                      </span>
                    ))}
                  </div>
                  <button 
                    className="add-tag-btn"
                    onClick={() => handleAddTag(tenant.tenant_id)}
                    title="添加标签"
                  >
                    <span>+</span>
                    <span>标签</span>
                  </button>
                </div>
              </div>
              
              {tenant.date_sequence && tenant.date_sequence.length > 0 && (
                <>
                  <TrendChart 
                    title="数据趋势对比"
                    data={tenant.date_sequence}
                    dataType={data?.data_type}
                    lastDataDate={data?.last_data_date}
                  />
                </>
              )}
            </div>
            ))}
          </div>
        </div>

        {/* 数据概览统计 */}
        <div className="card info-card">
          <div className="card-header">
            <h3 className="card-title">📊 数据概览</h3>
          </div>
          <div className="overview-stats">
            <div className="stat-row">
              <span className="stat-label">总租户数量：</span>
              <span className="stat-value">
                {(showingFiltered ? filteredNewTenants.length + filteredOldTenants.length : (data?.new_tenants?.length || 0) + (data?.old_tenants?.length || 0))} 个
              </span>
            </div>
            <div className="stat-row">
              <span className="stat-label">新客户数量：</span>
              <span className="stat-value">
                {showingFiltered ? filteredNewTenants.length : data?.new_tenants?.length || 0} 个 
                <small>（最近15天注册）</small>
              </span>
            </div>
            <div className="stat-row">
              <span className="stat-label">老客户数量：</span>
              <span className="stat-value">
                {showingFiltered ? filteredOldTenants.length : data?.old_tenants?.length || 0} 个 
                <small>（注册 ≥ 15天）</small>
              </span>
            </div>
            <div className="stat-row">
              <span className="stat-label">总数据差异：</span>
              <span className="stat-value" style={{
                color: getDiffColor(
                  [...(showingFiltered ? filteredNewTenants : data.new_tenants), 
                   ...(showingFiltered ? filteredOldTenants : data.old_tenants)]
                    .reduce((sum, tenant) => sum + tenant.last_30_day_diff, 0)
                )
              }}>
                {formatDiff(
                  [...(showingFiltered ? filteredNewTenants : data.new_tenants), 
                   ...(showingFiltered ? filteredOldTenants : data.old_tenants)]
                    .reduce((sum, tenant) => sum + tenant.last_30_day_diff, 0)
                )} 
                <small>（最近30天累计）</small>
              </span>
            </div>
            <div className="stat-row">
              <span className="stat-label">数据范围：</span>
              <span className="stat-value">最近90天数据，统计最近30天差异</span>
            </div>
            <div className="stat-row">
              <span className="stat-label">计算方式：</span>
              <span className="stat-value">数据差异 = wm_data - API数据</span>
            </div>
          </div>
        </div>
      </div>

      {/* Add Tag Modal */}
      {selectedTenantId && (
        <AddTagModal
          tenantId={selectedTenantId}
          platform={platform}
          isOpen={isAddTagModalOpen}
          onClose={handleCloseTagModal}
          onSuccess={handleTagAdded}
          existingTags={globalTags}
        />
      )}
    </div>
  )
}

export default Dashboard
