import React, { useState } from 'react';
import { Radio, Zap, Clock, Image, Video, CheckCircle2, AlertTriangle } from 'lucide-react';
import { formatTime } from '../utils/playback';

export function SyncControls({ mediaList, isSyncActive, isPendingSync, countdownSec, remainingSec, activeSync, onTriggerSync, isSubmitting }) {
  const [selectedMediaId, setSelectedMediaId] = useState('');
  const [duration, setDuration] = useState(15);

  const handleSyncSubmit = async (e) => {
    e.preventDefault();
    if (!selectedMediaId) return;
    try {
      await onTriggerSync(parseInt(selectedMediaId, 10), parseInt(duration, 10));
    } catch (err) {
      // handled via notification
    }
  };

  const selectedMedia = mediaList.find((m) => m.id === parseInt(selectedMediaId, 10));

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 shadow-xl">
      <div className="flex items-center justify-between pb-4 mb-4 border-b border-slate-800">
        <div className="flex items-center gap-2.5">
          <div className="p-2 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Radio className="w-5 h-5 animate-pulse" />
          </div>
          <div>
            <h3 className="text-base font-semibold text-white tracking-wide">
              Multi-Window Synchronization Hub
            </h3>
            <p className="text-xs text-slate-400">
              Temporarily override all 4 display windows with a single media asset
            </p>
          </div>
        </div>

        {/* Active Status Badge */}
        {isSyncActive && (
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-amber-500/20 border border-amber-500/40 text-amber-300 text-xs font-mono font-bold animate-pulse">
            <span className="w-2 h-2 rounded-full bg-amber-400 animate-ping" />
            SYNC OVERRIDE ACTIVE ({remainingSec}s REMAINING)
          </div>
        )}

        {isPendingSync && (
          <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-cyan-500/20 border border-cyan-500/40 text-cyan-300 text-xs font-mono font-bold animate-pulse">
            <Clock className="w-3.5 h-3.5" />
            SYNCHRONIZING IN {countdownSec}s...
          </div>
        )}
      </div>

      {/* Sync Active Banner */}
      {isSyncActive && activeSync?.media && (
        <div className="mb-5 p-4 rounded-lg bg-gradient-to-r from-amber-950/60 to-slate-900 border border-amber-500/30 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-lg bg-black/80 overflow-hidden border border-amber-500/40 shrink-0 flex items-center justify-center">
              {activeSync.media.type === 'image' && <Image className="w-6 h-6 text-blue-400" />}
              {activeSync.media.type === 'video' && <Video className="w-6 h-6 text-purple-400" />}
            </div>
            <div>
              <span className="text-[10px] font-mono uppercase tracking-widest text-amber-400 font-bold">
                Currently Synchronized Playback
              </span>
              <h4 className="text-sm font-bold text-white truncate max-w-sm">
                {activeSync.media.name}
              </h4>
              <p className="text-xs text-slate-400">
                All windows will restore their normal playlist position automatically after countdown.
              </p>
            </div>
          </div>

          <div className="text-right shrink-0">
            <span className="text-2xl font-mono font-black text-amber-400">
              {formatTime(remainingSec)}
            </span>
            <p className="text-[10px] text-slate-400 font-mono">until normal resume</p>
          </div>
        </div>
      )}

      {/* Sync Trigger Form */}
      <form onSubmit={handleSyncSubmit} className="grid grid-cols-1 md:grid-cols-12 gap-4 items-end">
        {/* Media Selector */}
        <div className="md:col-span-6">
          <label className="block text-xs font-medium text-slate-300 mb-1.5">
            Select Sync Media Asset
          </label>
          <select
            value={selectedMediaId}
            onChange={(e) => {
              setSelectedMediaId(e.target.value);
              const found = mediaList.find((m) => m.id === parseInt(e.target.value, 10));
              if (found) setDuration(found.duration);
            }}
            className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3.5 py-2.5 text-sm text-white focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500 transition-all"
            required
          >
            <option value="" disabled>
              -- Choose media to synchronize --
            </option>
            {mediaList.map((media) => (
              <option key={media.id} value={media.id}>
                [{media.type.toUpperCase()}] {media.name} ({media.duration}s)
              </option>
            ))}
          </select>
        </div>

        {/* Sync Duration Input */}
        <div className="md:col-span-3">
          <label className="block text-xs font-medium text-slate-300 mb-1.5">
            Override Duration (seconds)
          </label>
          <div className="relative">
            <input
              type="number"
              min="1"
              max="18000"
              value={duration}
              onChange={(e) => setDuration(Math.max(1, parseInt(e.target.value, 10) || 1))}
              className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3.5 py-2.5 text-sm text-white focus:outline-none focus:border-amber-500 focus:ring-1 focus:ring-amber-500 font-mono"
              required
            />
            <span className="absolute right-3 top-2.5 text-xs text-slate-500 font-mono">sec</span>
          </div>
        </div>

        {/* Quick Presets & Trigger Button */}
        <div className="md:col-span-3">
          <button
            type="submit"
            disabled={!selectedMediaId || isSubmitting}
            className={`w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-lg font-semibold text-sm transition-all duration-200 shadow-lg ${
              !selectedMediaId || isSubmitting
                ? 'bg-slate-800 text-slate-500 cursor-not-allowed border border-slate-700'
                : 'bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-400 hover:to-orange-400 text-slate-950 font-bold shadow-amber-500/20 active:scale-[0.98]'
            }`}
          >
            <Zap className="w-4 h-4 fill-current" />
            {isSubmitting ? 'Syncing...' : 'SYNC ALL WINDOWS'}
          </button>
        </div>
      </form>

      {/* Preset Quick Duration Buttons */}
      <div className="mt-3 flex items-center gap-2 text-xs text-slate-400">
        <span className="text-[11px] text-slate-500 font-mono">Quick Duration:</span>
        {[10, 15, 20, 30, 60].map((presetSec) => (
          <button
            key={presetSec}
            type="button"
            onClick={() => setDuration(presetSec)}
            className={`px-2 py-0.5 rounded border text-[11px] font-mono transition-colors ${
              duration === presetSec
                ? 'bg-amber-500/20 border-amber-500 text-amber-300'
                : 'bg-slate-950 border-slate-800 text-slate-400 hover:text-white'
            }`}
          >
            {presetSec}s
          </button>
        ))}
      </div>
    </div>
  );
}
