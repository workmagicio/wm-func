// ECharts 图表管理类
class ChartManager {
    constructor() {
        this.charts = new Map(); // 存储图表实例
        this.chartsContainer = document.getElementById('charts-grid');
        this.resizeListenerAdded = false;
        
        // 只添加一次resize监听器
        this.addResizeListener();
    }

    // 添加窗口resize监听器
    addResizeListener() {
        if (!this.resizeListenerAdded) {
            window.addEventListener('resize', () => {
                this.resizeAllCharts();
            });
            this.resizeListenerAdded = true;
            console.log('📐 窗口resize监听器已添加');
        }
    }

    // 初始化图表容器
    initChart(tenantData, retryCount = 0) {
        const maxRetries = 5; // 最大重试次数
        
        // 生成唯一的图表ID，避免同一租户不同平台的ID冲突
        const chartId = tenantData.chart_id || `chart-${tenantData.tenant_id}-${tenantData.platform || 'default'}`;
        
        console.log(`🎨 创建图表容器: ${chartId}`, {
            tenant_id: tenantData.tenant_id,
            platform: tenantData.platform,
            tenant_name: tenantData.tenant_name
        });
        
        // 检查是否已存在相同ID的图表
        if (document.getElementById(chartId)) {
            console.warn(`⚠️ 图表ID ${chartId} 已存在，跳过创建`);
            return;
        }
        
        // 创建图表容器
        const chartItem = document.createElement('div');
        chartItem.className = 'chart-item';
        chartItem.innerHTML = `
            <div class="chart-header">
                <h3 class="chart-title">${tenantData.tenant_name}</h3>
                <p class="chart-subtitle">租户ID: ${tenantData.tenant_id} | 平台: ${tenantData.platform}</p>
            </div>
            <div class="chart-body">
                <div id="${chartId}" class="chart-canvas"></div>
            </div>
        `;
        
        this.chartsContainer.appendChild(chartItem);
        
        // 延迟初始化图表，确保DOM和CSS完全就绪
        setTimeout(() => {
            const chartDom = document.getElementById(chartId);
            const containerRect = chartDom ? chartDom.getBoundingClientRect() : null;
            
            console.log(`🔍 图表容器 ${chartId} 检查:`, {
                exists: !!chartDom,
                offsetWidth: chartDom?.offsetWidth,
                offsetHeight: chartDom?.offsetHeight,
                rectWidth: containerRect?.width,
                rectHeight: containerRect?.height,
                retryCount: retryCount
            });
            
            if (chartDom && (chartDom.offsetWidth > 0 || containerRect?.width > 0)) {
                const chart = echarts.init(chartDom);
                
                // 设置图表配置
                const option = this.createChartOption(tenantData);
                chart.setOption(option);
                
                // 强制resize确保尺寸正确
                setTimeout(() => {
                    chart.resize();
                }, 50);
                
                // 保存图表实例
                this.charts.set(tenantData.tenant_id, chart);
                
                console.log(`✅ 图表 ${tenantData.tenant_id} 初始化完成 (重试${retryCount}次)`);
            } else if (retryCount < maxRetries) {
                console.warn(`⚠️ 图表容器 ${chartId} 尺寸异常，重试中... (${retryCount + 1}/${maxRetries})`);
                // 如果容器还没准备好，再次重试
                setTimeout(() => {
                    this.initChart(tenantData, retryCount + 1);
                }, 200 * (retryCount + 1)); // 递增延迟
            } else {
                console.error(`❌ 图表 ${chartId} 初始化失败，已达最大重试次数`);
                // 显示错误信息
                if (chartDom) {
                    chartDom.innerHTML = `
                        <div style="display: flex; align-items: center; justify-content: center; height: 300px; color: #e74c3c;">
                            <div style="text-align: center;">
                                <div style="font-size: 48px; margin-bottom: 16px;">⚠️</div>
                                <div>图表加载失败</div>
                                <div style="font-size: 12px; margin-top: 8px; color: #95a5a6;">容器尺寸异常</div>
                            </div>
                        </div>
                    `;
                }
            }
        }, 50 + (retryCount * 50)); // 根据重试次数调整延迟
        
        return chartItem;
    }

