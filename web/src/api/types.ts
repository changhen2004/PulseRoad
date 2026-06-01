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

export type FeedbackStatus = 'open' | 'resolved';

export interface Feedback {
  id: number;
  product_id: number;
  title: string;
  content: string;
  status: FeedbackStatus;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface CreateFeedbackPayload {
  title: string;
  content: string;
}

export interface UpdateFeedbackStatusPayload {
  status: FeedbackStatus;
}

export interface FeatureFlag {
  id: number;
  product_id: number;
  key: string;
  name: string;
  description: string;
  environment: string;
  enabled: boolean;
  rollout_percentage: number;
  created_by: number;
  created_at: string;
  updated_at: string;
}

export interface CreateFeatureFlagPayload {
  key: string;
  name: string;
  description: string;
  environment: string;
  rollout_percentage: number;
}

export interface UpdateFeatureFlagPayload {
  key?: string;
  name?: string;
  description: string;
  environment?: string;
  rollout_percentage: number;
}

export interface ToggleFeatureFlagPayload {
  enabled: boolean;
}

export interface EvaluateFeatureFlagPayload {
  product_id: number;
  key: string;
  environment: string;
  user_key: string;
}

export interface EvaluateFeatureFlagResult {
  key: string;
  environment: string;
  enabled: boolean;
  rollout_percentage: number;
  reason: string;
}
