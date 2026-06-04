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

export type TeamRole = 'owner' | 'member';

export interface TeamMember {
  user_id: number;
  email: string;
  name: string;
  role: TeamRole;
  created_at: string;
}

export interface TeamInvitation {
  id: number;
  team_id: number;
  team_name: string;
  email: string;
  role: TeamRole;
  status: 'pending' | 'accepted';
  invited_by: number;
  accepted_at?: string;
  created_at: string;
}

export interface InviteTeamMemberPayload {
  email: string;
  role: TeamRole;
}

export interface UpdateTeamMemberRolePayload {
  role: TeamRole;
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

export interface ProductSummary {
  product: Product;
  feedback_total: number;
  feedback_open: number;
  feedback_resolved: number;
  comment_total: number;
  vote_total: number;
  flag_total: number;
  flag_enabled: number;
}

export type FeedbackStatus = 'open' | 'resolved';

export interface Feedback {
  id: number;
  product_id: number;
  title: string;
  content: string;
  status: FeedbackStatus;
  created_by: number;
  vote_count: number;
  comment_count: number;
  voted: boolean;
  created_at: string;
  updated_at: string;
}

export interface FeedbackPage {
  items: Feedback[];
  page: number;
  page_size: number;
  total: number;
}

export interface CreateFeedbackPayload {
  title: string;
  content: string;
}

export interface UpdateFeedbackStatusPayload {
  status: FeedbackStatus;
}

export interface FeedbackComment {
  id: number;
  feedback_id: number;
  content: string;
  created_by: number;
  created_at: string;
}

export interface CreateFeedbackCommentPayload {
  content: string;
}

export interface FeedbackVoteResult {
  feedback_id: number;
  voted: boolean;
  vote_count: number;
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
