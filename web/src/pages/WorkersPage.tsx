import { useState, useEffect } from 'react';
import { workersApi, type Worker, type CreateWorkerRequest, type UpdateWorkerRequest } from '../lib/workers';
import { Button } from '../components/ui/button';
import { Card } from '../components/ui/card';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Alert } from '../components/ui/alert';
import { useAuth } from '../context/auth';

export default function WorkersPage() {
  const { user } = useAuth();
  const [workers, setWorkers] = useState<Worker[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingWorker, setEditingWorker] = useState<Worker | null>(null);

  useEffect(() => {
    loadWorkers();
  }, []);

  const loadWorkers = async () => {
    try {
      setLoading(true);
      const data = await workersApi.list();
      setWorkers(data);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load workers');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async (data: CreateWorkerRequest) => {
    try {
      await workersApi.create(data);
      setShowCreateForm(false);
      await loadWorkers();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create worker');
    }
  };

  const handleUpdate = async (id: string, data: UpdateWorkerRequest) => {
    try {
      await workersApi.update(id, data);
      setEditingWorker(null);
      await loadWorkers();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update worker');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(`Are you sure you want to delete worker ${id}?`)) {
      return;
    }
    try {
      await workersApi.delete(id);
      await loadWorkers();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete worker');
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return 'bg-green-500';
      case 'offline':
        return 'bg-gray-500';
      case 'maintenance':
        return 'bg-yellow-500';
      default:
        return 'bg-gray-500';
    }
  };

  if (loading) {
    return (
      <div className="container mx-auto p-6">
        <p>Loading workers...</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Worker Nodes</h1>
        {user?.is_admin && (
          <Button onClick={() => setShowCreateForm(true)}>Add Worker</Button>
        )}
      </div>

      {error && (
        <Alert className="mb-4 bg-red-50 text-red-900 border-red-200">
          {error}
        </Alert>
      )}

      {showCreateForm && (
        <WorkerForm
          onSubmit={handleCreate}
          onCancel={() => setShowCreateForm(false)}
        />
      )}

      {editingWorker && (
        <WorkerForm
          worker={editingWorker}
          onSubmit={(data) => handleUpdate(editingWorker.id, data)}
          onCancel={() => setEditingWorker(null)}
        />
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {workers.map((worker) => (
          <Card key={worker.id} className="p-4">
            <div className="flex justify-between items-start mb-3">
              <div>
                <h3 className="font-bold text-lg">{worker.name}</h3>
                <p className="text-sm text-gray-500">{worker.id}</p>
              </div>
              <div className="flex items-center gap-2">
                <span className={`w-3 h-3 rounded-full ${getStatusColor(worker.status)}`} />
                <span className="text-sm capitalize">{worker.status}</span>
              </div>
            </div>

            <div className="space-y-2 text-sm mb-4">
              <div>
                <span className="font-medium">Address:</span>
                <span className="ml-2 text-gray-600">{worker.address}</span>
              </div>
              <div>
                <span className="font-medium">Containers:</span>
                <span className="ml-2 text-gray-600">{worker.container_count || 0}</span>
              </div>
              {worker.last_heartbeat && (
                <div>
                  <span className="font-medium">Last Heartbeat:</span>
                  <span className="ml-2 text-gray-600">
                    {new Date(worker.last_heartbeat).toLocaleString()}
                  </span>
                </div>
              )}
            </div>

            {user?.is_admin && (
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setEditingWorker(worker)}
                >
                  Edit
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleDelete(worker.id)}
                  className="text-red-600 hover:text-red-700"
                >
                  Delete
                </Button>
              </div>
            )}
          </Card>
        ))}
      </div>

      {workers.length === 0 && (
        <div className="text-center py-12 text-gray-500">
          No workers registered yet.
        </div>
      )}
    </div>
  );
}

interface WorkerFormProps {
  worker?: Worker;
  onSubmit: (data: any) => void;
  onCancel: () => void;
}

function WorkerForm({ worker, onSubmit, onCancel }: WorkerFormProps) {
  const [id, setId] = useState(worker?.id || '');
  const [name, setName] = useState(worker?.name || '');
  const [address, setAddress] = useState(worker?.address || '');
  const [status, setStatus] = useState(worker?.status || 'offline');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (worker) {
      onSubmit({ name, address, status });
    } else {
      onSubmit({ id, name, address });
    }
  };

  return (
    <Card className="p-6 mb-6">
      <h2 className="text-xl font-bold mb-4">
        {worker ? 'Edit Worker' : 'Create Worker'}
      </h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        {!worker && (
          <div>
            <Label htmlFor="id">Worker ID</Label>
            <Input
              id="id"
              value={id}
              onChange={(e) => setId(e.target.value)}
              placeholder="worker-gpu-01"
              required
              disabled={!!worker}
            />
            <p className="text-sm text-gray-500 mt-1">
              Unique identifier (e.g., worker-gpu-01)
            </p>
          </div>
        )}

        <div>
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="GPU Worker 01"
            required
          />
        </div>

        <div>
          <Label htmlFor="address">Address</Label>
          <Input
            id="address"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder="192.168.1.100:50051"
            required
          />
          <p className="text-sm text-gray-500 mt-1">
            gRPC endpoint (host:port)
          </p>
        </div>

        {worker && (
          <div>
            <Label htmlFor="status">Status</Label>
            <select
              id="status"
              value={status}
              onChange={(e) => setStatus(e.target.value as any)}
              className="w-full border border-gray-300 rounded-md px-3 py-2"
            >
              <option value="online">Online</option>
              <option value="offline">Offline</option>
              <option value="maintenance">Maintenance</option>
            </select>
          </div>
        )}

        <div className="flex gap-2">
          <Button type="submit">
            {worker ? 'Update' : 'Create'}
          </Button>
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
        </div>
      </form>
    </Card>
  );
}

