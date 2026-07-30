# Architecture & File Explanation Guide

This document provides a complete file-by-file breakdown for the **Admin**, **Audience**, **Analytics**, **Billing**, **Delivery**, and **Bootstrap** modules in the Skykin Platform backend.

---

## 1. Admin Module (`internal/admin/`)

The **Admin Module** houses the business logic, ports, HTTP handlers, and DTOs for operator administrative actions in the Ad Portal (moderating campaigns, managing Audiencemart catalog segments and billing plans, reviewing segment candidates, and querying user intent profiles).

### Application Layer (`internal/admin/application/`)

* **`approve_candidate.go`**  
  Defines `ApproveCandidateUseCase` and the `SegmentPublisher` port interface. It validates input parameters (name, estimated CPM) for candidate approval and delegates transactional segment creation/publishing to the port implemented in `bootstrap`.
* **`reject_candidate.go`**  
  Defines `RejectCandidateUseCase` and the `CandidateRejecter` port interface. It records operator rejections for segment candidates with review notes.
* **`get_users_with_intents.go`**  
  Defines `GetUsersWithIntentsUseCase` and port interfaces (`UserLister`, `IntentBatchFetcher`) to retrieve paginated SDK users enriched with their latest predicted intent signals.

### Domain / Events Layer (`internal/admin/events/`)

* **`events.go`**  
  Contains domain event definitions and topic constants emitted when administrative actions occur (such as `TopicSubscriptionPlanCreated`, `TopicCampaignModerationPassed`, `TopicCampaignModerationRejected`, `TopicCampaignActivated`).

### HTTP Interfaces (`internal/admin/interfaces/http/`)

* **`campaign_handler.go`**  
  HTTP handler for operator campaign moderation (`/campaigns/pending`, `/campaigns/:id/validate`, `/campaigns/:id/activate`).
* **`catalog_handler.go`**  
  HTTP handler for managing catalog subscription plans, billing rates, and Audiencemart segment definitions (`/plans`, `/audience/segments`, etc.).
* **`segment_candidate_handler.go`**  
  HTTP handler for approving or rejecting candidate cohorts (`/audience/segment-candidates/:id/approve` and `/reject`). Converts candidate review errors into HTTP `201 Created` or `409 Conflict`.
* **`segment_candidate_dto.go`**  
  HTTP DTOs for candidate review requests and responses (`ApproveSegmentCandidateRequest`, `ApproveSegmentCandidateResponse`, `RejectSegmentCandidateRequest`).
* **`users_handler.go`**  
  HTTP handler for fetching paginated SDK users and their active intent predictions (`/sdk-users`).
* **`dto.go`**  
  Swagger/HTTP DTOs for catalog creation and moderation payloads (e.g. `ValidateCampaignRequest`, `CreatePlanRequest`, `CreateSegmentRequest`, `UpdateBillingRateRequest`).

### Routes & Validation (`internal/admin/routes/`, `internal/admin/validation/`)

* **`routes/routes.go`**  
  Wires the admin HTTP module (`Wire`) using services from `bootstrap`, `billing`, `audience`, and `campaigns`, and mounts routes under the ad portal admin group.
* **`validation/plan_segment.go`**  
  Input validation routines for plan and segment parameter limits (fees, CPM, daily budget caps).

---

## 2. Audience Module (`internal/audience/`)

The **Audience Module** manages Audiencemart catalog segments, user memberships, candidates created from sustained intent signals, and segment purchase entitlements.

### Domain Layer (`internal/audience/domain/`)

* **`audience_segment.go`**  
  Core domain entity `AudienceSegment` representing a purchasable audience cohort with CPM pricing, approximate size, and target intent signals.
* **`segment_candidate.go`**  
  Domain entities (`SegmentCandidate`, `UserInCandidate`, `CandidateStatus`) and repository ports (`CandidateRepository`, `MembershipRepository`, `UnitOfWork`) for candidate cohort lifecycle management.
* **`segment_purchase.go`**  
  Core domain entity `SegmentPurchase` representing an advertiser's entitlement to target an audience segment during a valid time window.
* **`repository.go`**  
  Port interfaces `SegmentRepository` and `PurchaseRepository` for reading/writing segments and purchases.

### Application Layer (`internal/audience/application/`)

* **`list_service.go`**  
  `ListService` manages reading catalog segments for advertisers and operator admins, calculating bundle prices, and exporting both `estimated_price_etb` and `price_etb` in `SegmentDTO`.
* **`process_approved_candidate.go`**  
  `ProcessApprovedCandidateUseCase` runs within an audience `UnitOfWork` transaction to atomically mark a candidate approved, create an `AudienceSegment`, and bulk-insert memberships from `segment_candidate_users`.
