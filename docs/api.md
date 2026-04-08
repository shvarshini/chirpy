# Chirpy API Reference 
Base URL: `http://localhost:8080`

All endpoints return JSON and expect `Content-Type: application/json` for request bodies unless otherwise noted.

---
## User & Authentication
### Create a User
**`POST /api/users`**
Creates a new user account.
**Request Body:**
```json
{
"email" : "test@example.com",
"password" : "password123"
}
```
**Response (201 Created):**
```json
{
"id" : "uuid-string",
"created-at" : "2023-10-01T12:00:00Z",
"updated-at" : "2023-10-01T12:00:00Z",
"email" :  "test@example.com",
"is_chirpy_red" : false
}
```
* `400 Bad Request`:JSON request is invalid.
* `500 Internal Server Error`: Database error or server error.
---

### Login
**`POST /api/login`**
Authenticates a user and returns a JWT.

*(Note: `expires_in_seconds` is optional and defaults to 1 hour if not provided).*
**Request Body:**
```json
{
"email" : "test@example.com",
"password" : "password123",
"expires_in_seconds": 3600 
}
```
**Response (200 OK):**
```json
{
"id" : "uuid-string",
"created-at" : "2023-10-01T12:00:00Z",
"updated-at" : "2023-10-01T12:00:00Z",
"email" :  "test@example.com",
"is_chirpy_red" : false,
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
* `400 Bad Request`:JSON request is invalid.
* `401 Unauthorized`:Invalid email or password.
* `500 Internal Server Error`: Database error or server error.
---

### Update User
**`PUT /api/users`**
Updates a user's email and password
**Headers:**
* `Authorization: Bearer <your_jwt>`
**Request Body:**
```json
{
"email" : "test@example.com",
"password" : "password123"
}
```
**Response (200 OK):**
```json
{
"id" : "uuid-string",
"created-at" : "2023-10-01T12:00:00Z",
"updated-at" : "2023-10-01T12:00:00Z",
"email" :  "test@example.com",
"is_chirpy_red" : false
}
```
* `400 Bad Request`:JSON request is invalid.
* `401 Unauthorized`: Missing or invalid JWT token.
* `500 Internal Server Error`: Database error or server error.
---
## Chirps
### Create a Chirp
**`POST /api/chirps`**

Creates a new chirp. The chirp body must be between 1 and 140 characters.

**Headers:**
* `Authorization: Bearer <your_jwt>`

**Request Body:**
```json
{
"body": "This is my very first chirp!"
}
```

**Responses:**

* `201 Created`: Successfully created.
  ```json
  {
  "id": "uuid-string",
  "created_at": "2023-10-01T12:00:00Z",
  "updated_at": "2023-10-01T12:00:00Z",
  "body": "This is my very first chirp!",
  "user_id": "uuid-string"
  }
  ```
* `400 Bad Request`: Chirp is empty, exceeds 140 characters, or JSON is invalid.
* `401 Unauthorized`: Missing or invalid JWT token.
* `500 Internal Server Error`: Database error or server failure.

---
### Get All Chirps
**`GET /api/chirps`**

Retrieves a list of all chirps.
* `author_id` (UUID): Filter the results to only show chirps from a specific user.
    * *Example: `/api/chirps?author_id=uuid-string`*
* `sort` (String): Sort by creation date. Accepts `asc` (default) or `desc`.
    * *Example: `/api/chirps?sort=desc`*
**Response (200 OK):**
```json
[
  {
    "id": "uuid-string",
    "created_at": "2023-10-01T12:00:00Z",
    "updated_at": "2023-10-01T12:00:00Z",
    "body": "This is my first chirp!",
    "user_id": "uuid-string"
  }
]
```
* `400 Bad Request`: Provided invalid author Id .
* `500 Internal Server Error`: Database error or server failure.
---
### Delete a Chirp
**`DELETE /api/chirps/{chirpID}`**

Deletes a chirp. The user requesting the deletion must be the author of the chirp.

**Headers:**
* `Authorization: Bearer <your_jwt>`

**Responses:**
* `204 No Content`: Successfully deleted.
* `400 Bad Request`: Provided invalid chirp Id .
* `401 Unauthorized`: Missing or invalid JWT token.
* `403 Forbidden`: User is authenticated but is not the author of the chirp.
* `404 Not Found`: Chirp does not exist.
* `500 Internal Server Error`: Database error or server error.

---
##  Webhooks

### Polka Upgrade Webhook
**`POST /api/polka/webhooks`**

Secured endpoint used by the Polka payment processor to notify the API of a user's account upgrade to "Chirpy Red" status.

**Headers:**
* `Authorization: ApiKey <polka_api_key>`

**Request Body:**
```json
{
"event": "user.upgraded",
"data": {
"user_id": "uuid-string"
}
}
```

**Responses:**
* `204 No Content`: Successfully processed (or ignored safely if the event type was not `user.upgraded`).
* `400 Bad Request`: Invalid JSON Request or invalid user Id.
* `401 Unauthorized`: Missing or invalid API key.
* `404 Not Found`: User ID provided in the payload does not exist in the database.
* `500 Internal Server Error`: Database error or server error.











