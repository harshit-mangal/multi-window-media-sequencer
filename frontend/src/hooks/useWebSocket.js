import { useState, useEffect } from 'react';
import { wsService } from '../services/websocket';

export function useWebSocket() {
  const [status, setStatus] = useState('disconnected');

  useEffect(() => {
    wsService.connect();
    const unsubscribe = wsService.onStatusChange((newStatus) => {
      setStatus(newStatus);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  return {
    status,
    isConnected: status === 'connected',
    wsService,
  };
}
