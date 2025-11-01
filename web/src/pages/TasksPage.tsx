import { useEffect, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { useAuth } from "../context/auth";
import { ApiError } from "../lib/api";
import { fetchContainers } from "../lib/containers";
import { createTask, cancelTask, fetchTasks, fetchQueueStatistics } from "../lib/tasks";
import type { Container, Task, TaskType, QueueStatistics } from "../types";

export default function TasksPage() {
  const { user, loading: authLoading } = useAuth();
  const [tasks, setTasks] = useState<Task[]>([]);
  const [containers, setContainers] = useState<Container[]>([]);
  const [queueStats, setQueueStats] = useState<QueueStatistics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const [selectedContainerId, setSelectedContainerId] = useState<number>(0);
  const [command, setCommand] = useState("");
  const [taskType, setTaskType] = useState<TaskType>("cpu");
  const [hours, setHours] = useState<number>(0);
  const [minutes, setMinutes] = useState<number>(5);
  const [priority, setPriority] = useState<number>(0);

  useEffect(() => {
    if (authLoading || !user) {
      return;
    }

    const load = async () => {
      try {
        setLoading(true);
        const [fetchedContainers, fetchedTasks, stats] = await Promise.all([
          fetchContainers(),
          fetchTasks(),
          fetchQueueStatistics(),
        ]);
        setContainers(fetchedContainers);
        setTasks(fetchedTasks);
        setQueueStats(stats);
        if (fetchedContainers.length > 0) {
          setSelectedContainerId(fetchedContainers[0].id);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "加载任务信息失败");
      } finally {
        setLoading(false);
      }
    };

    void load();
  }, [authLoading, user]);

  const handleCreateTask = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedContainerId || !command.trim()) {
      setError("请选择容器并输入命令");
      return;
    }

    const expectedDuration = hours * 3600 + minutes * 60;
    if (expectedDuration <= 0) {
      setError("预期时长必须大于 0");
      return;
    }

    try {
      setSubmitting(true);
      setError(null);
      const task = await createTask({
        container_id: selectedContainerId,
        command: command.trim(),
        task_type: taskType,
        expected_duration: expectedDuration,
        priority,
      });
      setTasks((list) => [task, ...list]);
      
      // Refresh queue stats
      const stats = await fetchQueueStatistics();
      setQueueStats(stats);
      
      // Reset form
      setCommand("");
      setHours(0);
      setMinutes(5);
      setPriority(0);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("提交任务失败，请稍后再试");
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleCancelTask = async (taskId: number) => {
    try {
      setError(null);
      await cancelTask(taskId);
      setTasks((list) =>
        list.map((t) => (t.id === taskId ? { ...t, status: "killed" as const } : t))
      );
      
      // Refresh queue stats
      const stats = await fetchQueueStatistics();
      setQueueStats(stats);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("取消任务失败");
      }
    }
  };

  const runningContainers = containers.filter((c) => c.status === "running");

  if (authLoading || loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-lg">加载中...</div>
      </div>
    );
  }

  const getStatusBadge = (status: Task["status"]) => {
    const colors = {
      pending: "bg-yellow-100 text-yellow-800",
      running: "bg-blue-100 text-blue-800",
      completed: "bg-green-100 text-green-800",
      failed: "bg-red-100 text-red-800",
      killed: "bg-gray-100 text-gray-800",
    };
    return (
      <span className={`px-2 py-1 rounded text-xs font-medium ${colors[status]}`}>
        {status}
      </span>
    );
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const hrs = Math.floor(mins / 60);
    const remainMins = mins % 60;
    
    if (hrs === 0) return `${mins} 分钟`;
    if (remainMins === 0) return `${hrs} 小时`;
    return `${hrs} 小时 ${remainMins} 分钟`;
  };

  const getTaskTypeBadge = (type: TaskType) => {
    const colors = {
      cpu: "bg-blue-100 text-blue-800",
      gpu: "bg-purple-100 text-purple-800",
    };
    const labels = {
      cpu: "CPU",
      gpu: "GPU",
    };
    return (
      <span className={`px-2 py-1 rounded text-xs font-medium ${colors[type]}`}>
        {labels[type]}
      </span>
    );
  };

  const formatTimestamp = (ts?: string) => {
    if (!ts) return "—";
    return new Date(ts).toLocaleString("zh-CN");
  };

  return (
    <div className="container mx-auto p-6 max-w-7xl">
      <h1 className="text-3xl font-bold mb-6">任务队列</h1>

      {error && (
        <Alert variant="destructive" className="mb-4">
          <AlertTitle>错误</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {queueStats && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>系统队列状态</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-6">
              <div className="space-y-2">
                <h3 className="font-semibold text-lg">CPU 任务</h3>
                <div className="space-y-1">
                  <p className="text-sm">
                    <span className="text-gray-600">运行中:</span>{" "}
                    <span className="font-medium">
                      {queueStats.running.cpu} / {queueStats.limits.cpu}
                    </span>
                  </p>
                  <p className="text-sm">
                    <span className="text-gray-600">等待中:</span>{" "}
                    <span className="font-medium">{queueStats.pending.cpu}</span>
                  </p>
                </div>
              </div>
              <div className="space-y-2">
                <h3 className="font-semibold text-lg">GPU 任务</h3>
                <div className="space-y-1">
                  <p className="text-sm">
                    <span className="text-gray-600">运行中:</span>{" "}
                    <span className="font-medium">
                      {queueStats.running.gpu} / {queueStats.limits.gpu}
                    </span>
                  </p>
                  <p className="text-sm">
                    <span className="text-gray-600">等待中:</span>{" "}
                    <span className="font-medium">{queueStats.pending.gpu}</span>
                  </p>
                </div>
              </div>
            </div>
            {(queueStats.pending.cpu > 0 || queueStats.pending.gpu > 0) && (
              <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded">
                <p className="text-sm text-yellow-800">
                  系统当前有 {queueStats.pending.total} 个任务在队列中等待，请耐心等待调度。
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      <Card className="mb-6">
        <CardHeader>
          <CardTitle>提交新任务</CardTitle>
        </CardHeader>
        <CardContent>
          {runningContainers.length === 0 ? (
            <Alert>
              <AlertDescription>
                暂无运行中的容器，请先创建并启动容器。
              </AlertDescription>
            </Alert>
          ) : (
            <form onSubmit={handleCreateTask} className="space-y-4">
              <div>
                <Label htmlFor="container">选择容器</Label>
                <select
                  id="container"
                  value={selectedContainerId}
                  onChange={(e) => setSelectedContainerId(Number(e.target.value))}
                  className="w-full p-2 border rounded"
                >
                  {runningContainers.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name} ({c.image})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <Label htmlFor="command">命令</Label>
                <Input
                  id="command"
                  value={command}
                  onChange={(e) => setCommand(e.target.value)}
                  placeholder="例如: python train.py"
                  required
                />
              </div>

              <div>
                <Label htmlFor="taskType">任务类型</Label>
                <select
                  id="taskType"
                  value={taskType}
                  onChange={(e) => setTaskType(e.target.value as TaskType)}
                  className="w-full p-2 border rounded"
                >
                  <option value="cpu">CPU 任务（最多 3 个并发）</option>
                  <option value="gpu">GPU 任务（独占执行）</option>
                </select>
                <p className="text-sm text-gray-500 mt-1">
                  GPU 任务同时只能运行 1 个，CPU 任务最多可同时运行 3 个
                </p>
              </div>

              <div>
                <Label>预期时长</Label>
                <div className="flex gap-3 items-center">
                  <div className="flex-1">
                    <Input
                      type="number"
                      placeholder="小时"
                      value={hours}
                      onChange={(e) => setHours(Number(e.target.value))}
                      min={0}
                    />
                  </div>
                  <span className="text-gray-500">小时</span>
                  <div className="flex-1">
                    <Input
                      type="number"
                      placeholder="分钟"
                      value={minutes}
                      onChange={(e) => setMinutes(Number(e.target.value))}
                      min={0}
                      max={59}
                    />
                  </div>
                  <span className="text-gray-500">分钟</span>
                </div>
                <p className="text-sm text-gray-500 mt-1">
                  总计: {formatDuration(hours * 3600 + minutes * 60)}
                </p>
              </div>

              <div>
                <Label htmlFor="priority">优先级（可选）</Label>
                <Input
                  id="priority"
                  type="number"
                  value={priority}
                  onChange={(e) => setPriority(Number(e.target.value))}
                />
                <p className="text-sm text-gray-500 mt-1">
                  数字越大优先级越高（默认为 0）
                </p>
              </div>

              <Button type="submit" disabled={submitting}>
                {submitting ? "提交中..." : "提交任务"}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>我的任务</CardTitle>
        </CardHeader>
        <CardContent>
          {tasks.length === 0 ? (
            <p className="text-gray-500">暂无任务</p>
          ) : (
            <div className="space-y-4">
              {tasks.map((task) => (
                <div key={task.id} className="border rounded p-4">
                  <div className="flex justify-between items-start mb-2">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="font-medium">任务 #{task.id}</span>
                        {getStatusBadge(task.status)}
                        {getTaskTypeBadge(task.task_type)}
                        {task.priority > 0 && (
                          <span className="text-xs text-gray-500">
                            优先级: {task.priority}
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-gray-700 mb-1">
                        <strong>命令:</strong> <code>{task.command}</code>
                      </p>
                      {task.container && (
                        <p className="text-sm text-gray-500 mb-1">
                          <strong>容器:</strong> {task.container.name} ({task.container.image})
                        </p>
                      )}
                      <p className="text-sm text-gray-500 mb-1">
                        <strong>预期时长:</strong> {formatDuration(task.expected_duration)}
                      </p>
                      <p className="text-sm text-gray-500 mb-1">
                        <strong>创建时间:</strong> {formatTimestamp(task.created_at)}
                      </p>
                      {task.started_at && (
                        <p className="text-sm text-gray-500 mb-1">
                          <strong>开始时间:</strong> {formatTimestamp(task.started_at)}
                        </p>
                      )}
                      {task.completed_at && (
                        <p className="text-sm text-gray-500 mb-1">
                          <strong>完成时间:</strong> {formatTimestamp(task.completed_at)}
                        </p>
                      )}
                      {task.exit_code !== undefined && task.exit_code !== null && (
                        <p className="text-sm text-gray-500 mb-1">
                          <strong>退出码:</strong> {task.exit_code}
                        </p>
                      )}
                      {task.output && (
                        <details className="mt-2">
                          <summary className="cursor-pointer text-sm text-blue-600">
                            查看输出
                          </summary>
                          <pre className="mt-2 p-2 bg-gray-50 border rounded text-xs overflow-x-auto">
                            {task.output}
                          </pre>
                        </details>
                      )}
                    </div>
                    {(task.status === "pending" || task.status === "running") && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleCancelTask(task.id)}
                      >
                        取消
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

