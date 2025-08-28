/**
 * Fairing图表管理器
 */
class FairingChartManager {
    constructor() {
        this.chartInstances = new Map();
    }

    /**
     * 创建租户图表
     */
    createTenantChart(containerId, tenantData) {
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
        const responseData = tenantData.response_data;
        const processedResponses = tenantData.processed_responses;
        const averageLine = Array(dates.length).fill(tenantData.daily_average);

        // 找出掉0的日期索引，用于添加竖线标记
        const zeroMarkLines = [];
        processedResponses.forEach((processedValue, index) => {
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
                        rotate: 0
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
                name: '响应数量',
                type: 'line',
                data: responseData.map((value, index) => {
                    const processedValue = processedResponses[index];
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
                            symbolSize: 12
                        };
                    } else if (value === 0) {
                        // 普通的0值
                        return {
                            value: value,
                            itemStyle: {
                                color: '#6c757d',
                                borderColor: '#545b62',
                                borderWidth: 1
                            },
                            symbol: 'circle',
                            symbolSize: 6
                        };
                    } else {
                        // 正常数据
                        return {
                            value: value,
                            itemStyle: {
                                color: '#007bff',
                                borderColor: '#0056b3',
                                borderWidth: 1
                            },
                            symbol: 'circle',
                            symbolSize: 6
                        };
                    }
                }),
                smooth: false,
                lineStyle: {
                    color: '#007bff',
                    width: 2
                },
                areaStyle: {
                    color: {
                        type: 'linear',
                        x: 0, y: 0, x2: 0, y2: 1,
                        colorStops: [
                            { offset: 0, color: 'rgba(0, 123, 255, 0.3)' },
                            { offset: 1, color: 'rgba(0, 123, 255, 0.05)' }
                        ]
                    }
                },
                markLine: {
                    silent: true,
                    animation: false,
                    data: zeroMarkLines,
                    symbol: 'none'
                }
            },
            {
                name: '平均值',
                type: 'line',
                data: averageLine,
                smooth: false,
                showSymbol: false,
                lineStyle: {
                    color: '#28a745',
                    type: 'dashed',
                    width: 2,
                    opacity: 0.8
                },
                emphasis: {
                    disabled: true
                }
            }
        ];

        // 准备日期标签（简化显示）
        const xAxisData = dates.map(date => {
            if (typeof date === 'string') {
                let cleanDate = date.split('T')[0].split(' ')[0];
                if (cleanDate.includes('+')) {
                    cleanDate = cleanDate.split('+')[0];
                }
                const dateObj = new Date(cleanDate);
                const month = String(dateObj.getMonth() + 1).padStart(2, '0');
                const day = String(dateObj.getDate()).padStart(2, '0');
                return `${month}-${day}`;
            }
            return date;
        });

        // 计算30天前的位置（用于标记线）
        const thirtyDaysAgoIndex = Math.max(0, dates.length - 30);

        // 图表配置
        const option = {
            title: {
                show: false
            },
            tooltip: {
                trigger: 'axis',
                backgroundColor: 'rgba(255, 255, 255, 0.95)',
                borderColor: '#e9ecef',
                borderWidth: 1,
                textStyle: {
                    color: '#495057',
                    fontSize: 12
                },
                formatter: (params) => {
                    const date = dates[params[0].dataIndex];
                    const responseCount = responseData[params[0].dataIndex];
                    const processedValue = processedResponses[params[0].dataIndex];
                    const average = Math.round(tenantData.daily_average);
                    
                    let statusText = '';
                    if (processedValue === -200) {
                        statusText = '<br/><span style="color: #ff9800">🔶 凹形异常</span>';
                    } else if (processedValue === -100) {
                        statusText = '<br/><span style="color: #dc3545">🚨 掉0异常</span>';
                    }
                    
                    return `
                        <div style="font-weight: 600; margin-bottom: 6px;">${date}</div>
                        <div>📋 响应数: <span style="color: #007bff; font-weight: 600;">${responseCount}</span></div>
                        <div>📊 平均值: <span style="color: #28a745; font-weight: 600;">${average}</span></div>
                        ${statusText}
                    `;
                },
                axisPointer: {
                    type: 'cross',
                    crossStyle: {
                        color: '#999'
                    }
                }
            },
            legend: {
                data: ['响应数量', '平均值'],
                top: 10,
                right: 20,
                textStyle: {
                    fontSize: 12,
                    color: '#495057'
                }
            },
            grid: {
                left: 60,
                right: 30,
                top: 60,
                bottom: 80,
                containLabel: false
            },
            xAxis: {
                type: 'category',
                data: xAxisData,
                axisLabel: {
                    fontSize: 10,
                    color: '#6c757d',
                    interval: Math.ceil(dates.length / 10), // 智能间隔显示
                    rotate: 45
                },
                axisLine: {
                    lineStyle: {
                        color: '#dee2e6'
                    }
                },
                splitLine: {
                    show: true,
                    lineStyle: {
                        color: '#f8f9fa',
                        type: 'solid'
                    }
                }
            },
            yAxis: {
                type: 'value',
                name: '响应数',
                nameTextStyle: {
                    color: '#6c757d',
                    fontSize: 11
                },
                axisLabel: {
                    fontSize: 10,
                    color: '#6c757d',
                    formatter: (value) => {
                        if (value >= 1000) {
                            return (value / 1000).toFixed(1) + 'K';
                        }
                        return value;
                    }
                },
                axisLine: {
                    lineStyle: {
                        color: '#dee2e6'
                    }
                },
                splitLine: {
                    lineStyle: {
                        color: '#f8f9fa',
                        type: 'dashed'
                    }
                }
            },
            series: series,
            animation: true,
            animationDuration: 800,
            animationEasing: 'cubicOut'
        };

        // 添加30天前标记线（如果数据超过30天）
        if (dates.length > 30) {
            option.series.push({
                name: '30天前',
                type: 'line',
                markLine: {
                    symbol: 'none',
                    label: {
                        show: true,
                        position: 'insideEndTop',
                        formatter: '30天前',
                        fontSize: 10,
                        color: '#ffc107',
                        backgroundColor: 'rgba(255, 255, 255, 0.9)',
                        padding: [2, 4],
                        borderColor: '#ffc107',
                        borderWidth: 1,
                        borderRadius: 2
                    },
                    lineStyle: {
                        color: '#ffc107',
                        type: 'dashed',
                        width: 1,
                        opacity: 0.8
                    },
                    data: [{
                        xAxis: thirtyDaysAgoIndex
                    }],
                    silent: true
                }
            });
        }

        // 设置图表配置并启用响应式
        chart.setOption(option);
        
        // 监听窗口大小变化
        const resizeHandler = () => {
            if (chart && !chart.isDisposed()) {
                chart.resize();
            }
        };
        window.addEventListener('resize', resizeHandler);

        // 存储清理函数
        chart._resizeHandler = resizeHandler;

        console.log(`Created chart for tenant ${tenantData.tenant_id} with ${dates.length} data points`);
    }

    /**
     * 销毁所有图表
     */
    disposeAllCharts() {
        this.chartInstances.forEach((chart, containerId) => {
            if (chart && !chart.isDisposed()) {
                // 移除窗口resize监听器
                if (chart._resizeHandler) {
                    window.removeEventListener('resize', chart._resizeHandler);
                }
                chart.dispose();
            }
        });
        this.chartInstances.clear();
        console.log('All charts disposed');
    }

    /**
     * 销毁指定图表
     */
    disposeChart(containerId) {
        if (this.chartInstances.has(containerId)) {
            const chart = this.chartInstances.get(containerId);
            if (chart && !chart.isDisposed()) {
                if (chart._resizeHandler) {
                    window.removeEventListener('resize', chart._resizeHandler);
                }
                chart.dispose();
            }
            this.chartInstances.delete(containerId);
            console.log(`Chart ${containerId} disposed`);
        }
    }

    /**
     * 重新调整所有图表大小
     */
    resizeAllCharts() {
        this.chartInstances.forEach((chart, containerId) => {
            if (chart && !chart.isDisposed()) {
                chart.resize();
            }
        });
    }

    /**
     * 获取图表实例
     */
    getChart(containerId) {
        return this.chartInstances.get(containerId);
    }

    /**
     * 获取所有图表实例
     */
    getAllCharts() {
        return Array.from(this.chartInstances.values());
    }
}
