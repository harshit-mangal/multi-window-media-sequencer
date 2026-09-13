import React, { useState, useCallback } from 'react';
import { Header } from './components/Header';
import { SyncControls } from './components/SyncControls';
import { WindowGrid } from './components/WindowGrid';
import { PlaylistEditor } from './components/PlaylistEditor';
import { MediaLibraryModal } from './components/MediaLibraryModal';
import { useWebSocket } from './hooks/useWebSocket';
import { useSync } from './hooks/useSync';
import { usePlaylist } from './hooks/usePlaylist';
import { Info, CheckCircle, AlertTriangle, X } from 'lucide-react';

export function App() {
  const [notifications, setNotifications] = useState([]);
  const [editingWindow, setEditingWindow] = useState(null);
  const [isMediaLibraryOpen, setIsMediaLibraryOpen] = useState(false);

  const addNotification = useCallback((notif) => {
    const id = Date.now() + Math.random();
    setNotifications((prev) => [...prev, { ...notif, id }]);
    setTimeout(() => {
      setNotifications((prev) => prev.filter((n) => n.id !== id));
    }, 5000);
  }, []);

  const { isConnected } = useWebSocket();
  const {
    activeSync,
    pendingSync,
    countdownSec,
    remainingSec,
    isSyncActive,
    isPendingSync,
    isSubmitting,
    triggerSync,
    checkCurrentSync,
  } = useSync(addNotification);

  const {
    windows,
    mediaList,
    isLoading,
    refreshWindows,
    refreshMedia,
    addItem,
    updateItem,
    removeItem,
  } = usePlaylist(addNotification);

  const handleRefreshAll = () => {
    refreshWindows();
    refreshMedia();
    checkCurrentSync();
  };

  // Re-fetch selected window to keep PlaylistEditor updated
  const currentEditingWindow = windows.find((w) => w.id === editingWindow?.id) || editingWindow;

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans selection:bg-cyan-500 selection:text-black">
      {/* Top Header */}
      <Header
        isConnected={isConnected}
        isSyncActive={isSyncActive}
        onOpenMediaLibrary={() => setIsMediaLibraryOpen(true)}
        onRefresh={handleRefreshAll}
      />

      {/* Main Dashboard */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
        {/* Sync Controls Hub */}
        <SyncControls
          mediaList={mediaList}
          isSyncActive={isSyncActive}
          isPendingSync={isPendingSync}
          countdownSec={countdownSec}
          remainingSec={remainingSec}
          activeSync={activeSync}
          onTriggerSync={triggerSync}
          isSubmitting={isSubmitting}
        />

        {/* Display Windows Grid */}
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-bold uppercase tracking-wider text-slate-400">
              Live Window Sequencers ({windows.length})
            </h2>
            <span className="text-xs font-mono text-slate-500">
              Continuous 5-Hour Cycle • Auto-Looping
            </span>
          </div>

          <WindowGrid
            windows={windows}
            activeSync={activeSync}
            onOpenPlaylistEditor={(win) => setEditingWindow(win)}
          />
        </div>
      </main>

      {/* Playlist Editor Drawer/Modal */}
      {editingWindow && (
        <PlaylistEditor
          window={currentEditingWindow}
          mediaList={mediaList}
          onClose={() => setEditingWindow(null)}
          onAddItem={addItem}
          onRemoveItem={removeItem}
          onUpdateItem={updateItem}
        />
      )}

      {/* Media Library Modal */}
      <MediaLibraryModal
        isOpen={isMediaLibraryOpen}
        onClose={() => setIsMediaLibraryOpen(false)}
        mediaList={mediaList}
        onMediaCreated={() => {
          refreshMedia();
          addNotification({
            type: 'success',
            title: 'Media Asset Created',
            message: 'New media asset successfully registered.',
          });
        }}
        onMediaDeleted={() => {
          refreshMedia();
          refreshWindows();
          addNotification({
            type: 'info',
            title: 'Media Deleted',
            message: 'Media asset removed from system.',
          });
        }}
      />

      {/* Floating Notifications / Toasts */}
      <div className="fixed bottom-4 right-4 z-50 space-y-2 max-w-md w-full pointer-events-none">
        {notifications.map((n) => (
          <div
            key={n.id}
            className={`pointer-events-auto flex items-start gap-3 p-3.5 rounded-xl shadow-2xl border backdrop-blur-md transition-all duration-300 animate-slide-in ${
              n.type === 'error'
                ? 'bg-red-950/90 border-red-800 text-red-200'
                : n.type === 'success'
                ? 'bg-emerald-950/90 border-emerald-800 text-emerald-200'
                : 'bg-slate-900/90 border-slate-700 text-slate-200'
            }`}
          >
            {n.type === 'error' && <AlertTriangle className="w-5 h-5 text-red-400 shrink-0 mt-0.5" />}
            {n.type === 'success' && <CheckCircle className="w-5 h-5 text-emerald-400 shrink-0 mt-0.5" />}
            {n.type === 'info' && <Info className="w-5 h-5 text-cyan-400 shrink-0 mt-0.5" />}

            <div className="flex-1 truncate">
              <h5 className="text-xs font-bold">{n.title}</h5>
              <p className="text-xs text-slate-400 truncate mt-0.5">{n.message}</p>
            </div>

            <button
              onClick={() => setNotifications((prev) => prev.filter((item) => item.id !== n.id))}
              className="text-slate-400 hover:text-white p-1"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
