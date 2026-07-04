import React, { useState } from 'react';
import { Mail, User, Link2, Loader2, AlertCircle, KeyRound, ArrowLeft } from 'lucide-react';

export default function AuthPage({ backendUrl, onAuthSuccess }) {
  const [isLogin, setIsLogin] = useState(true);
  const [step, setStep] = useState(1); // 1: Enter Email/Name, 2: Enter OTP
  const [formData, setFormData] = useState({ name: '', email: '', otp: '' });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  // Phase 1: Request the 6-Digit OTP via Email
  const handleRequestOTP = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const endpoint = isLogin ? '/api/auth/login-request' : '/api/auth/signup-request';
      
      const response = await fetch(`${backendUrl}${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(
          isLogin 
            ? { email: formData.email }
            : { name: formData.name, email: formData.email }
        ),
      });

      const data = await response.json();

      console.log(data);

      if (!response.ok) {
        throw new Error(data.error || 'Failed to send OTP email.');
      }

      // Successfully dispatched OTP, move to Step 2
      setStep(2);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  // Phase 2: Verify the 6-Digit OTP
  const handleVerifyOTP = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await fetch(`${backendUrl}/api/auth/verify-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: formData.email,
          otp: formData.otp
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Invalid or expired OTP.');
      }

      // Capture JWT token inside client storage securely
      if (data.token) {
        localStorage.setItem('token', data.token);
        if (onAuthSuccess) onAuthSuccess();
      } else if (!isLogin) {
        // If signup requires them to re-login based on your backend architecture
        setIsLogin(true);
        setStep(1);
        setError('Account created successfully! Please log in.');
        setFormData({ name: '', email: '', otp: '' });
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setIsLogin(!isLogin);
    setStep(1);
    setError('');
    setFormData({ name: '', email: '', otp: '' });
  };

  return (
    <div className="min-h-screen bg-slate-900 flex flex-col justify-center py-12 sm:px-6 lg:px-8 font-sans antialiased selection:bg-indigo-500 selection:text-white">
      <div className="sm:mx-auto sm:w-full sm:max-w-md text-center animate-fadeIn">
        <div className="inline-flex items-center justify-center h-12 w-12 rounded-xl bg-indigo-600 text-white shadow-lg shadow-indigo-600/30 mb-4">
          <Link2 className="h-6 w-6 rotate-45" />
        </div>
        <h2 className="text-3xl font-extrabold text-white tracking-tight">
          {step === 2 
            ? 'Verify your identity' 
            : isLogin 
              ? 'Welcome back' 
              : 'Create your account'}
        </h2>
        <p className="mt-2 text-sm text-slate-400">
          {step === 2 
            ? `We sent a 6-digit code to ${formData.email}`
            : isLogin 
              ? "Passwordless access to your link engine" 
              : "Start scaling your link tracking engine today"}
        </p>
      </div>

      <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-md">
        <div className="bg-slate-950 py-8 px-4 shadow-xl border border-slate-800 sm:rounded-3xl sm:px-10 relative overflow-hidden">
          
          {/* Subtle Background Glow */}
          <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-600/10 rounded-full blur-3xl pointer-events-none"></div>

          {error && (
            <div className={`mb-6 p-4 rounded-xl flex items-start space-x-3 text-sm border ${
              error.includes('successfully') 
                ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' 
                : 'bg-rose-500/10 border-rose-500/30 text-rose-400'
            }`}>
              <AlertCircle className="h-5 w-5 shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {step === 1 ? (
            <form className="space-y-5 animate-fadeIn" onSubmit={handleRequestOTP}>
              {/* NAME FIELD (Only shown on signup phase) */}
              {!isLogin && (
                <div>
                  <label className="block text-sm font-semibold uppercase tracking-wider text-slate-400">Full Name</label>
                  <div className="mt-2 relative rounded-md shadow-sm">
                    <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                      <User className="h-5 w-5 text-slate-500" />
                    </div>
                    <input
                      type="text"
                      name="name"
                      required={!isLogin}
                      value={formData.name}
                      onChange={handleChange}
                      className="block w-full pl-11 pr-3 py-3.5 bg-slate-900 border border-slate-800 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                      placeholder="Alex Mercer"
                    />
                  </div>
                </div>
              )}

              {/* EMAIL FIELD */}
              <div>
                <label className="block text-sm font-semibold uppercase tracking-wider text-slate-400">Email Address</label>
                <div className="mt-2 relative rounded-md shadow-sm">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <Mail className="h-5 w-5 text-slate-500" />
                  </div>
                  <input
                    type="email"
                    name="email"
                    required
                    value={formData.email}
                    onChange={handleChange}
                    className="block w-full pl-11 pr-3 py-3.5 bg-slate-900 border border-slate-800 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all"
                    placeholder="you@example.com"
                  />
                </div>
              </div>

              <div className="pt-2">
                <button
                  type="submit"
                  disabled={loading}
                  className="w-full flex justify-center items-center py-4 px-4 rounded-xl shadow-lg text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-500 disabled:bg-indigo-800 disabled:cursor-not-allowed transition-all"
                >
                  {loading ? (
                    <Loader2 className="h-5 w-5 animate-spin" />
                  ) : (
                    'Send Login Code'
                  )}
                </button>
              </div>

              <div className="mt-6 text-center">
                <button
                  type="button"
                  onClick={resetForm}
                  className="text-sm font-medium text-indigo-400 hover:text-indigo-300 transition-colors focus:outline-none underline underline-offset-4"
                >
                  {isLogin ? "Don't have an account? Sign up" : 'Already have an account? Log in'}
                </button>
              </div>
            </form>
          ) : (
            <form className="space-y-5 animate-fadeIn" onSubmit={handleVerifyOTP}>
              {/* OTP VERIFICATION FIELD */}
              <div>
                <label className="block text-sm font-semibold text-center uppercase tracking-wider text-slate-400">
                  Enter 6-Digit Code
                </label>
                <div className="mt-4 relative rounded-md shadow-sm">
                  <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
                    <KeyRound className="h-5 w-5 text-slate-500" />
                  </div>
                  <input
                    type="text"
                    name="otp"
                    required
                    maxLength="6"
                    value={formData.otp}
                    onChange={handleChange}
                    className="block w-full pl-11 pr-3 py-4 bg-slate-900 border border-slate-800 rounded-xl text-white placeholder-slate-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all text-center tracking-[1em] text-xl font-bold"
                    placeholder="••••••"
                    autoComplete="one-time-code"
                  />
                </div>
              </div>

              <div className="pt-2">
                <button
                  type="submit"
                  disabled={loading || formData.otp.length < 6}
                  className="w-full flex justify-center items-center py-4 px-4 rounded-xl shadow-lg text-sm font-semibold text-white bg-emerald-600 hover:bg-emerald-500 disabled:bg-slate-800 disabled:text-slate-500 disabled:cursor-not-allowed transition-all"
                >
                  {loading ? (
                    <Loader2 className="h-5 w-5 animate-spin" />
                  ) : (
                    'Verify & Secure Access'
                  )}
                </button>
              </div>

              <div className="mt-6 text-center">
                <button
                  type="button"
                  onClick={() => {
                    setStep(1);
                    setFormData({ ...formData, otp: '' });
                    setError('');
                  }}
                  className="flex items-center justify-center w-full text-sm font-medium text-slate-400 hover:text-slate-300 transition-colors focus:outline-none"
                >
                  <ArrowLeft className="h-4 w-4 mr-1.5" />
                  Back to email entry
                </button>
              </div>
            </form>
          )}

        </div>
      </div>
    </div>
  );
}