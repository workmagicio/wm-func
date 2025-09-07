import React from 'react'
import { Platform } from '../config/platforms'
import './PlatformSelector.css'

interface PlatformSelectorProps {
  platforms: Platform[]
  selectedPlatform: string
  onPlatformChange: (platformId: string) => void
}

const PlatformSelector: React.FC<PlatformSelectorProps> = ({
  platforms,
  selectedPlatform,
  onPlatformChange
}) => {
  return (
    <div className="platform-selector-container">
      <label className="platform-selector-label">📊 平台选择：</label>
      <div className="platform-selector-buttons">
        {platforms.map((platform) => (
          <button
            key={platform.id}
            className={`platform-button ${selectedPlatform === platform.id ? 'active' : ''}`}
            onClick={() => onPlatformChange(platform.id)}
            style={{
              '--platform-color': platform.color,
              '--platform-color-light': platform.color + '20',
              '--platform-color-dark': platform.color
            } as React.CSSProperties}
          >
            <span className="platform-icon">{platform.icon}</span>
            <span className="platform-name">{platform.displayName}</span>
          </button>
        ))}
      </div>
    </div>
  )
}

export default PlatformSelector
