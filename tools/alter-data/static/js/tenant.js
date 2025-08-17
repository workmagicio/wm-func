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
}
