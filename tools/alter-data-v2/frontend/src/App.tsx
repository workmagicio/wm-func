import React, { useState } from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import './App.css'
import Header from './components/Header'
import Dashboard from './pages/Dashboard'

function App() {
  const [selectedPlatform, setSelectedPlatform] = useState('googleAds')

  return (
    <Router>
      <div className="App">
        <Header />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard platform={selectedPlatform} />} />
            <Route path="/dashboard" element={<Dashboard platform={selectedPlatform} />} />
          </Routes>
        </main>
      </div>
    </Router>
  )
}

export default App
