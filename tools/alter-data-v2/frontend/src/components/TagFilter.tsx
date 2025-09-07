import React, { useState, useEffect, useMemo } from 'react'
import './TagFilter.css'

interface TenantData {
  tenant_id: number
  tags: string[]
  last_30_day_diff: number
  register_time: string
  date_sequence: any[]
}

interface TagFilterProps {
  newTenants: TenantData[]
  oldTenants: TenantData[]
  onFilterChange: (filteredNew: TenantData[], filteredOld: TenantData[]) => void
}

type FilterMode = 'all' | 'error' | 'normal' | 'custom'

interface FilterState {
  mode: FilterMode
  selectedTags: string[]
  searchText: string
}

const TagFilter: React.FC<TagFilterProps> = ({ newTenants, oldTenants, onFilterChange }) => {
  const [filterState, setFilterState] = useState<FilterState>({
    mode: 'all',
    selectedTags: [],
    searchText: ''
  })

  // 计算所有标签和统计信息
  const tagStats = useMemo(() => {
    const allTenants = [...newTenants, ...oldTenants]
    const tagCounts: { [tag: string]: number } = {}
    
    let errorCount = 0
    let normalCount = 0
    
    allTenants.forEach(tenant => {
      const tags = tenant.tags || []
      const hasErrorTag = tags.some(tag => tag.startsWith('err_'))
      const hasNormalTag = tags.some(tag => !tag.startsWith('err_'))
      
      // 如果有正常标签，就不统计到错误中，但仍然可以筛选错误标签
      if (hasErrorTag && !hasNormalTag) {
        errorCount++
      } else {
        normalCount++
      }
      
      tags.forEach(tag => {
        tagCounts[tag] = (tagCounts[tag] || 0) + 1
      })
    })
    
    // 分类标签
    const errorTags = Object.keys(tagCounts).filter(tag => tag.startsWith('err_')).sort()
    const normalTags = Object.keys(tagCounts).filter(tag => !tag.startsWith('err_')).sort()
    
    return {
      tagCounts,
      errorTags,
      normalTags,
      errorCount,
      normalCount,
      totalCount: allTenants.length
    }
  }, [newTenants, oldTenants])

  // 筛选逻辑
  const applyFilters = (tenants: TenantData[], filterState: FilterState): TenantData[] => {
    let filtered = tenants

    // 按模式筛选
    switch (filterState.mode) {
      case 'error':
        filtered = filtered.filter(tenant => {
          const tags = tenant.tags || []
          return tags.some(tag => tag.startsWith('err_'))
        })
        break
      case 'normal':
        filtered = filtered.filter(tenant => {
          const tags = tenant.tags || []
          return !tags.some(tag => tag.startsWith('err_'))
        })
        break
      case 'custom':
        if (filterState.selectedTags.length > 0) {
          filtered = filtered.filter(tenant => {
            const tags = tenant.tags || []
            return filterState.selectedTags.every(selectedTag => 
              tags.includes(selectedTag)
            )
          })
        }
        break
      default:
        // 'all' - 不过滤
        break
    }

    // 搜索筛选
    if (filterState.searchText) {
      const searchLower = filterState.searchText.toLowerCase()
      filtered = filtered.filter(tenant => {
        const tags = tenant.tags || []
        return (
          tenant.tenant_id.toString().includes(searchLower) ||
          tags.some(tag => tag.toLowerCase().includes(searchLower))
        )
      })
    }

    return filtered
  }

  // 当筛选条件变化时，应用筛选
  useEffect(() => {
    const filteredNew = applyFilters(newTenants, filterState)
    const filteredOld = applyFilters(oldTenants, filterState)
    onFilterChange(filteredNew, filteredOld)
  }, [newTenants, oldTenants, filterState, onFilterChange])

  const handleModeChange = (mode: FilterMode) => {
    setFilterState(prev => ({
      ...prev,
      mode,
      selectedTags: mode === 'custom' ? prev.selectedTags : []
    }))
  }

  const handleTagToggle = (tag: string) => {
    setFilterState(prev => {
      const newSelectedTags = prev.selectedTags.includes(tag)
        ? prev.selectedTags.filter(t => t !== tag)
        : [...prev.selectedTags, tag]
      
      return {
        ...prev,
        mode: 'custom',
        selectedTags: newSelectedTags
      }
    })
  }

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFilterState(prev => ({
      ...prev,
      searchText: e.target.value
    }))
  }

  const clearFilters = () => {
    setFilterState({
      mode: 'all',
      selectedTags: [],
      searchText: ''
    })
  }

  const removeSelectedTag = (tag: string) => {
    setFilterState(prev => ({
      ...prev,
      selectedTags: prev.selectedTags.filter(t => t !== tag)
    }))
  }

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
        {(filterState.mode !== 'all' || filterState.selectedTags.length > 0 || filterState.searchText) && (
          <button onClick={clearFilters} className="clear-btn">
            清空筛选
          </button>
        )}
      </div>

      {/* 标签选择器 */}
      {(tagStats.errorTags.length > 0 || tagStats.normalTags.length > 0) && (
        <div className="tag-selector">
          <div className="tag-section">
            {tagStats.errorTags.length > 0 && (
              <div className="tag-group">
                <h4 className="tag-group-title">异常标签</h4>
                <div className="tag-list">
                  {tagStats.errorTags.map(tag => (
                    <button
                      key={tag}
                      className={`tag-option error ${filterState.selectedTags.includes(tag) ? 'selected' : ''}`}
                      onClick={() => handleTagToggle(tag)}
                    >
                      {tag} ({tagStats.tagCounts[tag]})
                    </button>
                  ))}
                </div>
              </div>
            )}
            
            {tagStats.normalTags.length > 0 && (
              <div className="tag-group">
                <h4 className="tag-group-title">正常标签</h4>
                <div className="tag-list">
                  {tagStats.normalTags.map(tag => (
                    <button
                      key={tag}
                      className={`tag-option normal ${filterState.selectedTags.includes(tag) ? 'selected' : ''}`}
                      onClick={() => handleTagToggle(tag)}
                    >
                      {tag} ({tagStats.tagCounts[tag]})
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* 已选标签展示 */}
      {filterState.selectedTags.length > 0 && (
        <div className="selected-tags">
          <span className="selected-label">已选标签:</span>
          {filterState.selectedTags.map(tag => (
            <span 
              key={tag} 
              className={`selected-tag ${tag.startsWith('err_') ? 'error' : 'normal'}`}
            >
              {tag}
              <button 
                onClick={() => removeSelectedTag(tag)}
                className="remove-tag-btn"
              >
                ×
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

export default TagFilter
