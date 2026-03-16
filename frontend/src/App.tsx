import { useState, useEffect, useRef } from 'react'
import './index.css'

const API_BASE = import.meta.env.VITE_API_URL || ''
const WS_URL = import.meta.env.VITE_WS_URL || ''

function getWsUrl() {
  if (WS_URL) return WS_URL
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}/ws/alerts`
}

function getApiUrl() {
  if (API_BASE) return API_BASE
  const protocol = window.location.protocol
  return `${protocol}//${window.location.host}`
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

interface TamperAlert {
  detected_at: number
  tampered_block_id: number
  details: string
}

function App() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [alerts, setAlerts] = useState<TamperAlert[]>([])
  const [connected, setConnected] = useState(false)
  const [lastAlert, setLastAlert] = useState<TamperAlert | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const pollIntervalRef = useRef<number | null>(null)

  useEffect(() => {
    const wsUrl = getWsUrl()
    
    const connectWebSocket = () => {
      try {
        const ws = new WebSocket(wsUrl)
        
        ws.onopen = () => {
          console.log('WebSocket connected')
          setConnected(true)
        }
        
        ws.onmessage = (event) => {
          try {
            const alert = JSON.parse(event.data) as TamperAlert
            setAlerts(prev => [...prev.slice(-9), alert])
            setLastAlert(alert)
          } catch (err) {
            console.error('Failed to parse alert:', err)
          }
        }
        
        ws.onclose = () => {
          console.log('WebSocket disconnected')
          setConnected(false)
          setTimeout(connectWebSocket, 3000)
        }
        
        ws.onerror = (err) => {
          console.error('WebSocket error:', err)
        }
        
        wsRef.current = ws
      } catch (err) {
        console.error('WebSocket connection failed:', err)
        setTimeout(connectWebSocket, 3000)
      }
    }

    connectWebSocket()

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [])

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
      }
    }

    fetchLogs()
    pollIntervalRef.current = window.setInterval(fetchLogs, 2000)

    return () => {
      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current)
      }
    }
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

  return (
    <div className={`min-h-screen ${lastAlert ? 'tamper-flash bg-red-900/20' : 'bg-siem-dark'}`}>
      <div className="container mx-auto p-6">
        <header className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-white">SentinelChain</h1>
              <p className="text-gray-400">SIEM Blockchain Control Room</p>
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className={`w-3 h-3 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`}></div>
                <span className="text-sm text-gray-400">
                  {connected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              <div className="px-4 py-2 bg-siem-card rounded-lg border border-siem-border">
                <span className="text-2xl font-mono text-white">{logs.length}</span>
                <span className="text-gray-400 ml-2">Blocks</span>
              </div>
            </div>
          </div>
        </header>

        {lastAlert && (
          <div className="mb-6 p-4 bg-red-900/50 border border-red-600 rounded-lg animate-pulse">
            <div className="flex items-center gap-3">
              <span className="text-4xl">🚨</span>
              <div>
                <h2 className="text-xl font-bold text-red-400">TAMPER DETECTED</h2>
                <p className="text-gray-300">Block ID: {lastAlert.tampered_block_id}</p>
                <p className="text-sm text-gray-400">{lastAlert.details}</p>
                <p className="text-xs text-gray-500 mt-1">
                  Detected at: {new Date(lastAlert.detected_at / 1e6).toLocaleTimeString()}
                </p>
              </div>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2">
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

          <div>
            <div className="bg-siem-card rounded-lg border border-siem-border">
              <div className="p-4 border-b border-siem-border">
                <h2 className="text-lg font-semibold text-white">Alert History</h2>
              </div>
              <div className="max-h-[400px] overflow-y-auto p-4 space-y-3">
                {alerts.length === 0 ? (
                  <p className="text-gray-500 text-sm">No alerts detected</p>
                ) : (
                  alerts.slice().reverse().map((alert, idx) => (
                    <div key={idx} className="p-3 bg-red-900/20 border border-red-800 rounded-lg">
                      <p className="text-red-400 font-semibold text-sm">
                        Block #{alert.tampered_block_id}
                      </p>
                      <p className="text-gray-400 text-xs mt-1 truncate">
                        {alert.details}
                      </p>
                      <p className="text-gray-500 text-xs mt-1">
                        {new Date(alert.detected_at / 1e6).toLocaleTimeString()}
                      </p>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default App
