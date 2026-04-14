# Changelog

All notable changes to this project are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
