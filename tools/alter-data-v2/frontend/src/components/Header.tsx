import React from 'react'
import './Header.css'

const Header: React.FC = () => {
  return (
    <header className="header">
      <div className="header-content">
        <h1 className="header-title">数据差异分析系统 V2</h1>
        <div className="header-subtitle">
          实时监控不同平台的数据差异情况
        </div>
      </div>
    </header>
  )
}

export default Header
