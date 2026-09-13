import React, { useState } from 'react';
import { X, Plus, Trash2, Image, Video, MonitorOff, CheckCircle2 } from 'lucide-react';
import { api } from '../services/api';

export function MediaLibraryModal({ isOpen, onClose, mediaList, onMediaCreated, onMediaDeleted }) {
  const [name, setName] = useState('');
  const [type, setType] = useState('image');
  const [url, setUrl] = useState('');
  const [duration, setDuration] = useState(15);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    try {
      const created = await api.createMedia({
        name,
        type,
        url: type === 'blank' ? 'blank://black' : url,
        duration: parseInt(duration, 10),
      });
      onMediaCreated(created);
      setName('');
      setUrl('');
      setDuration(15);
    } catch (err) {
      setError(err.message || 'Failed to create media item');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this media asset?')) return;
    try {
      await api.deleteMedia(id);
      onMediaDeleted(id);
    } catch (err) {
      alert(err.message || 'Failed to delete media');
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-slate-900 border border-slate-700 rounded-2xl w-full max-w-3xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 bg-slate-950 border-b border-slate-800">
          <div>
            <h3 className="text-base font-bold text-white">Media Asset Library</h3>
            <p className="text-xs text-slate-400 mt-0.5">
              Manage images, videos, and blank media used across display sequences
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Media Grid */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Add New Media Form */}
          <div className="p-4 rounded-xl bg-slate-950/80 border border-slate-800">
            <h4 className="text-sm font-semibold text-white mb-3 flex items-center gap-2">
              <Plus className="w-4 h-4 text-cyan-400" />
              Register New Media Asset
            </h4>

            {error && (
              <div className="mb-3 p-2.5 rounded bg-red-950/50 border border-red-800/50 text-red-300 text-xs">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="grid grid-cols-1 md:grid-cols-12 gap-3">
              <div className="md:col-span-4">
                <label className="block text-[11px] font-medium text-slate-400 mb-1">Asset Name</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. City Lights"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-cyan-500"
                />
              </div>

              <div className="md:col-span-2">
                <label className="block text-[11px] font-medium text-slate-400 mb-1">Type</label>
                <select
                  value={type}
                  onChange={(e) => setType(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-2 py-1.5 text-xs text-white focus:outline-none focus:border-cyan-500"
                >
                  <option value="image">Image</option>
                  <option value="video">Video</option>
                  <option value="blank">Blank</option>
                </select>
              </div>

              {type !== 'blank' && (
                <div className="md:col-span-4">
                  <label className="block text-[11px] font-medium text-slate-400 mb-1">Media URL</label>
                  <input
                    type="url"
                    required
                    placeholder="https://...mp4 or .jpg"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-cyan-500"
                  />
                </div>
              )}

              <div className={type === 'blank' ? 'md:col-span-4' : 'md:col-span-2'}>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">Duration (s)</label>
                <input
                  type="number"
                  required
                  min="1"
                  max="18000"
                  value={duration}
                  onChange={(e) => setDuration(Math.max(1, parseInt(e.target.value, 10) || 1))}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-xs text-white focus:outline-none focus:border-cyan-500 font-mono"
                />
              </div>

              <div className="md:col-span-12 flex justify-end">
                <button
                  type="submit"
                  disabled={isSubmitting}
                  className="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white font-semibold text-xs transition-colors shadow-lg shadow-cyan-600/20"
                >
                  {isSubmitting ? 'Saving...' : 'Save Media Asset'}
                </button>
              </div>
            </form>
          </div>

          {/* Existing Media Items */}
          <div>
            <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
              Existing Media Assets ({mediaList.length})
            </h4>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {mediaList.map((m) => (
                <div
                  key={m.id}
                  className="flex items-center justify-between p-3 rounded-xl bg-slate-950/60 border border-slate-800"
                >
                  <div className="flex items-center gap-3 truncate">
                    <div className="w-10 h-10 rounded-lg bg-slate-900 border border-slate-800 flex items-center justify-center shrink-0">
                      {m.type === 'image' && <Image className="w-5 h-5 text-blue-400" />}
                      {m.type === 'video' && <Video className="w-5 h-5 text-purple-400" />}
                      {m.type === 'blank' && <MonitorOff className="w-5 h-5 text-slate-500" />}
                    </div>

                    <div className="truncate">
                      <p className="text-xs font-semibold text-white truncate">{m.name}</p>
                      <p className="text-[10px] font-mono text-slate-400">
                        {m.type.toUpperCase()} • {m.duration}s duration
                      </p>
                    </div>
                  </div>

                  <button
                    onClick={() => handleDelete(m.id)}
                    className="p-1.5 rounded text-slate-500 hover:text-red-400 hover:bg-red-950/30 transition-colors"
                    title="Delete Media"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
