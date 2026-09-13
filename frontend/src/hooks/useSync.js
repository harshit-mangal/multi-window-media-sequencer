import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { wsService } from '../services/websocket';
import { syncServerTime, getSynchronizedServerTime } from '../utils/playback';

export function useSync(onNotification) {
  const [activeSync, setActiveSync] = useState(null);
  const [pendingSync, setPendingSync] = useState(null);
  const [countdownSec, setCountdownSec] = useState(0);
  const [remainingSec, setRemainingSec] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Initialize and check current sync on mount (handles page reload during sync)
  const checkCurrentSync = useCallback(async () => {
    try {
      const resp = await api.getCurrentSync();
      if (resp && resp.serverTime) {
        syncServerTime(resp.serverTime);
      }

      if (resp && resp.active && resp.syncEvent) {
        const now = getSynchronizedServerTime();
        const startMs = new Date(resp.syncEvent.startAt).getTime();
        const endMs = new Date(resp.syncEvent.endAt).getTime();

        if (now < startMs) {
          // Upcoming sync in buffer window
          setPendingSync(resp.syncEvent);
          setActiveSync(null);
        } else if (now >= startMs && now < endMs) {
          // In-progress sync
          setActiveSync(resp.syncEvent);
          setPendingSync(null);
        } else {
          setActiveSync(null);
          setPendingSync(null);
        }
      } else {
        setActiveSync(null);
        setPendingSync(null);
      }
    } catch (err) {
      console.error('[SYNC_ERROR] Failed to fetch current sync:', err);
    }
  }, []);

  useEffect(() => {
    checkCurrentSync();

    // Listen for WebSocket SYNC_START event
    const unbindSyncStart = wsService.on('SYNC_START', (payload) => {
      console.log('[SYNC_EVENT] SYNC_START received:', payload);
      const startMs = new Date(payload.startAt).getTime();
      const now = getSynchronizedServerTime();

      if (now < startMs) {
        setPendingSync(payload);
        setActiveSync(null);
      } else {
        setActiveSync(payload);
        setPendingSync(null);
      }

      if (onNotification) {
        onNotification({
          type: 'info',
          title: 'Sync Event Triggered',
          message: `Synchronizing all windows to "${payload.media?.name || 'Selected Media'}" for ${payload.duration}s`,
        });
      }
    });

    // Listen for WebSocket SYNC_END event
    const unbindSyncEnd = wsService.on('SYNC_END', (payload) => {
      console.log('[SYNC_EVENT] SYNC_END received:', payload);
      setActiveSync(null);
      setPendingSync(null);

      if (onNotification) {
        onNotification({
          type: 'success',
          title: 'Sync Concluded',
          message: 'All display windows have returned to their normal playlist sequences.',
        });
      }
    });

    return () => {
      unbindSyncStart();
      unbindSyncEnd();
    };
  }, [checkCurrentSync, onNotification]);

  // Tick loop to evaluate pending / active transition and update remaining seconds
  useEffect(() => {
    const timer = setInterval(() => {
      const now = getSynchronizedServerTime();

      if (pendingSync) {
        const startMs = new Date(pendingSync.startAt).getTime();
        const endMs = new Date(pendingSync.endAt).getTime();

        if (now >= startMs) {
          // Transition pending -> active
          setActiveSync(pendingSync);
          setPendingSync(null);
        } else {
          const diffSec = Math.max(0, Math.ceil((startMs - now) / 1000));
          setCountdownSec(diffSec);
        }
      }

      if (activeSync) {
        const endMs = new Date(activeSync.endAt).getTime();
        if (now >= endMs) {
          setActiveSync(null);
        } else {
          const diffSec = Math.max(0, Math.ceil((endMs - now) / 1000));
          setRemainingSec(diffSec);
        }
      }
    }, 100);

    return () => clearInterval(timer);
  }, [pendingSync, activeSync]);

  const triggerSync = async (mediaId, duration) => {
    setIsSubmitting(true);
    try {
      const event = await api.startSync({ mediaId, duration });
      return event;
    } catch (err) {
      if (onNotification) {
        onNotification({
          type: 'error',
          title: 'Sync Trigger Failed',
          message: err.message || 'Could not initiate synchronization override.',
        });
      }
      throw err;
    } finally {
      setIsSubmitting(false);
    }
  };

  return {
    activeSync,
    pendingSync,
    countdownSec,
    remainingSec,
    isSyncActive: !!activeSync,
    isPendingSync: !!pendingSync,
    isSubmitting,
    triggerSync,
    checkCurrentSync,
  };
}