    // 创建图表配置
    createChartOption(tenantData) {
        // 先按日期排序数据
        const sortedData = this.sortDataByDate(tenantData);
        const { date_range, api_spend, ad_spend, difference } = sortedData;
        
        // 计算30天前的日期
        const thirtyDaysAgo = this.getDateNDaysAgo(30);
        
        return {
            title: {
                text: '数据对比趋势',
                left: 'center',
                textStyle: {
                    fontSize: 16,
                    fontWeight: 'bold',
                    color: '#2c3e50'
                }
            },
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'cross',
                    label: {
                        backgroundColor: '#6a7985'
                    }
                },
                formatter: function(params) {
                    let result = `<strong>${params[0].name}</strong><br/>`;
                    params.forEach((item, index) => {
                        const value = item.value;
                        const formattedValue = typeof value === 'number' ? 
                            value.toLocaleString() : value;
                        result += `${item.marker} ${item.seriesName}: $${formattedValue}<br/>`;
                    });
                    return result;
                }
            },
            legend: {
                data: ['API 消费', 'Overview spend', '差异值'],
                bottom: 10,
                textStyle: {
                    fontSize: 12
                },
                selected: {
                    'API 消费': true,
                    'Overview spend': true,
                    '差异值': false  // 差异值默认隐藏
                }
            },
            grid: {
                left: '5%',
                right: '5%',
                bottom: '25%',
                top: '15%',
                containLabel: true
            },
            xAxis: {
                type: 'category',
                data: date_range,
                axisLabel: {
                    rotate: 45,
                    fontSize: 11,
                    margin: 15,
                    color: '#666',
                    overflow: 'none',
                    showMaxLabel: true  // 只确保最后一天显示
                },
                axisTick: {
                    alignWithLabel: true
                },
                boundaryGap: false
            },
            yAxis: [
                {
                    type: 'value',
                    name: '消费金额 ($)',
                    position: 'left',
                    axisLabel: {
                        formatter: function(value) {
                            if (value >= 1000) {
                                return (value / 1000).toFixed(1) + 'K';
                            }
                            return value;
                        }
                    }
                },
                {
                    type: 'value',
                    name: '差异值 ($)',
                    position: 'right',
                    axisLabel: {
                        formatter: function(value) {
                            return value >= 0 ? '+' + value : value;
                        }
                    }
                }
            ],
            series: [
                {
                    name: 'API 消费',
                    type: 'line',
                    data: api_spend,
                    smooth: true,
                    symbol: 'circle',
                    symbolSize: 6,
                    lineStyle: {
                        width: 3,
                        color: '#3498db'
                    },
                    itemStyle: {
                        color: '#3498db'
                    },
                    areaStyle: {
                        color: {
                            type: 'linear',
                            x: 0, y: 0, x2: 0, y2: 1,
                            colorStops: [{
                                offset: 0, color: 'rgba(52, 152, 219, 0.3)'
                            }, {
                                offset: 1, color: 'rgba(52, 152, 219, 0.1)'
                            }]
                        }
                    },
                    markLine: {
                        silent: true,
                        lineStyle: {
                            color: '#f39c12',
                            width: 2,
                            type: 'dashed'
                        },
                        label: {
                            show: true,
                            position: 'insideEndTop',
                            formatter: '30天前',
                            color: '#f39c12',
                            fontWeight: 'bold',
                            fontSize: 12
                        },
                        data: [{
                            xAxis: thirtyDaysAgo
                        }]
                    }
                },
                {
                    name: 'Overview spend',
                    type: 'line',
                    data: ad_spend,
                    smooth: true,
                    symbol: 'circle',
                    symbolSize: 6,
                    lineStyle: {
                        width: 3,
                        color: '#e74c3c'
                    },
                    itemStyle: {
                        color: '#e74c3c'
                    },
                    areaStyle: {
                        color: {
                            type: 'linear',
                            x: 0, y: 0, x2: 0, y2: 1,
                            colorStops: [{
                                offset: 0, color: 'rgba(231, 76, 60, 0.3)'
                            }, {
                                offset: 1, color: 'rgba(231, 76, 60, 0.1)'
                            }]
                        }
                    }
                },
                {
                    name: '差异值',
                    type: 'bar',
                    data: difference,
                    yAxisIndex: 1,
                    itemStyle: {
                        color: function(params) {
                            return params.value >= 0 ? '#f39c12' : '#e67e22';
                        }
                    },
                    barWidth: '40%',
                    opacity: 0.7
                }
            ],
            animation: true,
            animationDuration: 1000,
            animationEasing: 'cubicOut'
        };
    }

    // 更新图表数据
    updateChartData(tenantData) {
        const chart = this.charts.get(tenantData.tenant_id);
        if (chart) {
            const option = this.createChartOption(tenantData);
            chart.setOption(option, true); // true 表示替换合并
        }
    }

    // 显示图表加载状态
    showChartLoading(tenantId) {
        const chart = this.charts.get(tenantId);
        if (chart) {
            chart.showLoading('default', {
                text: '数据加载中...',
                color: '#3498db',
                textColor: '#000',
                maskColor: 'rgba(255, 255, 255, 0.8)',
                zlevel: 0
            });
        }
    }

    // 隐藏图表加载状态
    hideChartLoading(tenantId) {
        const chart = this.charts.get(tenantId);
        if (chart) {
            chart.hideLoading();
        }
    }

    // 销毁所有图表
    destroyCharts() {
        console.log('🧹 开始清除所有图表...');
        
        // 销毁ECharts实例
        this.charts.forEach((chart, chartId) => {
            console.log(`  🗑️ 销毁图表: ${chartId}`);
            if (chart && typeof chart.dispose === 'function') {
                chart.dispose();
            }
        });
        this.charts.clear();
        
        // 强制清空容器
        if (this.chartsContainer) {
            this.chartsContainer.innerHTML = '';
        }
        
        // 等待DOM更新完成
        setTimeout(() => {
            console.log('✅ 图表清除完成，当前容器内容:', this.chartsContainer?.innerHTML || 'empty');
        }, 100);
    }

    // 重新渲染所有图表（窗口大小变化时使用）
    resizeCharts() {
        this.charts.forEach((chart) => {
            chart.resize();
        });
    }

    // 获取图表统计信息
    getChartStats() {
        return {
            totalCharts: this.charts.size,
            chartIds: Array.from(this.charts.keys())
        };
    }

    // 按日期排序租户数据
    sortDataByDate(tenantData) {
        const { date_range, api_spend, ad_spend, difference } = tenantData;
        
        // 创建包含所有数据的数组，用于一起排序
        const dataArray = date_range.map((date, index) => ({
            date: date,
            apiSpend: api_spend[index],
            adSpend: ad_spend[index],
            difference: difference[index]
        }));
        
        // 按日期排序
        dataArray.sort((a, b) => {
            const dateA = new Date(a.date);
            const dateB = new Date(b.date);
            return dateA - dateB; // 升序排序（最早的日期在前）
        });
        
        // 重新组装数据
        return {
            date_range: dataArray.map(item => item.date),
            api_spend: dataArray.map(item => item.apiSpend),
            ad_spend: dataArray.map(item => item.adSpend),
            difference: dataArray.map(item => item.difference)
        };
    }

    // 获取N天前的日期
    getDateNDaysAgo(days) {
        const date = new Date();
        date.setDate(date.getDate() - days);
        
        // 格式化为 YYYY-MM-DD 格式
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        
        return `${year}-${month}-${day}`;
    }
}