* **`process_intent_finding.go`**  
  `ProcessIntentFindingUseCase` handles intent consistency findings by creating or updating pending candidates via `UpsertPending` or enriching existing segments.
* **`purchase_service.go`**  
  `PurchaseService` validates target intent compatibility, prepares purchase quotes, and executes `ConfirmPurchaseTx` via `PurchaseRepository.CreatePurchaseTx`.
* **`reject_candidate.go`**  
  `RejectCandidateUseCase` marks a pending candidate as rejected within an audience `UnitOfWork`.
* **`segment_candidates.go`**  
  `ListSegmentCandidatesUseCase` queries pending, approved, or rejected candidates for admin review.
* **`targeting_resolver.go`**  
  `TargetingResolver` resolves allowed pseudonymous IDs for campaign targeting, verifying valid segment purchases and membership membership.
* **`finding_processor_adapter.go`**  
  Adapts `ProcessIntentFindingUseCase` to implement `analyticsApp.IntentFindingProcessor`.
* **`signals.go`**  
  Helper functions for parsing and matching intent signal JSON arrays against target intents.

### Infrastructure Layer (`internal/audience/infrastructure/`)

* **`candidate_repository.go`**  
  GORM implementation of `CandidateRepository`, persisting candidate metadata in `segment_candidates` and candidate user lists in `segment_candidate_users`.
* **`membership_repository.go`**  
  GORM implementation of `MembershipRepository`, handling bulk inserts and pseudonymous ID lookups in `segment_memberships`.
* **`postgres_repository.go`**  
  GORM implementation of `SegmentRepository` for catalog segment CRUD operations.
* **`postgres_purchase_repository.go`**  
  GORM implementation of `PurchaseRepository`, managing `segment_purchases` rows and supporting outer transaction handles.
* **`unit_of_work.go`**  
  GORM `UnitOfWork` implementation providing transactional scopes across candidate, segment, and membership repositories.
* **`persistence/audience_segment_row.go`**  
  GORM persistence model for `audience_segments`.
* **`persistence/segment_purchase_row.go`**  
  GORM persistence model for `segment_purchases`.

### Interfaces & Routes (`internal/audience/interfaces/`, `internal/audience/routes/`, `internal/audience/validation/`)

* **`interfaces/http/handler.go`**  
  Advertiser and admin HTTP handler for segment catalog operations.
* **`interfaces/http/segment_candidates.go`**  
  Admin HTTP handler for listing candidate cohorts.
* **`routes/routes.go`**  
  Mounts advertiser and admin audience routes.
* **`validation/segment.go`**  
  Input validation rules for creating segments and checking segment UUID formats.

---

## 3. Analytics Module (`internal/analytics/`)

The **Analytics Module** calculates platform-wide performance metrics, revenue overviews, delivery funnels, and runs automated intent consistency scans to detect sustained user interests.

### Domain Layer (`internal/analytics/domain/`)

* **`types.go`**  
  Domain data structures for analytics read models (`OverviewStats`, `CampaignPerformance`, `CampaignDetail`, `DeliveryAnalytics`, `RevenueOverview`, `AdvertiserSummary`).
* **`classification_config.go`**  
  Configuration parameters for intent consistency analysis (`MinConfidence`, `LookbackDays`, `MinDaysActive`, `MaxAgeDays`).
* **`intent_consistency.go`**  
  Data structures `ConsistentUser` and `IntentConsistencyFinding` representing sustained intent patterns.
* **`aggregate_report.go`**  
  Structs for anonymized device aggregate reports (`AggregateReport`, `AggregateItem`).
* **`aggregate_repository.go`**  
  Domain interface `AggregateRepository` for upserting daily anonymized intent aggregate counts.

### Application Layer (`internal/analytics/application/`)

* **`service.go`**  
  Application service wrapping `AnalyticsRepository` to expose overview, campaign performance, delivery, revenue, and advertiser summary methods.
* **`analyze_intent_consistency.go`**  
  `AnalyzeIntentConsistencyUseCase` orchestrates periodic scans over intent signals, reporting partial or total failures in `RunReport`.
* **`run_report.go`**  
  Summary data structures (`RunReport`, `IntentScanFailure`) for scan execution results.
* **`aggregate_ingest.go`**  
  `AggregateIngestService` validates and normalizes anonymous device aggregate reports before enqueueing them to Redis.

### Infrastructure Layer (`internal/analytics/infrastructure/`)

* **`postgres_repository.go`**  
  GORM read-model repository executing aggregate SQL queries across database tables with error checking.
