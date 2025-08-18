// 平台管理类
class PlatformManager {
    constructor() {
        this.currentPlatform = null;
        this.availablePlatforms = [];
        this.platformSelect = document.getElementById('platform-select');
        this.loadingIndicator = document.getElementById('loading-indicator');
        
        // 绑定事件
        this.bindEvents();
    }

    // 绑定事件
    bindEvents() {
        this.platformSelect.addEventListener('change', (e) => {
            const selectedPlatform = e.target.value;
            if (selectedPlatform && selectedPlatform !== this.currentPlatform) {
                this.switchPlatform(selectedPlatform);
            }
        });
    }

    // 加载平台列表
    async loadPlatforms() {
        try {
            this.showLoading(true);
            
            const response = await fetch('/api/platforms');
            const result = await response.json();
            
            if (result.success) {
                this.availablePlatforms = result.data;
                this.renderPlatformOptions();
                
                // 默认选择第一个平台
                if (this.availablePlatforms.length > 0) {
                    const defaultPlatform = this.availablePlatforms[0].name;
                    this.platformSelect.value = defaultPlatform;
                    await this.switchPlatform(defaultPlatform);
                }
            } else {
                throw new Error(result.message || '加载平台列表失败');
            }
        } catch (error) {
            console.error('加载平台失败:', error);
            this.showError('加载平台列表失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    // 渲染平台选项
    renderPlatformOptions() {
        this.platformSelect.innerHTML = '';
        
        // 添加默认选项
        const defaultOption = document.createElement('option');
        defaultOption.value = '';
        defaultOption.textContent = '请选择平台';
        this.platformSelect.appendChild(defaultOption);
        
        // 添加平台选项
        this.availablePlatforms.forEach(platform => {
            const option = document.createElement('option');
            option.value = platform.name;
            option.textContent = platform.display_name;
            
            // 添加状态标识
            option.setAttribute('data-platform', platform.name);
            this.platformSelect.appendChild(option);
        });
    }

    // 切换平台
    async switchPlatform(platformName) {
        if (!platformName) {
            // 如果没有选择平台，禁用刷新按钮
            if (window.dashboard) {
                window.dashboard.updateRefreshButton(false);
            }
            return;
        }
        
        try {
            this.showLoading(true);
            this.currentPlatform = platformName;
            
            // 通知主应用程序平台已切换
            if (window.dashboard) {
                await window.dashboard.loadPlatformData(platformName);
            }
            
            // 更新URL（可选，用于书签和分享）
            this.updateURL(platformName);
            
        } catch (error) {
            console.error('切换平台失败:', error);
            this.showError('切换平台失败: ' + error.message);
            
            // 出错时禁用刷新按钮
            if (window.dashboard) {
                window.dashboard.updateRefreshButton(false);
            }
        } finally {
            this.showLoading(false);
        }
    }

    // 获取当前平台
    getCurrentPlatform() {
        return this.currentPlatform;
    }

    // 获取平台显示名称
    getPlatformDisplayName(platformName) {
        const platform = this.availablePlatforms.find(p => p.name === platformName);
        return platform ? platform.display_name : platformName;
    }

    // 检查平台是否已实现
    isPlatformImplemented(platformName) {
        // 这里可以通过API检查，目前只有google实现了
        return platformName === 'google';
    }

    // 显示加载状态
    showLoading(show) {
        if (this.loadingIndicator) {
            this.loadingIndicator.style.display = show ? 'flex' : 'none';
        }
        
        this.platformSelect.disabled = show;
    }

    // 显示错误信息
    showError(message) {
        // 创建临时错误提示
        const existingError = document.querySelector('.platform-error');
        if (existingError) {
            existingError.remove();
        }
        
        const errorDiv = document.createElement('div');
        errorDiv.className = 'platform-error';
        errorDiv.style.cssText = `
            color: #e74c3c;
            font-size: 0.9rem;
            margin-top: 10px;
            padding: 10px;
            background: #fdf2f2;
            border: 1px solid #f8d7da;
            border-radius: 6px;
            text-align: center;
        `;
        errorDiv.textContent = message;
        
        const platformSelector = document.querySelector('.platform-selector');
        platformSelector.appendChild(errorDiv);
        
        // 3秒后自动移除
        setTimeout(() => {
            if (errorDiv.parentNode) {
                errorDiv.remove();
            }
        }, 3000);
    }

    // 更新URL（用于书签和分享）
    updateURL(platformName) {
        const url = new URL(window.location);
        if (platformName) {
            url.searchParams.set('platform', platformName);
            // 平台视图：清除租户参数，避免参数混乱
            url.searchParams.delete('tenant');
        } else {
            url.searchParams.delete('platform');
        }
        window.history.replaceState({}, '', url);
        console.log(`🔗 Platform URL updated: platform=${platformName || 'null'}`);
    }

    // 从URL获取平台参数
    getPlatformFromURL() {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get('platform');
    }

    // 重置平台选择
    reset() {
        this.currentPlatform = null;
        this.platformSelect.value = '';
        this.updateURL(null);
    }

    // 获取平台统计信息
    getStats() {
        return {
            totalPlatforms: this.availablePlatforms.length,
            currentPlatform: this.currentPlatform,
            implementedPlatforms: this.availablePlatforms.filter(p => 
                this.isPlatformImplemented(p.name)
            ).length
        };
    }

    // 刷新平台列表
    async refresh() {
        await this.loadPlatforms();
    }
}
