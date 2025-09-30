import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, ReferenceLine } from 'recharts';
import { getApiUrl } from './config';

const ChartView = ({ selectedTenant }) => {
  const [tenantData, setTenantData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  // 获取租户数据
  const fetchTenantData = async (name) => {
    try {
      setLoading(true);
      const response = await fetch(getApiUrl(`/api/alter-data/${name}`));
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      
      // 按照顺序拼接数据：ErrTenantData -> NewTenantData -> TenantData
      const allData = [
        ...(data.ErrTenantData || []), 
        ...(data.NewTenantData || []), 
        ...(data.TenantData || [])
      ];
      setTenantData(allData);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (selectedTenant && selectedTenant !== 'config') {
      fetchTenantData(selectedTenant);
    }
  }, [selectedTenant]);

  // 格式化数据用于图表显示
  const formatChartData = (data) => {
    return data.Data.map(item => ({
      date: item.Date,
      apiData: item.APiData,
      wmData: item.WMData
    })).reverse(); // 反转数组使日期从早到晚排列
  };

  // 自定义WM数据点渲染函数 - 只在最后7天且数据为0时显示空心点
  const customWmDot = (props, tenant) => {
    const { cx, cy, payload } = props;
    const wmValue = payload.wmData;
    const currentDate = payload.date;
    
    // 获取最近7天的日期范围
    const last7Days = tenant.Data.slice(0, 7).map(item => item.Date);
    
    // 只有在最近7天且WM数据为0时才显示空心点
    if (last7Days.includes(currentDate) && wmValue === 0) {
      return (
        <circle 
          cx={cx} 
          cy={cy} 
          r={3} 
          fill="transparent" 
          stroke="#82ca9d" 
          strokeWidth={2}
        />
      );
    }
    return null; // 其他情况不显示点
  };

  // 计算第30天的日期
  const get30DaysAgoDate = (lastSyncDate) => {
    const date = new Date(lastSyncDate);
    date.setDate(date.getDate() - 30);
    return date.toISOString().split('T')[0];
  };

  if (loading) {
    return (
      <div className="chart-view">
        <h2>租户数据图表</h2>
        <div className="loading">加载中...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="chart-view">
        <h2>租户数据图表</h2>
        <div className="error">错误: {error}</div>
      </div>
    );
  }

  if (!tenantData || tenantData.length === 0) {
    return (
      <div className="chart-view">
        <h2>租户数据图表</h2>
        <div>暂无数据</div>
      </div>
    );
  }

  return (
    <div className="chart-view">
      <h2>租户数据图表</h2>
      <div className="charts-grid">
        {tenantData.map((tenant, index) => (
          <div key={tenant.TenantId} className="tenant-chart">
          <h3>
            租户 ID: {tenant.TenantId} 
            {tenant.IsNewTenant && <span className="new-tenant-badge">新租户</span>}
          </h3>
          <div className="tenant-tags">
            {tenant.ErrTags && tenant.ErrTags.map((tag, tagIndex) => (
              <span key={`err-${tagIndex}`} className="tag error-tag">{tag}</span>
            ))}
            {tenant.Tags && tenant.Tags.map((tag, tagIndex) => (
              <span key={`tag-${tagIndex}`} className="tag">{tag}</span>
            ))}
          </div>
          
          <ResponsiveContainer width="100%" height={400}>
            <LineChart
              data={formatChartData(tenant)}
              margin={{
                top: 20,
                right: 30,
                left: 20,
                bottom: 5,
              }}
            >
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis 
                dataKey="date" 
                tick={{ fontSize: 12 }}
                interval="preserveStartEnd"
              />
              <YAxis tick={{ fontSize: 12 }} />
              <Tooltip 
                formatter={(value, name) => [
                  value.toLocaleString(), 
                  name === 'apiData' ? 'API 数据' : 'WM 数据'
                ]}
                labelFormatter={(label) => `日期: ${label}`}
              />
              <Legend 
                formatter={(value) => {
                  if (value === 'apiData') return 'API 数据';
                  if (value === 'wmData') return 'WM 数据';
                  return value;
                }}
              />
              {tenant.HaveApiData && (
                <Line 
                  type="monotone" 
                  dataKey="apiData" 
                  stroke="#8884d8" 
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 5 }}
                />
              )}
              <Line 
                type="monotone" 
                dataKey="wmData" 
                stroke="#82ca9d" 
                strokeWidth={2}
                dot={(props) => customWmDot(props, tenant)}
                activeDot={{ r: 5 }}
              />
              <ReferenceLine 
                x={tenant.LastSyncDate} 
                stroke="#ff7300" 
                strokeDasharray="5 5"
                strokeWidth={2}
                label={{ value: tenant.LastSyncDate, position: "top" }}
              />
              <ReferenceLine 
                x={get30DaysAgoDate(tenant.LastSyncDate)} 
                stroke="#e74c3c" 
                strokeDasharray="3 3"
                strokeWidth={2}
                label={{ value: "30天前", position: "top" }}
              />
            </LineChart>
          </ResponsiveContainer>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ChartView;
