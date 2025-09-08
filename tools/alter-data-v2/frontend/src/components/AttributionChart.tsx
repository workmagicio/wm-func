import React from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine
} from 'recharts'
import './AttributionChart.css'

interface AttributionDateSequence {
  date: string
  platform_data: { [key: string]: number }
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
  customer_type?: string
  register_time?: string
}

interface AttributionChartProps {
  title: string
  data: AttributionTenantData
  showCustomerAnalysis?: boolean
  allCustomersData?: AttributionTenantData[]
}

const AttributionChart: React.FC<AttributionChartProps> = ({ 
  title, 
  data,
  showCustomerAnalysis = false,
  allCustomersData = []
}) => {
  // 数据安全检查
  if (!data || !data?.date_sequence || !Array.isArray(data?.date_sequence) || data?.date_sequence?.length === 0) {
    return (
      <div className="attribution-chart">
        <h4 className="chart-title">{title}</h4>
        <div style={{ padding: '40px', textAlign: 'center', color: '#999' }}>
          暂无归因数据
        </div>
      </div>
    )
  }

  // 计算最近3天的日期范围
  const threeDaysAgo = new Date()
  threeDaysAgo.setDate(threeDaysAgo.getDate() - 3)

  // 转换数据格式，添加最近3天标记
  const chartData = data?.date_sequence?.map((item, index) => {
    const itemDate = new Date(item?.date || '')
    const isLast3Days = itemDate >= threeDaysAgo

    return {
      date: new Date(item?.date || '').toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }),
      '总归因': item?.total_attribution || 0,
      'Meta归因': item?.platform_data?.['Meta'] || 0,
      'Google归因': item?.platform_data?.['Google'] || 0,
      'TikTok归因': item?.platform_data?.['TikTok'] || 0,
      'Pinterest归因': item?.platform_data?.['Pinterest'] || 0,
      'Snapchat归因': item?.platform_data?.['Snapchat'] || 0,
      originalDate: item?.date || '',
      index: index,
      isRecent3Days: isLast3Days,
      isRecentZero: item?.is_recent_zero || false, // 标记是否为最近3天的零值
    }
  }) || []


  // 平台颜色配置
  const platformColors = {
    'Meta归因': '#4267B2',
    'Google归因': '#34A853',
    'TikTok归因': '#000000',
    'Pinterest归因': '#E60023',
    'Snapchat归因': '#FFFC00'
  }

  // 计算30天差异绝对值
  const calculate30DayDiff = (tenant: AttributionTenantData): number => {
    if (!tenant?.date_sequence || tenant.date_sequence.length === 0) return 0
    
    // 获取最近30天的数据
    const last30Days = tenant.date_sequence.slice(-30)
    const totalAttribution = last30Days.reduce((sum, day) => sum + (day?.total_attribution || 0), 0)
    const avgExpected = (tenant?.total_attribution_avg || 0) * last30Days.length
    
    return Math.abs(totalAttribution - avgExpected)
  }

  // 客户差异分析数据
  const customerAnalysisData = React.useMemo(() => {
    if (!showCustomerAnalysis || !allCustomersData || allCustomersData.length === 0) {
      return { newCustomers: [], oldCustomers: [] }
    }

    // 计算15天前的日期（用于新客户判断）
    const fifteenDaysAgo = new Date()
    fifteenDaysAgo.setDate(fifteenDaysAgo.getDate() - 15)

    // 分类并计算差异
    const customersWithDiff = allCustomersData.map(customer => ({
      ...customer,
      diff30Day: calculate30DayDiff(customer),
      isNew15Days: customer.register_time ? new Date(customer.register_time) >= fifteenDaysAgo : false
    }))

    // 按15天注册时间分类新客户
    const newCustomers = customersWithDiff
      .filter(customer => customer.isNew15Days)
      .sort((a, b) => b.diff30Day - a.diff30Day) // 按30天差异绝对值降序

    // 老客户按30天差异绝对值降序
    const oldCustomers = customersWithDiff
      .filter(customer => !customer.isNew15Days)
      .sort((a, b) => b.diff30Day - a.diff30Day)

    return { newCustomers, oldCustomers }
  }, [showCustomerAnalysis, allCustomersData])

  return (
    <div className="attribution-chart">
      <h4 className="chart-title">{title}</h4>
      
      {/* 数据统计面板 */}
      <div className="attribution-stats">
        <div className="stats-row">
          <div className="stat-item">
            <span className="stat-label">总归因平均:</span>
            <span className="stat-value">{data?.total_attribution_avg?.toFixed(1) || '0'}</span>
          </div>
          <div className="stat-item">
            <span className="stat-label">最近零值天数:</span>
            <span className={`stat-value ${data?.has_recent_zeros ? 'warning' : ''}`}>
              {data?.recent_zero_days || 0}天
            </span>
          </div>
        </div>
        
        {/* 平台汇总数据 */}
        <div className="platform-totals">
          {data?.platform_totals?.map(platform => (
            <div key={platform?.platform} className="platform-total-item">
              <span className="platform-name">{platform?.platform}:</span>
              <span className="platform-value">
                总计 {platform?.total_attribution || 0} | 日均 {platform?.daily_average?.toFixed(1) || '0'}
              </span>
            </div>
          )) || []}
        </div>

        {/* 标签显示 */}
        {data?.tags && data?.tags?.length > 0 && (
          <div className="attribution-tags">
            {data?.tags?.map((tag, index) => (
              <span 
                key={index} 
                className={`tag ${tag?.startsWith('err_') ? 'error-tag' : 'normal-tag'}`}
              >
                {tag}
              </span>
            )) || []}
          </div>
        )}
      </div>

      {/* 客户差异分析 */}
      {showCustomerAnalysis && (
        <div className="customer-analysis-section">
          <h5 className="analysis-title">📊 数据差异分析</h5>
          
          {/* 新客户分析 */}
          <div className="customer-group">
            <div className="group-header">
              <span className="group-icon">🌟</span>
              <span className="group-title">最近15天注册的客户</span>
              <span className="group-count">共 {customerAnalysisData.newCustomers.length} 个新客户（按30天差异绝对值降序排列）</span>
            </div>
            
            {customerAnalysisData.newCustomers.length > 0 ? (
              <div className="customer-list">
                {customerAnalysisData.newCustomers.slice(0, 5).map((customer: any) => (
                  <div key={customer.tenant_id} className="customer-item">
                    <span className="customer-id">租户 {customer.tenant_id}</span>
                    <span className="customer-register-time">
                      注册: {customer.register_time ? new Date(customer.register_time).toLocaleDateString('zh-CN') : '未知'}
                    </span>
                    <span className="customer-diff">
                      差异: {customer.diff30Day.toLocaleString()}
                    </span>
                  </div>
                ))}
                {customerAnalysisData.newCustomers.length > 5 && (
                  <div className="more-customers">
                    还有 {customerAnalysisData.newCustomers.length - 5} 个新客户...
                  </div>
                )}
              </div>
            ) : (
              <div className="no-customers">暂无最近15天注册的客户</div>
            )}
          </div>

          {/* 老客户分析 */}
          <div className="customer-group">
            <div className="group-header">
              <span className="group-icon">👥</span>
              <span className="group-title">老客户</span>
              <span className="group-count">共 {customerAnalysisData.oldCustomers.length} 个老客户（按30天差异绝对值降序排列）</span>
            </div>
            
            {customerAnalysisData.oldCustomers.length > 0 ? (
              <div className="customer-list">
                {customerAnalysisData.oldCustomers.slice(0, 5).map((customer: any) => (
                  <div key={customer.tenant_id} className="customer-item">
                    <span className="customer-id">租户 {customer.tenant_id}</span>
                    <span className="customer-register-time">
                      注册: {customer.register_time ? new Date(customer.register_time).toLocaleDateString('zh-CN') : '未知'}
                    </span>
                    <span className="customer-diff">
                      差异: {customer.diff30Day.toLocaleString()}
                    </span>
                  </div>
                ))}
                {customerAnalysisData.oldCustomers.length > 5 && (
                  <div className="more-customers">
                    还有 {customerAnalysisData.oldCustomers.length - 5} 个老客户...
                  </div>
                )}
              </div>
            ) : (
              <div className="no-customers">暂无老客户数据</div>
            )}
          </div>
        </div>
      )}

      {/* 图表 */}
      <ResponsiveContainer width="100%" height={400}>
        <LineChart data={chartData} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="date" />
          <YAxis />
          <Tooltip 
            formatter={(value, name) => [value, name]}
            labelFormatter={(label) => `日期: ${label}`}
          />
          <Legend />
          
          {/* 总归因平均线 */}
          {data?.total_attribution_avg > 0 && (
            <ReferenceLine 
              y={data.total_attribution_avg} 
              stroke="#ff7300" 
              strokeDasharray="5 5"
              label={{ 
                value: `总归因平均: ${data.total_attribution_avg.toFixed(1)}`, 
                position: 'top' as const,
                style: { fontSize: '12px', fill: '#ff7300' }
              }}
            />
          )}
          
          {/* 总归因数据线 */}
          <Line
            type="monotone"
            dataKey="总归因"
            stroke="#e74c3c"
            strokeWidth={3}
            dot={false}
            name="总归因"
          />
          
          {/* Shopify API数据线 */}
          <Line
            type="monotone"
            dataKey="Shopify订单"
            stroke="#3498db"
            strokeWidth={2}
            dot={false}
            name="Shopify订单"
          />
          
          {/* 各平台归因数据 */}
          {Object.entries(platformColors).map(([platform, color]) => (
            <Line
              key={platform}
              type="monotone"
              dataKey={platform}
              stroke={color}
              strokeWidth={1}
              dot={false}
              name={platform}
            />
          ))}
          
          {/* 最近3天零值标记 - 空心点 */}
          <Line
            type="monotone"
            dataKey="总归因"
            stroke="transparent"
            strokeWidth={0}
            dot={(props: any) => {
              const { payload } = props;
              if (payload && payload.isRecent3Days && payload.isRecentZero) {
                return (
                  <circle 
                    cx={props.cx} 
                    cy={props.cy} 
                    r={4} 
                    fill="#fff"           // 空心 - 白色填充
                    stroke="#e74c3c"     // 红色边框
                    strokeWidth={2} 
                  />
                );
              }
              return <></>;
            }}
            activeDot={false}
            connectNulls={false}
            name=""
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

export default AttributionChart
