import { useCallback, useEffect, useState } from 'react'
import './index.css'

const API_BASE = import.meta.env.VITE_API_URL || ''

function getApiUrl() {
  if (API_BASE) return API_BASE
  return window.location.origin
}

interface LogEntry {
  id: number
  timestamp: number
  device_id: string
  device_name: string
  source_ip: string
  event_type: string
  severity: string
  message: string
  hash: string
}

type SimulationMode = 'success' | 'failed' | 'bruteforce'

const DEVICE_PRESETS = [
  { id: 'soc-ws-01', label: 'SOC-Workstation-01', ip: '192.168.1.10' },
  { id: 'soc-ws-02', label: 'SOC-Workstation-02', ip: '192.168.1.11' },
  { id: 'finance-laptop', label: 'Finance-Laptop', ip: '192.168.1.20' },
  { id: 'eng-desktop', label: 'Engineering-Desktop', ip: '192.168.1.30' },
  { id: 'remote-tablet', label: 'Remote-Tablet', ip: '192.168.1.40' },
]

function App() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [username, setUsername] = useState('analyst')
  const [password, setPassword] = useState('')
  const [deviceId, setDeviceId] = useState(DEVICE_PRESETS[0].id)
  const [deviceName, setDeviceName] = useState(DEVICE_PRESETS[0].label)
  const [sourceIp, setSourceIp] = useState(DEVICE_PRESETS[0].ip)
  const [mode, setMode] = useState<SimulationMode>('failed')
  const [attempts, setAttempts] = useState(5)
  const [submitting, setSubmitting] = useState(false)
  const [submitStatus, setSubmitStatus] = useState<string>('')

  const fetchLogs = useCallback(async () => {
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
  }, [])

  useEffect(() => {
    fetchLogs()
    const interval = setInterval(fetchLogs, 2000)
    return () => clearInterval(interval)
  }, [fetchLogs])

  const submitSimulation = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setSubmitting(true)
    setSubmitStatus('')

    const apiBase = getApiUrl()

    let eventType = 'AUTH_FAILURE'
    let severity = 'WARNING'
    let message = `User ${username} failed login from ${deviceName}`

    if (mode === 'success') {
      eventType = 'LOGIN_SUCCESS'
      severity = 'INFO'
      message = `User ${username} logged in successfully from ${deviceName}`
    }

    if (mode === 'bruteforce') {
      eventType = 'BRUTE_FORCE_ATTEMPT'
      severity = 'CRITICAL'
      message = `${attempts} failed login attempts for ${username} from ${deviceName}`
    }

    try {
      const res = await fetch(`${apiBase}/api/log`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          timestamp: Date.now(),
          device_id: deviceId,
          device_name: deviceName,
          source_ip: sourceIp,
          event_type: eventType,
          severity,
          message,
        }),
      })

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`)
      }

      setPassword('')
      setSubmitStatus('Simulation event sent to blockchain log.')
      await fetchLogs()
    } catch (err) {
      console.error('Failed to submit simulation log:', err)
      setSubmitStatus('Failed to send event. Confirm backend is running on :8080.')
    } finally {
      setSubmitting(false)
    }
  }

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
          <div className="lg:col-span-1">
            <div className="bg-siem-card rounded-lg border border-siem-border p-4">
              <h2 className="text-lg font-semibold text-white mb-3">Login Attack Simulator</h2>
              <p className="text-sm text-gray-400 mb-4">
                Use this form to emulate login events from different devices in the same LAN.
              </p>

              <form onSubmit={submitSimulation} className="space-y-3">
                <div>
                  <label className="block text-xs text-gray-400 mb-1">Device Preset</label>
                  <select
                    className="w-full bg-gray-900 text-gray-200 border border-siem-border rounded px-3 py-2"
                    value={`${deviceId}|${deviceName}|${sourceIp}`}
                    onChange={(e) => {
                      const [presetId, name, ip] = e.target.value.split('|')
                      setDeviceId(presetId)
                      setDeviceName(name)
                      setSourceIp(ip)
                    }}
                  >
                    {DEVICE_PRESETS.map((device) => (
                      <option key={device.id} value={`${device.id}|${device.label}|${device.ip}`}>
                        {device.label} ({device.ip})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-xs text-gray-400 mb-1">Source IP</label>
                  <input
                    className="w-full bg-gray-900 text-gray-200 border border-siem-border rounded px-3 py-2"
                    value={sourceIp}
                    onChange={(e) => setSourceIp(e.target.value)}
                    placeholder="192.168.1.10"
                    required
                  />
                </div>

                <div>
                  <label className="block text-xs text-gray-400 mb-1">Username</label>
                  <input
                    className="w-full bg-gray-900 text-gray-200 border border-siem-border rounded px-3 py-2"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    placeholder="analyst"
                    required
                  />
                </div>

                <div>
                  <label className="block text-xs text-gray-400 mb-1">Password</label>
                  <input
                    type="password"
                    className="w-full bg-gray-900 text-gray-200 border border-siem-border rounded px-3 py-2"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••"
                    required
                  />
                </div>

                <div>
                  <label className="block text-xs text-gray-400 mb-1">Simulation Mode</label>
                  <select
                    className="w-full bg-gray-900 text-gray-200 border border-siem-border rounded px-3 py-2"
                    value={mode}
                    onChange={(e) => setMode(e.target.value as SimulationMode)}
                  >
                    <option value="failed">Failed Login (Warning)</option>
                    <option value="success">Successful Login (Info)</option>
                    <option value="bruteforce">Brute Force Burst (Critical)</option>
                  </select>
                </div>

                {mode === 'bruteforce' && (
                  <div>
                    <label className="block text-xs text-gray-400 mb-1">Attempts</label>
                    <input
                      type="number"
                      min={2}
                      max={100}
                      className="w-full bg-gray-900 text-gray-200 border border-siem-border rounded px-3 py-2"
                      value={attempts}
                      onChange={(e) => setAttempts(Number(e.target.value))}
                    />
                  </div>
                )}

                <button
                  type="submit"
                  disabled={submitting}
                  className="w-full bg-red-700 hover:bg-red-600 disabled:bg-gray-700 text-white font-semibold rounded px-4 py-2"
                >
                  {submitting ? 'Sending...' : 'Send Simulation Event'}
                </button>
              </form>

              {submitStatus && (
                <p className="mt-3 text-sm text-cyan-300">{submitStatus}</p>
              )}
            </div>
          </div>

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
                        <th className="p-3 text-left text-xs font-medium text-gray-400 uppercase">Device</th>
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
                        <td className="p-3 text-sm text-gray-300">
                          <div className="font-medium">{log.device_name}</div>
                          <div className="text-xs text-gray-500 font-mono">{log.device_id}</div>
                        </td>
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
