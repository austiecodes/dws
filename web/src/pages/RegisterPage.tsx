import { FormEvent, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import { Alert, AlertDescription, AlertTitle } from "../components/ui/alert";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Label } from "../components/ui/label";
import { useAuth } from "../context/auth";
import { ApiError, post } from "../lib/api";

export default function RegisterPage() {
  const navigate = useNavigate();
  const { refresh } = useAuth();

  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await post<{ user: unknown }>("/api/v1/auth/register", {
        email,
        display_name: displayName,
        password,
      });
      await refresh();
      navigate("/dashboard", { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("注册失败，请稍后再试。");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="mx-auto max-w-md">
      <Card className="mt-10">
        <CardHeader className="space-y-2 text-center">
          <CardTitle className="text-3xl font-semibold">创建账号</CardTitle>
          <p className="text-sm text-muted-foreground">
            填写基础信息，我们会为你创建一个隔离实验环境。
          </p>
        </CardHeader>
        <CardContent>
          {error ? (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>注册失败</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          ) : null}
          <form className="space-y-4" onSubmit={handleSubmit}>
            <div className="space-y-2">
              <Label htmlFor="displayName">昵称</Label>
              <Input
                id="displayName"
                placeholder="实验小队"
                value={displayName}
                onChange={(event) => setDisplayName(event.target.value)}
                required
              />
            </div>
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
                placeholder="至少 8 位字符"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
                minLength={8}
                autoComplete="new-password"
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "创建中..." : "立即注册"}
            </Button>
          </form>
          <p className="mt-6 text-center text-sm text-muted-foreground">
            已经有账号？{" "}
            <Button variant="link" size="sm" className="px-0" asChild>
              <Link to="/login">直接登录</Link>
            </Button>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

