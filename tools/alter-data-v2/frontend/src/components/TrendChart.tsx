import React from 'react'
import { Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, BarChart, Bar, ReferenceLine, Area, ComposedChart } from 'recharts'
import './TrendChart.css'

interface DateSequence {
  date: string
  api_data: number
  data: number
  remove_data: number
}

interface TrendChartProps {
  title: string
  data: DateSequence[]
  showDifference?: boolean
  dataType?: string
  lastDataDate?: string
}

const TrendChart: React.FC<TrendChartProps> = ({ title, data, showDifference = false, dataType, lastDataDate }) => {
  // 数据安全检查
  if (!data || !Array.isArray(data) || data.length === 0) {
    return (
      <div className="trend-chart">
        <h4 className="chart-title">{title}</h4>
        <div style={{ padding: '40px', textAlign: 'center', color: '#999' }}>
          暂无数据
        </div>
      </div>
    )
  }

  // 判断是否为WM-only数据类型
  const isWmOnly = dataType === 'wm_only'
  
  // 计算过去30天的日期范围
  const thirtyDaysAgo = new Date()
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)

  // 计算最近3天的日期范围
  const threeDaysAgo = new Date()
  threeDaysAgo.setDate(threeDaysAgo.getDate() - 3)

  // 转换数据格式，添加差异计算和最近3天标记
  let chartData = data.map((item, index) => {
    const itemDate = new Date(item.date)
    const isLast30Days = itemDate >= thirtyDaysAgo
    const isLast3Days = itemDate >= threeDaysAgo
    const hasRemoveData = item.remove_data !== 0

    return {
      date: new Date(item.date).toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }),
      'API数据': isWmOnly ? null : item.api_data,  // WM-only类型不显示API数据
      'wm_data': item.data,
      'Data+RemoveData': (isWmOnly || !isLast30Days || !hasRemoveData) ? null : item.data + item.remove_data,  // WM-only类型不显示
      '差异': isWmOnly ? null : item.data - item.api_data,  // WM-only类型不显示差异
      originalDate: item.date,
      index: index,
      isRecent3Days: isLast3Days, // 标记是否为最近3天
    }
  })

  // 对于WM-only类型，在最后添加一个空的数据点以扩展横坐标
  if (isWmOnly && chartData.length > 0) {
    const lastDataPoint = chartData[chartData.length - 1]
    const lastDate = new Date(lastDataPoint.originalDate)
    const nextDate = new Date(lastDate)
    nextDate.setDate(lastDate.getDate() + 1)
    
    chartData.push({
      date: nextDate.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }),
      'API数据': null,
      'wm_data': null, // 空数据点，不显示
      'Data+RemoveData': null,
      '差异': null,
      originalDate: nextDate.toISOString().split('T')[0],
      index: chartData.length,
      isRecent3Days: false,
    })
  }

  // 检查是否需要显示 Data+RemoveData 线（是否有有效数据）
  const shouldShowRemoveDataLine = chartData.some(item => item['Data+RemoveData'] !== null)

    // 计算参考线位置（从后往前第30天和第60天）
  const getReferenceLinesData = () => {
    if (!chartData.length) return []

    const lines: Array<{type: string, date: string, index: number}> = []
    const latestDate = new Date(chartData[chartData.length - 1].originalDate)
    
    // 第30天参考线
    const date30DaysAgo = new Date(latestDate)
    date30DaysAgo.setDate(latestDate.getDate() - 30)
    
    // 第60天参考线
    const date60DaysAgo = new Date(latestDate)
    date60DaysAgo.setDate(latestDate.getDate() - 60)
    
    // 找到对应的数据点
    chartData.forEach((item, index) => {
      const itemDate = new Date(item.originalDate)
      
      // 找最接近30天前的数据点
      if (Math.abs(itemDate.getTime() - date30DaysAgo.getTime()) < 24 * 60 * 60 * 1000) {
        lines.push({
          type: '30天前',
          date: item.date,
          index: index
        })
      }
      
      // 找最接近60天前的数据点
      if (Math.abs(itemDate.getTime() - date60DaysAgo.getTime()) < 24 * 60 * 60 * 1000) {
        lines.push({
          type: '60天前',
          date: item.date,
          index: index
        })
      }
    })
    
    return lines
  }

  const referenceLines = getReferenceLinesData() || []

  // 获取WM-only类型的最后数据日期参考线
  const getWmOnlyReferenceLines = () => {
    if (!isWmOnly || !chartData || chartData.length === 0) return []
    
    // 从后往前找最后一次数据大于0的日期
    let lastValidDataIndex = -1
    for (let i = chartData.length - 1; i >= 0; i--) {
      if (chartData[i].wm_data > 0) {
        lastValidDataIndex = i
        break
      }
    }
    
    // 如果没有找到大于0的数据，不显示参考线
    if (lastValidDataIndex === -1) return []
    
    const lastValidDataPoint = chartData[lastValidDataIndex]
    
    // 格式化日期显示为两行
    const dateObj = new Date(lastValidDataPoint.originalDate)
    const year = dateObj.getFullYear()
    const monthDay = dateObj.toLocaleDateString('zh-CN', {
      month: '2-digit', 
      day: '2-digit'
    }).replace(/\//g, '-')
    
    return [{
      type: `${year}\n${monthDay}`,
      date: lastValidDataPoint.date,
      index: lastValidDataIndex
    }]
  }

  const wmOnlyReferenceLines = getWmOnlyReferenceLines() || []

  // 格式化Y轴数值为k格式
  const formatYAxisValue = (value: number) => {
    if (value >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`
    } else if (value >= 1000) {
      return `${(value / 1000).toFixed(1)}k`
    }
    return value.toString()
  }

  // 格式化工具提示中的数值
  const formatTooltipValue = (value: number) => {
    if (value >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`
    } else if (value >= 1000) {
      return `${(value / 1000).toFixed(1)}k`
    }
    return value.toLocaleString('en-US', { 
      maximumFractionDigits: 0,
      useGrouping: true 
    })
  }

  // 获取数据项的显示名称
  const getDataLabel = (dataKey: string) => {
    const labelMap: { [key: string]: string } = {
      'API数据': 'API 数据',
      'wm_data': 'Overview 数据',
      'Data+RemoveData': 'Data+RemoveData',
      '差异': '数据差异'
    }
    return labelMap[dataKey] || dataKey
  }

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      // 获取原始日期数据
      const dataPoint = chartData.find(item => item.date === label)
      const originalDate = dataPoint?.originalDate
      const formattedDate = originalDate ? new Date(originalDate).toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit', 
        day: '2-digit'
      }).replace(/\//g, '-') : label

      return (
        <div className="echarts-tooltip">
          <div className="tooltip-date"><strong>{formattedDate}</strong></div>
          {payload.map((entry: any, index: number) => (
            entry.value !== null && (
              <div key={index} className="tooltip-item">
                <span 
                  className="tooltip-dot"
                  style={{
                    backgroundColor: entry.color
                  }}
                />
                <span className="tooltip-text">
                  {getDataLabel(entry.dataKey)}: {formatTooltipValue(entry.value)}
                </span>
              </div>
            )
          ))}
        </div>
      )
    }
    return null
  }

  if (showDifference) {
    // 显示差异的柱状图
    return (
      <div className="trend-chart">
        <h4 className="chart-title">{title}</h4>
        <ResponsiveContainer width="100%" height={280}>
          <BarChart data={chartData} margin={{ top: 20, right: 30, left: 10, bottom: 5 }}>
            <defs>
              <linearGradient id="barGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.8}/>
                <stop offset="100%" stopColor="#3b82f6" stopOpacity={0.4}/>
              </linearGradient>
            </defs>
            <CartesianGrid 
              strokeDasharray="1 1" 
              stroke="#f0f0f0" 
              strokeWidth={1}
              opacity={0.6}
            />
            <XAxis 
              dataKey="date" 
              tick={{ fontSize: 12, fill: '#8c8c8c' }}
              axisLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
              tickLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
              tickMargin={8}
            />
            <YAxis 
              tickFormatter={formatYAxisValue}
              tick={{ fontSize: 12, fill: '#8c8c8c' }}
              axisLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
              tickLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
              tickMargin={8}
            />
            <Tooltip content={<CustomTooltip />} />
            <Legend />
            <Bar 
              dataKey="差异" 
              fill="url(#barGradient)" 
              name="数据差异"
              radius={[2, 2, 0, 0]}
            />
            {/* 添加参考线 */}
            {referenceLines.map((line, index) => (
              <ReferenceLine 
                key={index}
                x={line.date}
                stroke={line.type === '30天前' ? '#ef4444' : '#8b5cf6'}
                strokeWidth={2}
                strokeDasharray="5 5"
                label={{ 
                  value: line.type, 
                  position: 'top',
                  style: { 
                    fontSize: '12px',
                    fill: line.type === '30天前' ? '#ef4444' : '#8b5cf6',
                    fontWeight: 'bold'
                  }
                }}
              />
            ))}
          </BarChart>
        </ResponsiveContainer>
      </div>
    )
  }

  // 显示趋势对比的折线图
  return (
    <div className="trend-chart">
      <h4 className="chart-title">{title}</h4>
      <ResponsiveContainer width="100%" height={280}>
        <ComposedChart data={chartData} margin={{ top: 30, right: 80, left: 10, bottom: 5 }}>
          <defs>
            <linearGradient id="areaGradient1" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#3498db" stopOpacity={0.3}/>
              <stop offset="100%" stopColor="#3498db" stopOpacity={0.05}/>
            </linearGradient>
            <linearGradient id="areaGradient2" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#e74c3c" stopOpacity={0.3}/>
              <stop offset="100%" stopColor="#e74c3c" stopOpacity={0.05}/>
            </linearGradient>
          </defs>
          <CartesianGrid 
            strokeDasharray="1 1" 
            stroke="#f0f0f0" 
            strokeWidth={1}
            opacity={0.6}
          />
          <XAxis 
            dataKey="date" 
            tick={{ fontSize: 12, fill: '#8c8c8c' }}
            axisLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
            tickLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
            tickMargin={8}
          />
          <YAxis 
            tickFormatter={formatYAxisValue}
            tick={{ fontSize: 12, fill: '#8c8c8c' }}
            axisLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
            tickLine={{ stroke: '#d9d9d9', strokeWidth: 1 }}
            tickMargin={8}
          />
          <Tooltip content={<CustomTooltip />} />
          <Legend />
          {!isWmOnly && (
            <Area 
              type="monotone" 
              dataKey="API数据" 
              name="API 数据"
              stroke="#3498db" 
              strokeWidth={2.5}
              fill="url(#areaGradient1)"
              dot={false}
              activeDot={{ r: 4, fill: '#3498db', strokeWidth: 2, stroke: '#fff' }}
            />
          )}
          <Area 
            type="monotone" 
            dataKey="wm_data"
            name="Overview 数据"
            stroke="#e74c3c" 
            strokeWidth={2.5}
            fill="url(#areaGradient2)"
            dot={false}
            activeDot={{ r: 4, fill: '#e74c3c', strokeWidth: 2, stroke: '#fff' }}
            connectNulls={false}
          />
          {shouldShowRemoveDataLine && (
            <Line 
              type="monotone" 
              dataKey="Data+RemoveData"
              name="Data+RemoveData"
              stroke="#f87171" 
              strokeWidth={3}
              strokeDasharray="10 5"
              dot={false}
              activeDot={{ r: 4, fill: '#f87171' }}
              connectNulls={false}
            />
          )}
          
          {/* 显示最近3天的数据点 */}
          <Line
            type="monotone"
            dataKey="API数据"
            stroke="transparent"
            strokeWidth={0}
            dot={(props: any) => {
              const { payload } = props;
              if (payload && payload.isRecent3Days) {
                return <circle cx={props.cx} cy={props.cy} r={3} fill="#3498db" stroke="#fff" strokeWidth={2} />;
              }
              return <></>;
            }}
            activeDot={false}
            connectNulls={false}
          />
          <Line
            type="monotone"
            dataKey="wm_data"
            stroke="transparent"
            strokeWidth={0}
            dot={(props: any) => {
              const { payload } = props;
              if (payload && payload.isRecent3Days) {
                const isZero = payload['wm_data'] === 0;
                return (
                  <circle 
                    cx={props.cx} 
                    cy={props.cy} 
                    r={3} 
                    fill={isZero ? '#fff' : '#e74c3c'} 
                    stroke="#e74c3c" 
                    strokeWidth={2} 
                  />
                );
              }
              return <></>;
            }}
            activeDot={false}
            connectNulls={false}
          />
          {/* WM-only类型显示最后数据日期参考线 */}
          {isWmOnly && wmOnlyReferenceLines.map((line, index) => (
            <ReferenceLine 
              key={`wm-${index}`}
              x={line.date}
              stroke="#10b981"
              strokeWidth={2}
              strokeDasharray="5 5"
              label={{ 
                value: line.type, 
                position: 'top',
                offset: 5,
                style: { 
                  fontSize: '10px',
                  fill: '#10b981',
                  fontWeight: 'bold',
                  textAnchor: 'middle',
                  dominantBaseline: 'auto',
                  backgroundColor: 'rgba(255, 255, 255, 0.8)',
                  padding: '2px 4px',
                  borderRadius: '3px'
                }
              }}
            />
          ))}
          
          {/* 非WM-only类型显示30天/60天参考线 */}
          {!isWmOnly && referenceLines.map((line, index) => (
            <ReferenceLine 
              key={index}
              x={line.date}
              stroke={line.type === '30天前' ? '#ef4444' : '#8b5cf6'}
              strokeWidth={2}
              strokeDasharray="5 5"
              label={{ 
                value: line.type, 
                position: 'top',
                style: { 
                  fontSize: '12px',
                  fill: line.type === '30天前' ? '#ef4444' : '#8b5cf6',
                  fontWeight: 'bold'
                }
              }}
            />
          ))}
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  )
}

export default TrendChart

