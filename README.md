## gapi

A REST API built in Go to explore application architecture, authentication, authorization, HTTP middleware, concurrency, and data consistency.

# Features

- User registration and login
- 
- Password hashing and verification using Argon2id
- 
- Server-side session management
- 
- Cookie-based authentication
- 
- Authentication middleware
- 
- Owner-based authorization for item modifications
- 
- Item CRUD operations
- 
- In-memory stores with mutex-based synchronization
- 
- JSON request decoding and body-size limits

# Architecture

The application separates HTTP handling, middleware, business operations, and data storage.

- API handlers: Process HTTP requests and construct responses.
- 
- Middleware: Handles authentication and request identity.
- 
- Stores: Own application data and synchronization.
- 
- Password package: Handles password hashing and verification.
- 
- Root router: Composes the application's routes and middleware.
