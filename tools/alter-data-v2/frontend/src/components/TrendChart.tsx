import React from 'react'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, BarChart, Bar, ReferenceLine } from 'recharts'
import './TrendChart.css'

interface DateSequence {
  date: string
  api_data: number
  data: number
}

interface TrendChartProps {
  title: string
  data: DateSequence[]
  showDifference?: boolean
}

const TrendChart: React.FC<TrendChartProps> = ({ title, data, showDifference = false }) => {
  // 转换数据格式，添加差异计算
  const chartData = data.map((item, index) => ({
    date: new Date(item.date).toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }),
    'API数据': item.api_data,
    'wm_data': item.data,
    '差异': item.data - item.api_data,
    originalDate: item.date,
    index: index,
  }))

  // 计算参考线位置（从后往前第30天和第60天）
  const getReferenceLinesData = () => {
    if (!chartData.length) return []
    
    const lines = []
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

  const referenceLines = getReferenceLinesData()

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
    return value.toLocaleString()
  }

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="chart-tooltip">
          <p className="tooltip-label">{`日期: ${label}`}</p>
          {payload.map((entry: any, index: number) => (
            <p key={index} style={{ color: entry.color }}>
              {`${entry.dataKey}: ${formatTooltipValue(entry.value)}`}
            </p>
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
        <ResponsiveContainer width="100%" height={320}>
          <BarChart data={chartData} margin={{ top: 20, right: 30, left: -30, bottom: 5 }}>
            <defs>
              <linearGradient id="barGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.8}/>
                <stop offset="100%" stopColor="#3b82f6" stopOpacity={0.4}/>
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis 
              dataKey="date" 
              tick={{ fontSize: 12 }}
              stroke="#6b7280"
            />
            <YAxis 
              tickFormatter={formatYAxisValue}
              tick={{ fontSize: 12 }}
              stroke="#6b7280"
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
      <ResponsiveContainer width="100%" height={320}>
        <LineChart data={chartData} margin={{ top: 20, right: 30, left: -30, bottom: 5 }}>
          <defs>
            <linearGradient id="areaGradient1" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#f87171" stopOpacity={0.1}/>
              <stop offset="100%" stopColor="#f87171" stopOpacity={0.02}/>
            </linearGradient>
            <linearGradient id="areaGradient2" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#60a5fa" stopOpacity={0.1}/>
              <stop offset="100%" stopColor="#60a5fa" stopOpacity={0.02}/>
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
          <XAxis 
            dataKey="date" 
            tick={{ fontSize: 12 }}
            stroke="#6b7280"
          />
          <YAxis 
            tickFormatter={formatYAxisValue}
            tick={{ fontSize: 12 }}
            stroke="#6b7280"
          />
          <Tooltip content={<CustomTooltip />} />
          <Legend />
          <Line 
            type="monotone" 
            dataKey="API数据" 
            stroke="#f87171" 
            strokeWidth={3}
            dot={false}
            activeDot={{ r: 4, fill: '#f87171' }}
          />
          <Line 
            type="monotone" 
            dataKey="wm_data"
            stroke="#60a5fa" 
            strokeWidth={3}
            dot={false}
            activeDot={{ r: 4, fill: '#60a5fa' }}
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
                position: 'topLeft',
                style: { 
                  fontSize: '12px',
                  fill: line.type === '30天前' ? '#ef4444' : '#8b5cf6',
                  fontWeight: 'bold'
                }
              }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

export default TrendChart
