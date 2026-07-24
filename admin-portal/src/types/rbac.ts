export interface Permission {
  id: string;
  name: string;
  resource: string;
  action: string;
  description: string;
}

export interface Role {
  id: string;
  name: string;
  description: string;
  is_system: boolean;
  permissions: Permission[];
}

export interface CreateRoleRequest {
  name: string;
  description: string;
}
