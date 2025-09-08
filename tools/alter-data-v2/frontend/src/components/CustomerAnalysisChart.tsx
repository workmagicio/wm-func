import React from 'react'
import './CustomerAnalysisChart.css'

interface AttributionTenantData {
  tenant_id: number
  date_sequence: any[]
  platform_totals: any[]
  total_attribution_avg: number
  tags: string[]
  recent_zero_days: number
  has_recent_zeros: boolean
  customer_type?: string
  register_time?: string
}

interface CustomerWithDiff extends AttributionTenantData {
  diff30Day: number
  isNew15Days: boolean
}

interface CustomerAnalysisChartProps {
  allCustomersData: AttributionTenantData[]
}

const CustomerAnalysisChart: React.FC<CustomerAnalysisChartProps> = ({ 
  allCustomersData 
}) => {
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
    if (!allCustomersData || allCustomersData.length === 0) {
      return { newCustomers: [], oldCustomers: [] }
    }

    // 计算15天前的日期（用于新客户判断）
    const fifteenDaysAgo = new Date()
    fifteenDaysAgo.setDate(fifteenDaysAgo.getDate() - 15)

    // 分类并计算差异
    const customersWithDiff: CustomerWithDiff[] = allCustomersData.map(customer => ({
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
  }, [allCustomersData])

  if (!allCustomersData || allCustomersData.length === 0) {
    return (
      <div className="customer-analysis-chart">
        <h4 className="analysis-title">📊 数据差异分析</h4>
        <div className="no-data">
          暂无客户数据进行分析
        </div>
      </div>
    )
  }

  return (
    <div className="customer-analysis-chart">
      <h4 className="analysis-title">📊 数据差异分析</h4>
      
      {/* 概览统计 */}
      <div className="analysis-overview">
        <div className="overview-item">
          <span className="overview-label">🌟 最近15天新客户:</span>
          <span className="overview-value">{customerAnalysisData.newCustomers.length} 个</span>
        </div>
        <div className="overview-item">
          <span className="overview-label">👥 老客户:</span>
          <span className="overview-value">{customerAnalysisData.oldCustomers.length} 个</span>
        </div>
        <div className="overview-item">
          <span className="overview-label">📈 总客户数:</span>
          <span className="overview-value">{allCustomersData.length} 个</span>
        </div>
      </div>
      
      {/* 新客户分析 */}
      <div className="customer-group">
        <div className="group-header">
          <span className="group-icon">🌟</span>
          <span className="group-title">最近15天注册的客户</span>
          <span className="group-count">共 {customerAnalysisData.newCustomers.length} 个新客户（按30天差异绝对值降序排列）</span>
        </div>
        
        {customerAnalysisData.newCustomers.length > 0 ? (
          <div className="customer-list">
            {customerAnalysisData.newCustomers.map((customer, index) => (
              <div key={customer.tenant_id} className="customer-item">
                <span className="customer-rank">#{index + 1}</span>
                <span className="customer-id">租户 {customer.tenant_id}</span>
                <span className="customer-register-time">
                  注册: {customer.register_time ? new Date(customer.register_time).toLocaleDateString('zh-CN') : '未知'}
                </span>
                <span className="customer-avg">
                  日均: {customer.total_attribution_avg?.toFixed(1) || '0'}
                </span>
                <span className="customer-diff">
                  30天差异: {customer.diff30Day.toLocaleString()}
                </span>
                {customer.has_recent_zeros && (
                  <span className="zero-warning">⚠️ 最近有零值</span>
                )}
              </div>
            ))}
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
            {customerAnalysisData.oldCustomers.slice(0, 10).map((customer, index) => (
              <div key={customer.tenant_id} className="customer-item">
                <span className="customer-rank">#{index + 1}</span>
                <span className="customer-id">租户 {customer.tenant_id}</span>
                <span className="customer-register-time">
                  注册: {customer.register_time ? new Date(customer.register_time).toLocaleDateString('zh-CN') : '未知'}
                </span>
                <span className="customer-avg">
                  日均: {customer.total_attribution_avg?.toFixed(1) || '0'}
                </span>
                <span className="customer-diff">
                  30天差异: {customer.diff30Day.toLocaleString()}
                </span>
                {customer.has_recent_zeros && (
                  <span className="zero-warning">⚠️ 最近有零值</span>
                )}
              </div>
            ))}
            {customerAnalysisData.oldCustomers.length > 10 && (
              <div className="more-customers">
                还有 {customerAnalysisData.oldCustomers.length - 10} 个老客户...
              </div>
            )}
          </div>
        ) : (
          <div className="no-customers">暂无老客户数据</div>
        )}
      </div>
    </div>
  )
}

export default CustomerAnalysisChart
