/**
 * Amazon订单分析主应用
 */
class AmazonOrdersApp {
    constructor() {
        this.chartManager = new AmazonOrdersChartManager();
        this.currentData = null;
        this.isLoading = false;
    }

    /**
     * 初始化应用
     */
    async init() {
        console.log('初始化Amazon订单分析应用');
        await this.loadData();
    }

    /**
     * 加载数据
     */
    async loadData(forceRefresh = false) {
        if (this.isLoading) return;
        
        this.showLoading(true);
        this.isLoading = true;

        try {
            const url = forceRefresh ? '/api/amazon-orders?refresh=true' : '/api/amazon-orders';
            console.log(`Loading Amazon orders data from: ${url}`);
            
            const response = await fetch(url);
            const result = await response.json();

            if (result.success) {
                this.currentData = result.data;
                this.updateUI(result);
                console.log(`Loaded ${result.data.length} tenant data`);
            } else {
                this.showError(result.message);
            }
        } catch (error) {
            console.error('Error loading data:', error);
            this.showError('数据加载失败: ' + error.message);
        } finally {
            this.showLoading(false);
            this.isLoading = false;
        }
    }

    /**
     * 刷新数据
     */
    async refreshData() {
        console.log('Refreshing Amazon orders data...');
        await this.loadData(true);
    }

    /**
     * 更新UI
     */
    updateUI(result) {
        // 更新统计摘要
        this.updateStatsSummary(result.data);
        
        // 更新缓存信息
        this.updateCacheInfo(result.cache_info);
        
        // 更新数据截止日期
        this.updateDataEndDate(result.data);
        
        // 渲染图表
        this.renderCharts(result.data);
    }

    /**
     * 更新统计摘要
     */
    updateStatsSummary(data) {
        const totalTenants = data.length;
        let anomalyTenants = 0;
        let criticalCount = 0;
        let warningCount = 0;
        let totalOrders = 0;

        data.forEach(tenant => {
            if (tenant.has_anomalies) {
                anomalyTenants++;
            }
            
            if (tenant.warning_level === 'critical') {
                criticalCount++;
            } else if (tenant.warning_level === 'warning') {
                warningCount++;
            }
            
            totalOrders += tenant.total_orders;
        });

        // 更新统计值
        document.getElementById('total-tenants').textContent = totalTenants.toLocaleString();
        document.getElementById('anomaly-tenants').textContent = anomalyTenants.toLocaleString();
        document.getElementById('critical-count').textContent = criticalCount.toLocaleString();
        document.getElementById('warning-count').textContent = warningCount.toLocaleString();
        document.getElementById('total-orders').textContent = totalOrders.toLocaleString();

        // 显示统计摘要
        document.getElementById('stats-summary').style.display = 'flex';
    }

    /**
     * 更新缓存信息
     */
    updateCacheInfo(cacheInfo) {
        if (cacheInfo) {
            const cacheTime = new Date(cacheInfo.updated_at);
            const formattedTime = this.formatDateTime(cacheTime);
            
            document.getElementById('cache-time').textContent = formattedTime;
            document.getElementById('cache-info').style.display = 'flex';
        }
    }

    /**
     * 更新数据截止日期
     */
    updateDataEndDate(data) {
        if (data && data.length > 0) {
            // 找到所有租户的最后日期中最新的一个
            let latestEndDate = '';
            data.forEach(tenant => {
                if (tenant.date_range && tenant.date_range.length > 0) {
                    const endDate = tenant.date_range[tenant.date_range.length - 1];
                    if (endDate > latestEndDate) {
                        latestEndDate = endDate;
                    }
                }
            });
            
            if (latestEndDate) {
                // 格式化日期显示（去掉时区信息）
                let displayDate = latestEndDate;
                if (typeof latestEndDate === 'string') {
                    displayDate = latestEndDate.split('T')[0].split(' ')[0];
                    if (displayDate.includes('+')) {
                        displayDate = displayDate.split('+')[0];
                    }
                }
                document.getElementById('data-end-date').textContent = displayDate;
            }
        }
    }

    /**
     * 渲染图表
     */
    renderCharts(data) {
        const container = document.getElementById('charts-container');
        
        if (!data || data.length === 0) {
            document.getElementById('no-data-message').style.display = 'block';
            return;
        }

        // 清空容器
        container.innerHTML = '';

        // 按优先级排序：凹形异常 > 掉0天数 > 总异常情况 > 订单量
        const sortedData = [...data].sort((a, b) => {
            // 第一优先级：凹形异常数量多的排在前面
            if (a.concave_count !== b.concave_count) {
                return b.concave_count - a.concave_count;
            }
            
            // 第二优先级：掉0天数多的排在前面
            if (a.zero_days_count !== b.zero_days_count) {
                return b.zero_days_count - a.zero_days_count;
            }
            
            // 第三优先级：有异常的排在前面
            if (a.has_anomalies !== b.has_anomalies) {
                return b.has_anomalies - a.has_anomalies;
            }
            
            // 最后按总订单量降序排序
            return b.total_orders - a.total_orders;
        });

        // 双列布局渲染图表
        for (let i = 0; i < sortedData.length; i += 2) {
            const row = document.createElement('div');
            row.className = 'charts-row';

            // 左侧图表
            const leftTenant = sortedData[i];
            const leftChart = this.createTenantChartElement(leftTenant, i);
            row.appendChild(leftChart);

            // 右侧图表（如果存在）
            if (i + 1 < sortedData.length) {
                const rightTenant = sortedData[i + 1];
                const rightChart = this.createTenantChartElement(rightTenant, i + 1);
                row.appendChild(rightChart);
            }

            container.appendChild(row);
        }

        // 初始化所有图表
        setTimeout(() => {
            sortedData.forEach((tenant, index) => {
                this.chartManager.renderTenantChart(tenant, `chart-${index}`);
            });
        }, 100);
    }

