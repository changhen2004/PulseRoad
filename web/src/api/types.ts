export interface User {
  id: number;
  email: string;
  name: string;
  created_at: string;
}

export interface AuthResult {
  token: string;
  user: User;
}

export interface RegisterPayload {
  email: string;
  password: string;
  name: string;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface Team {
  id: number;
  name: string;
  description: string;
  created_by: number;
  role: string;
  created_at: string;
}

export interface CreateTeamPayload {
  name: string;
  description: string;
}

export interface Product {
  id: number;
  team_id: number;
  name: string;
  description: string;
  created_by: number;
  created_at: string;
}

export interface CreateProductPayload {
  name: string;
  description: string;
}