* **`aggregate_repository.go`**  
  GORM implementation of `AggregateRepository` applying `ON CONFLICT (intent_name, date_bucket)` upserts to `intent_aggregate_counts`.
* **`aggregate_worker.go`**  
  Background worker draining `queue:analytics_aggregate` from Redis and executing batch upserts.
* **`aggregate_queue.go`**  
  Redis queue producer for pushing anonymized aggregate reports onto `queue:analytics_aggregate`.
* **`persistence/aggregate_row.go`**  
  GORM persistence model `IntentAggregateCountRow` for table `intent_aggregate_counts`.

### Interfaces & Routes (`internal/analytics/interfaces/http/`, `internal/analytics/routes/`)

* **`interfaces/http/handler.go`**  
  HTTP handlers exposing analytics endpoints (`/overview`, `/campaigns`, `/delivery`, `/revenue`, `/advertisers`).
* **`routes/routes.go`**  
  Wires and registers read-only analytics HTTP routes.

---

## 4. Billing Module (`internal/billing/`)

The **Billing Module** handles advertiser subscription plans, billing channels, pricing rates, billing event calculation from SDK telemetry, and subscription entitlement enforcement.

### Domain Layer (`internal/billing/domain/`)

* **`subscription_plan.go`**  
  Entity `SubscriptionPlan` defining monthly fees, active campaign caps, daily budget caps, included impressions, and feature flags (`SMSPlusEnabled`, `AudiencemartEnabled`).
* **`advertiser_subscription.go`**  
  Entity `AdvertiserSubscription` representing an advertiser's active subscription period.
* **`billing_rate.go`**  
  Entity `BillingRate` mapping billing event types (e.g., click, impression) and models (CPM, CPC, CPI, CPA, REV_SHARE) to ETB rates.
* **`billing_event.go`**  
  Entity `BillingEvent` recording a single billable interaction for an advertiser campaign.
* **`channel.go`**  
  Entity `Channel` representing delivery channels (`IN_APP_BANNER`, `PUSH`, `SMS_PLUS`, `NATIVE_FEED`).
* **`invoice.go`**  
  Entity `Invoice` representing aggregated advertiser charges.
* **`repository.go`**  
  Repository interfaces (`SubscriptionRepository`, `BillingRateRepository`, `ChannelRepository`, `CampaignQuotaReader`, `BillingEventRepository`).

### Application Layer (`internal/billing/application/`)

* **`event_processor.go`**  
  `EventProcessor` contains domain pricing rules (`CalculateCharge`, `BillingModelForEvent`, `BudgetExceeded`), resolves plan rates, persists `BillingEvent` batches, updates daily spend counters, and triggers budget exhaustion flags.
* **`plan_service.go`**  
  `PlanService` manages catalog subscription plans and emits plan creation events.
* **`billing_admin_service.go`**  
  `BillingAdminService` manages operator updates for billing rates.
* **`subscription_service.go`**  
  `SubscriptionService` handles advertiser plan subscriptions.
* **`subscription_enforcer.go`**  
  `SubscriptionEnforcer` verifies that advertisers remain within active campaign quotas, budget limits, and channel feature entitlements.
* **`seed_plan_rates.go`**  
  Seeding helper for initial subscription plans and rate cards.

### Infrastructure Layer (`internal/billing/infrastructure/`)

* **`postgres_subscription_repository.go`**  
  GORM implementation for `subscription_plans` and `advertiser_subscriptions`.
* **`postgres_billing_rate_repository.go`**  
  GORM implementation for `billing_rates`.
* **`postgres_channel_repository.go`**  
  GORM implementation for `channels`.
* **`billing_event_repository.go`**  
  GORM implementation for bulk-inserting `billing_events`.
* **`persistence/*.go`**  
  GORM persistence rows (`subscription_plan_row.go`, `advertiser_subscription_row.go`, `billing_rate_row.go`, `billing_event_row.go`, `channel_row.go`, `invoice_row.go`).

### Workers, Events, Interfaces & Routes (`internal/billing/`)

* **`worker/billing_consumer.go`**  
  Redis stream worker reading `stream:billing_events` under consumer group `billing_processor_group` and passing messages to `EventProcessor`.
* **`interfaces/events/plan_consumer.go`**  
  Messaging consumer reacting to admin plan creation events to seed default billing rates.
* **`interfaces/http/handler.go`**  
  HTTP handler for subscription and plan endpoints.
* **`routes/routes.go`**  
  Wires the billing HTTP module and mounts read/write subscription endpoints.
* **`validation/plan.go`**  
  Input validation for plan creation and rate update payloads.

---

## 5. Delivery Module (`internal/delivery/`)

