import React from 'react';
import { MediaWindow } from './MediaWindow';
import { MonitorPlay } from 'lucide-react';

export function WindowGrid({ windows, activeSync, onOpenPlaylistEditor }) {
  if (!windows || windows.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-24 bg-slate-900/50 rounded-2xl border border-slate-800 text-slate-400">
        <MonitorPlay className="w-16 h-16 text-slate-700 mb-4 animate-bounce" />
        <h4 className="text-lg font-semibold text-slate-300">No Display Windows Configured</h4>
        <p className="text-sm text-slate-500 mt-1">Please ensure backend database is initialized with seed data.</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {windows.map((window) => (
        <MediaWindow
          key={window.id}
          window={window}
          activeSync={activeSync}
          onOpenPlaylistEditor={onOpenPlaylistEditor}
        />
      ))}
    </div>
  );
}
