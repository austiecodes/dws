import { get, post } from './api';

export interface Worker {
  id: string;
  name: string;
  address: string;
  status: 'online' | 'offline' | 'maintenance';
  last_heartbeat?: string;
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
  container_count?: number;
}

export interface CreateWorkerRequest {
  id: string;
  name: string;
  address: string;
  metadata?: Record<string, any>;
}

export interface UpdateWorkerRequest {
  name?: string;
  address?: string;
  status?: 'online' | 'offline' | 'maintenance';
  metadata?: Record<string, any>;
}

async function apiDelete(path: string): Promise<void> {
  const response = await fetch(path, {
    method: 'DELETE',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
    },
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || 'Request failed');
  }
}

async function apiPut<T>(path: string, body: unknown): Promise<T> {
  const response = await fetch(path, {
    method: 'PUT',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    const data = await response.json().catch(() => ({}));
    throw new Error(data.error || 'Request failed');
  }
  return response.json();
}

export const workersApi = {
  list: async (): Promise<Worker[]> => {
    const response = await get<{ workers: Worker[] }>('/api/v1/workers');
    return response.workers || [];
  },

  get: async (id: string): Promise<Worker> => {
    const response = await get<{ worker: Worker }>(`/api/v1/workers/${id}`);
    return response.worker;
  },

  create: async (data: CreateWorkerRequest): Promise<Worker> => {
    const response = await post<{ worker: Worker }>('/api/v1/workers', data);
    return response.worker;
  },

  update: async (id: string, data: UpdateWorkerRequest): Promise<void> => {
    await apiPut(`/api/v1/workers/${id}`, data);
  },

  delete: async (id: string): Promise<void> => {
    await apiDelete(`/api/v1/workers/${id}`);
  },
};

