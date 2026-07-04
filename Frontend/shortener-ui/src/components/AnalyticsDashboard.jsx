import React, { useState, useEffect } from 'react';
import { Search, Globe, Laptop, RefreshCw, Layers, Zap, AlertCircle } from 'lucide-react';

export default function AnalyticsDashboard({ initialId = '', backendUrl }) {
  const [shortId, setShortId] = useState(initialId);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchStats = async (targetId) => {
    if (!targetId) return;
    setLoading(true);
    setError('');
    try {
      const token = localStorage.getItem('token');

      const response = await fetch(`${backendUrl}/api/v1/stats/${targetId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}` // ATTACH IT HERE TOO
      }
    });
      const data = await response.json();
      
      if (!response.ok) throw new Error(data.error || 'Failed to fetch analytics metrics.');
      setStats(data);
    } catch (err) {
      setError(err.message);
      setStats(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (initialId) fetchStats(initialId);
  }, [initialId]);

  const handleSubmit = (e) => {
    e.preventDefault();
    const cleanedInput = shortId.trim();
    if (!cleanedInput) return;

    const targetId = cleanedInput.split('/').filter(Boolean).pop();
    fetchStats(targetId);
  };

  // Small utility helper to process high-to-low percentage bars
  const renderMetricRows = (metricMap) => {
    if (!metricMap || Object.keys(metricMap).length === 0) {
      return <div className="text-slate-500 text-sm italic py-2">No data recorded for this metric.</div>;
    }

    const items = Object.entries(metricMap).sort((a, b) => b[1] - a[1]);
    const maxVal = Math.max(...items.map(i => i[1]));

    return (
      <div className="space-y-3.5 mt-2">
        {items.map(([key, val]) => {
          const percentage = maxVal > 0 ? (val / maxVal) * 100 : 0;
          return (
            <div key={key} className="space-y-1">
              <div className="flex justify-between text-xs font-medium">
                <span className="text-slate-300">{key === "" ? "Direct / Unknown" : key}</span>
                <span className="text-slate-400 font-mono">{val.toLocaleString()} hits</span>
              </div>
              <div className="w-full bg-slate-900 rounded-full h-2 overflow-hidden border border-slate-800/50">
                <div 
                  className="bg-indigo-500 h-full rounded-full transition-all duration-500" 
                  style={{ width: `${percentage}%` }}
                ></div>
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="space-y-8 max-w-5xl mx-auto p-4">
      {/* Header Search Module */}
      <div className="bg-slate-950 border border-slate-800 rounded-2xl p-6 flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="space-y-1 text-center md:text-left">
          <h2 className="text-xl font-bold text-white">Link Analytics Dashboard</h2>
          <p className="text-slate-400 text-sm">Enter a short link path or code to fetch visitor performance data.</p>
        </div>
        <form onSubmit={handleSubmit} className="w-full md:w-auto flex bg-slate-900 border border-slate-800 rounded-xl overflow-hidden p-1 max-w-md shrink-0">
          <input
            type="text"
            placeholder="Paste code or URL (e.g., devlink)"
            value={shortId}
            onChange={(e) => setShortId(e.target.value)}
            className="w-full md:w-64 bg-transparent px-3 text-slate-200 text-sm focus:outline-none placeholder-slate-500"
          />
          <button
            type="submit"
            disabled={loading || !shortId.trim()}
            className="bg-indigo-600 hover:bg-indigo-500 disabled:bg-indigo-800/50 disabled:text-slate-400 text-white px-4 py-2 rounded-lg flex items-center space-x-1.5 transition-all text-sm font-medium"
          >
            <Search className="h-4 w-4" />
            <span>Inspect</span>
          </button>
        </form>
      </div>

      {/* Loading Block */}
      {loading && (
        <div className="flex flex-col items-center justify-center py-20 space-y-4">
          <RefreshCw className="h-8 w-8 text-indigo-500 animate-spin" />
          <span className="text-sm text-slate-400 font-mono">Aggregating real-time metrics...</span>
        </div>
      )}

      {/* Error Interface */}
      {error && !loading && (
        <div className="bg-red-950/30 border border-red-900/50 rounded-2xl p-6 text-center space-y-2 max-w-md mx-auto">
          <AlertCircle className="h-8 w-8 text-red-500 mx-auto" />
          <p className="text-red-400 font-medium text-sm">{error}</p>
        </div>
      )}

      {/* Analytics Result Layout Grid */}
      {stats && !loading && (
        <div className="space-y-6 animate-fadeIn">
          {/* Top Line Performance Metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="bg-slate-950 border border-slate-800 rounded-2xl p-5 relative overflow-hidden">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400 block">Total Clicks</span>
              <span className="text-3xl font-extrabold font-mono mt-2 block text-white">
                {(stats.total_clicks || 0).toLocaleString()}
              </span>
              <div className="absolute top-3 right-3 bg-indigo-500/10 p-1.5 rounded-lg text-indigo-400">
                <Zap className="h-4 w-4" />
              </div>
            </div>

            <div className="bg-slate-950 border border-slate-800 rounded-2xl p-5 relative overflow-hidden">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400 block">Unique Visitors</span>
              <span className="text-3xl font-extrabold font-mono mt-2 block text-emerald-400">
                {(stats.unique_visitor || 0).toLocaleString()}
              </span>
              <div className="absolute top-3 right-3 bg-emerald-500/10 p-1.5 rounded-lg text-emerald-400">
                <Layers className="h-4 w-4" />
              </div>
            </div>

            {/* Combined Status & Expiry Panel */}
            <div className="bg-slate-950 border border-slate-800 rounded-2xl p-5 relative overflow-hidden col-span-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-400 block">Link Status Registry</span>
              
              <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold font-mono mt-3 ${
                stats.expires_at && new Date(stats.expires_at) < new Date()
                  ? 'bg-rose-950/50 text-rose-400 border border-rose-900/50'
                  : stats.state === 'Live' 
                    ? 'bg-emerald-950/50 text-emerald-400 border border-emerald-900/50' 
                    : 'bg-amber-950/50 text-amber-400 border border-amber-900/50'
              }`}>
                ● Status: {stats.expires_at && new Date(stats.expires_at) < new Date() ? 'Expired' : (stats.state || 'Live')}
              </span>

              {/* Expiry Timestamp row below status badge */}
              <div className="mt-3 pt-3 border-t border-slate-900/60 flex flex-wrap items-center justify-center gap-x-3 gap-y-1 w-full">
  <span className="text-[11px] font-semibold uppercase tracking-wider text-slate-500">Link Lifespan:</span>
  {stats.expires_at ? (
    <div className="flex items-center space-x-2 text-xs font-mono">
      <span className={new Date(stats.expires_at) < new Date() ? 'text-rose-400 line-through' : 'text-emerald-400 font-medium'}>
        {new Date(stats.expires_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' })}
      </span>
      <span className="text-slate-600">@</span>
      <span className="text-slate-400 text-[11px]">
        {new Date(stats.expires_at).toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })}
      </span>
    </div>
  ) : (
    <span className="text-xs font-mono text-slate-400 italic">Permanent Link (No Expiration)</span>
  )}
</div>

              <div className="text-[10px] text-slate-500 mt-2.5 truncate font-mono" title={stats.long_url}>
                Dest: {stats.long_url || 'Target resolved'}
              </div>
            </div>
          </div>

          {/* Breakdown Analytics Sub-Matrices */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Geolocation distribution list */}
            <div className="bg-slate-950 border border-slate-800 rounded-2xl p-6">
              <div className="flex items-center space-x-2 pb-3 border-b border-slate-900">
                <Globe className="h-4 w-4 text-indigo-400" />
                <h3 className="font-semibold text-sm uppercase tracking-wider text-slate-300">Geographic Regions</h3>
              </div>
              <div className="pt-2 max-h-64 overflow-y-auto pr-2">
                {renderMetricRows(stats.countries)}
              </div>
            </div>

            {/* Referrer logs */}
            <div className="bg-slate-950 border border-slate-800 rounded-2xl p-6">
              <div className="flex items-center space-x-2 pb-3 border-b border-slate-900">
                <Layers className="h-4 w-4 text-emerald-400" />
                <h3 className="font-semibold text-sm uppercase tracking-wider text-slate-300">Top Referrers</h3>
              </div>
              <div className="pt-2 max-h-64 overflow-y-auto pr-2">
                {renderMetricRows(stats.referrers)}
              </div>
            </div>

            {/* Device breakdown matrix */}
            <div className="bg-slate-950 border border-slate-800 rounded-2xl p-6">
              <div className="flex items-center space-x-2 pb-3 border-b border-slate-900">
                <Laptop className="h-4 w-4 text-sky-400" />
                <h3 className="font-semibold text-sm uppercase tracking-wider text-slate-300">Devices & Platforms</h3>
              </div>
              <div className="pt-2 max-h-64 overflow-y-auto pr-2">
                {renderMetricRows(stats.devices)}
              </div>
            </div>

            {/* UTM Tracking Trends */}
            <div className="bg-slate-950 border border-slate-800 rounded-2xl p-6">
              <div className="flex items-center space-x-2 pb-3 border-b border-slate-900">
                <Search className="h-4 w-4 text-amber-400" />
                <h3 className="font-semibold text-sm uppercase tracking-wider text-slate-300">Campaign Sources (UTM)</h3>
              </div>
              <div className="pt-2 max-h-64 overflow-y-auto pr-2">
                {renderMetricRows(stats.utm_sources)}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}