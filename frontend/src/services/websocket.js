const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';

class WebSocketService {
  constructor() {
    this.url = WS_URL;
    this.ws = null;
    this.listeners = new Map();
    this.statusListeners = new Set();
    this.reconnectAttempts = 0;
    this.maxReconnectDelay = 10000;
    this.reconnectTimer = null;
    this.status = 'disconnected'; // 'connecting', 'connected', 'disconnected'
  }

  connect() {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.setStatus('connecting');

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('[WEBSOCKET_CONNECTED] Connected to backend WebSocket');
        this.reconnectAttempts = 0;
        this.setStatus('connected');
      };

      this.ws.onmessage = (event) => {
        try {
          // Could receive multiple newline-delimited JSON frames
          const lines = event.data.split('\n');
          for (const line of lines) {
            if (!line.trim()) continue;
            const data = JSON.parse(line);
            this.handleMessage(data);
          }
        } catch (err) {
          console.error('[WEBSOCKET_ERROR] Error parsing message:', err, event.data);
        }
      };

      this.ws.onclose = (event) => {
        console.log('[WEBSOCKET_DISCONNECTED] Connection closed:', event.code, event.reason);
        this.setStatus('disconnected');
        this.scheduleReconnect();
      };

      this.ws.onerror = (error) => {
        console.error('[WEBSOCKET_ERROR] WebSocket encountered error:', error);
      };
    } catch (err) {
      console.error('[WEBSOCKET_ERROR] Failed to establish connection:', err);
      this.setStatus('disconnected');
      this.scheduleReconnect();
    }
  }

  scheduleReconnect() {
    if (this.reconnectTimer) return;

    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(1.5, this.reconnectAttempts), this.maxReconnectDelay);
    console.log(`[WEBSOCKET] Reconnecting in ${(delay / 1000).toFixed(1)}s (Attempt ${this.reconnectAttempts})...`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }

  handleMessage(msg) {
    if (!msg || !msg.type) return;

    const eventType = msg.type;
    const listeners = this.listeners.get(eventType);
    if (listeners) {
      listeners.forEach((callback) => {
        try {
          callback(msg.payload, msg);
        } catch (err) {
          console.error(`[WEBSOCKET_ERROR] Listener failed for event ${eventType}:`, err);
        }
      });
    }

    // Also trigger global wildcard listeners
    const globalListeners = this.listeners.get('*');
    if (globalListeners) {
      globalListeners.forEach((callback) => {
        try {
          callback(msg);
        } catch (err) {
          console.error('[WEBSOCKET_ERROR] Wildcard listener failed:', err);
        }
      });
    }
  }

  on(eventType, callback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set());
    }
    this.listeners.get(eventType).add(callback);

    // Return unbind function for easy React cleanup
    return () => {
      const set = this.listeners.get(eventType);
      if (set) {
        set.delete(callback);
        if (set.size === 0) {
          this.listeners.delete(eventType);
        }
      }
    };
  }

  onStatusChange(callback) {
    this.statusListeners.add(callback);
    callback(this.status);
    return () => {
      this.statusListeners.delete(callback);
    };
  }

  setStatus(status) {
    this.status = status;
    this.statusListeners.forEach((cb) => cb(status));
  }

  send(data) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(typeof data === 'string' ? data : JSON.stringify(data));
    } else {
      console.warn('[WEBSOCKET] Cannot send message: WebSocket not open');
    }
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.setStatus('disconnected');
  }
}

export const wsService = new WebSocketService();
