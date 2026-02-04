import { Link } from "react-router-dom";

import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { useAuth } from "../context/auth";

export default function LandingPage() {
  const { user } = useAuth();

  return (
    <div className="mx-auto max-w-4xl space-y-8 py-10">
      <div className="space-y-4 text-center">
        <h1 className="text-4xl font-bold tracking-tight">
          Deep Learning Lab 任务调度平台
        </h1>
        <p className="text-lg text-muted-foreground">
          多用户远程开发 · 资源隔离 · 实验任务排队与优先级控制。管理你的 GPU 实验、查看任务状态、快速切换队列。
        </p>
        <div className="flex justify-center gap-3">
          {user ? (
            <Button size="lg" asChild>
              <Link to="/dashboard">进入控制台</Link>
            </Button>
          ) : (
            <>
              <Button size="lg" asChild>
                <Link to="/login">登录平台</Link>
              </Button>
              <Button size="lg" variant="outline" asChild>
                <Link to="/register">注册账号</Link>
              </Button>
            </>
          )}
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>资源隔离</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            每个任务自动创建容器，隔离环境互不影响，支持自定义镜像与预热策略。
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>任务调度</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            队列 + 优先级机制，配合插队能力，确保紧急实验及时运行。
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>全流程追踪</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            平台 API 与前端实时展示运行日志、容器信息及资源使用，方便复现与调试。
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

