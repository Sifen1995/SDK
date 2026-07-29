export interface IntentSummary {
  intent_name: string;
  confidence: number;
  predicted_at: string;
}

export interface SdkUser {
  user_id: number;
  created_at: string;
  latest_intent?: IntentSummary | null;
}

export interface SdkUsersResponse {
  users: SdkUser[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}
