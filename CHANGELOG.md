# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] — 2026-04-15

### Changed (breaking)
- Minimum Go version is now **1.25** (was 1.22). Go 1.22 and 1.23 are past end-of-support. CI matrix now runs on Go 1.25 and 1.26.

### Added
- Runnable examples under [`examples/`](./examples): `send-message` (CLI) and `webhook-server` (HTTP receiver with signature verification).
- Supply-chain hardening:
  - CodeQL SAST workflow (security-and-quality queries) on push, PR, and weekly cron.
  - Go fuzz targets for `ParseEvent`, `VerifyWebhook`, and `MessagePart`, running 30s each in CI.
  - Release workflow keyless-signs SHA256 checksums of a source tarball with cosign/sigstore; signature, certificate, and checksums attached to the GitHub release.
  - All GitHub Actions pinned to commit SHAs; workflow tokens scoped to minimum permissions.

### Fixed
- Library test coverage raised from 77.9% → 86.7%. Added tests for `Webhooks.{List,Create,Update,Delete}`, `Chats.SendVoiceMemo`, `Attachments.Get`, `NewMediaPartByURL`, and `WithUserAgent`.

## [0.3.0] — 2026-04-14

### Changed (breaking)
- Webhook subscription JSON fields corrected to match the actual API (docs mislabel them):
  - `WebhookSubscription.URL` → `TargetURL` (`target_url`)
  - `WebhookSubscription.Events` → `SubscribedEvents` (`subscribed_events`)
  - `WebhookSubscription.Secret` → `SigningSecret` (`signing_secret`)
  - `WebhookSubscription.Description` removed (not returned by the API)
  - `WebhookSubscription.PhoneNumbers` added (`phone_numbers`)
  - Same renames applied to `CreateWebhookSubscriptionRequest` and `UpdateWebhookSubscriptionRequest`.

  Verified against the Linq sandbox on 2026-04-14: the API rejected `url` with `"target_url is required"` and `events` with `"subscribed_events is required"`. Response bodies carry `signing_secret` and `phone_numbers`.

## [0.2.0] — 2026-04-13

### Changed (breaking)
- `CreateChatResult.Message` removed. The initial message is nested inside the chat on the real Linq response, not a top-level field. Access it via `res.Chat.Message` (now a `*Message` on `Chat`). Fixes [#6](https://github.com/vsima/linq-go-sdk/issues/6).

## [0.1.0] — 2026-04-13

### Added
- Initial release: full coverage of the Linq Partner API V3 (chats, messages, reactions, attachments, phone numbers, webhooks).
- Typed errors with `APIError`, `IsNotFound`, `IsUnauthorized`, `IsRateLimited` helpers.
- Webhook event parsing via `ParseEvent`.
