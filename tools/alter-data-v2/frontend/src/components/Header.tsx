import React from 'react'
import { Link, useLocation } from 'react-router-dom'
import './Header.css'

const Header: React.FC = () => {
  const location = useLocation()

  return (
    <header className="header">
      <div className="header-container">
        <div className="header-left">
          <h1 className="header-title">数据分析平台</h1>
        </div>
        <nav className="header-nav">
          <Link 
            to="/dashboard" 
            className={`nav-link ${location.pathname === '/' || location.pathname === '/dashboard' ? 'active' : ''}`}
          >
            📊 数据差异分析
          </Link>
          <Link 
            to="/attribution" 
            className={`nav-link ${location.pathname === '/attribution' ? 'active' : ''}`}
          >
            🎯 归因数据分析
          </Link>
        </nav>
      </div>
    </header>
  )
}

export default Header