The **Delivery Module** owns SDK ad delivery, anonymous/consented telemetry ingestion, anonymous CPC click tracking, and delivery log persistence (`campaign_delivery_logs`, `delivery_jobs`).

### Domain Layer (`internal/delivery/domain/`)

* **`delivery_log.go`**  
  Domain entity `DeliveryLog` representing a delivery lifecycle event (`DISPATCHED`, `RENDERED`, `CLICKED`, `CONVERTED`) and interface `DeliveryLogRepository`.
* **`delivery_job.go`**  
  Domain entity `DeliveryJob` tracking daily user/campaign delivery frequency.
* **`repository.go`**  
  Interface `DeliveryRepository` for querying delivery history and recording jobs.

### Application Layer (`internal/delivery/application/`)

* **`telemetry_ingest.go`**  
  `TelemetryIngestService` owns telemetry validation, high-speed Redis SETNX deduplication for impressions/clicks, and enqueueing to Redis stream `stream:billing_events`.
* **`delivery_log_mapper.go`**  
  `MapStreamToDeliveryLog` converts incoming stream fields into domain `DeliveryLog` records, validating install HMAC tokens.
* **`cpc_click_service.go`**  
  `CPCClickService` validates signed CPC click tokens and enqueues click events directly to `stream:billing_events`.

### Infrastructure Layer (`internal/delivery/infrastructure/`)

* **`delivery_log_repository.go`**  
  GORM repository writing batches to `campaign_delivery_logs`.
* **`delivery_repository.go`**  
  GORM repository checking delivery frequency and recording `delivery_jobs`.
* **`persistence/delivery_log_row.go`**  
  GORM persistence model for `campaign_delivery_logs`.
* **`persistence/delivery_job_row.go`**  
  GORM persistence model for `delivery_jobs`.

### Workers, HTTP & Routes (`internal/delivery/`)

* **`worker/delivery_log_consumer.go`**  
  Background stream consumer reading `stream:billing_events` under consumer group `delivery_log_processor_group` and persisting logs to `campaign_delivery_logs` and `delivery_jobs`.
* **`http/telemetry_handler.go`**  
  Gin HTTP handler exposing `/telemetry/track` and `/telemetry/track-anonymous`.
* **`http/cpc_click_handler.go`**  
  HTTP handler exposing `/telemetry/anonymous-click`.
* **`http/campaign_handler.go`**  
  Gin HTTP handler exposing SDK ad delivery routes.
* **`http/dto.go`**  
  Response DTOs for anonymous campaign delivery.
* **`http/routes.go`**  
  Mounts delivery SDK routes on Gin.

---

## 6. Platform Bootstrap Module (`internal/platform/bootstrap/`)

The **Bootstrap Module** acts as the application's composition root. It instantiates concrete infrastructure repositories, wires them into application services via domain ports, and initializes background workers.

* **`admin.go`**  
  Constructs `GetUsersWithIntentsUseCase` by adapting `UserRepository`, `PseudonymousMappingRepository`, and `IntentRepository`.
* **`admin_events.go`**  
  Registers async event consumers for admin-emitted events.
* **`analytics.go`**  
  Wires `AnalyzeIntentConsistencyUseCase` with `FindingProcessorAdapter` and sets up the periodic analysis ticker and aggregate worker.
* **`campaign_quota.go`**  
  Provides `NewCampaignQuotaReader`, adapting campaign repository to `billingdomain.CampaignQuotaReader`.
* **`campaigns.go`**  
  Constructs `StartTargetingJob`, wiring campaign, intent, delivery, channel, segment, purchase, and membership repositories.
* **`consent.go`**  
  Wires the consent registration system and event consumers.
* **`delivery_sdk.go`**  
  Wires `NewDeliverySDKSystem` and starts stream workers `StartBillingStreamWorker` and `StartDeliveryLogStreamWorker`.
* **`intents.go`**  
  Wires `NewIntentSystem` (intent ingestion, profile repository, ad selector) and starts the intent log flusher worker.
* **`permissions.go`**  
  Wires RBAC permission checker and permission management services.
* **`segment_review.go`**  
  Adapts audience transactional use cases (`ProcessApprovedCandidateUseCase`, `RejectCandidateUseCase`) to implement admin port interfaces (`SegmentPublisher`, `CandidateRejecter`).
* **`user_resolver.go`**  
  Constructs `PseudonymousConsentGate` to verify consent mappings for incoming pseudonymous IDs without exposing internal user IDs.

---

## Verification Summary

All packages across the workspace compile cleanly without errors:

```bash
go build ./...
```

All package unit tests pass successfully:

```bash
go test ./internal/...
```
