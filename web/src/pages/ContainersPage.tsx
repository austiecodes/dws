import { useEffect, useMemo, useState } from "react";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { useAuth } from "../context/auth";
import { ApiError } from "../lib/api";
import {
  createContainer,
  deleteContainer,
  fetchContainerImages,
  fetchContainers,
  startContainer,
  stopContainer,
} from "../lib/containers";
import type { Container } from "../types";

export default function ContainersPage() {
  const { user, loading: authLoading } = useAuth();
  const [containers, setContainers] = useState<Container[]>([]);
  const [images, setImages] = useState<string[]>([]);
  const [selectedImage, setSelectedImage] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    if (authLoading || !user) {
      return;
    }

    const load = async () => {
      try {
        setLoading(true);
        const [availableImages, existingContainers] = await Promise.all([
          fetchContainerImages(),
          fetchContainers(),
        ]);
        setImages(availableImages);
        setContainers(existingContainers);
        if (availableImages.length > 0) {
          setSelectedImage((image) => image || availableImages[0]);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "加载容器信息失败");
      } finally {
        setLoading(false);
      }
    };

    void load();
  }, [authLoading, user]);

  const handleCreate = async () => {
    if (!selectedImage) {
      setError("请选择镜像");
      return;
    }

    try {
      setCreating(true);
      setError(null);
      const container = await createContainer(selectedImage, password || undefined);
      setContainers((list) => [container, ...list]);
      setPassword(""); // 清空密码输入
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("创建容器失败，请稍后再试");
      }
    } finally {
      setCreating(false);
    }
  };

  const handleStop = async (uuid: string) => {
    try {
      setError(null);
      await stopContainer(uuid);
      // 更新本地状态
      setContainers((list) =>
        list.map((c) => (c.uuid === uuid ? { ...c, status: "stopped" } : c))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "停止容器失败");
    }
  };

  const handleStart = async (uuid: string) => {
    try {
      setError(null);
      await startContainer(uuid);
      // 更新本地状态
      setContainers((list) =>
        list.map((c) => (c.uuid === uuid ? { ...c, status: "running" } : c))
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "启动容器失败");
    }
  };

  const handleDelete = async (uuid: string) => {
    if (!confirm("确定要删除此容器吗？此操作不可恢复。")) {
      return;
    }

    try {
      setError(null);
      await deleteContainer(uuid);
      // 从列表中移除
      setContainers((list) => list.filter((c) => c.uuid !== uuid));
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除容器失败");
    }
  };

  const hasImages = images.length > 0;

  const description = useMemo(() => {
    if (!hasImages) {
      return "管理员尚未配置可用镜像，请联系平台维护人员";
    }
    return "选择镜像后将自动创建容器，并把 22 端口映射到宿主机供 SSH/VSCode 使用";
  }, [hasImages]);

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl font-semibold">创建容器</CardTitle>
          <p className="text-sm text-muted-foreground">{description}</p>
        </CardHeader>
        <CardContent className="space-y-4">
          {error ? (
            <Alert variant="destructive">
              <AlertTitle>操作失败</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}

          <div className="space-y-2">
            <label className="block text-sm font-medium text-muted-foreground" htmlFor="image">
              选择镜像
            </label>
            <select
              id="image"
              value={selectedImage}
              onChange={(event) => setSelectedImage(event.target.value)}
              className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              disabled={!hasImages || creating}
            >
              {images.map((image) => (
                <option key={image} value={image}>
                  {image}
                </option>
              ))}
            </select>
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-muted-foreground" htmlFor="password">
              SSH 密码（可选）
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="留空使用默认密码 dws"
              className="h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              disabled={creating}
            />
            <p className="text-xs text-muted-foreground">
              自定义密码后请妥善保管，平台不会存储您的密码
            </p>
          </div>

          <Button onClick={handleCreate} disabled={!hasImages || creating} className="w-full md:w-auto">
            {creating ? "创建中..." : "创建容器"}
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-xl">我的容器</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p className="text-sm text-muted-foreground">正在加载 ...</p>
          ) : containers.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无容器，创建一个开始实验。</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-border text-sm">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="px-4 py-2 text-left font-medium text-muted-foreground">名称</th>
                    <th className="px-4 py-2 text-left font-medium text-muted-foreground">镜像</th>
                    <th className="px-4 py-2 text-left font-medium text-muted-foreground">状态</th>
                    <th className="px-4 py-2 text-left font-medium text-muted-foreground">SSH 连接</th>
                    <th className="px-4 py-2 text-left font-medium text-muted-foreground">创建时间</th>
                    <th className="px-4 py-2 text-left font-medium text-muted-foreground">操作</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  {containers.map((container) => {
                    const sshCommand = `ssh root@localhost -p ${container.host_ssh_port}`;
                    const passwordHint = "密码: ***";
                    return (
                      <tr key={container.uuid}>
                        <td className="px-4 py-2 font-medium">{container.name}</td>
                        <td className="px-4 py-2">{container.image}</td>
                        <td className="px-4 py-2">
                          <span
                            className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
                              container.status === "running"
                                ? "bg-green-100 text-green-700"
                                : container.status === "stopped"
                                ? "bg-yellow-100 text-yellow-700"
                                : "bg-gray-100 text-gray-700"
                            }`}
                          >
                            {container.status}
                          </span>
                        </td>
                        <td className="px-4 py-2">
                          {container.status === "running" ? (
                            <div className="space-y-1">
                              <code className="block rounded bg-muted px-2 py-1 text-xs">
                                {sshCommand}
                              </code>
                              <p className="text-xs text-muted-foreground">{passwordHint}</p>
                            </div>
                          ) : (
                            <span className="text-xs text-muted-foreground">容器未运行</span>
                          )}
                        </td>
                        <td className="px-4 py-2 text-muted-foreground">
                          {new Date(container.created_at).toLocaleString()}
                        </td>
                        <td className="px-4 py-2">
                          <div className="flex gap-2">
                            {container.status === "running" && (
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={() => handleStop(container.uuid)}
                              >
                                停止
                              </Button>
                            )}
                            {container.status === "stopped" && (
                              <Button
                                size="sm"
                                variant="default"
                                onClick={() => handleStart(container.uuid)}
                              >
                                启动
                              </Button>
                            )}
                            <Button
                              size="sm"
                              variant="destructive"
                              onClick={() => handleDelete(container.uuid)}
                            >
                              删除
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
