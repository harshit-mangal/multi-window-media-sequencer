import React, { useState, useEffect } from 'react';
import { MediaPlayer } from './MediaPlayer';
import { computeWindowPlayback, formatTime } from '../utils/playback';
import { Settings2, Clock, ListMusic, ArrowRight, Radio } from 'lucide-react';

export function MediaWindow({ window, activeSync, onOpenPlaylistEditor }) {
  const [playbackState, setPlaybackState] = useState(() =>
    computeWindowPlayback(window?.playlist, activeSync)
  );

  // Fast tick loop for smooth timer & progress bar rendering (100ms)
  useEffect(() => {
    const updatePlayback = () => {
      const state = computeWindowPlayback(window?.playlist, activeSync);
      setPlaybackState(state);
    };

    updatePlayback();
    const interval = setInterval(updatePlayback, 100);
    return () => clearInterval(interval);
  }, [window?.playlist, activeSync]);

  const {
    isSyncOverride,
    media,
    itemElapsedSec,
    itemRemainingSec,
    progressPercent,
    nextItem,
    cycleNumber,
    cycleElapsedSec,
    playlistTotalSec,
    currentItemIndex,
  } = playbackState;

  const totalItems = window?.playlist?.length || 0;

  return (
    <div className="relative flex flex-col bg-slate-900 border border-slate-800 rounded-xl overflow-hidden shadow-2xl transition-all duration-300 hover:border-slate-700 group">
      {/* Window Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-slate-950/80 border-b border-slate-800/80">
        <div className="flex items-center gap-2.5">
          <div className="relative flex items-center justify-center">
            <Radio className={`w-4 h-4 ${isSyncOverride ? 'text-amber-400 animate-pulse' : 'text-emerald-400'}`} />
          </div>
          <div>
            <h3 className="text-sm font-semibold text-white tracking-wide truncate max-w-[180px]">
              {window.name}
            </h3>
            <span className="text-[11px] font-mono text-slate-400">
              Cycle #{cycleNumber} • {formatTime(cycleElapsedSec)} / 5:00:00
            </span>
          </div>
        </div>

        <div className="flex items-center gap-1.5">
          <button
            onClick={() => onOpenPlaylistEditor(window)}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
            title="Edit Window Playlist"
          >
            <Settings2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Media Playback Viewport */}
      <div className="relative w-full aspect-video bg-black">
        <MediaPlayer
          media={media}
          isSyncOverride={isSyncOverride}
          itemElapsedSec={itemElapsedSec}
          itemRemainingSec={itemRemainingSec}
        />

        {/* Floating Media Info Bar */}
        {media && (
          <div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/90 via-black/50 to-transparent p-3 pt-6 flex items-end justify-between text-white">
            <div className="truncate pr-2">
              <p className="text-xs font-medium text-slate-200 truncate">
                {isSyncOverride ? `[SYNC] ${media.name}` : media.name}
              </p>
              <p className="text-[10px] text-slate-400 font-mono">
                {isSyncOverride
                  ? `Sync Override • ${formatTime(itemRemainingSec)} left`
                  : `Item ${currentItemIndex + 1} of ${totalItems} • Total Loop: ${formatTime(playlistTotalSec)}`}
              </p>
            </div>

            <div className="flex items-center gap-1 text-right shrink-0">
              <span className="text-xs font-mono font-semibold text-emerald-400">
                {formatTime(itemElapsedSec)}
              </span>
              <span className="text-xs font-mono text-slate-500">/</span>
              <span className="text-xs font-mono text-slate-300">
                {formatTime(media.duration || 0)}
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Playback Progress Bar */}
      <div className="w-full h-1.5 bg-slate-800 overflow-hidden">
        <div
          className={`h-full transition-all duration-100 ease-linear ${
            isSyncOverride ? 'bg-gradient-to-r from-amber-500 to-amber-300' : 'bg-gradient-to-r from-cyan-500 to-emerald-400'
          }`}
          style={{ width: `${progressPercent}%` }}
        />
      </div>

      {/* Footer / Next Item Info */}
      <div className="px-3.5 py-2 bg-slate-950/60 flex items-center justify-between text-[11px] text-slate-400 border-t border-slate-800/40 font-mono">
        <div className="flex items-center gap-1.5 truncate">
          <ListMusic className="w-3.5 h-3.5 text-slate-500 shrink-0" />
          <span className="text-slate-500">Next:</span>
          <span className="text-slate-300 truncate">
            {nextItem ? nextItem.name : 'Loop Restart'}
          </span>
        </div>

        <div className="flex items-center gap-1 text-slate-500 shrink-0">
          <Clock className="w-3 h-3" />
          <span>-{formatTime(itemRemainingSec)}</span>
        </div>
      </div>
    </div>
  );
}
