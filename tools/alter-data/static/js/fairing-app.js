/**
 * Fairing分析主应用
 */
class FairingApp {
    constructor() {
        this.chartManager = new FairingChartManager();
        this.currentData = null;
        this.isLoading = false;
    }

    /**
     * 初始化应用
     */
    async init() {
        console.log('初始化Fairing分析应用');
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
            const url = forceRefresh ? '/api/fairing?refresh=true' : '/api/fairing';
            console.log(`Loading Fairing data from: ${url}`);
            
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
        console.log('Refreshing Fairing data...');
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
        let totalResponses = 0;

        data.forEach(tenant => {
            if (tenant.has_anomalies) {
                anomalyTenants++;
            }
            
            if (tenant.warning_level === 'critical') {
                criticalCount++;
            } else if (tenant.warning_level === 'warning') {
                warningCount++;
            }
            
            totalResponses += tenant.total_responses;
        });

        // 更新统计值
        document.getElementById('total-tenants').textContent = totalTenants.toLocaleString();
        document.getElementById('anomaly-tenants').textContent = anomalyTenants.toLocaleString();
        document.getElementById('critical-count').textContent = criticalCount.toLocaleString();
        document.getElementById('warning-count').textContent = warningCount.toLocaleString();
        document.getElementById('total-responses').textContent = totalResponses.toLocaleString();

        // 显示统计摘要
        document.getElementById('stats-summary').style.display = 'flex';
        
        console.log(`Statistics: ${totalTenants} tenants, ${anomalyTenants} with anomalies, ${totalResponses} total responses`);
    }

    /**
     * 更新缓存信息
     */
    updateCacheInfo(cacheInfo) {
        if (cacheInfo) {
            const cacheTime = new Date(cacheInfo.updated_at).toLocaleString('zh-CN');
            document.getElementById('cache-time').textContent = cacheTime;
            
            // 缓存状态标识
            const cacheBadge = document.getElementById('cache-badge');
            if (cacheInfo.is_expired) {
                cacheBadge.textContent = '已过期';
                cacheBadge.className = 'cache-badge expired';
            } else {
                cacheBadge.textContent = '永久缓存';
                cacheBadge.className = 'cache-badge active';
            }
            
            console.log(`Cache info: updated at ${cacheTime}, ${cacheInfo.data_count} items`);
        }
    }

    /**
     * 更新数据截止日期
     */
    updateDataEndDate(data) {
        if (data && data.length > 0) {
            // 找到最新的数据日期
            let latestDate = null;
            data.forEach(tenant => {
                if (tenant.date_range && tenant.date_range.length > 0) {
                    const tenantLatestDate = tenant.date_range[tenant.date_range.length - 1];
                    if (!latestDate || tenantLatestDate > latestDate) {
                        latestDate = tenantLatestDate;
                    }
                }
            });
            
            if (latestDate) {
                document.getElementById('data-end-date').textContent = latestDate;
            }
        }
    }

    /**
     * 渲染图表
     */
    renderCharts(data) {
        const container = document.getElementById('charts-container');
        
        if (!data || data.length === 0) {
            this.showNoData(true);
            return;
        }

        this.showNoData(false);
        
        // 清空现有图表
        container.innerHTML = '';
        
        // 为每个租户创建图表
        data.forEach((tenantData, index) => {
            const chartCard = this.createChartCard(tenantData, index);
            container.appendChild(chartCard);
        });

        console.log(`Rendered ${data.length} charts`);
    }

    /**
     * 创建图表卡片
     */
    createChartCard(tenantData, index) {
        const card = document.createElement('div');
        card.className = 'chart-card';
        card.style.animationDelay = `${index * 0.1}s`;

        // 预警级别样式
        let warningClass = 'normal';
        let warningIcon = '🟢';
        if (tenantData.warning_level === 'critical') {
            warningClass = 'critical';
            warningIcon = '🔴';
        } else if (tenantData.warning_level === 'warning') {
            warningClass = 'warning';
            warningIcon = '🟡';
        }

        card.innerHTML = `
            <div class="chart-header">
                <div class="chart-title">
                    📋 ${tenantData.tenant_name}
                    <span class="chart-subtitle">(${tenantData.date_range ? tenantData.date_range.length : 0}天数据)</span>
                </div>
                <div class="chart-info">
                    <span class="info-badge ${warningClass}">
                        ${warningIcon} ${tenantData.warning_level.toUpperCase()}
                    </span>
                    ${tenantData.zero_days_count > 0 ? `<span class="info-badge warning">🚨 掉0: ${tenantData.zero_days_count}天</span>` : ''}
                    ${tenantData.concave_count > 0 ? `<span class="info-badge critical">🔶 凹形: ${tenantData.concave_count}个</span>` : ''}
                </div>
            </div>
            <div class="chart-content" id="chart-${tenantData.tenant_id}"></div>
            <div class="stats-grid">
                <div class="stat-item-small">
                    <span class="stat-label">总响应数</span>
                    <span class="stat-value">${tenantData.total_responses.toLocaleString()}</span>
                </div>
                <div class="stat-item-small">
                    <span class="stat-label">日均响应</span>
                    <span class="stat-value">${Math.round(tenantData.daily_average)}</span>
                </div>
                <div class="stat-item-small">
                    <span class="stat-label">数据天数</span>
                    <span class="stat-value">${tenantData.date_range ? tenantData.date_range.length : 0}</span>
                </div>
                <div class="stat-item-small">
                    <span class="stat-label">异常天数</span>
                    <span class="stat-value">${tenantData.zero_days_count + tenantData.concave_count}</span>
                </div>
            </div>
        `;

        // 创建图表
        setTimeout(() => {
            this.chartManager.createTenantChart(`chart-${tenantData.tenant_id}`, tenantData);
        }, 100 + index * 50);

        return card;
    }

    /**
     * 显示/隐藏加载状态
     */
    showLoading(show) {
        const loadingIndicator = document.getElementById('loading-indicator');
        const refreshButton = document.getElementById('refresh-button');
        
        loadingIndicator.style.display = show ? 'flex' : 'none';
        refreshButton.disabled = show;
        
        if (show) {
            refreshButton.classList.add('loading');
        } else {
            refreshButton.classList.remove('loading');
        }
    }

    /**
     * 显示/隐藏无数据消息
     */
    showNoData(show) {
        document.getElementById('no-data-message').style.display = show ? 'block' : 'none';
        document.getElementById('stats-summary').style.display = show ? 'none' : 'flex';
    }

    /**
     * 显示错误信息
     */
    showError(message) {
        console.error('Error:', message);
        
        const container = document.getElementById('charts-container');
        container.innerHTML = `
            <div class="error-message">
                <h3>❌ 加载失败</h3>
                <p>${message}</p>
                <button onclick="fairingApp.loadData()" class="retry-button">重试</button>
            </div>
        `;
        
        // 隐藏统计摘要
        document.getElementById('stats-summary').style.display = 'none';
    }

    /**
     * 获取当前数据
     */
    getCurrentData() {
        return this.currentData;
    }

    /**
     * 搜索租户
     */
    searchTenant(tenantId) {
        if (!this.currentData) return null;
        return this.currentData.find(tenant => tenant.tenant_id == tenantId);
    }
}
