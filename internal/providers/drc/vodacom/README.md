# Vodacom DRC Provider

This package contains the **Vodacom DRC provider integration** used by TAAS for:

- Airtime purchase
- Bundle offers lookup
- Bundle allocation / purchase
- Shared provider credentials loading
- Transaction persistence support

It is the provider-facing layer for **Vodacom DRC** inside the Telecom-as-a-Service application.

---

## What this provider does

The Vodacom DRC integration currently supports two different provider communication styles:

1. **Bundle flow** – JSON/HTTP-based flow
2. **Airtime flow** – SOAP/XML-based flow using a Huawei-style request/response structure

Although both flows belong to the same provider, they do **not** behave the same way internally.

---

## Provider flows

## 1. Bundle flow

The bundle flow is a **three-step JSON API process**:

1. Generate an access token
2. Fetch bundle offers for a target MSISDN
3. Allocate/purchase the selected bundle using the returned session and transaction details

### Internal flow

```text
TAAS API
   ↓
Vodacom bundle service
   ↓
Generate Token
   ↓
Get Offers
   ↓
Allocate Bundle
```

### Bundle steps explained

#### Step 1 – Generate token

Before any bundle request can be made, the provider issues an access token using client credentials.

This token is then used in the authorization header for subsequent bundle requests.

#### Step 2 – Get offers

The system fetches bundle offers for a customer MSISDN.

The provider response typically includes:

- Available offers
- Bundle IDs
- Offer descriptions
- Price or amount
- Session ID
- Provider transaction ID
- Event detail / event description

These values are important because the next allocation call depends on them.

#### Step 3 – Allocate bundle

Once an offer is chosen, TAAS sends an allocation request using:

- Access token
- Customer MSISDN
- Selected bundle ID
- Session ID from the offers response
- Provider transaction ID from the offers response
- Language / request metadata

TAAS then interprets the response and stores the raw provider result in the transaction record.

---

## 2. Airtime flow

The airtime flow is different from bundles.

Airtime is currently implemented using a **SOAP/XML request** that follows a Huawei-style telecom integration pattern.

This means the request is sent as an XML SOAP envelope rather than a JSON payload.

### Internal flow

```text
TAAS API
   ↓
Vodacom airtime service
   ↓
Build SOAP envelope
   ↓
Submit SOAP request
   ↓
Parse SOAP response / fault
```

### Airtime request structure

The airtime purchase request contains telecom-style request metadata such as:

- Command ID
- Originator conversation ID
- Caller information
- Initiator information
- Security credential
- Short code
- Recharge target MSISDN
- Amount
- Currency

These values are packaged into a SOAP envelope and sent to the provider endpoint.

### Airtime response handling

The airtime service checks for:

- HTTP transport failures
- Non-2xx HTTP responses
- SOAP faults
- Empty provider responses
- Successful raw XML responses

If a SOAP fault is returned, the fault string is used as the failure reason and stored on the transaction.

---

## Huawei-style behavior

The airtime integration behaves like a Huawei-style telecom SOAP service.

That generally means:

- Requests are XML SOAP envelopes
- Errors may come back as SOAP Faults
- A failure may appear in the XML body even if the HTTP status is not obviously descriptive
- Important provider details may appear in:
  - `faultcode`
  - `faultstring`
  - SOAP body payload
  - provider raw XML

SOAP faults are a standard SOAP mechanism for structured error reporting and usually contain a fault code and a human-readable fault string.[web:261][web:265][web:268]

In this provider implementation, SOAP raw responses are preserved in transaction records so provider failures can be inspected later.

---

## Files in this provider

```text
internal/providers/drc/vodacom/
├── airtime/
│   ├── dto.go
│   └── service.go
├── bundle/
│   ├── dto.go
│   └── service.go
├── client/
│   └── client.go
├── shared/
│   └── helpers.go
└── README.md
```

### File responsibilities

#### `client/client.go`

Shared provider client for Vodacom DRC.

Responsibilities:

- Load provider credentials from the database
- Generate bearer access tokens
- Send JSON provider requests
- Send SOAP provider requests
- Provide common provider metadata defaults

#### `bundle/service.go`

Implements the Vodacom DRC bundle flow.

Responsibilities:

- Load partner/provider configuration
- Generate bearer token
- Fetch bundle offers
- Purchase/allocate bundles
- Store transaction state and raw provider payloads
- Interpret provider event description and response codes

#### `bundle/dto.go`

Defines request and response DTOs for bundle operations.

Includes:

