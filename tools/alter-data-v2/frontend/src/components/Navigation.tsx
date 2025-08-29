import React from 'react'
import './Navigation.css'

interface NavigationProps {
  selectedPlatform: string
  onPlatformChange: (platform: string) => void
}

const Navigation: React.FC<NavigationProps> = ({ selectedPlatform, onPlatformChange }) => {
  const platforms = [
    { id: 'googleAds', name: 'Google Ads', icon: '🔍' },
  ]

  return (
    <nav className="navigation">
      <div className="nav-header">
        <h3>平台选择</h3>
      </div>
      <ul className="nav-list">
        {platforms.map((platform) => (
          <li key={platform.id} className="nav-item">
            <button
              className={`nav-button ${selectedPlatform === platform.id ? 'active' : ''}`}
              onClick={() => onPlatformChange(platform.id)}
            >
              <span className="nav-icon">{platform.icon}</span>
              <span className="nav-text">{platform.name}</span>
            </button>
          </li>
        ))}
      </ul>
      
      <div className="nav-footer">
        <div className="nav-info">
          <p className="nav-info-text">选择平台查看数据差异分析</p>
        </div>
      </div>
    </nav>
  )
}

export default Navigation
