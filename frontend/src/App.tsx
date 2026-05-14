import React, { useEffect, useState, useRef } from 'react';
import { 
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area 
} from 'recharts';
import { Activity, Users, DollarSign, TrendingUp } from 'lucide-react';
import { motion } from 'framer-motion';

interface Metric {
  window_start: string;
  window_end: string;
  event_count: number;
  total_value: number;
  avg_value: number;
}

const Dashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [current, setCurrent] = useState<Metric | null>(null);
  const [status, setStatus] = useState<'connecting' | 'connected' | 'disconnected'>('connecting');
  const [logs, setLogs] = useState<{ id: number; message: string; time: string }[]>([]);
  const ws = useRef<WebSocket | null>(null);
  const logContainerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    connect();
    return () => ws.current?.close();
  }, []);

  const connect = () => {
    ws.current = new WebSocket('ws://localhost:8080/ws');
    
    ws.current.onopen = () => setStatus('connected');
    ws.current.onclose = () => {
      setStatus('disconnected');
      setTimeout(connect, 3000); // Reconnect logic
    };
    
    ws.current.onmessage = (event) => {
      const data: Metric = JSON.parse(event.data);
      const time = new Date().toLocaleTimeString();
      setCurrent(data);
      setMetrics(prev => [...prev.slice(-20), { ...data, timestamp: time }]);
      setLogs(prev => [
        { id: Date.now(), message: `INGEST: ${data.event_count} events | AVG: ${data.avg_value.toFixed(2)} | VOL: ${data.total_value.toFixed(0)}`, time },
        ...prev.slice(0, 49)
      ]);
    };
  };

  return (
    <div className="min-h-screen bg-[#050505] text-white font-sans selection:bg-white/10">
      <div className="max-w-6xl mx-auto p-8 md:p-16">
      {/* Header */}
      <header className="flex justify-between items-center mb-12">
        <div>
          <h1 className="text-2xl font-medium tracking-tight text-white">
            Real-time Monitoring
          </h1>
          <p className="text-gray-500 text-sm mt-1.5 font-light">Engine telemetry and sliding-window analysis</p>
        </div>
        <div className="flex items-center gap-3 bg-[#111] px-4 py-1.5 rounded-full border border-white/10">
          <div className={`w-1.5 h-1.5 rounded-full ${status === 'connected' ? 'bg-white shadow-[0_0_8px_rgba(255,255,255,0.4)]' : 'bg-gray-600'}`} />
          <span className="text-[10px] font-semibold uppercase tracking-[0.2em] text-gray-400">{status}</span>
        </div>
      </header>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
        <StatCard 
          icon={<Activity size={18} className="text-white" />} 
          label="Event Rate" 
          value={`${current?.event_count || 0}`} 
          unit="eps"
        />
        <StatCard 
          icon={<DollarSign size={18} className="text-white" />} 
          label="Volume" 
          value={`${(current?.total_value || 0).toLocaleString()}`} 
          unit="usd"
        />
        <StatCard 
          icon={<TrendingUp size={18} className="text-white" />} 
          label="Average" 
          value={`${(current?.avg_value || 0).toFixed(2)}`} 
          unit="per"
        />
        <StatCard 
          icon={<Users size={18} className="text-white" />} 
          label="Window" 
          value="30" 
          unit="sec"
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <ChartContainer title="Event Frequency">
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={metrics}>
              <defs>
                <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#fff" stopOpacity={0.1}/>
                  <stop offset="95%" stopColor="#fff" stopOpacity={0}/>
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#ffffff05" vertical={false} />
              <XAxis dataKey="timestamp" stroke="#333" fontSize={10} tickLine={false} axisLine={false} />
              <YAxis stroke="#333" fontSize={10} tickLine={false} axisLine={false} />
              <Tooltip 
                contentStyle={{ backgroundColor: '#000', border: '1px solid #222', borderRadius: '4px' }}
                itemStyle={{ color: '#fff', fontSize: '12px' }}
              />
              <Area type="monotone" dataKey="event_count" stroke="#fff" fillOpacity={1} fill="url(#colorCount)" strokeWidth={1} />
            </AreaChart>
          </ResponsiveContainer>
        </ChartContainer>

        <ChartContainer title="Monetary Volume">
          <ResponsiveContainer width="100%" height={300}>
            <LineChart data={metrics}>
              <CartesianGrid strokeDasharray="3 3" stroke="#ffffff05" vertical={false} />
              <XAxis dataKey="timestamp" stroke="#333" fontSize={10} tickLine={false} axisLine={false} />
              <YAxis stroke="#333" fontSize={10} tickLine={false} axisLine={false} />
              <Tooltip 
                contentStyle={{ backgroundColor: '#000', border: '1px solid #222', borderRadius: '4px' }}
                itemStyle={{ color: '#fff', fontSize: '12px' }}
              />
              <Line type="monotone" dataKey="total_value" stroke="#fff" strokeWidth={1} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </ChartContainer>
      </div>
      
      {/* Logging Section */}
      <div className="mt-20">
        <h3 className="text-xs font-light mb-8 text-gray-500 uppercase tracking-[0.4em] flex items-center gap-3">
          <div className="w-8 h-[1px] bg-white/10" />
          System Logs
        </h3>
        <div 
          className="bg-[#0a0a0a] border border-white/5 p-8 rounded-sm h-64 overflow-y-auto font-mono text-[11px] leading-relaxed scrollbar-hide"
          ref={logContainerRef}
        >
          {logs.map((log) => (
            <div key={log.id} className="flex gap-4 py-1 border-b border-white/[0.02] last:border-0 hover:bg-white/[0.01] transition-colors">
              <span className="text-gray-600 shrink-0">{log.time}</span>
              <span className="text-gray-400 shrink-0">DEBUG</span>
              <span className="text-gray-300 break-all">{log.message}</span>
            </div>
          ))}
          {logs.length === 0 && (
            <div className="text-gray-700 italic">Waiting for telemetry stream...</div>
          )}
        </div>
      </div>
    </div>
  </div>
);
};

const StatCard = ({ icon, label, value, unit }: any) => (
  <motion.div 
    initial={{ opacity: 0 }}
    whileInView={{ opacity: 1 }}
    viewport={{ once: true }}
    className="bg-transparent border-b border-white/5 pb-8 flex flex-col justify-between hover:border-white/20 transition-all duration-500 group"
  >
    <div className="flex items-center gap-3 mb-6">
      <div className="p-2 bg-white/5 rounded-md grayscale opacity-50 group-hover:opacity-100 transition-opacity">{icon}</div>
      <span className="text-gray-500 text-xs font-light uppercase tracking-widest">{label}</span>
    </div>
    <div className="flex items-baseline gap-2">
      <span className="text-4xl font-light tracking-tighter">{value}</span>
      <span className="text-[10px] text-gray-600 font-medium uppercase tracking-[0.3em]">{unit}</span>
    </div>
  </motion.div>
);

const ChartContainer = ({ title, children }: any) => (
  <motion.div 
    initial={{ opacity: 0 }}
    whileInView={{ opacity: 1 }}
    viewport={{ once: true }}
    className="bg-[#0a0a0a] border border-white/5 p-10 rounded-sm"
  >
    <h3 className="text-xs font-light mb-10 text-gray-500 uppercase tracking-[0.4em] flex items-center gap-3">
      <div className="w-8 h-[1px] bg-white/10" />
      {title}
    </h3>
    {children}
  </motion.div>
);

export default Dashboard;