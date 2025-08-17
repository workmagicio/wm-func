// 主应用程序类
class Dashboard {
    constructor() {
        this.platformManager = new PlatformManager();
        this.tenantManager = new TenantManager();
        this.chartManager = new ChartManager();
        this.currentData = [];
        this.currentCacheInfo = null;
        this.currentViewMode = 'platform'; // 'platform' or 'tenant'
        
        // DOM 元素
        this.chartsContainer = document.getElementById('charts-container');
        this.noDataMessage = document.getElementById('no-data-message');
        this.errorMessage = document.getElementById('error-message');
        this.errorText = document.getElementById('error-text');
        
        // 平台视图元素
        this.refreshButton = document.getElementById('refresh-button');
        this.cacheInfo = document.getElementById('cache-info');
        this.cacheTime = document.getElementById('cache-time');
        this.cacheBadge = document.getElementById('cache-badge');
        this.platformViewContainer = document.getElementById('platform-view-container');
        
        // 租户视图元素
        this.refreshTenantButton = document.getElementById('refresh-tenant-button');
        this.tenantCacheInfo = document.getElementById('tenant-cache-info');
        this.tenantCacheTime = document.getElementById('tenant-cache-time');
        this.tenantCacheBadge = document.getElementById('tenant-cache-badge');
        this.tenantViewContainer = document.getElementById('tenant-view-container');
        
        // 视图模式切换元素
        this.platformViewRadio = document.getElementById('platform-view');
        this.tenantViewRadio = document.getElementById('tenant-view');
        
        // 状态
        this.isLoading = false;
        this.isRefreshing = false;
        
        // 绑定事件
        this.bindViewModeEvents();
    }

    // 绑定视图模式切换事件
    bindViewModeEvents() {
        this.platformViewRadio.addEventListener('change', () => {
            if (this.platformViewRadio.checked) {
                this.switchViewMode('platform');
            }
        });
        
        this.tenantViewRadio.addEventListener('change', () => {
            if (this.tenantViewRadio.checked) {
                this.switchViewMode('tenant');
            }
        });
    }

    // 切换视图模式
    async switchViewMode(mode) {
        console.log(`🔄 切换到${mode === 'platform' ? '平台' : '租户'}视图`);
        
        this.currentViewMode = mode;
        
        // 隐藏所有内容
        this.hideMessages();
        this.chartsContainer.style.display = 'none';
        
        if (mode === 'platform') {
            // 显示平台视图，隐藏租户视图
            this.platformViewContainer.style.display = 'block';
            this.tenantViewContainer.style.display = 'none';
            
            // 检查URL中的平台参数
            const platformFromURL = this.platformManager.getPlatformFromURL();
            if (platformFromURL && this.platformManager.platformSelect.value !== platformFromURL) {
                this.platformManager.platformSelect.value = platformFromURL;
                await this.loadPlatformData(platformFromURL);
            } else if (this.platformManager.getCurrentPlatform()) {
                await this.loadPlatformData(this.platformManager.getCurrentPlatform());
            } else {
                this.updateRefreshButton(false);
            }
        } else {
            // 显示租户视图，隐藏平台视图
            this.platformViewContainer.style.display = 'none';
            this.tenantViewContainer.style.display = 'block';
            
            // 加载租户列表
            await this.tenantManager.loadTenants();
            
            // 检查URL中的租户参数，如果没有则使用默认租户
            const tenantFromURL = this.tenantManager.getTenantFromURL();
            if (tenantFromURL) {
                // 输入框只显示租户ID
                this.tenantManager.tenantInput.value = tenantFromURL;
                await this.loadTenantCrossPlatformData(tenantFromURL);
            }
            // 注意：默认租户134301已在loadTenants()中设置
        }
    }

