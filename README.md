# Chirpy API

## What it does 
Chirpy is RESTfulAPI built from ground up in go. It serves as backend for microblogging. This service handles user authentication, post creation (chirp), database migrations, and secure webhook processing.

## Why should you care 
Chirpy is designed to demonstrate strong  foundational backend engineering principles. It stands out because it features:
* **Raw Performance:** Built using Go's lightning-fast library (`net/http` ServeMux)
* **Type-Safe Database Interactions:** Uses [SQLC](https://sqlc.dev/) to generate type-safe Go code directly from raw SQL queries,eliminating the magic and overhead of an ORM.
* **Security Best Practices** Implements short-live JWT authentications, secure password hashing and strict API key verifications for external webhooks.
* **Clean Architecture:** Separates routing, handlers and database logic into a maintainable, production-ready structure.

## How to Install and run
### 1.Prerequisites
Ensure you have the following installed on your machine:
* Go (v1.22 or higher) 
* PostgreSQL
* [Goose](https://pressly.github.io/goose/) (for running database migrations)
* [SQLC] (https://sqlc.dev/) (optional, for regenerating database queries)

### 2. Environment Variables
create `.env` file in the root directory and add the following keys:
```env
# The port the server will run on
PORT = 8080
# Database connection string
DB_URL="postgres://your_username:your_password@localhost:5432/chirpy"
# Security secrets
JWT_SECRET="generate-a-secure-random-256-bit-string"
POLKA_KEY="your-webhook-api-key"
```


### 3. Database Setup
Create a PostgreSQL database named `chirpy`.Then, run the Goose migrations to set up the database schema:
```bash
cd sql/schema
goose postgres $DB_URL up
```

### 4. Running the server 
Start the server from root directory of the project:
```bash
go build -o out && ./out
```
The server will start and listen for requests on `http://localhost:8080`.
---

## API Endpoints Overview
* **`/api/users`**: Account creation and profile updates.
* **`/api/login`**: JWT-based authentication.
* **`/api/chirps`**: Create, read and delete chirps (author filtering and creation-time based sorting for read)
* **`/api/polka/webhooks`**: Secure endpoint for external service upgrade (IsChirpyRed status)


