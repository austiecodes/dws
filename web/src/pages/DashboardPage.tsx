import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { useAuth } from "../context/auth";

export default function DashboardPage() {
  const { user, logout, refresh } = useAuth();

  if (!user) {
    return null;
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl font-semibold">
            欢迎回来，{user.display_name}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 text-sm text-muted-foreground">
          <p>当前登录邮箱：{user.email}</p>
          <p>
            角色：
            {user.is_admin ? "管理员（可以管理所有任务与资源）" : "普通实验用户"}
          </p>
          <div className="flex gap-2">
            <Button variant="secondary" onClick={() => refresh()}>
              刷新会话
            </Button>
            <Button variant="ghost" onClick={() => logout()}>
              退出登录
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

