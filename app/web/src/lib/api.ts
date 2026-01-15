import { MachineStatus, DoseMode } from '@/types/status';

export const API_BASE = import.meta.env.DEV ? 'http://localhost:8080/api' : '/api';

export async function fetchStatus(): Promise<MachineStatus> {
  const response = await fetch(`${API_BASE}/status`);
  if (!response.ok) {
    throw new Error('Failed to fetch status');
  }
  return response.json();
}

export async function setMode(mode: DoseMode): Promise<void> {
  const response = await fetch(`${API_BASE}/mode`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ mode }),
  });
  if (!response.ok) {
    throw new Error('Failed to set mode');
  }
}
