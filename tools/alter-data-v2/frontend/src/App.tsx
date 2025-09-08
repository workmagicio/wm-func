import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import './App.css'
import Header from './components/Header'
import Dashboard from './pages/Dashboard'
import AttributionAnalysis from './pages/AttributionAnalysis'

function App() {
  return (
    <Router>
      <div className="App">
        <Header />
        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/attribution" element={<AttributionAnalysis />} />
          </Routes>
        </main>
      </div>
    </Router>
  )
}

export default App
