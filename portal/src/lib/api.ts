const BASE = '/api/v1/portal';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token');
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, { ...options, headers });
  const body = await res.json();

  if (!res.ok) {
    throw new Error(body.message || 'Request failed');
  }
  return body;
}

export interface Developer {
  id: string;
  name: string;
  email: string;
}

export interface Application {
  id: string;
  app_name: string;
  platform: string;
  bundle_id: string;
  status: string;
  created_at: string;
}

export interface Credentials {
  application_id: string;
  publishable_key: string;
  secret_key: string;
  rate_limit: number;
}

interface SuccessResponse<T> {
  status: string;
  message: string;
  data: T;
}

export const api = {
  register(name: string, email: string, password: string) {
    return request<SuccessResponse<{ developer: Developer }>>('/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password }),
    });
  },

  login(email: string, password: string) {
    return request<SuccessResponse<{ token: string; developer: Developer }>>('/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  },

  getApplications() {
    return request<SuccessResponse<{ applications: Application[] }>>('/applications');
  },

  createApplication(app_name: string, platform: string, bundle_id: string) {
    return request<SuccessResponse<{ application: Application; credentials: Credentials }>>('/applications', {
      method: 'POST',
      body: JSON.stringify({ app_name, platform, bundle_id }),
    });
  },
};
