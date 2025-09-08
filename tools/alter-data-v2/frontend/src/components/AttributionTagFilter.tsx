import React, { useState, useEffect, useMemo } from 'react'
import './TagFilter.css'

interface AttributionTenantData {
  tenant_id: number
  tags: string[]
  total_attribution_avg: number
  recent_zero_days: number
  has_recent_zeros: boolean
  date_sequence: any[]
  platform_totals: any[]
}

interface AttributionTagFilterProps {
  data: AttributionTenantData[]
  onFilterChange: (filtered: AttributionTenantData[]) => void
}

type FilterMode = 'all' | 'error' | 'normal' | 'custom' | 'exclude'

interface FilterState {
  mode: FilterMode
  selectedTags: string[]
  excludedTags: string[]
  searchText: string
}

const AttributionTagFilter: React.FC<AttributionTagFilterProps> = ({ data, onFilterChange }) => {
  const [filterState, setFilterState] = useState<FilterState>({
    mode: 'all',
    selectedTags: [],
    excludedTags: [],
    searchText: ''
  })

  // 计算所有标签和统计信息
  const tagStats = useMemo(() => {
    const tagCounts: { [tag: string]: number } = {}
    
    let errorCount = 0
    let normalCount = 0
    
    if (!Array.isArray(data)) {
      return {
        allTags: [],
        errorTags: [],
        normalTags: [],
        tagCounts: {},
        errorCount: 0,
        normalCount: 0,
        totalCount: 0
      }
    }
    
    data.forEach(tenant => {
      const tags = tenant.tags || []
      const hasErrorTag = tags.some(tag => tag.startsWith('err_'))
      const hasNormalTag = tags.some(tag => !tag.startsWith('err_'))
      
      if (hasErrorTag && !hasNormalTag) {
        errorCount++
      } else {
        normalCount++
      }
      
      tags.forEach(tag => {
        tagCounts[tag] = (tagCounts[tag] || 0) + 1
      })
    })
    
    const allTags = Object.keys(tagCounts).sort()
    const errorTags = allTags.filter(tag => tag.startsWith('err_'))
    const normalTags = allTags.filter(tag => !tag.startsWith('err_'))
    
    return {
      allTags,
      errorTags,
      normalTags,
      tagCounts,
      errorCount,
      normalCount,
      totalCount: data.length
    }
  }, [data])

  // 应用筛选逻辑
  const filteredData = useMemo(() => {
    if (!Array.isArray(data)) {
      return []
    }
    
    let filtered = data

    // 按模式筛选
    if (filterState.mode === 'error') {
      filtered = filtered.filter(tenant => 
        tenant.tags?.some(tag => tag.startsWith('err_')) && 
        !tenant.tags?.some(tag => !tag.startsWith('err_'))
      )
    } else if (filterState.mode === 'normal') {
      filtered = filtered.filter(tenant => 
        !tenant.tags?.some(tag => tag.startsWith('err_')) ||
        tenant.tags?.some(tag => !tag.startsWith('err_'))
      )
    }

    // 按选中的标签筛选
    if (filterState.selectedTags.length > 0) {
      filtered = filtered.filter(tenant =>
        filterState.selectedTags.every(selectedTag =>
          tenant.tags?.includes(selectedTag)
        )
      )
    }

    // 排除指定的标签
    if (filterState.excludedTags.length > 0) {
      filtered = filtered.filter(tenant =>
        !filterState.excludedTags.some(excludedTag =>
          tenant.tags?.includes(excludedTag)
        )
      )
    }

    // 按搜索文本筛选
    if (filterState.searchText.trim()) {
      const searchLower = filterState.searchText.toLowerCase()
      filtered = filtered.filter(tenant =>
        tenant.tenant_id.toString().includes(searchLower) ||
        tenant.tags?.some(tag => tag.toLowerCase().includes(searchLower))
      )
    }

    return filtered
  }, [data, filterState])

  // 当筛选结果改变时通知父组件
  useEffect(() => {
    onFilterChange(filteredData)
  }, [filteredData, onFilterChange])

  const handleModeChange = (mode: FilterMode) => {
    setFilterState(prev => ({ ...prev, mode, selectedTags: [] }))
  }

  const handleTagToggle = (tag: string) => {
    setFilterState(prev => ({
      ...prev,
      selectedTags: prev.selectedTags.includes(tag)
        ? prev.selectedTags.filter(t => t !== tag)
        : [...prev.selectedTags, tag]
    }))
  }

  const handleExcludeTagToggle = (tag: string) => {
    setFilterState(prev => {
      const newExcludedTags = prev.excludedTags.includes(tag)
        ? prev.excludedTags.filter(t => t !== tag)
        : [...prev.excludedTags, tag]
      
      return {
        ...prev,
        excludedTags: newExcludedTags
      }
    })
  }

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilterState(prev => ({ ...prev, searchText: e.target.value }))
  }

  const clearFilters = () => {
    setFilterState({
      mode: 'all',
      selectedTags: [],
      excludedTags: [],
      searchText: ''
    })
  }

  const isFiltered = filterState.mode !== 'all' || filterState.selectedTags.length > 0 || filterState.excludedTags.length > 0 || filterState.searchText.trim() !== ''

  return (
    <div className="tag-filter-section">
      {/* 快捷筛选按钮 */}
      <div className="quick-filters">
        <button
          className={`filter-btn ${filterState.mode === 'all' ? 'active' : ''}`}
          onClick={() => handleModeChange('all')}
        >
          全部 ({tagStats.totalCount})
        </button>
        <button
          className={`filter-btn error ${filterState.mode === 'error' ? 'active' : ''}`}
          onClick={() => handleModeChange('error')}
        >
          异常 ({tagStats.errorCount})
        </button>
        <button
          className={`filter-btn normal ${filterState.mode === 'normal' ? 'active' : ''}`}
          onClick={() => handleModeChange('normal')}
        >
          正常 ({tagStats.normalCount})
        </button>
        
        {/* 排除标签快捷指示器 */}
        {filterState.excludedTags.length > 0 && (
          <div className="excluded-tags-indicator">
            <span className="excluded-label">已排除:</span>
            {filterState.excludedTags.slice(0, 2).map(tag => (
              <span key={tag} className="excluded-tag-mini">
                🚫 {tag}
              </span>
            ))}
            {filterState.excludedTags.length > 2 && (
              <span className="excluded-more">+{filterState.excludedTags.length - 2}</span>
            )}
          </div>
        )}
      </div>

      {/* 搜索框 */}
      <div className="search-section">
        <input
          type="text"
          placeholder="搜索租户ID或标签..."
          value={filterState.searchText}
          onChange={handleSearchChange}
          className="search-input"
        />
        {isFiltered && (
          <button onClick={clearFilters} className="clear-btn">
            清空筛选
          </button>
        )}
      </div>

      {/* 标签选择器 - 重新设计为更便捷的方式 */}
      {(tagStats.errorTags.length > 0 || tagStats.normalTags.length > 0) && (
        <div className="tag-selector">
          <div className="tag-section">
            <div className="tag-actions-header">
              <h4>标签筛选</h4>
              <div className="tag-help">
                <span className="help-text">💡 点击标签包含，点击🚫排除</span>
              </div>
            </div>

            {/* 异常标签 */}
            {tagStats.errorTags.length > 0 && (
              <div className="tag-group">
                <h4 className="tag-group-title">异常标签</h4>
                <div className="tag-list">
                  {tagStats.errorTags.map(tag => (
                    <div key={tag} className="tag-item-wrapper">
                      <button
                        className={`tag-option error ${filterState.selectedTags.includes(tag) ? 'selected' : ''} ${filterState.excludedTags.includes(tag) ? 'excluded' : ''}`}
                        onClick={() => handleTagToggle(tag)}
                        disabled={filterState.excludedTags.includes(tag)}
                      >
                        {tag} ({tagStats.tagCounts[tag]})
                      </button>
                      <button
                        className={`exclude-btn ${filterState.excludedTags.includes(tag) ? 'active' : ''}`}
                        onClick={() => handleExcludeTagToggle(tag)}
                        title={filterState.excludedTags.includes(tag) ? '取消排除' : '排除此标签'}
                      >
                        {filterState.excludedTags.includes(tag) ? '❌' : '🚫'}
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}
            
            {/* 正常标签 */}
            {tagStats.normalTags.length > 0 && (
              <div className="tag-group">
                <h4 className="tag-group-title">正常标签</h4>
                <div className="tag-list">
                  {tagStats.normalTags.map(tag => (
                    <div key={tag} className="tag-item-wrapper">
                      <button
                        className={`tag-option normal ${filterState.selectedTags.includes(tag) ? 'selected' : ''} ${filterState.excludedTags.includes(tag) ? 'excluded' : ''}`}
                        onClick={() => handleTagToggle(tag)}
                        disabled={filterState.excludedTags.includes(tag)}
                      >
                        {tag} ({tagStats.tagCounts[tag]})
                      </button>
                      <button
                        className={`exclude-btn ${filterState.excludedTags.includes(tag) ? 'active' : ''}`}
                        onClick={() => handleExcludeTagToggle(tag)}
                        title={filterState.excludedTags.includes(tag) ? '取消排除' : '排除此标签'}
                      >
                        {filterState.excludedTags.includes(tag) ? '❌' : '🚫'}
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* 已选标签展示 */}
      {(filterState.selectedTags.length > 0 || filterState.excludedTags.length > 0) && (
        <div className="selected-tags">
          {filterState.selectedTags.length > 0 && (
            <>
              <span className="selected-label">包含标签:</span>
              {filterState.selectedTags.map(tag => (
                <span 
                  key={tag} 
                  className={`selected-tag ${tag.startsWith('err_') ? 'error' : 'normal'}`}
                >
                  {tag}
                  <button 
                    onClick={() => setFilterState(prev => ({ ...prev, selectedTags: prev.selectedTags.filter(t => t !== tag) }))}
                    className="remove-tag-btn"
                  >
                    ×
                  </button>
                </span>
              ))}
            </>
          )}
          
          {filterState.excludedTags.length > 0 && (
            <>
              <span className="selected-label">排除标签:</span>
              {filterState.excludedTags.map(tag => (
                <span 
                  key={tag} 
                  className={`selected-tag exclude ${tag.startsWith('err_') ? 'error' : 'normal'}`}
                >
                  🚫 {tag}
                  <button 
                    onClick={() => setFilterState(prev => ({ ...prev, excludedTags: prev.excludedTags.filter(t => t !== tag) }))}
                    className="remove-tag-btn"
                  >
                    ×
                  </button>
                </span>
              ))}
            </>
          )}
        </div>
      )}

      {/* 筛选结果 */}
      <div className="filter-result">
        显示 {filteredData.length} / {data.length} 个租户
        {isFiltered && (
          <span className="filtered-indicator">（已筛选）</span>
        )}
      </div>
    </div>
  )
}

export default AttributionTagFilter
