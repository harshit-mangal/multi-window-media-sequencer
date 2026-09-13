import { useState, useEffect, useCallback } from 'react';
import { api } from '../services/api';
import { wsService } from '../services/websocket';

export function usePlaylist(onNotification) {
  const [windows, setWindows] = useState([]);
  const [mediaList, setMediaList] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchMedia = useCallback(async () => {
    try {
      const data = await api.getMedia();
      setMediaList(data || []);
    } catch (err) {
      console.error('[MEDIA_ERROR] Failed to fetch media items:', err);
    }
  }, []);

  const fetchWindows = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await api.getWindows();
      setWindows(data || []);
      setError(null);
    } catch (err) {
      console.error('[WINDOWS_ERROR] Failed to fetch windows:', err);
      setError(err.message || 'Failed to load windows');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const refreshWindow = useCallback(async (windowId) => {
    try {
      const playlist = await api.getPlaylist(windowId);
      setWindows((prevWindows) =>
        prevWindows.map((w) => (w.id === windowId ? { ...w, playlist: playlist || [] } : w))
      );
    } catch (err) {
      console.error(`[PLAYLIST_ERROR] Failed to refresh playlist for window ${windowId}:`, err);
    }
  }, []);

  useEffect(() => {
    fetchWindows();
    fetchMedia();

    // Listen for WebSocket PLAYLIST_UPDATED event
    const unbindPlaylistUpdate = wsService.on('PLAYLIST_UPDATED', (payload) => {
      console.log('[PLAYLIST_EVENT] PLAYLIST_UPDATED for Window ID:', payload.windowId);
      if (payload.windowId) {
        refreshWindow(payload.windowId);
      } else {
        fetchWindows();
      }
    });

    return () => {
      unbindPlaylistUpdate();
    };
  }, [fetchWindows, fetchMedia, refreshWindow]);

  const addItem = async (windowId, mediaId, position) => {
    try {
      await api.addToPlaylist(windowId, { mediaId, position });
      await refreshWindow(windowId);
      if (onNotification) {
        onNotification({
          type: 'success',
          title: 'Media Added',
          message: `Added media to window playlist.`,
        });
      }
    } catch (err) {
      if (onNotification) {
        onNotification({
          type: 'error',
          title: 'Add Media Failed',
          message: err.message,
        });
      }
      throw err;
    }
  };

  const updateItem = async (windowId, itemId, position, mediaId) => {
    try {
      await api.updatePlaylistItem(windowId, itemId, { position, mediaId });
      await refreshWindow(windowId);
    } catch (err) {
      if (onNotification) {
        onNotification({
          type: 'error',
          title: 'Update Failed',
          message: err.message,
        });
      }
      throw err;
    }
  };

  const removeItem = async (windowId, itemId) => {
    try {
      await api.removeFromPlaylist(windowId, itemId);
      await refreshWindow(windowId);
      if (onNotification) {
        onNotification({
          type: 'info',
          title: 'Item Removed',
          message: `Item deleted from playlist.`,
        });
      }
    } catch (err) {
      if (onNotification) {
        onNotification({
          type: 'error',
          title: 'Removal Failed',
          message: err.message,
        });
      }
      throw err;
    }
  };

  return {
    windows,
    mediaList,
    isLoading,
    error,
    refreshWindows: fetchWindows,
    refreshMedia: fetchMedia,
    addItem,
    updateItem,
    removeItem,
  };
}
