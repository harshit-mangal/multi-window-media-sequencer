// Client-side deterministic 5-hour cycle playback calculation

const CYCLE_DURATION_SEC = 18000; // 5 hours = 18,000 seconds
let serverTimeOffsetMs = 0; // Difference between server time and local client time

/**
 * Synchronize client clock with backend server timestamp
 */
export function syncServerTime(serverTimestampStr) {
  if (!serverTimestampStr) return;
  const serverMs = new Date(serverTimestampStr).getTime();
  const localMs = Date.now();
  serverTimeOffsetMs = serverMs - localMs;
}

/**
 * Get synchronized current server time in milliseconds
 */
export function getSynchronizedServerTime() {
  return Date.now() + serverTimeOffsetMs;
}

/**
 * Compute deterministic playback position for a window's playlist
 * @param {Array} playlist - List of playlist items with media details
 * @param {Object|null} activeSync - Active temporary sync override if any
 * @returns {Object} playback state
 */
export function computeWindowPlayback(playlist = [], activeSync = null) {
  const nowMs = getSynchronizedServerTime();
  const nowSec = Math.floor(nowMs / 1000);

  const cycleNumber = Math.floor(nowSec / CYCLE_DURATION_SEC);
  const cycleElapsedSec = nowSec % CYCLE_DURATION_SEC;

  // 1. Check if temporary sync override is active
  if (activeSync) {
    const startMs = new Date(activeSync.startAt).getTime();
    const endMs = new Date(activeSync.endAt).getTime();

    if (nowMs >= startMs && nowMs < endMs) {
      const syncElapsedSec = Math.floor((nowMs - startMs) / 1000);
      const syncRemainingSec = Math.max(0, Math.floor((endMs - nowMs) / 1000));
      const syncTotalDuration = activeSync.duration || Math.floor((endMs - startMs) / 1000);

      return {
        isSyncOverride: true,
        media: activeSync.media,
        syncId: activeSync.id || activeSync.syncId,
        itemElapsedSec: syncElapsedSec,
        itemRemainingSec: syncRemainingSec,
        progressPercent: Math.min(100, Math.max(0, (syncElapsedSec / syncTotalDuration) * 100)),
        cycleNumber,
        cycleElapsedSec,
        playlistTotalSec: 0,
        currentItemIndex: -1,
        nextItem: null,
      };
    }
  }

  // 2. Normal deterministic playlist playback
  if (!playlist || playlist.length === 0) {
    return {
      isSyncOverride: false,
      media: null,
      itemElapsedSec: 0,
      itemRemainingSec: 0,
      progressPercent: 0,
      cycleNumber,
      cycleElapsedSec,
      playlistTotalSec: 0,
      currentItemIndex: -1,
      nextItem: null,
    };
  }

  // Filter valid media items
  const validItems = playlist.filter((item) => item.media && item.media.duration > 0);
  if (validItems.length === 0) {
    return {
      isSyncOverride: false,
      media: null,
      itemElapsedSec: 0,
      itemRemainingSec: 0,
      progressPercent: 0,
      cycleNumber,
      cycleElapsedSec,
      playlistTotalSec: 0,
      currentItemIndex: -1,
      nextItem: null,
    };
  }

  const playlistTotalSec = validItems.reduce((sum, item) => sum + item.media.duration, 0);
  const playLoopElapsedSec = cycleElapsedSec % playlistTotalSec;

  let accumulatedSec = 0;
  for (let i = 0; i < validItems.length; i++) {
    const item = validItems[i];
    const duration = item.media.duration;

    if (playLoopElapsedSec >= accumulatedSec && playLoopElapsedSec < accumulatedSec + duration) {
      const itemElapsedSec = playLoopElapsedSec - accumulatedSec;
      const itemRemainingSec = duration - itemElapsedSec;
      const progressPercent = Math.min(100, Math.max(0, (itemElapsedSec / duration) * 100));

      const nextIndex = (i + 1) % validItems.length;
      const nextItem = validItems[nextIndex];

      return {
        isSyncOverride: false,
        media: item.media,
        playlistItemId: item.id,
        currentItemIndex: i,
        itemElapsedSec,
        itemRemainingSec,
        progressPercent,
        nextItem: nextItem ? nextItem.media : null,
        cycleNumber,
        cycleElapsedSec,
        playlistTotalSec,
      };
    }
    accumulatedSec += duration;
  }

  // Fallback
  return {
    isSyncOverride: false,
    media: validItems[0].media,
    playlistItemId: validItems[0].id,
    currentItemIndex: 0,
    itemElapsedSec: 0,
    itemRemainingSec: validItems[0].media.duration,
    progressPercent: 0,
    nextItem: validItems.length > 1 ? validItems[1].media : validItems[0].media,
    cycleNumber,
    cycleElapsedSec,
    playlistTotalSec,
  };
}

/**
 * Format seconds into HH:MM:SS or MM:SS
 */
export function formatTime(seconds) {
  if (isNaN(seconds) || seconds < 0) seconds = 0;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);

  if (h > 0) {
    return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  }
  return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}
