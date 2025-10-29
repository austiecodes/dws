import { FormEvent, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { useAuth } from "../context/auth";
import { ApiError, post } from "../lib/api";

export default function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { refresh } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await post<{ user: unknown }>("/api/v1/auth/login", {
        email,
        password,
      });
      await refresh();
      const target =
        (location.state as { from?: { pathname?: string } } | undefined)?.from
          ?.pathname ?? "/dashboard";
      navigate(target, { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("登录失败，请稍后再试。");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-md">
      <Card className="mt-10">
        <CardHeader className="space-y-2 text-center">
          <CardTitle className="text-3xl font-semibold">登录平台</CardTitle>
          <p className="text-sm text-muted-foreground">
            输入绑定的邮箱和密码，继续你的实验任务。
          </p>
        </CardHeader>
        <CardContent>
          {error ? (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>登录失败</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          <form className="space-y-4" onSubmit={handleSubmit}>
            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <Input
                id="email"
                type="email"
                placeholder="you@example.com"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                required
                autoComplete="email"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">密码</Label>
              <Input
                id="password"
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
                autoComplete="current-password"
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "登录中..." : "立即登录"}
            </Button>
          </form>
          <p className="mt-6 text-center text-sm text-muted-foreground">
            还没有账号？{" "}
            <Button variant="link" size="sm" className="px-0" asChild>
              <Link to="/register">注册一个</Link>
            </Button>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