    /**
     * 创建租户图表元素
     */
    createTenantChartElement(tenant, index) {
        const chartContainer = document.createElement('div');
        chartContainer.className = `tenant-chart-card ${tenant.warning_level}`;
        
        // 获取预警级别的显示文本和样式
        const warningInfo = this.getWarningInfo(tenant.warning_level);
        
        chartContainer.innerHTML = `
            <div class="chart-tenant-title">
                <h4>${tenant.tenant_name}</h4>
                <span class="tenant-warning-badge ${tenant.warning_level}">
                    ${warningInfo.text}
                </span>
            </div>
            <div class="chart-tenant-stats">
                <div class="tenant-stat-item">
                    <span class="tenant-stat-label">总订单:</span>
                    <span class="tenant-stat-value">${tenant.total_orders.toLocaleString()}</span>
                </div>
                <div class="tenant-stat-item">
                    <span class="tenant-stat-label">日均:</span>
                    <span class="tenant-stat-value">${tenant.daily_average.toFixed(1)}</span>
                </div>

                ${tenant.zero_days_count > 0 ? `
                <div class="tenant-stat-item">
                    <span class="tenant-stat-label">掉0天数:</span>
                    <span class="tenant-stat-value highlight-critical">${tenant.zero_days_count}</span>
                    <span class="recent-zero-dates">${this.getRecentZeroDates(tenant)}</span>
                </div>
                ` : ''}
                ${tenant.concave_count > 0 ? `
                <div class="tenant-stat-item">
                    <span class="tenant-stat-label">凹形异常:</span>
                    <span class="tenant-stat-value highlight-critical">${tenant.concave_count}</span>
                </div>
                ` : ''}
            </div>
            <div id="chart-${index}" style="height: 280px; width: 100%;"></div>
        `;

        return chartContainer;
    }

    /**
     * 获取预警信息
     */
    getWarningInfo(warningLevel) {
        const warningMap = {
            'critical': { text: 'Critical', icon: '🔴' },
            'warning': { text: 'Warning', icon: '🟡' },
            'normal': { text: 'Normal', icon: '🟢' }
        };
        
        return warningMap[warningLevel] || warningMap['normal'];
    }

    /**
     * 显示加载状态
     */
    showLoading(show) {
        const loadingIndicator = document.getElementById('loading-indicator');
        const refreshButton = document.getElementById('refresh-button');
        
        loadingIndicator.style.display = show ? 'flex' : 'none';
        refreshButton.disabled = show;
        
        if (show) {
            refreshButton.querySelector('.refresh-text').textContent = '加载中...';
        } else {
            refreshButton.querySelector('.refresh-text').textContent = '刷新数据';
        }
    }

    /**
     * 显示错误信息
     */
    showError(message) {
        const container = document.getElementById('charts-container');
        container.innerHTML = `
            <div class="error-message">
                <h3>❌ 数据加载失败</h3>
                <p>${message}</p>
                <button onclick="amazonOrdersApp.loadData()" class="retry-button">重试</button>
            </div>
        `;
    }

    /**
     * 获取最近一周的掉0日期
     */
    getRecentZeroDates(tenant) {
        if (!tenant.processed_orders || !tenant.date_range || tenant.zero_days_count === 0) {
            return '';
        }

        const zeroDates = [];
        
        // 计算一周前的日期
        const now = new Date();
        const oneWeekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
        
        // 找出所有掉0的日期（processed_orders中值为-100的）
        for (let i = 0; i < tenant.processed_orders.length; i++) {
            if (tenant.processed_orders[i] === -100) {
                let date = tenant.date_range[i];
                // 格式化日期（去掉时区信息）
                if (typeof date === 'string') {
                    date = date.split('T')[0].split(' ')[0];
                    if (date.includes('+')) {
                        date = date.split('+')[0];
                    }
                }
                
                // 检查日期是否在最近一周内
                const zeroDate = new Date(date);
                if (zeroDate >= oneWeekAgo && zeroDate <= now) {
                    zeroDates.push(date);
                }
            }
        }

        // 按日期排序（最新的在前）
        zeroDates.sort().reverse();

        if (zeroDates.length > 0) {
            return `(${zeroDates.join(', ')})`;
        }
        
        return '';
    }

    /**
     * 格式化日期时间
     */
    formatDateTime(date) {
        const options = {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
        };
        return date.toLocaleString('zh-CN', options);
    }
}
