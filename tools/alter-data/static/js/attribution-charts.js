/**
 * 归因订单图表管理器
 * 负责处理智能非线性纵坐标和图表渲染
 */
class AttributionChartManager {
    constructor() {
        this.charts = new Map();
        this.chartMode = 'auto'; // auto, linear, log, segmented
    }

    /**
     * 设置图表模式
     */
    setChartMode(mode) {
        this.chartMode = mode;
        console.log(`图表模式切换至: ${mode}`);
    }

    /**
     * 获取Y轴配置（固定使用分段坐标）
     */
    getOptimalYAxisConfig(allPlatformData) {
        // 收集所有数值（包括负值-100）
        const allValues = [];
        let hasNegativeMarkers = false;
        
        Object.values(allPlatformData).forEach(series => {
            series.forEach(value => {
                if (value === -100) {
                    hasNegativeMarkers = true;
                } else if (value > 0) {
                    allValues.push(value);
                }
            });
        });

        const maxVal = allValues.length > 0 ? Math.max(...allValues) : 100;
        const minVal = allValues.length > 0 ? Math.min(...allValues) : 0;

        console.log(`数据分析: 最大值=${maxVal}, 最小值=${minVal}, 包含掉0标记=${hasNegativeMarkers}, 使用分段坐标`);

        // 固定使用分段坐标
        return this.getSegmentedYAxisConfigWithNegative(minVal, maxVal, hasNegativeMarkers);
    }

    /**
     * 自动选择最优配置（支持负值标记）
     */
    getAutoYAxisConfigWithNegative(ratio, minVal, maxVal, hasNegativeMarkers) {
        if (ratio > 100) {
            console.log('选择对数坐标 (ratio > 100)');
            return this.getLogYAxisConfigWithNegative(minVal, maxVal, hasNegativeMarkers);
        } else if (ratio > 10) {
            console.log('选择分段坐标 (ratio > 10)');
            return this.getSegmentedYAxisConfigWithNegative(minVal, maxVal, hasNegativeMarkers);
        } else {
            console.log('选择线性坐标 (ratio <= 10)');
            return this.getLinearYAxisConfigWithNegative(hasNegativeMarkers);
        }
    }

    /**
     * 线性坐标配置（支持负值标记）
     */
    getLinearYAxisConfigWithNegative(hasNegativeMarkers) {
        return {
            type: 'value',
            scale: false,
            min: hasNegativeMarkers ? -120 : 0, // 如果有-100标记，设置最小值为-120
            axisLabel: {
                formatter: (value) => this.formatNumberWithNegative(value)
            },
            splitLine: {
                lineStyle: {
                    type: 'dashed',
                    opacity: 0.5
                }
            },
            mode: 'linear'
        };
    }

    /**
     * 对数坐标配置（支持负值标记）
     */
    getLogYAxisConfigWithNegative(minVal, maxVal, hasNegativeMarkers) {
        // 对数坐标不能直接处理负值，改用线性坐标
        if (hasNegativeMarkers) {
            console.log('检测到负值标记，对数模式降级为线性模式');
            return this.getLinearYAxisConfigWithNegative(hasNegativeMarkers);
        }
        
        const logMin = Math.max(1, minVal * 0.5);
        return {
            type: 'log',
            logBase: 10,
            min: logMin,
            max: maxVal * 1.2,
            axisLabel: {
                formatter: this.formatLogValue
            },
            splitLine: {
                lineStyle: {
                    type: 'dashed',
                    opacity: 0.3
                }
            },
            mode: 'log'
        };
    }

    /**
     * 分段坐标配置（支持负值标记）
     */
    getSegmentedYAxisConfigWithNegative(minVal, maxVal, hasNegativeMarkers) {
        return {
            type: 'value',
            scale: true,
            min: hasNegativeMarkers ? -120 : 0,
            max: maxVal * 1.1,
            splitNumber: 8,
            axisLabel: {
                formatter: (value) => this.formatNumberWithNegative(value)
            },
            splitLine: {
                lineStyle: {
                    type: 'dashed',
                    opacity: 0.4
                }
            },
            mode: 'segmented'
        };
    }

    /**
     * 格式化大数字显示（整数）
     */
    formatLargeNumber(value) {
        if (value === 0) return '0';
        if (value < 1000) return Math.round(value).toString();
        if (value < 1000000) return Math.round(value / 1000) + 'K';
        if (value < 1000000000) return Math.round(value / 1000000) + 'M';
        return Math.round(value / 1000000000) + 'B';
    }

    /**
     * 格式化包含负值标记的数字显示
     */
    formatNumberWithNegative(value) {
        if (value === -200) return '数据缺失';
        if (value === -100) return '掉0';
        if (value === 0) return '0';
        return this.formatLargeNumber(value);
    }

    /**
     * 格式化对数值显示（整数）
     */
    formatLogValue(value) {
        if (value === 0) return '0';
        if (value < 1000) return Math.round(value).toString();
        if (value < 1000000) return Math.round(value / 1000) + 'K';
        if (value < 1000000000) return Math.round(value / 1000000) + 'M';
        return Math.round(value / 1000000000) + 'B';
    }

    /**
     * 获取图表模式显示名称
     */
    getChartModeDisplayName(mode) {
        return '分段坐标'; // 固定返回分段坐标
    }

    /**
     * 创建租户图表
     */
    createTenantChart(containerElement, tenantData) {
        const chart = echarts.init(containerElement);
        this.charts.set(tenantData.tenant_id, chart);

        const yAxisConfig = this.getOptimalYAxisConfig(tenantData.platform_data);
        const option = this.buildChartOption(tenantData, yAxisConfig);
        
        chart.setOption(option);

        // 响应式处理
        const resizeHandler = () => chart.resize();
        window.addEventListener('resize', resizeHandler);

        return chart;
    }

