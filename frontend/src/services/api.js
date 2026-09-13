const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

async function request(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(url, config);
    const result = await response.json();

    if (!response.ok) {
      const errorMsg = result?.error?.message || `HTTP error! status: ${response.status}`;
      const err = new Error(errorMsg);
      err.code = result?.error?.code;
      err.details = result?.error?.details;
      throw err;
    }

    return result.data !== undefined ? result.data : result;
  } catch (err) {
    console.error(`[API_ERROR] ${options.method || 'GET'} ${endpoint}:`, err);
    throw err;
  }
}

export const api = {
  // Windows
  getWindows: () => request('/windows'),
  getWindowById: (id) => request(`/windows/${id}`),

  // Playlists
  getPlaylist: (windowId) => request(`/windows/${windowId}/playlist`),
  addToPlaylist: (windowId, data) =>
    request(`/windows/${windowId}/playlist`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  updatePlaylistItem: (windowId, itemId, data) =>
    request(`/windows/${windowId}/playlist/${itemId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  removeFromPlaylist: (windowId, itemId) =>
    request(`/windows/${windowId}/playlist/${itemId}`, {
      method: 'DELETE',
    }),

  // Media
  getMedia: () => request('/media'),
  getMediaById: (id) => request(`/media/${id}`),
  createMedia: (data) =>
    request('/media', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  deleteMedia: (id) =>
    request(`/media/${id}`, {
      method: 'DELETE',
    }),

  // Sync
  startSync: (data) =>
    request('/sync', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  getCurrentSync: () => request('/sync/current'),
};
