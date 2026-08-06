CREATE TABLE rbac_permissions (
    id          UUID         PRIMARY KEY
                             DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    resource    VARCHAR(50)  NOT NULL,
    action      VARCHAR(50)  NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rbac_permissions_resource
    ON rbac_permissions (resource);

CREATE TABLE rbac_roles (
    id          UUID         PRIMARY KEY
                             DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    is_system   BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE rbac_role_permissions (
    role_id       UUID NOT NULL
                  REFERENCES rbac_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL
                  REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    granted_by    UUID,
    granted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_rbac_role_permissions_role
    ON rbac_role_permissions (role_id);

INSERT INTO rbac_permissions
    (name, resource, action, description)
VALUES
    ('campaigns:create',  'campaigns', 'create',
     'Create new campaigns'),
    ('campaigns:read',    'campaigns', 'read',
     'View campaigns and creatives'),
    ('campaigns:update',  'campaigns', 'update',
     'Edit existing campaigns'),
    ('campaigns:delete',  'campaigns', 'delete',
     'Archive or delete campaigns'),
    ('analytics:read',    'analytics', 'read',
     'View analytics dashboards'),
    ('analytics:export',  'analytics', 'export',
     'Export reports as CSV or PDF'),
    ('billing:read',      'billing',   'read',
     'View invoices and usage'),
    ('billing:manage',    'billing',   'manage',
     'Manage subscriptions and payment methods'),
    ('users:read',        'users',     'read',
     'View registered users and their intents'),
    ('geofences:manage',  'geofences', 'manage',
     'Create and manage geofence zones'),
    ('segments:read',     'segments',  'read',
     'Browse Audiencemart segments'),
    ('segments:purchase', 'segments',  'purchase',
     'Purchase audience segments for campaigns');

INSERT INTO rbac_roles (name, description, is_system)
VALUES
    ('operator_admin',
     'Full platform access — all permissions',    TRUE),
    ('advertiser',
     'Campaign management and performance access', TRUE),
    ('analyst',
     'Read-only analytics and reporting access',  TRUE);

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r, rbac_permissions p
WHERE r.name = 'advertiser'
AND   p.name IN (
    'campaigns:create', 'campaigns:read',
    'campaigns:update', 'analytics:read',
    'billing:read', 'segments:read',
    'segments:purchase', 'geofences:manage');

INSERT INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM rbac_roles r, rbac_permissions p
WHERE r.name = 'analyst'
AND   p.name IN (
    'analytics:read', 'analytics:export',
    'campaigns:read', 'users:read');
