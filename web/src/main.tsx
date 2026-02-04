import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import App from "./App";
import { RequireAuth } from "./components/RequireAuth";
import { AuthProvider } from "./context/auth";
import "./index.css";
import DashboardPage from "./pages/DashboardPage";
import LandingPage from "./pages/LandingPage";
import LoginPage from "./pages/LoginPage";
import RegisterPage from "./pages/RegisterPage";
import ContainersPage from "./pages/ContainersPage";
import TasksPage from "./pages/TasksPage";
import WorkersPage from "./pages/WorkersPage";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<App />}>
            <Route index element={<LandingPage />} />
            <Route path="login" element={<LoginPage />} />
            <Route path="register" element={<RegisterPage />} />
            <Route element={<RequireAuth />}>
              <Route path="dashboard" element={<DashboardPage />} />
              <Route path="containers" element={<ContainersPage />} />
              <Route path="tasks" element={<TasksPage />} />
              <Route path="workers" element={<WorkersPage />} />
            </Route>
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>
);
