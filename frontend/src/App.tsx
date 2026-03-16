import { useState, useEffect } from 'react'
import './index.css'

const API_BASE = import.meta.env.VITE_API_URL || ''

function getApiUrl() {
  if (API_BASE) return API_BASE
  return window.location.origin
}

interface LogEntry {
  id: number
  timestamp: number
  source_ip: string
  event_type: string
  severity: string
  message: string
  hash: string
}

function App() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchLogs = async () => {
      try {
        const apiBase = getApiUrl()
        const res = await fetch(`${apiBase}/api/logs`)
        if (res.ok) {
          const data = await res.json()
          setLogs(data)
        }
      } catch (err) {
        console.error('Failed to fetch logs:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchLogs()
    const interval = setInterval(fetchLogs, 2000)
    return () => clearInterval(interval)
  }, [])

  const getSeverityColor = (severity: string) => {
    switch (severity.toUpperCase()) {
      case 'CRITICAL': return 'text-red-400 bg-red-900/30'
      case 'ERROR': return 'text-orange-400 bg-orange-900/30'
      case 'WARNING': return 'text-yellow-400 bg-yellow-900/30'
      case 'INFO': return 'text-blue-400 bg-blue-900/30'
      default: return 'text-gray-400 bg-gray-900/30'
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-siem-dark flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-siem-dark">
      <div className="container mx-auto p-6">
        <header className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-white">SentinelChain</h1>
              <p className="text-gray-400">SIEM Blockchain Control Room</p>
            </div>
            <div className="flex items-center gap-4">
              <div className="px-4 py-2 bg-siem-card rounded-lg border border-siem-border">
                <span className="text-2xl font-mono text-white">{logs.length}</span>
                <span className="text-gray-400 ml-2">Blocks</span>
              </div>
            </div>
          </div>
        </header>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-3">
            <div className="bg-siem-card rounded-lg border border-siem-border overflow-hidden">
              <div className="p-4 border-b border-siem-border">
                <h2 className="text-lg font-semibold text-white">Blockchain Ledger</h2>
              </div>
              <div className="max-h-[500px] overflow-y-auto">
                <table className="w-full">
                  <thead className="bg-gray-900/50 sticky top-0">
                    <tr>
                      <th className="p-3 text-left text-xs font-medium text-gray-400 uppercase">ID</th>
                      <th className="p-3 text-left text-xs font-medium text-gray-400 uppercase">Source</th>
                      <th className="p-3 text-left text-xs font-medium text-gray-400 uppercase">Event</th>
                      <th className="p-3 text-left text-xs font-medium text-gray-400 uppercase">Severity</th>
                      <th className="p-3 text-left text-xs font-medium text-gray-400 uppercase">Hash</th>
                    </tr>
                  </thead>
                  <tbody>
                    {logs.slice().reverse().map((log) => (
                      <tr key={log.id} className="border-t border-siem-border hover:bg-gray-800/30">
                        <td className="p-3 font-mono text-sm text-gray-300">{log.id}</td>
                        <td className="p-3 text-sm text-gray-300">{log.source_ip}</td>
                        <td className="p-3 text-sm text-gray-300">{log.event_type}</td>
                        <td className="p-3">
                          <span className={`px-2 py-1 rounded text-xs ${getSeverityColor(log.severity)}`}>
                            {log.severity}
                          </span>
                        </td>
                        <td className="p-3 font-mono text-xs text-gray-500 truncate max-w-[150px]">
                          {log.hash}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
