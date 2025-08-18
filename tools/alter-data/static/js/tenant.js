// 租户管理类
class TenantManager {
    constructor() {
        this.currentTenant = null;
        this.availableTenants = [];
        this.filteredTenants = [];
        this.selectedIndex = -1;
        
        // DOM 元素
        this.tenantInput = document.getElementById('tenant-input');
        this.tenantDropdown = document.getElementById('tenant-dropdown');
        this.tenantDropdownContent = document.getElementById('tenant-dropdown-content');
        this.tenantSearchButton = document.getElementById('tenant-search-button');
        this.loadingIndicator = document.getElementById('tenant-loading-indicator');
        
        // 绑定事件
        this.bindEvents();
        
        // 初始化最近注册租户
        this.initRecentTenants();
        
        // 初始化经常访问租户
        this.initFrequentTenants();
    }

    // 绑定事件
    bindEvents() {
        // 输入框事件
        this.tenantInput.addEventListener('input', (e) => {
            this.handleInput(e.target.value);
        });
        
        this.tenantInput.addEventListener('focus', () => {
            this.showDropdown();
        });
        
        this.tenantInput.addEventListener('keydown', (e) => {
            this.handleKeyDown(e);
        });
        
        // 搜索按钮事件
        this.tenantSearchButton.addEventListener('click', () => {
            this.selectCurrentInput();
        });
        
        // 点击外部关闭下拉列表
        document.addEventListener('click', (e) => {
            if (!this.tenantInput.contains(e.target) && !this.tenantDropdown.contains(e.target)) {
                this.hideDropdown();
            }
        });
    }

    // 加载租户列表
    async loadTenants() {
        try {
            this.showLoading(true);
            
            const response = await fetch('/api/tenants');
            const result = await response.json();
            
            if (result.success) {
                this.availableTenants = result.data;
                this.filteredTenants = [...this.availableTenants];
                
                // 更新输入框占位符
                this.tenantInput.placeholder = `输入或搜索租户ID... (共${this.availableTenants.length}个)`;
                
                // 设置默认租户 134301
                if (!this.currentTenant) {
                    await this.setDefaultTenant(134301);
                }
            } else {
                throw new Error(result.message || '加载租户列表失败');
            }
        } catch (error) {
            console.error('加载租户失败:', error);
            this.showError('加载租户列表失败: ' + error.message);
        } finally {
            this.showLoading(false);
        }
    }

    // 处理输入事件
    handleInput(value) {
        this.filterTenants(value);
        this.renderDropdownOptions();
        this.showDropdown();
        this.selectedIndex = -1;
    }
    
