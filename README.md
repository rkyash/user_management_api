# User Management API

# User Management API

A complete RESTful API for user management with JWT authentication, role-based authorization, and comprehensive API documentation using Scalar UI. Built with Go, Gin framework, and PostgreSQL.

## Features

- User authentication with JWT (access and refresh tokens)
- User registration with email verification (simulated)
- User profile management
- Role-based access control (Admin, User)
- Password reset functionality (simulated)
- Interactive API documentation with Scalar UI
- Swagger/OpenAPI specification
- PostgreSQL database with GORM
- Docker support
- Structured logging and monitoring

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
├── docs/                    # API Documentation
│   ├── docs.go             # Generated Swagger docs
│   ├── swagger.json        # OpenAPI specification
│   └── swagger.yaml        # YAML version of API spec
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
├── statics/               # Static files for documentation
│   └── docs/
│       └── index.html    # Scalar UI template
├── logs/                 # Application logs
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

## API Documentation

The API comes with two different documentation interfaces:

### Scalar UI Documentation
- **URL**: http://localhost:8080/
- Modern and interactive API documentation
- Features:
  - Clean and intuitive interface
  - Try-out functionality with request builder
  - Authentication support
  - Dark/Light theme
  - OpenAPI specification viewer
  - Code samples for various languages
  - Real-time API testing

### Swagger UI Documentation
- **URL**: http://localhost:8080/swagger/index.html
- Traditional Swagger interface with:
  - Complete API specification
  - Request/Response examples
  - Authentication flow documentation
  - Model schemas
  - Interactive endpoint testing

### OpenAPI Specification
- Available at: http://localhost:8080/docs/swagger.json
- Can be imported into any OpenAPI-compatible tool
- Detailed request/response schemas
- Authentication specifications
- Error responses documentation

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

## Updating API Documentation

When you make changes to the API endpoints, follow these steps to update the documentation:

1. Update the Swagger annotations in your handler functions
2. Regenerate the Swagger documentation:
   ```bash
   swag init -g cmd/api/main.go
   ```
3. The documentation will be automatically updated in both Scalar UI and Swagger UI

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

Note: When contributing, please ensure you update the API documentation for any endpoint changes.

## License

This project is licensed under the MIT License.
