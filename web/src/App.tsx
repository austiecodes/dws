import { Link, Outlet } from "react-router-dom";

import { Button } from "./components/ui/button";
import { useAuth } from "./context/auth";

export default function App() {
  const { user, logout } = useAuth();

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="border-b bg-white">
        <div className="container flex h-16 items-center justify-between gap-4">
          <Link to="/" className="text-lg font-semibold">
            DWS 实验平台
          </Link>
          <nav className="flex items-center gap-2 text-sm font-medium">
            <Button variant="ghost" asChild>
              <Link to="/dashboard">控制台</Link>
            </Button>
            {user ? (
              <Button variant="ghost" asChild>
                <Link to="/containers">容器</Link>
              </Button>
            ) : null}
            {user ? (
              <Button variant="outline" size="sm" onClick={() => logout()}>
                退出
              </Button>
            ) : (
              <>
                <Button variant="ghost" size="sm" asChild>
                  <Link to="/login">登录</Link>
                </Button>
                <Button size="sm" asChild>
                  <Link to="/register">注册</Link>
                </Button>
              </>
            )}
          </nav>
        </div>
      </header>
      <main className="flex flex-1 flex-col bg-muted/20">
        <div className="container flex-1 py-10">
          <Outlet />
        </div>
      </main>
      <footer className="border-t bg-white text-center text-xs text-muted-foreground">
        <div className="container py-4">
          © {new Date().getFullYear()} Deep Web Service Lab Platform
        </div>
      </footer>
    </div>
  );
}