    // 处理键盘事件
    handleKeyDown(e) {
        const options = this.tenantDropdownContent.querySelectorAll('.tenant-option, .tenant-manual-option');
        
        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault();
                this.selectedIndex = Math.min(this.selectedIndex + 1, options.length - 1);
                this.updateSelection();
                break;
                
            case 'ArrowUp':
                e.preventDefault();
                this.selectedIndex = Math.max(this.selectedIndex - 1, -1);
                this.updateSelection();
                break;
                
            case 'Enter':
                e.preventDefault();
                if (this.selectedIndex >= 0 && options[this.selectedIndex]) {
                    options[this.selectedIndex].click();
                } else {
                    this.selectCurrentInput();
                }
                break;
                
            case 'Escape':
                this.hideDropdown();
                this.tenantInput.blur();
                break;
        }
    }
    
    // 过滤租户列表
    filterTenants(query) {
        if (!query || !query.trim()) {
            // 输入为空时显示所有租户
            this.filteredTenants = [...this.availableTenants];
            return;
        }
        
        const searchQuery = query.toLowerCase();
        this.filteredTenants = this.availableTenants.filter(tenant => {
            return tenant.tenant_id.toString().includes(searchQuery) ||
                   tenant.tenant_name.toLowerCase().includes(searchQuery);
        });
    }
    
    // 渲染下拉选项
    renderDropdownOptions() {
        this.tenantDropdownContent.innerHTML = '';
        
        const inputValue = this.tenantInput.value.trim();
        
        // 如果有输入值且不是纯数字，显示"使用输入值"选项
        if (inputValue && !/^\d+$/.test(inputValue)) {
            const manualOption = document.createElement('div');
            manualOption.className = 'tenant-manual-option';
            manualOption.innerHTML = `📝 使用租户ID: ${inputValue}`;
            manualOption.addEventListener('click', () => {
                this.selectTenantById(inputValue);
            });
            this.tenantDropdownContent.appendChild(manualOption);
        }
        
        // 显示过滤后的租户选项
        if (this.filteredTenants.length > 0) {
            const displayCount = inputValue ? 10 : 15; // 空输入时显示更多
            this.filteredTenants.slice(0, displayCount).forEach(tenant => {
                const option = document.createElement('div');
                option.className = 'tenant-option';
                option.innerHTML = `
                    <span class="tenant-option-name">${tenant.tenant_name}</span>
                    <span class="tenant-option-id">ID: ${tenant.tenant_id}</span>
                `;
                option.addEventListener('click', () => {
                    this.selectTenant(tenant);
                });
                this.tenantDropdownContent.appendChild(option);
            });
            
            // 如果有更多结果，显示提示
            if (this.filteredTenants.length > displayCount) {
                const moreResults = document.createElement('div');
                moreResults.className = 'tenant-no-results';
                moreResults.textContent = `还有 ${this.filteredTenants.length - displayCount} 个租户，继续输入以筛选...`;
                this.tenantDropdownContent.appendChild(moreResults);
            }
        } else if (!inputValue) {
            // 没有输入值但也没有租户（异常情况）
            const noResults = document.createElement('div');
            noResults.className = 'tenant-no-results';
            noResults.textContent = '暂无可用租户...';
            this.tenantDropdownContent.appendChild(noResults);
        } else {
            // 有输入但没有匹配结果
            const noResults = document.createElement('div');
            noResults.className = 'tenant-no-results';
            noResults.textContent = '未找到匹配的租户';
            this.tenantDropdownContent.appendChild(noResults);
        }
    }
    
    // 更新选择高亮
    updateSelection() {
        const options = this.tenantDropdownContent.querySelectorAll('.tenant-option, .tenant-manual-option');
        options.forEach((option, index) => {
            option.classList.toggle('highlighted', index === this.selectedIndex);
        });
        
        // 滚动到选中项
        if (this.selectedIndex >= 0 && options[this.selectedIndex]) {
            options[this.selectedIndex].scrollIntoView({
                block: 'nearest',
                behavior: 'smooth'
            });
        }
    }
    
    // 显示下拉列表
    showDropdown() {
        const inputValue = this.tenantInput.value;
        console.log('🔍 显示下拉列表，当前输入值:', `"${inputValue}"`);
        
        this.filterTenants(inputValue);
        console.log('📋 过滤后的租户数量:', this.filteredTenants.length);
        
        this.renderDropdownOptions();
        this.tenantDropdown.style.display = 'block';
    }
    
    // 隐藏下拉列表
    hideDropdown() {
        this.tenantDropdown.style.display = 'none';
        this.selectedIndex = -1;
    }
    
    // 选择租户
    selectTenant(tenant) {
        this.tenantInput.value = tenant.tenant_id;
        this.hideDropdown();
        this.switchTenant(tenant.tenant_id);
    }
    
    // 通过ID选择租户
    selectTenantById(tenantId) {
        const id = parseInt(tenantId);
        if (isNaN(id) || id <= 0) {
            alert('请输入有效的租户ID（正整数）');
            return;
        }
        
        // 输入框只显示租户ID
        this.tenantInput.value = id;
        
        this.hideDropdown();
        this.switchTenant(id);
    }
    
    // 选择当前输入
    selectCurrentInput() {
        const inputValue = this.tenantInput.value.trim();
        if (!inputValue) {
            alert('请输入租户ID或选择租户');
            return;
        }
        
        // 检查是否是纯数字
        if (/^\d+$/.test(inputValue)) {
            this.selectTenantById(inputValue);
        } else {
            // 查找匹配的租户名称
            const tenant = this.availableTenants.find(t => 
                t.tenant_name.toLowerCase() === inputValue.toLowerCase()
            );
            if (tenant) {
                this.selectTenant(tenant);
            } else {
                alert('未找到匹配的租户，请输入有效的租户ID');
            }
        }
    }

    // 切换租户
    async switchTenant(tenantID) {
        if (!tenantID) return;
        
        try {
            this.showLoading(true);
            this.currentTenant = tenantID;
            
            // 强制清除所有现有图表
            if (window.dashboard && window.dashboard.chartManager) {
                console.log('🧹 租户切换：强制清除现有图表');
                window.dashboard.chartManager.destroyCharts();
                
                // 等待图表清除完成
                await new Promise(resolve => setTimeout(resolve, 200));
            }
            
            // 通知主应用程序租户已切换
            if (window.dashboard) {
                await window.dashboard.loadTenantCrossPlatformData(tenantID);
            }
            
            // 更新URL（可选，用于书签和分享）
            this.updateURL(tenantID);
            
        } catch (error) {
            console.error('切换租户失败:', error);
            this.showError('切换租户失败: ' + error.message);
            
            // 出错时禁用刷新按钮
            if (window.dashboard) {
                window.dashboard.updateRefreshButton(false);
            }
        } finally {
            this.showLoading(false);
        }
    }

    // 获取当前租户
    getCurrentTenant() {
        return this.currentTenant;
    }

    // 获取租户显示名称
    getTenantDisplayName(tenantID) {
        const tenant = this.availableTenants.find(t => t.tenant_id === tenantID);
        return tenant ? tenant.tenant_name : `Tenant ${tenantID}`;
    }

    // 显示加载状态
    showLoading(show) {
        if (this.loadingIndicator) {
            this.loadingIndicator.style.display = show ? 'flex' : 'none';
        }
        
        this.tenantInput.disabled = show;
        this.tenantSearchButton.disabled = show;
    }

    // 显示错误信息
    showError(message) {
        // 创建临时错误提示
        const existingError = document.querySelector('.tenant-error');
        if (existingError) {
            existingError.remove();
        }
        
        const errorDiv = document.createElement('div');
        errorDiv.className = 'tenant-error';
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
        
        const tenantSelector = document.querySelector('.tenant-selector');
        tenantSelector.appendChild(errorDiv);
        
        // 3秒后自动移除
        setTimeout(() => {
            if (errorDiv.parentNode) {
                errorDiv.remove();
            }
        }, 3000);
    }

    // 更新URL（用于书签和分享）
    updateURL(tenantID) {
        const url = new URL(window.location);
        if (tenantID) {
            url.searchParams.set('tenant', tenantID);
        } else {
            url.searchParams.delete('tenant');
        }
        window.history.replaceState({}, '', url);
    }

    // 从URL获取租户参数
    getTenantFromURL() {
        const urlParams = new URLSearchParams(window.location.search);
        const tenantParam = urlParams.get('tenant');
        return tenantParam ? parseInt(tenantParam) : null;
    }

    // 重置租户选择
    reset() {
        this.currentTenant = null;
        this.tenantInput.value = '';
        this.hideDropdown();
        this.updateURL(null);
    }

    // 获取租户统计信息
    getStats() {
        return {
            totalTenants: this.availableTenants.length,
            currentTenant: this.currentTenant,
        };
    }

    // 设置默认租户
    async setDefaultTenant(defaultTenantID) {
        try {
            // 输入框只显示租户ID
            this.tenantInput.value = defaultTenantID;
            
            const defaultTenant = this.availableTenants.find(t => t.tenant_id === defaultTenantID);
            if (defaultTenant) {
                console.log(`✅ 设置默认租户: ${defaultTenant.tenant_name} (ID: ${defaultTenantID})`);
            } else {
                console.log(`⚠️ 默认租户 ${defaultTenantID} 不在列表中，但仍可查询`);
            }
            
            // 切换到默认租户
            await this.switchTenant(defaultTenantID);
            
        } catch (error) {
            console.error('设置默认租户失败:', error);
            // 即使设置失败，也只显示ID
            this.tenantInput.value = defaultTenantID;
        }
    }

    // 刷新租户列表
    async refresh() {
        await this.loadTenants();
    }

    // 初始化最近注册租户功能
    async initRecentTenants() {
        const recentTenantsSection = document.getElementById('recent-tenants-section');
        const recentTenantsRefresh = document.getElementById('recent-tenants-refresh');
        
        if (!recentTenantsSection || !recentTenantsRefresh) {
            console.warn('⚠️ 最近注册租户相关元素未找到');
            return;
        }

        // 绑定刷新按钮事件
        recentTenantsRefresh.addEventListener('click', () => {
            this.loadRecentTenants(true);
        });

        // 初始加载最近注册租户
        await this.loadRecentTenants(false);
    }

    // 加载最近注册租户
    async loadRecentTenants(forceRefresh = false) {
        const grid = document.getElementById('recent-tenants-grid');
        
        if (!grid) {
            console.error('❌ 最近租户网格元素未找到');
            return;
        }

        try {
            // 显示加载状态
            grid.innerHTML = '<div class="recent-tenants-loading"><span class="spinner"></span><span>加载中...</span></div>';
            
            const url = `/api/tenants/recent${forceRefresh ? '?refresh=true' : ''}`;
            const response = await fetch(url);
            const data = await response.json();

            if (data.success && data.data && data.data.length > 0) {
                this.renderRecentTenants(data.data);
                console.log(`✅ 加载了 ${data.data.length} 个最近注册租户`);
            } else {
                grid.innerHTML = '<div class="recent-tenants-empty">🔍 暂无最近注册的租户</div>';
                console.log('ℹ️ 暂无最近注册的租户');
            }
        } catch (error) {
            console.error('❌ 加载最近注册租户失败:', error);
            grid.innerHTML = '<div class="recent-tenants-error">⚠️ 加载失败，请稍后重试</div>';
        }
    }

    // 渲染最近注册租户
    renderRecentTenants(recentTenants) {
        const grid = document.getElementById('recent-tenants-grid');
        if (!grid) return;

        // 清空网格
        grid.innerHTML = '';

        // 为每个最近注册租户创建卡片
        recentTenants.forEach(tenant => {
            const card = document.createElement('div');
            card.className = 'recent-tenant-card';
            card.dataset.tenantId = tenant.tenant_id;
            
            // 格式化注册时间
            const registerTime = tenant.register_time ? 
                new Date(tenant.register_time).toLocaleDateString('zh-CN', {
                    month: 'short',
                    day: 'numeric'
                }) : '未知';

            // 计算注册天数
            const daysAgo = tenant.register_time ? 
                Math.floor((new Date() - new Date(tenant.register_time)) / (1000 * 60 * 60 * 24)) : null;
            
            const timeLabel = daysAgo !== null ? 
                (daysAgo === 0 ? '今天注册' : 
                 daysAgo === 1 ? '昨天注册' : 
                 `${daysAgo}天前`) : '未知';

            card.innerHTML = `
                <div class="recent-tenant-id">${tenant.tenant_id}</div>
            `;

            // 添加点击事件
            card.addEventListener('click', () => {
                this.selectRecentTenant(tenant);
            });

            grid.appendChild(card);
        });

        console.log(`📋 渲染了 ${recentTenants.length} 个最近注册租户卡片`);
    }

    // 选择最近注册租户
    selectRecentTenant(tenant) {
        console.log(`🎯 选择最近注册租户: ${tenant.tenant_id}`);
        
        // 更新输入框
        this.tenantInput.value = tenant.tenant_id;
        
        // 隐藏下拉列表
        this.hideDropdown();
        
        // 更新卡片选中状态
        this.updateRecentTenantSelection(tenant.tenant_id);
        
        // 切换到选中的租户
        this.switchTenant(tenant.tenant_id);
    }

    // 更新最近租户卡片的选中状态
    updateRecentTenantSelection(selectedTenantId) {
        const cards = document.querySelectorAll('.recent-tenant-card');
        cards.forEach(card => {
            const tenantId = card.dataset.tenantId;
            if (tenantId == selectedTenantId) {
                card.classList.add('selected');
            } else {
                card.classList.remove('selected');
            }
        });
    }

    // 初始化经常访问租户功能
    async initFrequentTenants() {
        const frequentTenantsSection = document.getElementById('frequent-tenants-section');
        const frequentTenantsRefresh = document.getElementById('frequent-tenants-refresh');
        
        if (!frequentTenantsSection || !frequentTenantsRefresh) {
            console.warn('⚠️ 经常访问租户相关元素未找到');
            return;
        }

        // 绑定刷新按钮事件
        frequentTenantsRefresh.addEventListener('click', () => {
            this.loadFrequentTenants(true);
        });

        // 初始加载经常访问租户
        await this.loadFrequentTenants(false);
    }

    // 加载经常访问租户
    async loadFrequentTenants(forceRefresh = false) {
        const grid = document.getElementById('frequent-tenants-grid');
        
        if (!grid) {
            console.error('❌ 经常访问租户网格元素未找到');
            return;
        }

        try {
            // 显示加载状态
            grid.innerHTML = '<div class="frequent-tenants-loading"><span class="spinner"></span><span>加载中...</span></div>';
            
            const url = `/api/tenants/frequent${forceRefresh ? '?refresh=true' : ''}`;
            const response = await fetch(url);
            const data = await response.json();

            if (data.success && data.data && data.data.length > 0) {
                this.renderFrequentTenants(data.data);
                console.log(`✅ 加载了 ${data.data.length} 个经常访问租户`);
            } else {
                grid.innerHTML = '<div class="frequent-tenants-empty">🔍 暂无经常访问的租户</div>';
                console.log('ℹ️ 暂无经常访问的租户');
            }
        } catch (error) {
            console.error('❌ 加载经常访问租户失败:', error);
            grid.innerHTML = '<div class="frequent-tenants-error">⚠️ 加载失败，请稍后重试</div>';
        }
    }

    // 渲染经常访问租户
    renderFrequentTenants(frequentTenants) {
        const grid = document.getElementById('frequent-tenants-grid');
        if (!grid) return;

        // 清空网格
        grid.innerHTML = '';

        // 为每个经常访问租户创建卡片
        frequentTenants.forEach(tenant => {
            const card = document.createElement('div');
            card.className = 'frequent-tenant-card';
            card.dataset.tenantId = tenant.tenant_id;
            
            // 格式化访问次数
            const accessCount = tenant.access_count || 0;
            const accessText = accessCount > 99 ? '99+' : accessCount.toString();

            // 计算最后访问时间
            const lastAccess = tenant.last_access ? 
                new Date(tenant.last_access) : null;
            
            let timeLabel = '未知';
            if (lastAccess) {
                const diffHours = Math.floor((new Date() - lastAccess) / (1000 * 60 * 60));
                if (diffHours < 1) {
                    timeLabel = '刚访问';
                } else if (diffHours < 24) {
                    timeLabel = `${diffHours}h前`;
                } else {
                    const diffDays = Math.floor(diffHours / 24);
                    timeLabel = `${diffDays}天前`;
                }
            }

            card.innerHTML = `
                <div class="frequent-tenant-id">${tenant.tenant_id}</div>
                <div class="frequent-tenant-stats">
                    <span class="access-count">${accessText}次</span>
                    <span class="last-access">${timeLabel}</span>
                </div>
            `;

            // 添加点击事件
            card.addEventListener('click', () => {
                this.selectFrequentTenant(tenant);
            });

            grid.appendChild(card);
        });

        console.log(`📋 渲染了 ${frequentTenants.length} 个经常访问租户卡片`);
    }

    // 选择经常访问租户
    selectFrequentTenant(tenant) {
        console.log(`🎯 选择经常访问租户: ${tenant.tenant_id}`);
        
        // 更新输入框
        this.tenantInput.value = tenant.tenant_id;
        
        // 隐藏下拉列表
        this.hideDropdown();
        
        // 更新卡片选中状态
        this.updateFrequentTenantSelection(tenant.tenant_id);
        
        // 切换到选中的租户
        this.switchTenant(tenant.tenant_id);
    }

    // 更新经常访问租户卡片的选中状态
    updateFrequentTenantSelection(selectedTenantId) {
        const cards = document.querySelectorAll('.frequent-tenant-card');
        cards.forEach(card => {
            const tenantId = card.dataset.tenantId;
            if (tenantId == selectedTenantId) {
                card.classList.add('selected');
            } else {
                card.classList.remove('selected');
            }
        });
    }
}
