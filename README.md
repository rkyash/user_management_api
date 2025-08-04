# User Management API

A complete RESTful API for user management with authentication, authorization, and logging built with Go, Gin, GORM, and PostgreSQL.

## Features

- User authentication with JWT (access and refresh tokens)
- User registration with email verification (simulated)
- User profile management
- Role-based access control (Admin, User)
- Password reset functionality (simulated)
- Structured logging
- Request/response logging
- PostgreSQL database with GORM
- Docker support

## Prerequisites

- Go 1.22 or higher
- PostgreSQL
- Docker and Docker Compose (optional)

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go
├── config/
│   ├── config.go
│   └── config.yaml
├── internal/
│   ├── auth/
│   │   └── auth.go
│   ├── handlers/
│   │   ├── admin_handler.go
│   │   ├── auth_handler.go
│   │   └── user_handler.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── logging.go
│   └── models/
│       └── models.go
├── logs/
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Configuration

Update the `config/config.yaml` file with your desired settings:

```yaml
server:
  port: "8080"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "postgres"
  dbname: "user_management"
  sslmode: "disable"

jwt:
  accessSecret: "your-access-secret-key"
  refreshSecret: "your-refresh-secret-key"
  accessExpiry: 15    # 15 minutes
  refreshExpiry: 7    # 7 days

log:
  level: "debug"
  file: "logs/app.log"
```

## Running the Application

### Using Docker

1. Build and start the containers:
   ```bash
   docker-compose up -d
   ```

2. The API will be available at http://localhost:8080

### Without Docker

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Set up the PostgreSQL database and update the configuration in `config/config.yaml`

3. Run the application:
   ```bash
   go run cmd/api/main.go
   ```

## API Endpoints

### Authentication
- POST `/api/v1/auth/register` - Register a new user
- POST `/api/v1/auth/login` - Login user
- POST `/api/v1/auth/refresh` - Refresh access token
- POST `/api/v1/auth/logout` - Logout user

### User Management
- GET `/api/v1/users/profile` - Get user profile
- PUT `/api/v1/users/profile` - Update user profile
- PUT `/api/v1/users/change-password` - Change password
- DELETE `/api/v1/users/account` - Delete user account

### Admin Routes
- GET `/api/v1/admin/users` - List all users
- PUT `/api/v1/admin/users/:id/role` - Change user role

### Health Check
- GET `/api/v1/health` - API health status

## Security Features

- Password hashing with bcrypt
- JWT token-based authentication
- Role-based access control
- Request rate limiting
- CORS configuration
- Secure headers
- SQL injection prevention through GORM

## Logging

The application uses structured logging with the following features:

- Request/response logging
- Authentication event logging
- User action logging
- Error logging with stack traces
- Daily rotating log files
- JSON formatted logs

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License.
