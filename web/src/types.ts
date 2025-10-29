export interface User {
  id: number;
  email: string;
  display_name: string;
  is_admin: boolean;
}

export interface Container {
  id: number;
  uuid: string;
  name: string;
  image: string;
  host_ssh_port: number;
  status: string;
  created_at: string;
}

export interface Task {
  id: number;
  user_id: number;
  container_id: number;
  command: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'killed';
  expected_duration: number;
  priority: number;
  started_at?: string;
  completed_at?: string;
  output: string;
  exit_code?: number;
  created_at: string;
  container?: Container;
}

export interface CreateTaskRequest {
  container_id: number;
  command: string;
  expected_duration: number;
  priority?: number;
}