    /**
     * 构建图表配置选项
     */
    buildChartOption(tenantData, yAxisConfig) {
        const { tenant_id, tenant_name, date_range, platform_data, platforms } = tenantData;

        // 准备数据系列
        const series = [];
        const colors = ['#007bff', '#28a745', '#ffc107', '#dc3545', '#6f42c1', '#fd7e14', '#20c997'];
        
        platforms.forEach((platform, index) => {
            const data = platform_data[platform] || [];
            series.push({
                name: platform,
                type: 'line',
                data: data,
                smooth: true,
                symbol: 'circle',
                symbolSize: 6,
                lineStyle: {
                    width: 3,
                    color: colors[index % colors.length]
                },
                itemStyle: {
                    color: colors[index % colors.length]
                },
                emphasis: {
                    focus: 'series'
                },
                // 特殊处理异常标记的显示
                markPoint: {
                    data: data.map((value, dataIndex) => {
                        if (value === -200) {
                            return {
                                coord: [dataIndex, value],
                                symbol: 'diamond',
                                symbolSize: 50,
                                itemStyle: {
                                    color: '#e74c3c'
                                },
                                label: {
                                    show: true,
                                    formatter: '凹形',
                                    color: '#fff',
                                    fontSize: 11,
                                    fontWeight: 'bold'
                                }
                            };
                        } else if (value === -100) {
                            return {
                                coord: [dataIndex, value],
                                symbol: 'pin',
                                symbolSize: 40,
                                itemStyle: {
                                    color: '#ff4757'
                                },
                                label: {
                                    show: true,
                                    formatter: '掉0',
                                    color: '#fff',
                                    fontSize: 10
                                }
                            };
                        }
                        return null;
                    }).filter(item => item !== null)
                }
            });
        });

        return {
            title: {
                text: `${tenant_name} 归因订单趋势`,
                textStyle: {
                    fontSize: 16,
                    fontWeight: 'bold'
                },
                left: 'center'
            },
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'cross',
                    crossStyle: {
                        color: '#999'
                    }
                },
                formatter: function(params) {
                    let tooltip = `<strong>${params[0].axisValue}</strong><br/>`;
                    params.forEach(param => {
                        const value = param.value;
                        let formattedValue;
                        if (value === -200) {
                            formattedValue = '<span style="color: #e74c3c; font-weight: bold;">🔶 凹字形异常</span>';
                        } else if (value === -100) {
                            formattedValue = '<span style="color: #ff4757; font-weight: bold;">🚨 掉0现象</span>';
                        } else if (value && value > 0) {
                            formattedValue = value.toLocaleString();
                        } else {
                            formattedValue = '0';
                        }
                        tooltip += `<span style="color:${param.color}">●</span> ${param.seriesName}: ${formattedValue}<br/>`;
                    });
                    return tooltip;
                }
            },
            legend: {
                data: platforms,
                bottom: 10,
                type: 'scroll',
                pageIconColor: '#007bff',
                pageIconInactiveColor: '#ccc'
            },
            grid: {
                left: '3%',
                right: '4%',
                bottom: '15%',
                top: '15%',
                containLabel: true
            },
            xAxis: {
                type: 'category',
                data: date_range,
                axisLine: {
                    onZero: false
                },
                axisLabel: {
                    formatter: function(value) {
                        // 只显示月-日格式，例如：07-20
                        const dateParts = value.split('T')[0].split('-'); // 去掉时间部分，只保留日期
                        return dateParts.slice(1).join('-'); // 显示月-日
                    },
                    rotate: 45
                }
            },
            yAxis: {
                ...yAxisConfig,
                name: '订单数',
                nameLocation: 'middle',
                nameGap: 50,
                nameTextStyle: {
                    fontWeight: 'bold'
                }
            },
            series: series,
            dataZoom: [
                {
                    type: 'inside',
                    start: 0,
                    end: 100
                },
                {
                    show: true,
                    type: 'slider',
                    top: '90%',
                    start: 0,
                    end: 100
                }
            ]
        };
    }

    /**
     * 更新所有图表的模式
     */
    updateAllChartsMode() {
        console.log(`更新所有图表模式: ${this.chartMode}`);
        // 触发重新渲染，具体实现由调用方处理
    }

    /**
     * 销毁指定图表
     */
    destroyChart(tenantId) {
        const chart = this.charts.get(tenantId);
        if (chart) {
            chart.dispose();
            this.charts.delete(tenantId);
        }
    }

    /**
     * 销毁所有图表
     */
    destroyAllCharts() {
        this.charts.forEach(chart => chart.dispose());
        this.charts.clear();
    }

    /**
     * 计算汇总统计信息
     */
    calculateSummaryStats(allTenantsData) {
        if (!allTenantsData || allTenantsData.length === 0) {
            return {
                totalTenants: 0,
                totalOrders: 0,
                activePlatforms: 0
            };
        }

        const platformSet = new Set();
        let totalOrders = 0;

        allTenantsData.forEach(tenant => {
            // 统计平台
            tenant.platforms.forEach(platform => platformSet.add(platform));
            
            // 统计总订单数
            Object.values(tenant.total_orders || {}).forEach(orders => {
                totalOrders += orders;
            });
        });

        return {
            totalTenants: allTenantsData.length,
            totalOrders: totalOrders,
            activePlatforms: platformSet.size
        };
    }
}
