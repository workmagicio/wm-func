/**
 * Amazon订单图表管理器
 */
class AmazonOrdersChartManager {
    constructor() {
        this.chartInstances = new Map();
    }

    /**
     * 渲染租户图表
     */
    renderTenantChart(tenantData, containerId) {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container not found: ${containerId}`);
            return;
        }

        // 销毁旧图表实例
        if (this.chartInstances.has(containerId)) {
            this.chartInstances.get(containerId).dispose();
        }

        // 创建新图表实例
        const chart = echarts.init(container);
        this.chartInstances.set(containerId, chart);

        // 准备数据
        const dates = tenantData.date_range;
        const ordersData = tenantData.orders_data;
        const processedOrders = tenantData.processed_orders;
        const averageLine = Array(dates.length).fill(tenantData.daily_average);

        // 找出掉0的日期索引，用于添加竖线标记
        const zeroMarkLines = [];
        processedOrders.forEach((processedValue, index) => {
            if (processedValue === -100) { // 掉0异常
                const date = dates[index];
                let displayDate = date;
                if (typeof date === 'string') {
                    displayDate = date.split('T')[0].split(' ')[0];
                    if (displayDate.includes('+')) {
                        displayDate = displayDate.split('+')[0];
                    }
                }
                // 格式化为月-日格式 (MM-DD)
                const dateObj = new Date(displayDate);
                const month = String(dateObj.getMonth() + 1).padStart(2, '0');
                const day = String(dateObj.getDate()).padStart(2, '0');
                const shortDate = `${month}-${day}`;
                
                zeroMarkLines.push({
                    xAxis: index,
                    label: {
                        show: true,
                        position: 'insideEndTop',
                        formatter: shortDate,
                        fontSize: 11,
                        color: '#dc3545',
                        backgroundColor: 'rgba(255, 255, 255, 0.9)',
                        padding: [3, 6],
                        borderColor: '#dc3545',
                        borderWidth: 1,
                        borderRadius: 3,
                        rotate: 0  // 确保标签水平显示
                    },
                    lineStyle: {
                        color: '#dc3545',
                        type: 'dashed',
                        width: 2
                    }
                });
            }
        });

        // 构建系列数据
        const series = [
            {
                name: '订单数量',
                type: 'line',
                data: ordersData.map((value, index) => {
                    const processedValue = processedOrders[index];
                    // 根据异常标记设置样式
                    if (processedValue === -200) {
                        // 凹形异常
                        return {
                            value: value,
                            itemStyle: {
                                color: '#ff9800',
                                borderColor: '#e68900',
                                borderWidth: 2,
                                shadowColor: 'rgba(255, 152, 0, 0.5)',
                                shadowBlur: 8
                            },
                            symbol: 'diamond',
                            symbolSize: 10
                        };
                    } else if (processedValue === -100) {
                        // 掉0异常
                        return {
                            value: value,
                            itemStyle: {
                                color: '#dc3545',
                                borderColor: '#c82333',
                                borderWidth: 2,
                                shadowColor: 'rgba(220, 53, 69, 0.5)',
                                shadowBlur: 8
                            },
                            symbol: 'triangle',
                            symbolSize: 10
                        };
                    } else {
                        // 正常数据
                        return {
                            value: value,
                            itemStyle: {
                                color: '#0d6efd'
                            }
                        };
                    }
                }),
                smooth: true,
                lineStyle: {
                    width: 2,
                    color: '#0d6efd'
                },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        { offset: 0, color: 'rgba(13, 110, 253, 0.3)' },
                        { offset: 1, color: 'rgba(13, 110, 253, 0.05)' }
                    ])
                },
                emphasis: {
                    focus: 'series',
                    lineStyle: {
                        width: 3
                    }
                },
                markLine: zeroMarkLines.length > 0 ? {
                    data: zeroMarkLines,
                    symbol: 'none',
                    silent: false
                } : undefined
            },
            {
                name: '平均值',
                type: 'line',
                data: averageLine,
                lineStyle: {
                    type: 'dashed',
                    width: 1,
                    color: '#6c757d'
                },
                itemStyle: {
                    opacity: 0
                },
                symbol: 'none',
                silent: true
            }
        ];

        // 图表配置
        const option = {
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'cross',
                    crossStyle: {
                        color: '#999'
                    }
                },
                formatter: (params) => {
                    if (!params || params.length === 0) return '';
                    
                    const dataIndex = params[0].dataIndex;
                    const date = dates[dataIndex];
                    const orders = ordersData[dataIndex];
                    const processedValue = processedOrders[dataIndex];
                    const average = tenantData.daily_average;
                    
                    // 格式化日期显示（去掉时区信息）
                    let displayDate = date;
                    if (typeof date === 'string') {
                        displayDate = date.split('T')[0].split(' ')[0];
                        if (displayDate.includes('+')) {
                            displayDate = displayDate.split('+')[0];
                        }
                    }
                    
                    let tooltipHtml = `
                        <div style="margin-bottom: 8px;">
                            <strong>${displayDate}</strong>
                        </div>
                        <div style="display: flex; align-items: center; margin-bottom: 4px;">
                            <span style="display: inline-block; width: 8px; height: 8px; background: #0d6efd; border-radius: 50%; margin-right: 8px;"></span>
                            订单数量: <strong>${orders.toLocaleString()}</strong>
                        </div>
                        <div style="display: flex; align-items: center; margin-bottom: 4px;">
                            <span style="display: inline-block; width: 8px; height: 8px; background: #6c757d; border-radius: 50%; margin-right: 8px;"></span>
                            日均值: <strong>${average.toFixed(1)}</strong>
                        </div>
                    `;
                    
                    // 添加异常标记说明
                    if (processedValue === -200) {
                        tooltipHtml += `
                            <div style="color: #ff9800; margin-top: 8px; padding: 4px; background: rgba(255, 152, 0, 0.1); border-radius: 4px;">
                                🔶 凹形异常：数据缺失模式
                            </div>
                        `;
                    } else if (processedValue === -100) {
                        tooltipHtml += `
                            <div style="color: #dc3545; margin-top: 8px; padding: 4px; background: rgba(220, 53, 69, 0.1); border-radius: 4px;">
                                🚨 掉0异常：订单突然变为0
                            </div>
                        `;
                    }
                    
                    return tooltipHtml;
                }
            },
            grid: {
                left: 50,
                right: 30,
                top: 30,
                bottom: 50
            },
            xAxis: {
                type: 'category',
                data: dates,
                axisLine: {
                    lineStyle: {
                        color: '#e9ecef'
                    }
                },
                axisTick: {
                    lineStyle: {
                        color: '#e9ecef'
                    }
                },
                axisLabel: {
                    color: '#6c757d',
                    fontSize: 11,
                    interval: Math.floor(dates.length / 8), // 显示约8个标签
                    rotate: 0,
                    formatter: (value) => {
                        // 处理各种日期格式，只保留YYYY-MM-DD
                        if (typeof value === 'string') {
                            // 去掉时区信息和时间信息
                            let dateOnly = value.split('T')[0].split(' ')[0];
                            // 去掉可能的时区标识符如+08:00
                            if (dateOnly.includes('+')) {
                                dateOnly = dateOnly.split('+')[0];
                            }
                            return dateOnly;
                        }
                        return value;
                    }
                }
            },
            yAxis: {
                type: 'value',
                axisLine: {
                    show: false
                },
                axisTick: {
                    show: false
                },
                axisLabel: {
                    color: '#6c757d',
                    fontSize: 11,
                    formatter: (value) => {
                        if (value >= 1000) {
                            return (value / 1000).toFixed(0) + 'k';
                        }
                        return value.toString();
                    }
                },
                splitLine: {
                    lineStyle: {
                        color: 'rgba(233, 236, 239, 0.6)',
                        type: 'dashed'
                    }
                }
            },
            series: series,
            legend: {
                show: false
            },
            animation: true,
            animationDuration: 1000,
            animationEasing: 'cubicOut'
        };

        // 设置配置并渲染
        chart.setOption(option);

        // 响应式处理
        const resizeObserver = new ResizeObserver(() => {
            chart.resize();
        });
        resizeObserver.observe(container);

        return chart;
    }

    /**
     * 销毁所有图表实例
     */
    disposeAll() {
        this.chartInstances.forEach(chart => {
            chart.dispose();
        });
        this.chartInstances.clear();
    }

    /**
     * 销毁指定图表实例
     */
    dispose(containerId) {
        if (this.chartInstances.has(containerId)) {
            this.chartInstances.get(containerId).dispose();
            this.chartInstances.delete(containerId);
        }
    }

    /**
     * 调整所有图表大小
     */
    resizeAll() {
        this.chartInstances.forEach(chart => {
            chart.resize();
        });
    }
}
