/**
 * 归因订单分析主应用
 */
class AttributionDashboard {
    constructor() {
        this.chartManager = new AttributionChartManager();
        this.currentData = null;
        this.currentDays = 100;
        this.isLoading = false;
    }

    /**
     * 初始化应用
     */
    async init() {
        console.log('初始化归因订单分析应用');
        
        this.setupEventListeners();
        await this.loadData();
    }

    /**
     * 设置事件监听器
     */
    setupEventListeners() {
        // 天数选择器
        const daysSelect = document.getElementById('days-select');
        if (daysSelect) {
            daysSelect.addEventListener('change', (e) => {
                this.currentDays = parseInt(e.target.value);
                this.loadData();
            });
        }

        // 图表模式选择器（固定分段坐标）
        const chartModeSelect = document.getElementById('chart-mode-select');
        if (chartModeSelect) {
            chartModeSelect.addEventListener('change', (e) => {
                const mode = e.target.value;
                this.chartManager.setChartMode(mode);
                this.updateCurrentChartModeDisplay(mode);
                this.renderCharts(); // 重新渲染图表
            });
            // 固定设置为分段坐标
            this.chartManager.setChartMode('segmented');
            this.updateCurrentChartModeDisplay('segmented');
        }

        // 刷新按钮
        const refreshButton = document.getElementById('refresh-button');
        if (refreshButton) {
            refreshButton.addEventListener('click', () => {
                this.refreshData();
            });
        }
    }

    /**
     * 加载数据
     */
    async loadData(forceRefresh = false) {
        if (this.isLoading) return;

        this.isLoading = true;
        this.showLoading(true);

        try {
            const url = `/api/attribution-orders?days=${this.currentDays}${forceRefresh ? '&refresh=true' : ''}`;
            console.log(`请求归因订单数据: ${url}`);

            const response = await fetch(url);
            const result = await response.json();

            if (result.success) {
                this.currentData = result.data;
                this.updateCacheInfo(result.cache_info);
                this.updateStats();
                this.renderCharts();
                console.log(`成功加载 ${this.currentData.length} 个租户的数据`);
            } else {
                throw new Error(result.message || '加载数据失败');
            }
        } catch (error) {
            console.error('加载归因订单数据失败:', error);
            this.showError(`加载数据失败: ${error.message}`);
        } finally {
            this.isLoading = false;
            this.showLoading(false);
        }
    }

    /**
     * 强制刷新数据
     */
    async refreshData() {
        await this.loadData(true);
    }

    /**
     * 更新统计信息
     */
    updateStats() {
        const stats = this.chartManager.calculateSummaryStats(this.currentData);
        
        this.updateElement('total-tenants', stats.totalTenants.toLocaleString());
        this.updateElement('total-orders', stats.totalOrders.toLocaleString());
        this.updateElement('active-platforms', stats.activePlatforms);

        // 显示统计区域
        const statsElement = document.getElementById('stats-summary');
        if (statsElement) {
            statsElement.style.display = 'flex';
        }
    }

    /**
     * 渲染图表
     */
    renderCharts() {
        if (!this.currentData || this.currentData.length === 0) {
            this.showNoData();
            return;
        }

        // 清理现有图表
        this.chartManager.destroyAllCharts();

        // 获取图表容器
        const container = document.getElementById('charts-container');
        if (!container) return;

        // 清空容器
        container.innerHTML = '';

        // 为每个租户创建图表
        this.currentData.forEach(tenantData => {
            const chartDiv = this.createChartElement(tenantData);
            container.appendChild(chartDiv);

            // 创建图表
            const chartContainer = chartDiv.querySelector('.chart-content');
            this.chartManager.createTenantChart(chartContainer, tenantData);
        });

        // 隐藏无数据消息
        const noDataMessage = document.getElementById('no-data-message');
        if (noDataMessage) {
            noDataMessage.style.display = 'none';
        }
    }

