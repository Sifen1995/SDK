# Skykin Ad Portal

Advertiser-facing web app for campaign management. Connects to the Go backend at `/api/v1/ad-portal`.

## Run locally

1. Start the backend (`make up` from SDK root — port **8081**).
2. Install and run this app:

```bash
cd ad-portal
npm install
npm run dev
```

Open **http://localhost:3001** (developer portal runs on 3000).

## Features

- Register / login with role selection (`advertiser`, `read_only_analyst`)
- Profile page (`GET /ad-portal/me`)
- Campaign list, create, detail, activate, preview
- Image URL: direct link or Google Drive share link (converted to direct URL)
- Operator admin: team user management (`POST /ad-portal/admin/users`)

## Roles

| Role | Self-register | Create campaigns |
|------|---------------|------------------|
| Advertiser | Yes | Yes |
| Read-only Analyst | Yes | No (view only) |
| Operator Admin | No (seeded) | Yes + team admin |

## API base

Vite proxies `/api` → `http://localhost:8081`. All requests use `/api/v1/ad-portal`.
