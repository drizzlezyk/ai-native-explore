# User External API Documentation

This document describes the external user APIs used by xihe-server for authentication and user management.

## Base URLs

- **Production**: `https://omapi.osinfra.cn`
- **Development**: `https://xiheapi.test.osinfra.cn`

---

## Authentication APIs

### Get User Info by Access Token

Retrieve user information using an access token.

**Endpoint**: `GET /oidc/user`

**Authentication**: Bearer token in Authorization header

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| Authorization | `{access_token}` | Yes |
| Accept | `application/json` | Yes |
| Content-Type | `application/json` | Yes |
| User-Agent | `xihe-server-authing` | Yes |

**Response**:

```json
{
  "username": "string",
  "picture": "string",
  "email": "string",
  "sub": "string",
  "phone_number": "string"
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| username | string | User's username |
| picture | string | URL to user's avatar |
| email | string | User's email address |
| sub | string | User ID (unique identifier) |
| phone_number | string | User's phone number |

---

### Get Access Token by Code

Exchange OAuth authorization code for access token.

**Endpoint**: `POST /oidc/token`

**Content-Type**: `application/x-www-form-urlencoded`

**Request Body**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| client_id | string | Yes | Application ID |
| client_secret | string | Yes | Application Secret |
| grant_type | string | Yes | Must be `authorization_code` |
| code | string | Yes | Authorization code from OAuth callback |
| redirect_uri | string | Yes | OAuth redirect URI |

**Response**:

```json
{
  "access_token": "string",
  "id_token": "string"
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| access_token | string | Access token for API requests |
| id_token | string | ID token for user authentication |

---

### Get User by Code

Complete login flow - exchange code for user info.

**Endpoint**: `POST /oidc/token` (then `GET /oidc/user`)

This is the login API that combines token exchange and user info retrieval.

**Request**: Same as "Get Access Token by Code"

**Response**:

```json
{
  "access_token": "string",
  "id_token": "string",
  "user_info": {
    "username": "string",
    "picture": "string",
    "email": "string",
    "sub": "string",
    "phone_number": "string"
  }
}
```

---

### Get Manager Token

Get an admin-level token for managing users.

**Endpoint**: `POST /manager/token`

**Content-Type**: `application/json`

**Request Body**:

```json
{
  "grant_type": "token",
  "app_id": "string",
  "app_secret": "string"
}
```

**Response**:

```json
{
  "msg": "OK",
  "refresh_token": "string",
  "token_expire": 600,
  "refresh_token_expire": 1200,
  "status": 200,
  "token": "string"
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| msg | string | Response message |
| refresh_token | string | Refresh token |
| token_expire | integer | Token expiration time in seconds |
| refresh_token_expire | integer | Refresh token expiration time in seconds |
| status | integer | HTTP status code |
| token | string | Manager access token |

---

### Get User Info (External)

Get user information for display purposes.

**Endpoint**: `GET /oneid/personal/center/user`

**Query Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| community | string | Yes | Community name (e.g., `mindspore`) |
| client_id | string | No | Community app ID |

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| token | `{user_token}` | Yes |
| cookie | `_Y_G_={yg_cookie_value}` | Yes |
| Referer | `{referer_url}` | Yes |

**Response**:

```json
{
  "data": {
    "company": "string",
    "email": "string",
    "nickname": "string",
    "phone": "string",
    "photo": "string",
    "signedUp": "string",
    "username": "string",
    "identities": [
      {
        "identity": "string",
        "accessToken": "string",
        "login_name": "string",
        "user_name": "string"
      }
    ]
  }
}
```

**Response Fields**:

| Field | Type | Description |
|-------|------|-------------|
| data.company | string | User's company name |
| data.email | string | User's bound email |
| data.nickname | string | User's nickname |
| data.phone | string | User's bound phone number |
| data.photo | string | URL to user's avatar |
| data.signedUp | string | Account creation date (ISO format) |
| data.username | string | User's username |
| data.identities | array | List of bound third-party accounts |
| data.identities[].identity | string | Third-party source (gitee/github/openatom) |
| data.identities[].accessToken | string | Third-party access token |
| data.identities[].login_name | string | Third-party login name |
| data.identities[].user_name | string | Third-party user name |

---

### Get User by Manager Token

Get user information using manager-level privileges.

**Endpoint**: `GET /oneid/manager/personal/center/user`

**Query Parameters**:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| community | string | Yes | Community name (e.g., `mindspore`) |

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| token | `{manager_token}` | Yes |
| user-token | `{user_token}` | Yes |
| cookie | `_Y_G_={yg_cookie_value}` | Yes |
| Referer | `{referer_url}` | Yes |

**Response**:

```json
{
  "msg": "success",
  "code": 200,
  "data": {
    "signedUp": "string",
    "identities": [],
    "phoneCountryCode": "+86",
    "phone": "string",
    "nickname": "string",
    "photo": "string",
    "company": "string",
    "privacySignedTime": "string",
    "userId": "string",
    "email": "string",
    "username": "string"
  }
}
```

---

### Send Email Verification Code

Send a verification code to user's email.

**Endpoint**: `POST /manager/sendcode`

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| token | `{manager_token}` | Yes |
| Content-Type | `application/json` | Yes |

**Request Body**:

```json
{
  "userId": "string"
}
```

---

### Bind Email Account

Bind an email account to a user.

**Endpoint**: `POST /manager/bind/account`

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| token | `{manager_token}` | Yes |
| Content-Type | `application/json` | Yes |

**Request Body**:

```json
{
  "userId": "string"
}
```

---

### Privacy Revoke

Revoke a user's privacy agreement.

**Endpoint**: `POST /manager/privacy/revoke`

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| token | `{manager_token}` | Yes |
| Content-Type | `application/json` | Yes |

**Request Body**:

```json
{
  "userId": "string"
}
```

---

### Modify User Account

Modify user's account information.

**Endpoint**: `POST /manager/update/account`

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| token | `{manager_token}` | Yes |
| Content-Type | `application/json` | Yes |

---

### Fallback Modify Account Info

Fallback endpoint for updating account information.

**Endpoint**: `POST /manager/update/accountInfo`

**Request Headers**:

| Header | Value | Required |
|--------|-------|----------|
| token | `{manager_token}` | Yes |
| Content-Type | `application/json` | Yes |

---

## Usage Examples

### cURL Example: Get User Info

```bash
# Variables (replace with actual values)
UT_TOKEN="<utCookie.Value>"
YG_VALUE="<ygCookie.Value>"
REFERER="https://mindspore-website.test.osinfra.cn"

curl -X GET "https://xiheapi.test.osinfra.cn/oneid/personal/center/user?community=mindspore" \
  -H "token: ${UT_TOKEN}" \
  -H "cookie: _Y_G_=${YG_VALUE}" \
  -H "Referer: ${REFERER}" \
  -v
```

### cURL Example: Get Manager Token

```bash
APP_ID="your_app_id"
APP_SECRET="your_app_secret"

curl -X POST "https://xiheapi.test.osinfra.cn/oneid/manager/token" \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "token",
    "app_id": "'${APP_ID}'",
    "app_secret": "'${APP_SECRET}'"
  }'
```

### cURL Example: Get User by Manager Token

```bash
UT_TOKEN="<utCookie.Value>"
YG_VALUE="<ygCookie.Value>"
MANAGER_TOKEN="<manager_token>"
REFERER="https://mindspore-website.test.osinfra.cn"

curl -X GET "https://xiheapi.test.osinfra.cn/oneid/manager/personal/center/user?community=mindspore" \
  -H "token: ${MANAGER_TOKEN}" \
  -H "user-token: ${UT_TOKEN}" \
  -H "cookie: _Y_G_=${YG_VALUE}" \
  -H "Referer: ${REFERER}" \
  -v
```