    // 初始化应用
    async init() {
        try {
            console.log('🚀 初始化数据监控看板...');
            
            // 显示加载状态
            this.showLoading(true);
            
            // 加载平台列表
            await this.platformManager.loadPlatforms();
            
            // 检查URL参数决定初始视图模式
            const platformFromURL = this.platformManager.getPlatformFromURL();
            const tenantFromURL = this.tenantManager.getTenantFromURL();
            
            if (tenantFromURL) {
                // 如果URL中有租户参数，切换到租户视图
                this.tenantViewRadio.checked = true;
                await this.switchViewMode('tenant');
                
                // 加载租户列表后设置租户输入框（只显示ID）
                setTimeout(() => {
                    this.tenantManager.tenantInput.value = tenantFromURL;
                }, 100);
            } else if (platformFromURL) {
                // 如果URL中有平台参数，保持平台视图
                this.platformManager.platformSelect.value = platformFromURL;
                await this.loadPlatformData(platformFromURL);
            } else {
                // 默认平台视图，禁用刷新按钮
                this.updateRefreshButton(false);
            }
            
            console.log('✅ 应用初始化完成');
            
        } catch (error) {
            console.error('❌ 应用初始化失败:', error);
            this.showError('应用初始化失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    // 加载平台数据
    async loadPlatformData(platformName, forceRefresh = false) {
        if (!platformName) {
            this.showNoData('请选择一个平台查看数据');
            this.updateRefreshButton(false);
            return;
        }

        try {
            console.log(`📊 加载平台数据: ${platformName} (强制刷新: ${forceRefresh})`);
            this.showLoading(true);
            this.hideMessages();
            
            // 构建URL
            let url = `/api/data/${platformName}`;
            if (forceRefresh) {
                url += '?refresh=true';
            }
            
            const response = await fetch(url);
            const result = await response.json();
            
            if (result.success) {
                this.currentData = result.data;
                this.currentCacheInfo = result.cache_info;
                
                // 更新缓存信息显示
                this.updateCacheInfo(this.currentCacheInfo, forceRefresh);
                
                // 启用刷新按钮
                this.updateRefreshButton(true);
                
                if (this.currentData.length > 0) {
                    this.renderCharts(this.currentData);
                    console.log(`✅ 成功加载 ${this.currentData.length} 个租户的数据`);
                    
                    // 显示成功消息
                    if (forceRefresh) {
                        this.showTemporaryMessage('数据已刷新', 'success');
                    }
                } else {
                    this.showNoData(`平台 ${this.platformManager.getPlatformDisplayName(platformName)} 暂无数据`);
                }
            } else {
                throw new Error(result.message || '加载数据失败');
            }
            
        } catch (error) {
            console.error('❌ 加载平台数据失败:', error);
            this.showError(`加载平台数据失败: ${error.message}`);
            this.updateRefreshButton(false);
        } finally {
            this.showLoading(false);
            this.setRefreshButtonLoading(false);
        }
    }

    // 渲染图表
    renderCharts(tenantDataList) {
        // 清除现有图表
        this.chartManager.destroyCharts();
        
        // 创建新图表
        tenantDataList.forEach(tenantData => {
            this.chartManager.initChart(tenantData);
        });
        
        // 显示图表容器
        this.chartsContainer.style.display = 'block';
        
        // 更新页面标题
        this.updatePageTitle(tenantDataList.length);
    }

    // 显示加载状态
    showLoading(show) {
        this.isLoading = show;
        
        if (show) {
            // 可以添加全局加载指示器
            document.body.style.cursor = 'wait';
        } else {
            document.body.style.cursor = 'default';
        }
    }

    // 显示无数据消息
    showNoData(message) {
        this.hideMessages();
        this.chartsContainer.style.display = 'none';
        this.noDataMessage.style.display = 'block';
        
        if (message) {
            const messageP = this.noDataMessage.querySelector('p');
            if (messageP) {
                messageP.textContent = message;
            }
        }
    }

    // 显示错误消息
    showError(message) {
        this.hideMessages();
        this.chartsContainer.style.display = 'none';
        this.errorMessage.style.display = 'block';
        
        if (this.errorText) {
            this.errorText.textContent = message;
        }
        
        console.error('Dashboard Error:', message);
    }

    // 隐藏所有消息
    hideMessages() {
        this.noDataMessage.style.display = 'none';
        this.errorMessage.style.display = 'none';
    }

    // 重试加载
    async retryLoad() {
        const currentPlatform = this.platformManager.getCurrentPlatform();
        if (currentPlatform) {
            await this.loadPlatformData(currentPlatform);
        } else {
            await this.init();
        }
    }

    // 刷新数据
    async refresh() {
        const currentPlatform = this.platformManager.getCurrentPlatform();
        if (currentPlatform) {
            await this.loadPlatformData(currentPlatform);
        }
    }

    // 刷新当前选择的内容
    async refreshCurrentSelection() {
        if (this.currentViewMode === 'platform') {
            await this.refreshCurrentPlatform();
        } else {
            await this.refreshCurrentTenant();
        }
    }

    // 强制刷新当前平台数据
    async refreshCurrentPlatform() {
        const currentPlatform = this.platformManager.getCurrentPlatform();
        if (!currentPlatform) {
            alert('请先选择一个平台');
            return;
        }

        if (this.isRefreshing) {
            return; // 防止重复点击
        }

        try {
            this.isRefreshing = true;
            this.setRefreshButtonLoading(true);
            await this.loadPlatformData(currentPlatform, true);
        } catch (error) {
            console.error('刷新失败:', error);
            this.showTemporaryMessage('刷新失败: ' + error.message, 'error');
        } finally {
            this.isRefreshing = false;
            this.setRefreshButtonLoading(false);
        }
    }

    // 强制刷新当前租户数据
    async refreshCurrentTenant() {
        const currentTenant = this.tenantManager.getCurrentTenant();
        if (!currentTenant) {
            alert('请先选择一个租户');
            return;
        }

        if (this.isRefreshing) {
            return; // 防止重复点击
        }

        try {
            this.isRefreshing = true;
            this.setRefreshTenantButtonLoading(true);
            await this.loadTenantCrossPlatformData(currentTenant, true);
        } catch (error) {
            console.error('刷新失败:', error);
            this.showTemporaryMessage('刷新失败: ' + error.message, 'error');
        } finally {
            this.isRefreshing = false;
            this.setRefreshTenantButtonLoading(false);
        }
    }

    // 更新缓存信息显示
    updateCacheInfo(cacheInfo, wasRefreshed = false) {
        if (!cacheInfo) {
            this.cacheInfo.style.display = 'none';
            return;
        }

        this.cacheInfo.style.display = 'block';
        
        // 格式化时间
        const updateTime = new Date(cacheInfo.updated_at);
        const now = new Date();
        const diffMinutes = Math.floor((now - updateTime) / 1000 / 60);
        
        let timeText;
        if (diffMinutes < 1) {
            timeText = '刚刚更新';
        } else if (diffMinutes < 60) {
            timeText = `${diffMinutes}分钟前更新`;
        } else {
            const diffHours = Math.floor(diffMinutes / 60);
            if (diffHours < 24) {
                timeText = `${diffHours}小时前更新`;
            } else {
                timeText = updateTime.toLocaleDateString();
            }
        }
        
        this.cacheTime.textContent = timeText;
        
        // 设置状态徽章
        this.cacheBadge.className = 'cache-badge';
        if (wasRefreshed || diffMinutes < 1) {
            this.cacheBadge.textContent = '最新';
            this.cacheBadge.classList.add('fresh');
        } else if (cacheInfo.is_expired) {
            this.cacheBadge.textContent = '已过期';
            this.cacheBadge.classList.add('expired');
        } else {
            this.cacheBadge.textContent = '缓存';
            this.cacheBadge.classList.add('cached');
        }
    }

    // 更新刷新按钮状态
    updateRefreshButton(enabled) {
        if (this.refreshButton) {
            this.refreshButton.disabled = !enabled;
        }
    }

    // 设置刷新按钮加载状态
    setRefreshButtonLoading(loading) {
        if (!this.refreshButton) return;
        
        if (loading) {
            this.refreshButton.classList.add('loading');
            this.refreshButton.disabled = true;
            const textElement = this.refreshButton.querySelector('.refresh-text');
            if (textElement) {
                textElement.textContent = '刷新中...';
            }
        } else {
            this.refreshButton.classList.remove('loading');
            this.refreshButton.disabled = false;
            const textElement = this.refreshButton.querySelector('.refresh-text');
            if (textElement) {
                textElement.textContent = '刷新数据';
            }
        }
    }

    // 显示临时消息
    showTemporaryMessage(message, type = 'info', duration = 3000) {
        // 创建临时消息元素
        const messageDiv = document.createElement('div');
        messageDiv.className = `temp-message temp-message-${type}`;
        messageDiv.textContent = message;
        messageDiv.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 12px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 600;
            z-index: 1000;
            opacity: 0;
            transition: opacity 0.3s ease;
            background: ${type === 'success' ? '#2ecc71' : type === 'error' ? '#e74c3c' : '#3498db'};
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
        `;
        
        document.body.appendChild(messageDiv);
        
        // 显示动画
        setTimeout(() => {
            messageDiv.style.opacity = '1';
        }, 100);
        
        // 自动隐藏
        setTimeout(() => {
            messageDiv.style.opacity = '0';
            setTimeout(() => {
                if (document.body.contains(messageDiv)) {
                    document.body.removeChild(messageDiv);
                }
            }, 300);
        }, duration);
    }

    // 加载租户跨平台数据
    async loadTenantCrossPlatformData(tenantID, forceRefresh = false) {
        if (!tenantID) {
            this.showNoData('请选择一个租户查看数据');
            this.updateRefreshTenantButton(false);
            return;
        }

        try {
            console.log(`👤 加载租户跨平台数据: ${tenantID} (强制刷新: ${forceRefresh})`);
            this.showLoading(true);
            this.hideMessages();
            
            // 构建URL
            let url = `/api/tenant/${tenantID}`;
            if (forceRefresh) {
                url += '?refresh=true';
            }
            
            const response = await fetch(url);
            const result = await response.json();
            
            if (result.success) {
                this.currentData = this.convertCrossPlatformDataToTenantData(result.data);
                this.currentCacheInfo = result.cache_info;
                
                // 更新租户缓存信息显示
                this.updateTenantCacheInfo(this.currentCacheInfo, forceRefresh);
                
                // 启用刷新按钮
                this.updateRefreshTenantButton(true);
                
                if (this.currentData.length > 0) {
                    this.renderCrossPlatformCharts(result.data);
                    console.log(`✅ 成功加载租户 ${tenantID} 的跨平台数据`);
                    
                    // 显示成功消息
                    if (forceRefresh) {
                        this.showTemporaryMessage('数据已刷新', 'success');
                    }
                } else {
                    this.showNoData(`租户 ${result.tenant_name} 暂无数据`);
                }
            } else {
                throw new Error(result.message || '加载数据失败');
            }
            
        } catch (error) {
            console.error('❌ 加载租户数据失败:', error);
            this.showError(`加载租户数据失败: ${error.message}`);
            this.updateRefreshTenantButton(false);
        } finally {
            this.showLoading(false);
            this.setRefreshTenantButtonLoading(false);
        }
    }

    // 将跨平台数据转换为图表数据格式
    convertCrossPlatformDataToTenantData(crossPlatformData) {
        const tenantDataList = [];
        
        for (const [platform, platformData] of Object.entries(crossPlatformData.platform_data)) {
            tenantDataList.push(...platformData);
        }
        
        return tenantDataList;
    }

    // 渲染跨平台图表
    renderCrossPlatformCharts(crossPlatformData) {
        // 清除现有图表
        this.chartManager.destroyCharts();
        
        console.log('🎯 开始渲染跨平台图表...');
        console.log('平台数据:', crossPlatformData.platform_data);
        
        // 为每个平台创建图表
        for (const [platform, platformData] of Object.entries(crossPlatformData.platform_data)) {
            console.log(`📊 处理平台: ${platform}, 数据数量: ${platformData.length}`);
            
            if (platformData.length > 0) {
                platformData.forEach((tenantData, index) => {
                    console.log(`  📈 创建图表 ${index + 1} for ${platform}:`, tenantData.tenant_name);
                    
                    // 修改图表标题以突出显示平台，并确保唯一的图表ID
                    const modifiedTenantData = {
                        ...tenantData,
                        tenant_name: `${crossPlatformData.tenant_name} - ${platform}`,
                        platform: platform,
                        // 添加唯一标识符避免图表ID冲突
                        chart_id: `tenant_${crossPlatformData.tenant_id}_${platform}_${index}`
                    };
                    
                    this.chartManager.initChart(modifiedTenantData);
                });
            } else {
                console.log(`  ⚠️ 平台 ${platform} 没有数据`);
            }
        }
        
        // 显示图表容器
        this.chartsContainer.style.display = 'block';
        
        // 更新页面标题
        const platformCount = Object.keys(crossPlatformData.platform_data).length;
        this.updatePageTitle(`${platformCount}个平台`, 'tenant');
        
        console.log(`✅ 跨平台图表渲染完成，共${platformCount}个平台`);
    }

    // 更新租户缓存信息显示
    updateTenantCacheInfo(cacheInfo, wasRefreshed = false) {
        if (!cacheInfo) {
            this.tenantCacheInfo.style.display = 'none';
            return;
        }

        this.tenantCacheInfo.style.display = 'block';
        
        // 格式化时间
        const updateTime = new Date(cacheInfo.updated_at);
        const now = new Date();
        const diffMinutes = Math.floor((now - updateTime) / 1000 / 60);
        
        let timeText;
        if (diffMinutes < 1) {
            timeText = '刚刚更新';
        } else if (diffMinutes < 60) {
            timeText = `${diffMinutes}分钟前更新`;
        } else {
            const diffHours = Math.floor(diffMinutes / 60);
            if (diffHours < 24) {
                timeText = `${diffHours}小时前更新`;
            } else {
                timeText = updateTime.toLocaleDateString();
            }
        }
        
        this.tenantCacheTime.textContent = timeText;
        
        // 设置状态徽章
        this.tenantCacheBadge.className = 'cache-badge';
        if (wasRefreshed || diffMinutes < 1) {
            this.tenantCacheBadge.textContent = '最新';
            this.tenantCacheBadge.classList.add('fresh');
        } else if (cacheInfo.is_expired) {
            this.tenantCacheBadge.textContent = '已过期';
            this.tenantCacheBadge.classList.add('expired');
        } else {
            this.tenantCacheBadge.textContent = '缓存';
            this.tenantCacheBadge.classList.add('cached');
        }
    }

    // 更新租户刷新按钮状态
    updateRefreshTenantButton(enabled) {
        if (this.refreshTenantButton) {
            this.refreshTenantButton.disabled = !enabled;
        }
    }

    // 设置租户刷新按钮加载状态
    setRefreshTenantButtonLoading(loading) {
        if (!this.refreshTenantButton) return;
        
        if (loading) {
            this.refreshTenantButton.classList.add('loading');
            this.refreshTenantButton.disabled = true;
            const textElement = this.refreshTenantButton.querySelector('.refresh-text');
            if (textElement) {
                textElement.textContent = '刷新中...';
            }
        } else {
            this.refreshTenantButton.classList.remove('loading');
            this.refreshTenantButton.disabled = false;
            const textElement = this.refreshTenantButton.querySelector('.refresh-text');
            if (textElement) {
                textElement.textContent = '刷新数据';
            }
        }
    }

    // 更新页面标题
    updatePageTitle(chartCount, mode = 'platform') {
        if (mode === 'tenant') {
            const tenantName = this.tenantManager.getTenantDisplayName(
                this.tenantManager.getCurrentTenant()
            );
            document.title = `数据监控看板 - ${tenantName} (${chartCount})`;
        } else {
            const platform = this.platformManager.getPlatformDisplayName(
                this.platformManager.getCurrentPlatform()
            );
            document.title = `数据监控看板 - ${platform} (${chartCount}个租户)`;
        }
    }

    // 获取应用统计信息
    getStats() {
        return {
            platform: this.platformManager.getStats(),
            charts: this.chartManager.getChartStats(),
            currentData: {
                tenantCount: this.currentData.length,
                totalDataPoints: this.currentData.reduce((total, tenant) => 
                    total + tenant.date_range.length, 0
                )
            },
            isLoading: this.isLoading
        };
    }

    // 导出数据（可选功能）
    exportData() {
        if (this.currentData.length === 0) {
            alert('暂无数据可导出');
            return;
        }
        
        const dataStr = JSON.stringify(this.currentData, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        
        const link = document.createElement('a');
        link.href = url;
        link.download = `dashboard-data-${this.platformManager.getCurrentPlatform()}-${new Date().toISOString().split('T')[0]}.json`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        URL.revokeObjectURL(url);
        
        console.log('✅ 数据导出完成');
    }
}

// 全局工具函数
window.dashboardUtils = {
    // 格式化数字
    formatNumber: function(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    },
    
    // 格式化日期
    formatDate: function(dateStr) {
        const date = new Date(dateStr);
        return date.toLocaleDateString('zh-CN');
    },
    
    // 计算百分比差异
    calculatePercentageDiff: function(apiSpend, adSpend) {
        if (adSpend === 0) return apiSpend === 0 ? 0 : 100;
        return ((apiSpend - adSpend) / adSpend * 100).toFixed(2);
    },
    
    // 获取状态颜色
    getStatusColor: function(difference) {
        const absDiff = Math.abs(difference);
        if (absDiff === 0) return '#2ecc71'; // 绿色 - 完全一致
        if (absDiff <= 100) return '#f39c12'; // 橙色 - 小差异
        return '#e74c3c'; // 红色 - 大差异
    }
};

// 键盘快捷键支持
document.addEventListener('keydown', function(e) {
    // Ctrl/Cmd + R: 强制刷新数据
    if ((e.ctrlKey || e.metaKey) && e.key === 'r') {
        e.preventDefault();
        if (window.dashboard) {
            window.dashboard.refreshCurrentSelection();
        }
    }
    
    // Ctrl/Cmd + Shift + R: 普通刷新
    if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === 'R') {
        e.preventDefault();
        if (window.dashboard) {
            window.dashboard.refresh();
        }
    }
    
    // Ctrl/Cmd + T: 切换视图模式
    if ((e.ctrlKey || e.metaKey) && e.key === 't') {
        e.preventDefault();
        if (window.dashboard) {
            const newMode = window.dashboard.currentViewMode === 'platform' ? 'tenant' : 'platform';
            const radio = document.getElementById(`${newMode}-view`);
            if (radio) {
                radio.checked = true;
                window.dashboard.switchViewMode(newMode);
            }
        }
    }
    
    // Ctrl/Cmd + E: 导出数据
    if ((e.ctrlKey || e.metaKey) && e.key === 'e') {
        e.preventDefault();
        if (window.dashboard) {
            window.dashboard.exportData();
        }
    }
});

// 窗口大小变化时重新调整图表
window.addEventListener('resize', debounce(function() {
    if (window.dashboard && window.dashboard.chartManager) {
        window.dashboard.chartManager.resizeCharts();
    }
}, 200));

// 防抖函数
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}
