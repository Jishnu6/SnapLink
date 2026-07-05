import React, { useState, useEffect } from 'react';
import { Link2, BarChart3, Copy, Check, ArrowRight, Sparkles, Clock, LogOut } from 'lucide-react';
import AnalyticsDashboard from './components/AnalyticsDashboard';
import AuthPage from './components/AuthPage';

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [activeTab, setActiveTab] = useState('generate'); // 'generate' | 'analytics'
  const [longUrl, setLongUrl] = useState('');
  const [customAlias, setCustomAlias] = useState('');
  const [expiryDays, setExpiryDays] = useState('60');
  const [shortenedUrl, setShortenedUrl] = useState('');
  const [searchId, setSearchId] = useState('');
  const [copied, setCopied] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const BACKEND_URL = 'http://localhost:3000'; // Adjust to your Fiber port

  // Check storage on boot up lifecycle
  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      setIsAuthenticated(true);
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    setIsAuthenticated(false);
    setActiveTab('generate');
    setSearchId('');
  };

  const handleShorten = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setShortenedUrl('');

    const token = localStorage.getItem('token');

    try {
      const response = await fetch(`${BACKEND_URL}/api/v1/shorten`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}` // Critical JWT pipeline inclusion
        },
        body: JSON.stringify({
          url: longUrl,
          custom_alias: customAlias || undefined,
          expiry_duration: parseInt(expiryDays)
        }),
      });

      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Something went wrong');

      setShortenedUrl(`${BACKEND_URL}/${data.id}`);
      setSearchId(data.id); // Auto-fill analytics search target
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(shortenedUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // If unauthorized, lock rendering straight into the Auth Engine Gate
  if (!isAuthenticated) {
    return (
      <AuthPage 
        backendUrl={BACKEND_URL} 
        onAuthSuccess={() => setIsAuthenticated(true)} 
      />
    );
  }

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100 selection:bg-indigo-500 selection:text-white">
      {/* Header navbar */}
      <header className="border-b border-slate-800 bg-slate-900/50 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-4 h-16 flex items-center justify-between">
          <div className="flex items-center space-x-3 cursor-pointer" onClick={() => setActiveTab('generate')}>
            <div className="bg-indigo-600 p-2 rounded-xl text-white shadow-lg shadow-indigo-600/30">
              <Link2 className="h-5 w-5 rotate-45" />
            </div>
            <span className="font-bold text-xl tracking-tight bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              SnapLink Engine
            </span>
          </div>
          <div className="flex items-center space-x-6">
            <nav className="flex space-x-1 bg-slate-950 p-1 rounded-xl border border-slate-800">
              <button
                onClick={() => setActiveTab('generate')}
                className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                  activeTab === 'generate' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <Sparkles className="h-4 w-4" />
                <span>Generator</span>
              </button>
              <button
                onClick={() => setActiveTab('analytics')}
                className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                  activeTab === 'analytics' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                <BarChart3 className="h-4 w-4" />
                <span>Analytics</span>
              </button>
            </nav>
            
            <button
              onClick={handleLogout}
              className="flex items-center space-x-2 text-sm font-medium text-slate-400 hover:text-rose-400 transition-colors group"
            >
              <LogOut className="h-4 w-4 transform group-hover:-translate-x-1 transition-transform" />
              <span className="hidden sm:inline">Logout</span>
            </button>
          </div>
        </div>
      </header>

      {/* Main Container Layout */}
      <main className="max-w-6xl mx-auto px-4 py-12">
        {activeTab === 'generate' ? (
          <div className="max-w-2xl mx-auto space-y-8">
            <div className="text-center space-y-3">
              <h1 className="text-4xl font-extrabold tracking-tight sm:text-5xl bg-gradient-to-r from-white via-slate-200 to-slate-500 bg-clip-text text-transparent">
                Shorten Links. Track Telemetry.
              </h1>
              <p className="text-slate-400 text-lg">
              SnapLink delivers lightning-fast URL redirection powered by a high-performance hybrid architecture. Gain actionable insights into your audience with real-time analytics and enterprise-grade UTM tracking, all built for scale and security.
              </p>
            </div>

            <div className="bg-slate-950 border border-slate-800 rounded-3xl p-6 sm:p-8 shadow-2xl relative overflow-hidden">
              <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-600/10 rounded-full blur-3xl"></div>
              
              <form onSubmit={handleShorten} className="space-y-6">
                {/* Destination Link */}
                <div className="space-y-2">
                  <label className="text-xs font-semibold uppercase tracking-wider text-slate-400">Destination Long URL</label>
                  <div className="relative">
                    <input
                      type="url"
                      required
                      placeholder="https://example.com/deep/path/to/product?utm_source=campaign"
                      value={longUrl}
                      onChange={(e) => setLongUrl(e.target.value)}
                      className="w-full bg-slate-900 border border-slate-800 rounded-xl px-4 py-3.5 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                    />
                  </div>
                </div>

                {/* Configurations grid */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-xs font-semibold uppercase tracking-wider text-slate-400">Custom Alias (Optional)</label>
                    <input
                      type="text"
                      placeholder="promo2026"
                      value={customAlias}
                      onChange={(e) => setCustomAlias(e.target.value)}
                      className="w-full bg-slate-900 border border-slate-800 rounded-xl px-4 py-3.5 text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                    />
                  </div>
                  <div className="space-y-2">
                    <label className="text-xs font-semibold uppercase tracking-wider text-slate-400">Link Expiry Duration</label>
                    <div className="relative">
                      <select
                        value={expiryDays}
                        onChange={(e) => setExpiryDays(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-800 rounded-xl px-4 py-3.5 text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all appearance-none"
                      >
                        <option value="7">7 Days Grace Period</option>
                        <option value="30">30 Days (1 Month)</option>
                        <option value="60">60 Days (Standard 2 Months)</option>
                        <option value="365">365 Days (1 Year Link)</option>
                      </select>
                      <Clock className="absolute right-4 top-4 h-4 w-4 text-slate-500 pointer-events-none" />
                    </div>
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={loading}
                  className="w-full bg-indigo-600 hover:bg-indigo-500 disabled:bg-indigo-800 disabled:cursor-not-allowed text-white font-semibold py-4 rounded-xl shadow-lg shadow-indigo-600/20 transition-all flex items-center justify-center space-x-2"
                >
                  {loading ? (
                    <div className="h-5 w-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                  ) : (
                    <>
                      <span>Compress URL Target</span>
                      <ArrowRight className="h-4 w-4" />
                    </>
                  )}
                </button>
              </form>

              {/* Error Alert Display */}
              {error && (
                <div className="mt-6 bg-red-950/40 border border-red-900/50 text-red-400 rounded-xl p-4 text-sm">
                  {error}
                </div>
              )}

              {/* Success output block */}
              {shortenedUrl && (
                <div className="mt-8 pt-6 border-t border-slate-800/80 space-y-3 animate-fadeIn">
                  <span className="text-xs font-semibold uppercase tracking-wider text-emerald-400 flex items-center space-x-1">
                    <Check className="h-3 w-3" /> <span>Link generated dynamically</span>
                  </span>
                  <div className="flex bg-slate-900 border border-slate-800 rounded-xl overflow-hidden p-1">
                    <input
                      type="text"
                      readOnly
                      value={shortenedUrl}
                      className="w-full bg-transparent px-3 text-slate-200 text-sm focus:outline-none"
                    />
                    <button
                      onClick={copyToClipboard}
                      className="bg-slate-800 hover:bg-slate-700 text-slate-300 px-4 py-2.5 rounded-lg flex items-center space-x-1.5 transition-all text-sm font-medium shrink-0"
                    >
                      {copied ? <Check className="h-4 w-4 text-emerald-400" /> : <Copy className="h-4 w-4" />}
                      <span>{copied ? 'Copied' : 'Copy'}</span>
                    </button>
                  </div>
                  <div className="flex justify-end">
                    <button 
                      onClick={() => setActiveTab('analytics')}
                      className="text-xs text-indigo-400 hover:text-indigo-300 font-medium underline underline-offset-4"
                    >
                      Inspect live telemetry metrics →
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <AnalyticsDashboard initialId={searchId} backendUrl={BACKEND_URL} />
        )}
      </main>
    </div>
  );
}