- Offers request
- Offers response
- Bundle purchase request
- Bundle purchase response
- Provider-facing fields such as:
  - session ID
  - provider transaction ID
  - event description
  - provider status code

#### `airtime/service.go`

Implements the SOAP airtime purchase flow.

Responsibilities:

- Build SOAP XML envelope
- Populate Huawei-style request fields
- Submit SOAP request
- Detect SOAP faults
- Store provider raw response
- Mark transactions as success or failed

#### `airtime/dto.go`

Defines external request and response DTOs for airtime purchase.

#### `shared/helpers.go`

Contains small helper utilities shared across the provider package, such as pointer helpers or bundle ID normalization.

---

## Credentials expected by this provider

The provider expects its configuration to come from the `partner_provider_credentials.meta` JSON blob.

Typical fields include:

### JSON / bundle-related fields

- `base_url`
- `basic_token`
- `offer_username`
- `offer_password`

### SOAP / airtime-related fields

- `airtime_soap_url`
- `third_party_id`
- `third_party_password`
- `initiator_identifier`
- `security_credential`
- `short_code`
- `caller_type`
- `key_owner`
- `command_id`

### TLS / transport option

- `insecure_skip_tls_verify`

---

## TAAS endpoints backed by this provider

The public TAAS HTTP layer calls into this provider package.

### Bundle offers

```http
POST /api/drc/vodacom/bundles/offers
```

This endpoint:

- Authenticates the TAAS partner
- Calls the provider token endpoint
- Calls the provider offers endpoint
- Returns normalized offers plus provider session data

### Bundle purchase

```http
POST /api/drc/vodacom/bundles/purchase
```

This endpoint:

- Requires idempotency
- Uses bundle/session/provider transaction details
- Calls provider allocation
- Persists provider raw response
- Returns normalized purchase status

### Airtime purchase

```http
POST /api/drc/vodacom/airtime/purchase
```

This endpoint:

- Requires idempotency
- Builds a SOAP request
- Sends the SOAP request to the provider
- Detects faults and transport errors
- Stores raw XML response for auditing

### Transaction lookup

```http
GET /api/drc/vodacom/transactions/:id
```

This returns the stored transaction record, including provider-facing details captured during processing.

---

## Bundle response mapping

The bundle implementation normalizes provider fields into TAAS response fields.

Provider-originated values commonly mapped include:

- Offer list
- Session ID
- Provider transaction ID
- Event detail
- Event description
- Output response code

This makes it easier for TAAS clients to work with a consistent JSON shape while still preserving provider context.

---

## Airtime error mapping

The airtime implementation converts provider SOAP behavior into TAAS transaction states.

### Failure sources handled

- SOAP request transport failure
- HTTP status failure
- SOAP fault response
- Empty SOAP response
- Transaction persistence failure

### Stored debug context

When possible, the transaction record stores raw provider XML or error details so failures can be audited later.

This is especially important for SOAP integrations where faults may be embedded in structured XML instead of plain JSON error bodies.[web:261][web:264]

---

## Why bundles and airtime are separate

Even though both belong to Vodacom DRC, the two services are intentionally separated because they have different transport and response behavior:

| Flow | Transport | Request type | Response type | Error style |
|---|---|---|---|---|
| Bundles | HTTP | JSON | JSON | Event fields / response codes |
| Airtime | HTTP | SOAP XML | SOAP XML | SOAP faults / raw XML |

This separation keeps each service simpler, easier to test, and easier to extend.

---

## Notes for contributors

If you are extending this provider, keep these rules in mind:

- Keep bundle logic in `bundle/`
- Keep airtime SOAP logic in `airtime/`
- Put shared HTTP/provider helpers in `client/`
- Put small cross-cutting utilities in `shared/`
- Preserve raw provider responses when possible
- Do not hardcode provider secrets in code
- Read credentials from `partner_provider_credentials.meta`

---

## Future extension ideas

This provider can be extended with:

- Bundle status lookup
- Airtime transaction reconciliation
- Reversal support
- Cached access tokens
- Better provider response normalization
- Structured provider error catalog
- Automated integration tests against sandbox environments

---

## Security note

Do not commit real provider credentials, real private IP addresses, or real production secrets into source control.

Use placeholders in documentation and load all sensitive values from environment variables or database-backed provider credential records.

---

## Open-source note

This provider is part of a broader Telecom-as-a-Service platform and is open to improvement. Contributions are welcome for reliability, documentation, observability, testing, and additional telecom integrations.