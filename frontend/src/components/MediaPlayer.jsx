import React, { useRef, useEffect, useState } from 'react';
import { Image, Video, MonitorOff, AlertCircle } from 'lucide-react';

export function MediaPlayer({ media, isSyncOverride, itemElapsedSec, itemRemainingSec }) {
  const videoRef = useRef(null);
  const [loadError, setLoadError] = useState(false);

  // Reset error when media URL changes
  useEffect(() => {
    setLoadError(false);
  }, [media?.url]);

  // Synchronize video playback position if applicable
  useEffect(() => {
    if (media?.type === 'video' && videoRef.current && !isNaN(itemElapsedSec)) {
      const vid = videoRef.current;
      const currentVideoTime = vid.currentTime;
      const actualDuration = vid.duration;

      // Handle looping within video file length if video is shorter than playlist item duration
      let targetTime = itemElapsedSec;
      if (actualDuration && !isNaN(actualDuration) && isFinite(actualDuration) && actualDuration > 0) {
        targetTime = itemElapsedSec % actualDuration;
      }

      // If drift is greater than 1.5 seconds, resync video currentTime
      if (Math.abs(currentVideoTime - targetTime) > 1.5) {
        try {
          vid.currentTime = targetTime;
        } catch (e) {
          // ignore seek errors on unbuffered media
        }
      }

      // Ensure video is actively playing if browser had paused it
      if (vid.paused) {
        vid.play().catch(() => {});
      }
    }
  }, [media, itemElapsedSec]);

  if (!media) {
    return (
      <div className="relative w-full h-full flex flex-col items-center justify-center bg-slate-900/90 text-slate-400 p-4 select-none">
        <MonitorOff className="w-12 h-12 text-slate-600 mb-2 animate-pulse" />
        <p className="text-sm font-medium tracking-wide">No Active Media</p>
        <p className="text-xs text-slate-500 mt-1">Add items to window playlist</p>
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="relative w-full h-full flex flex-col items-center justify-center bg-red-950/40 text-red-400 p-4">
        <AlertCircle className="w-10 h-10 text-red-500 mb-2" />
        <p className="text-sm font-semibold">Media Load Error</p>
        <p className="text-xs text-red-300 mt-1 text-center truncate max-w-xs">{media.url}</p>
      </div>
    );
  }

  if (media.type === 'blank') {
    return (
      <div className="relative w-full h-full flex flex-col items-center justify-center bg-black text-slate-600 select-none">
        <div className="w-4 h-4 rounded-full bg-slate-800 animate-ping opacity-75 mb-2" />
        <span className="text-xs tracking-widest uppercase font-mono text-slate-500">
          [ BLANK PLAYBACK ]
        </span>
      </div>
    );
  }

  return (
    <div className="relative w-full h-full bg-black overflow-hidden flex items-center justify-center">
      {media.type === 'image' && (
        <img
          key={media.url}
          src={media.url}
          alt={media.name}
          className="w-full h-full object-cover transition-opacity duration-700 ease-in-out"
          onError={() => setLoadError(true)}
        />
      )}

      {media.type === 'video' && (
        <video
          ref={videoRef}
          key={media.url}
          src={media.url}
          autoPlay
          muted
          playsInline
          loop
          className="w-full h-full object-cover"
          onError={() => setLoadError(true)}
        />
      )}

      {/* Media Type Watermark Badge */}
      <div className="absolute top-3 left-3 z-10 flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-black/60 backdrop-blur-md border border-white/10 text-white/80 text-xs font-mono">
        {media.type === 'image' && <Image className="w-3.5 h-3.5 text-blue-400" />}
        {media.type === 'video' && <Video className="w-3.5 h-3.5 text-purple-400" />}
        <span className="capitalize">{media.type}</span>
      </div>

      {/* Sync Active Overlay Banner */}
      {isSyncOverride && (
        <div className="absolute top-3 right-3 z-10 flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-amber-500/90 text-slate-950 text-xs font-bold uppercase tracking-wider animate-pulse shadow-lg shadow-amber-500/20">
          <span className="w-2 h-2 rounded-full bg-slate-950 animate-ping" />
          SYNC OVERRIDE
        </div>
      )}
    </div>
  );
}
