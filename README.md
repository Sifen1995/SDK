| Component | vCPU | RAM | Disk |
|-----------|------|-----|------|
| **Postgres + PostGIS** | 1–2 | **2–4 GB** | **20+ GB** SSD (grows with events/visits) |
| **Redis** | 1 | **512 MB–1 GB** | small + AOF volume |
| **API** | 1–2 | **512 MB–1 GB** | negligible |
| **All-in-one VM** | **2–4** | **4–8 GB** | **40+ GB** SSD |