import React, { useState, useEffect } from 'react'
import TrendChart from '../components/TrendChart'
import './Dashboard.css'

interface DateSequence {
  date: string
  api_data: number
  data: number
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
  }
  message: string
}

interface DashboardProps {
  platform: string
}

const Dashboard: React.FC<DashboardProps> = ({ platform }) => {
  const [data, setData] = useState<ApiResponse['data'] | null>(null)
  const [loading, setLoading] = useState<boolean>(true)
  const [error, setError] = useState<string | null>(null)

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
        setData(result.data)
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
    if (platform === 'googleAds') {
      fetchData()
    }
  }, [platform])

  const handleRefresh = () => {
    fetchData(true)
  }

  const getPlatformName = (platformId: string) => {
    const platforms = {
      googleAds: 'Google Ads',
      facebookMarketing: 'Facebook Marketing',
      tiktokMarketing: 'TikTok Marketing'
    }
    return platforms[platformId as keyof typeof platforms] || platformId
  }

  const formatDiff = (diff: number) => {
    return `${diff.toLocaleString()}`
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
        <div className="error">
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
          <div className="platform-selector">
            <label className="platform-label">📊 平台选择：</label>
            <select 
              value={platform} 
              onChange={(e) => window.location.reload()} // 现在只有googleAds，所以暂时用刷新
              className="platform-select"
            >
              <option value="googleAds">Google Ads</option>
            </select>
          </div>
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
        {/* 数据概览 */}
        <div className="overview-cards">
          <div className="overview-card">
            <div className="overview-title">最近15天注册</div>
            <div className="overview-value">{data.new_tenants.length}</div>
            <div className="overview-subtitle">新客户数量</div>
          </div>
          <div className="overview-card">
            <div className="overview-title">老客户</div>
            <div className="overview-value">{data.old_tenants.length}</div>
            <div className="overview-subtitle">注册 ≥ 15天</div>
          </div>
          <div className="overview-card">
            <div className="overview-title">总差异</div>
            <div className="overview-value" style={{
              color: getDiffColor(
                [...data.new_tenants, ...data.old_tenants]
                  .reduce((sum, tenant) => sum + tenant.last_30_day_diff, 0)
              )
            }}>
              {formatDiff(
                [...data.new_tenants, ...data.old_tenants]
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
            <div className="section-subtitle">共 {data.new_tenants.length} 个新客户</div>
          </div>
          
          <div className="tenant-grid">
            {data.new_tenants.map((tenant) => (
            <div key={tenant.tenant_id} className="tenant-card">
              <div className="tenant-card-header">
                <div className="tenant-info">
                  <h4 className="tenant-id">租户 ID: {tenant.tenant_id}</h4>
                  {tenant.register_time && (
                    <div className="register-time">注册时间: {formatRegisterTime(tenant.register_time)}</div>
                  )}
                  <div className="tenant-tags">
                    {tenant.tags.filter(tag => tag && tag.trim() !== '').map((tag, index) => (
                      <span key={index} className="tag tag-new">{tag}</span>
                    ))}
                  </div>
                </div>
                <div className="tenant-diff">
                  <div 
                    className="diff-value" 
                    style={{ color: getDiffColor(tenant.last_30_day_diff) }}
                  >
                    {formatDiff(tenant.last_30_day_diff)}
                  </div>
                  <div className="diff-label">30天差异</div>
                </div>
              </div>
              
              {tenant.date_sequence.length > 0 && (
                <>
                  <TrendChart 
                    title="数据趋势对比"
                    data={tenant.date_sequence}
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
            <div className="section-subtitle">共 {data.old_tenants.length} 个老客户（按30天差异降序排列）</div>
          </div>
          
          <div className="tenant-grid">
            {data.old_tenants.map((tenant) => (
            <div key={tenant.tenant_id} className="tenant-card">
              <div className="tenant-card-header">
                <div className="tenant-info">
                  <h4 className="tenant-id">租户 ID: {tenant.tenant_id}</h4>
                  {tenant.register_time && (
                    <div className="register-time">注册时间: {formatRegisterTime(tenant.register_time)}</div>
                  )}
                  <div className="tenant-tags">
                    {tenant.tags.filter(tag => tag && tag.trim() !== '').map((tag, index) => (
                      <span key={index} className="tag tag-old">{tag}</span>
                    ))}
                  </div>
                </div>
                <div className="tenant-diff">
                  <div 
                    className="diff-value" 
                    style={{ color: getDiffColor(tenant.last_30_day_diff) }}
                  >
                    {formatDiff(tenant.last_30_day_diff)}
                  </div>
                  <div className="diff-label">30天差异</div>
                </div>
              </div>
              
              {tenant.date_sequence.length > 0 && (
                <>
                  <TrendChart 
                    title="数据趋势对比"
                    data={tenant.date_sequence}
                  />
                </>
              )}
            </div>
            ))}
          </div>
        </div>

        {/* 数据说明 */}
        <div className="card info-card">
          <div className="card-header">
            <h3 className="card-title">📊 数据说明</h3>
          </div>
          <ul className="info-list">
            <li>数据差异 = wm_data(Data) - API数据(ApiData)</li>
            <li>正值表示wm_data高于API数据，负值表示wm_data低于API数据</li>
            <li>新客户：最近15天注册的客户</li>
            <li>老客户：注册时间 ≥ 15天，按30天差异降序排列</li>
            <li>数据范围：最近90天，统计范围：最近30天</li>
            <li>图表展示：趋势对比图显示API数据与wm_data的对比，差异图显示每日数据差异</li>
          </ul>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