    /**
     * 创建图表DOM元素
     */
    createChartElement(tenantData) {
        const chartDiv = document.createElement('div');
        chartDiv.className = 'tenant-chart';
        chartDiv.setAttribute('data-tenant-id', tenantData.tenant_id);

        // 计算租户总订单数
        const totalOrders = Object.values(tenantData.total_orders || {}).reduce((sum, count) => sum + count, 0);
        const platformCount = tenantData.platforms.length;
        const currentMode = this.chartManager.getChartModeDisplayName(this.chartManager.chartMode);
        
        // 检查是否有凹字形异常
        const hasConcave = tenantData.has_concave || false;
        const concaveCount = tenantData.concave_count || 0;
        
        // 构建标题和异常标记
        let titleHtml = `<h3 class="chart-title">${tenantData.tenant_name}`;
        if (hasConcave) {
            titleHtml += ` <span class="concave-badge">⚠️ 数据缺失异常(${concaveCount})</span>`;
        }
        titleHtml += `</h3>`;

        chartDiv.innerHTML = `
            <div class="chart-header">
                <div>
                    ${titleHtml}
                    <div class="chart-subtitle">租户ID: ${tenantData.tenant_id}</div>
                </div>
                <div class="chart-stats">
                    <div class="chart-stat">
                        <div class="chart-stat-label">总订单</div>
                        <div class="chart-stat-value">${totalOrders.toLocaleString()}</div>
                    </div>
                    <div class="chart-stat">
                        <div class="chart-stat-label">平台数</div>
                        <div class="chart-stat-value">${platformCount}</div>
                    </div>
                </div>
            </div>
            <div class="chart-content">
                <div class="chart-mode-indicator">${currentMode}</div>
            </div>
        `;

        return chartDiv;
    }

    /**
     * 显示无数据状态
     */
    showNoData() {
        const container = document.getElementById('charts-container');
        if (!container) return;

        container.innerHTML = '';
        
        const noDataMessage = document.getElementById('no-data-message');
        if (noDataMessage) {
            noDataMessage.style.display = 'block';
        }

        // 隐藏统计区域
        const statsElement = document.getElementById('stats-summary');
        if (statsElement) {
            statsElement.style.display = 'none';
        }
    }

    /**
     * 显示/隐藏加载状态
     */
    showLoading(show) {
        const loadingElement = document.getElementById('loading-indicator');
        const refreshButton = document.getElementById('refresh-button');

        if (loadingElement) {
            loadingElement.style.display = show ? 'flex' : 'none';
        }

        if (refreshButton) {
            refreshButton.disabled = show;
            if (show) {
                refreshButton.querySelector('.refresh-text').textContent = '加载中...';
            } else {
                refreshButton.querySelector('.refresh-text').textContent = '刷新数据';
            }
        }
    }

    /**
     * 显示错误信息
     */
    showError(message) {
        console.error('应用错误:', message);
        
        const container = document.getElementById('charts-container');
        if (container) {
            container.innerHTML = `
                <div class="chart-error">
                    <div class="chart-error-icon">⚠️</div>
                    <div class="chart-error-message">${message}</div>
                </div>
            `;
        }
    }

    /**
     * 更新缓存信息
     */
    updateCacheInfo(cacheInfo) {
        if (!cacheInfo) return;

        const cacheInfoElement = document.getElementById('cache-info');
        const cacheTimeElement = document.getElementById('cache-time');
        const cacheBadgeElement = document.getElementById('cache-badge');

        if (cacheInfoElement && cacheTimeElement && cacheBadgeElement) {
            // 格式化时间显示，和主页面保持一致
            const updateTime = new Date(cacheInfo.updated_at);
            const formattedTime = updateTime.toLocaleString('zh-CN', {
                year: 'numeric',
                month: '2-digit', 
                day: '2-digit',
                hour: '2-digit',
                minute: '2-digit',
                second: '2-digit'
            });
            
            cacheTimeElement.textContent = `上次加载: ${formattedTime}`;
            
            // 固定显示为有效，因为缓存永不过期
            cacheBadgeElement.textContent = '有效';
            cacheBadgeElement.className = 'cache-badge valid';

            cacheInfoElement.style.display = 'block';
        }
    }

    /**
     * 更新当前图表模式显示
     */
    updateCurrentChartModeDisplay(mode) {
        const currentModeElement = document.getElementById('current-chart-mode');
        if (currentModeElement) {
            const displayName = this.chartManager.getChartModeDisplayName(mode);
            currentModeElement.textContent = displayName;
        }
    }

    /**
     * 更新元素内容
     */
    updateElement(id, content) {
        const element = document.getElementById(id);
        if (element) {
            element.textContent = content;
        }
    }
}
