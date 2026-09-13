import React, { useState } from 'react';
import { X, Plus, Trash2, ArrowUp, ArrowDown, Image, Video, MonitorOff, Clock, Sparkles } from 'lucide-react';
import { formatTime } from '../utils/playback';

export function PlaylistEditor({ window, mediaList, onClose, onAddItem, onRemoveItem, onUpdateItem }) {
  const [selectedMediaId, setSelectedMediaId] = useState('');
  const [isAdding, setIsAdding] = useState(false);

  if (!window) return null;

  const playlist = window.playlist || [];
  const totalDuration = playlist.reduce((sum, item) => sum + (item.media?.duration || 0), 0);

  const handleAddMedia = async (e) => {
    e.preventDefault();
    if (!selectedMediaId) return;
    setIsAdding(true);
    try {
      await onAddItem(window.id, parseInt(selectedMediaId, 10));
      setSelectedMediaId('');
    } finally {
      setIsAdding(false);
    }
  };

  const handleMoveUp = async (index) => {
    if (index <= 0) return;
    const currentItem = playlist[index];
    const prevItem = playlist[index - 1];
    await onUpdateItem(window.id, currentItem.id, prevItem.position);
  };

  const handleMoveDown = async (index) => {
    if (index >= playlist.length - 1) return;
    const currentItem = playlist[index];
    const nextItem = playlist[index + 1];
    await onUpdateItem(window.id, currentItem.id, nextItem.position);
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-700 rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 bg-slate-950 border-b border-slate-800">
          <div>
            <h3 className="text-base font-bold text-white flex items-center gap-2">
              <span>{window.name}</span>
              <span className="text-xs font-normal text-slate-400 font-mono">
                ({playlist.length} items • Loop: {formatTime(totalDuration)})
              </span>
            </h3>
            <p className="text-xs text-slate-400 mt-0.5">
              Configure ordered media playback sequence for this window
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Playlist Items List */}
        <div className="flex-1 overflow-y-auto p-6 space-y-3">
          {playlist.length === 0 ? (
            <div className="text-center py-12 text-slate-500 border border-dashed border-slate-800 rounded-xl">
              <MonitorOff className="w-10 h-10 mx-auto mb-2 text-slate-600" />
              <p className="text-sm font-medium">Playlist is empty</p>
              <p className="text-xs text-slate-600 mt-1">Add media assets using the selector below.</p>
            </div>
          ) : (
            playlist.map((item, index) => (
              <div
                key={item.id}
                className="flex items-center justify-between p-3 rounded-xl bg-slate-950/70 border border-slate-800 hover:border-slate-700 transition-colors group"
              >
                <div className="flex items-center gap-3 truncate">
                  <div className="w-7 h-7 rounded-lg bg-slate-800 flex items-center justify-center text-xs font-mono text-slate-300 shrink-0">
                    {index + 1}
                  </div>

                  <div className="w-10 h-10 rounded-lg bg-black/60 border border-slate-800 overflow-hidden flex items-center justify-center shrink-0">
                    {item.media?.type === 'image' && <Image className="w-5 h-5 text-blue-400" />}
                    {item.media?.type === 'video' && <Video className="w-5 h-5 text-purple-400" />}
                    {item.media?.type === 'blank' && <MonitorOff className="w-5 h-5 text-slate-500" />}
                  </div>

                  <div className="truncate">
                    <p className="text-sm font-semibold text-white truncate">
                      {item.media?.name || 'Untitled Media'}
                    </p>
                    <div className="flex items-center gap-2 text-xs font-mono text-slate-400">
                      <span className="capitalize">{item.media?.type}</span>
                      <span>•</span>
                      <span className="text-emerald-400">{item.media?.duration}s duration</span>
                    </div>
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-1 shrink-0">
                  <button
                    onClick={() => handleMoveUp(index)}
                    disabled={index === 0}
                    className="p-1.5 rounded text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Move Up"
                  >
                    <ArrowUp className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleMoveDown(index)}
                    disabled={index === playlist.length - 1}
                    className="p-1.5 rounded text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed"
                    title="Move Down"
                  >
                    <ArrowDown className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => onRemoveItem(window.id, item.id)}
                    className="p-1.5 rounded text-red-400 hover:text-red-300 hover:bg-red-950/40 ml-1 transition-colors"
                    title="Remove Item"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))
          )}
        </div>

        {/* Add Media Form Footer */}
        <div className="p-4 bg-slate-950 border-t border-slate-800">
          <form onSubmit={handleAddMedia} className="flex gap-3">
            <select
              value={selectedMediaId}
              onChange={(e) => setSelectedMediaId(e.target.value)}
              className="flex-1 bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-cyan-500"
              required
            >
              <option value="" disabled>
                -- Select media asset to append --
              </option>
              {mediaList.map((media) => (
                <option key={media.id} value={media.id}>
                  [{media.type.toUpperCase()}] {media.name} ({media.duration}s)
                </option>
              ))}
            </select>

            <button
              type="submit"
              disabled={!selectedMediaId || isAdding}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-cyan-600/20"
            >
              <Plus className="w-4 h-4" />
              {isAdding ? 'Adding...' : 'Add to Playlist'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
