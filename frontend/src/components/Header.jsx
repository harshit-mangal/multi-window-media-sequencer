import React, { useState, useEffect } from 'react';
import { Layers, Activity, FolderPlus, RefreshCw, Radio, Sparkles } from 'lucide-react';
import { getSynchronizedServerTime, formatTime } from '../utils/playback';

export function Header({ isConnected, isSyncActive, onOpenMediaLibrary, onRefresh }) {
  const [cycleInfo, setCycleInfo] = useState({ cycleNumber: 0, cycleElapsedSec: 0 });

  useEffect(() => {
    const timer = setInterval(() => {
      const nowMs = getSynchronizedServerTime();
      const nowSec = Math.floor(nowMs / 1000);
      const cycleNumber = Math.floor(nowSec / 18000);
      const cycleElapsedSec = nowSec % 18000;
      setCycleInfo({ cycleNumber, cycleElapsedSec });
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  return (
    <header className="bg-slate-900/90 border-b border-slate-800 sticky top-0 z-40 backdrop-blur-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Logo & Title */}
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center shadow-lg shadow-cyan-500/20">
            <Layers className="w-5 h-5 text-white" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-base font-bold text-white tracking-tight">
                Media Sequencer
              </h1>
              <span className="px-2 py-0.5 rounded-full bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 text-[10px] font-mono font-semibold">
                v1.0 SYNC
              </span>
            </div>
            <p className="text-xs text-slate-400 hidden sm:block">
              Multi-Window Deterministic 5-Hour Playback System
            </p>
          </div>
        </div>

        {/* Global Cycle Clock & Status */}
        <div className="flex items-center gap-4">
          <div className="hidden md:flex items-center gap-2.5 px-3 py-1.5 rounded-lg bg-slate-950 border border-slate-800 text-xs font-mono">
            <Activity className="w-3.5 h-3.5 text-cyan-400 animate-pulse" />
            <span className="text-slate-400">Cycle #{cycleInfo.cycleNumber}:</span>
            <span className="text-white font-semibold">
              {formatTime(cycleInfo.cycleElapsedSec)}
            </span>
            <span className="text-slate-600">/ 05:00:00</span>
          </div>

          {/* WebSocket Status Indicator */}
          <div className="flex items-center gap-2 px-2.5 py-1 rounded-full bg-slate-950 border border-slate-800 text-xs font-mono">
            <span
              className={`w-2 h-2 rounded-full ${
                isConnected ? 'bg-emerald-400 shadow-sm shadow-emerald-400 animate-pulse' : 'bg-red-400'
              }`}
            />
            <span className="text-slate-300 text-[11px]">
              {isConnected ? 'LIVE WS' : 'OFFLINE'}
            </span>
          </div>

          {/* Action Buttons */}
          <button
            onClick={onOpenMediaLibrary}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium border border-slate-700 transition-colors"
          >
            <FolderPlus className="w-3.5 h-3.5 text-cyan-400" />
            <span>Media Library</span>
          </button>

          <button
            onClick={onRefresh}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
            title="Refresh Data"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
        </div>
      </div>
    </header>
  );
}
