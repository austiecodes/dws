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
