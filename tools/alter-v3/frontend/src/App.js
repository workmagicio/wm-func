import React, { useState, useEffect } from 'react';
import './App.css';
import { getApiUrl, API_CONFIG } from './config';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { prism } from 'react-syntax-highlighter/dist/esm/styles/prism';
import ChartView from './ChartView';

function App() {
  const [configData, setConfigData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedMenu, setSelectedMenu] = useState(() => {
    // 从localStorage读取上次选择的菜单，如果没有则默认为'config'
    return localStorage.getItem('selectedMenu') || 'config';
  });
  const [showAddForm, setShowAddForm] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editingConfigName, setEditingConfigName] = useState('');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [copyMessage, setCopyMessage] = useState('');
  const [newConfig, setNewConfig] = useState({
    name: '',
    base_platform: '',
    api_data_query: '',
    wm_data_query: '',
    icon: '',
    total_data_count: 90,
    tenants: 'all'
  });

  useEffect(() => {
    fetchConfigData();
  }, []);

  // 处理菜单切换并保存到localStorage
  const handleMenuChange = (menuName) => {
    setSelectedMenu(menuName);
    localStorage.setItem('selectedMenu', menuName);
  };

  const fetchConfigData = async () => {
    try {
      setLoading(true);
      const response = await fetch(getApiUrl(API_CONFIG.ENDPOINTS.CONFIG));
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      setConfigData(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleAddConfig = async () => {
    try {
      // Add or update config (backend handles overwrite automatically)
      const response = await fetch(getApiUrl(API_CONFIG.ENDPOINTS.CONFIG), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newConfig)
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      // Reset form and hide it
      resetForm();
      
      // Refresh data
      fetchConfigData();
    } catch (err) {
      setError(err.message);
    }
  };

  const resetForm = () => {
    setNewConfig({
      name: '',
      base_platform: '',
      api_data_query: '',
      wm_data_query: '',
      icon: '',
      total_data_count: 90,
      tenants: 'all'
    });
    setShowAddForm(false);
    setIsEditing(false);
    setEditingConfigName('');
  };

  const handleEditConfig = (config) => {
    setNewConfig({
      name: config.name,
      base_platform: config.base_platform,
      api_data_query: config.api_data_query,
      wm_data_query: config.wm_data_query,
      icon: config.icon,
      total_data_count: config.total_data_count,
      tenants: config.tenants || 'all'
    });
    setIsEditing(true);
    setEditingConfigName(config.name);
    setShowAddForm(true);
  };

  const copyToClipboard = async (text) => {
    try {
      await navigator.clipboard.writeText(text);
      showCopyMessage('SQL copied to clipboard!');
    } catch (err) {
      console.error('Failed to copy: ', err);
      // 备用方案：使用传统的复制方法
      const textArea = document.createElement('textarea');
      textArea.value = text;
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      try {
        document.execCommand('copy');
        showCopyMessage('SQL copied to clipboard!');
      } catch (fallbackErr) {
        console.error('Fallback copy failed: ', fallbackErr);
        showCopyMessage('Failed to copy SQL');
      }
      document.body.removeChild(textArea);
    }
  };

  const showCopyMessage = (message) => {
    setCopyMessage(message);
    setTimeout(() => {
      setCopyMessage('');
    }, 2000); // 2秒后消失
  };

  const handleDeleteConfig = async (name) => {
    if (!window.confirm(`Are you sure you want to delete config "${name}"?`)) {
      return;
    }
    
    try {
      const response = await fetch(`${getApiUrl(API_CONFIG.ENDPOINTS.CONFIG)}/${name}`, {
        method: 'DELETE'
      });
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      // Refresh data
      fetchConfigData();
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="App">
      <div className={`sidebar ${sidebarCollapsed ? 'collapsed' : ''}`}>
        <div className="sidebar-header">
          <h2 className="nav-title">Navigation</h2>
          <button 
            className="collapse-btn"
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            title={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {sidebarCollapsed ? '▶' : '◀'}
          </button>
        </div>
        
        {/* Config menu item */}
        <div 
          className={`nav-item ${selectedMenu === 'config' ? 'active' : ''}`}
          onClick={() => handleMenuChange('config')}
          title="Config"
        >
          <span className="icon">⚙️</span>
          <span className="name">Config</span>
        </div>
        
        {loading && !sidebarCollapsed && <div className="loading">Loading...</div>}
        {error && !sidebarCollapsed && <div className="error">Error: {error}</div>}
        
        {configData.map((item, index) => (
          <div 
            key={index} 
            className={`nav-item ${selectedMenu === item.name ? 'active' : ''}`}
            onClick={() => handleMenuChange(item.name)}
            title={item.name}
          >
            <span className="icon">{item.icon}</span>
            <span className="name">{item.name}</span>
          </div>
        ))}
      </div>
      <div className={`main-content ${sidebarCollapsed ? 'sidebar-collapsed' : ''}`}>
        {selectedMenu === 'config' && (
          <div>
            {copyMessage && (
              <div className="copy-message">
                {copyMessage}
              </div>
            )}
            
            <div className="config-header">
              <h3>Configuration</h3>
              <button 
                className="btn btn-primary"
                onClick={() => {
                  if (showAddForm) {
                    resetForm();
                  } else {
                    setShowAddForm(true);
                  }
                }}
              >
                {showAddForm ? 'Cancel' : 'Add New Config'}
              </button>
            </div>
            
            {showAddForm && (
              <div className="add-config-form">
                <h4>{isEditing ? 'Edit Configuration' : 'Add New Configuration'}</h4>
                <div className="form-grid">
                  <input
                    type="text"
                    placeholder="Name"
                    value={newConfig.name}
                    onChange={(e) => setNewConfig({...newConfig, name: e.target.value})}
                    disabled={isEditing}
                  />
                  <input
                    type="text"
                    placeholder="Base Platform"
                    value={newConfig.base_platform}
                    onChange={(e) => setNewConfig({...newConfig, base_platform: e.target.value})}
                  />
                  <input
                    type="text"
                    placeholder="Icon (emoji)"
                    value={newConfig.icon}
                    onChange={(e) => setNewConfig({...newConfig, icon: e.target.value})}
                  />
                  <input
                    type="number"
                    placeholder="Total Data Count"
                    value={newConfig.total_data_count}
                    onChange={(e) => setNewConfig({...newConfig, total_data_count: parseInt(e.target.value)})}
                  />
                  <input
                    type="text"
                    placeholder="Tenants (comma-separated IDs or 'all')"
                    value={newConfig.tenants}
                    onChange={(e) => setNewConfig({...newConfig, tenants: e.target.value})}
                  />
                </div>
                <textarea
                  placeholder="API Data Query (SQL)"
                  value={newConfig.api_data_query}
                  onChange={(e) => setNewConfig({...newConfig, api_data_query: e.target.value})}
                />
                <textarea
                  placeholder="WM Data Query (SQL)"
                  value={newConfig.wm_data_query}
                  onChange={(e) => setNewConfig({...newConfig, wm_data_query: e.target.value})}
                />
                <div className="form-buttons">
                  <button className="btn btn-success" onClick={handleAddConfig}>
                    {isEditing ? 'Update Configuration' : 'Add Configuration'}
                  </button>
                  <button className="btn btn-secondary" onClick={resetForm}>
                    Cancel
                  </button>
                </div>
              </div>
            )}
            
            {loading && <div className="loading">Loading...</div>}
            {error && <div className="error">Error: {error}</div>}
            {!loading && !error && configData.length > 0 && (
              <table className="config-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Icon</th>
                    <th>Base Platform</th>
                    <th>Tenants</th>
                    <th>API Data Query</th>
                    <th>WM Data Query</th>
                    <th>Total Data Count</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {configData.map((item, index) => (
                    <tr key={index}>
                      <td>{item.name}</td>
                      <td>{item.icon}</td>
                      <td>{item.base_platform}</td>
                      <td>{item.tenants || 'all'}</td>
                      <td className="query-cell">
                        {item.api_data_query ? (
                          <div 
                            className="query-container"
                            onClick={() => copyToClipboard(item.api_data_query)}
                            title="Click to copy SQL"
                          >
                            <SyntaxHighlighter
                              language="sql"
                              style={prism}
                              customStyle={{
                                margin: 0,
                                padding: '8px',
                                fontSize: '12px',
                                maxHeight: '200px',
                                overflow: 'auto',
                                whiteSpace: 'pre-wrap',
                                wordWrap: 'break-word'
                              }}
                            >
                              {item.api_data_query}
                            </SyntaxHighlighter>
                          </div>
                        ) : (
                          '-'
                        )}
                      </td>
                      <td className="query-cell">
                        {item.wm_data_query ? (
                          <div 
                            className="query-container"
                            onClick={() => copyToClipboard(item.wm_data_query)}
                            title="Click to copy SQL"
                          >
                            <SyntaxHighlighter
                              language="sql"
                              style={prism}
                              customStyle={{
                                margin: 0,
                                padding: '8px',
                                fontSize: '12px',
                                maxHeight: '200px',
                                overflow: 'auto',
                                whiteSpace: 'pre-wrap',
                                wordWrap: 'break-word'
                              }}
                            >
                              {item.wm_data_query}
                            </SyntaxHighlighter>
                          </div>
                        ) : (
                          '-'
                        )}
                      </td>
                      <td>{item.total_data_count}</td>
                      <td>
                        <div className="action-buttons">
                          <button 
                            className="btn btn-warning btn-small"
                            onClick={() => handleEditConfig(item)}
                          >
                            Edit
                          </button>
                          <button 
                            className="btn btn-danger btn-small"
                            onClick={() => handleDeleteConfig(item.name)}
                          >
                            Delete
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
            {!loading && !error && configData.length === 0 && (
              <div>暂无配置数据</div>
            )}
          </div>
        )}
        {selectedMenu !== 'config' && (
          <div>
            <ChartView selectedTenant={selectedMenu} />
          </div>
        )}
      </div>
    </div>
  );
}

export default App;
