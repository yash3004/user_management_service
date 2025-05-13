# User Management Service

A service for managing users, roles, policies, and projects.

## Super User Feature

The service automatically creates a super user on startup with full permissions to manage policies, roles, and projects. This super user can be used to bootstrap the system and create additional users and roles.

### Super User Credentials

By default, the super user is created with the following credentials:

- Email: admin@example.com
- Password: superuser123

**Important:** Change these credentials in production by modifying the `DefaultSuperUserConfig` function in the `internal/superuser/superuser.go` file.

### Authentication

To authenticate as the super user:

1. Send a POST request to `/auth/login` with the following JSON payload:

```json
{
  "email": "admin@example.com",
  "password": "superuser123"
}
```

2. The response will include a JWT token:

```json
{
  "token": "your-jwt-token",
  "user_id": "user-uuid",
  "email": "admin@example.com",
  "first_name": "Super",
  "last_name": "User",
  "role": "SuperAdmin"
}
```

3. Include this token in the `Authorization` header for subsequent requests:

```
Authorization: Bearer your-jwt-token
```

### Super User Permissions

The super user has the following permissions:

- Full access to manage policies (`/api/policies/*`)
- Full access to manage roles (`/api/roles/*`)
- Full access to manage projects (`/api/projects/*`)

## API Endpoints

### Authentication

- `POST /auth/login` - Authenticate a user and get a JWT token

### Projects

- `POST /api/projects/create` - Create a new project
- `GET /api/projects/get/{id}` - Get a project by ID
- `GET /api/projects/list` - List all projects
- `PUT /api/projects/update/{id}` - Update a project
- `DELETE /api/projects/delete/{id}` - Delete a project

### Roles

- Role management endpoints (to be implemented)

### Policies

- Policy management endpoints (to be implemented)

## Development

### Running the Service

```bash
go run cmd/server/main.go
```

### Building the Service

```bash
go build -o server cmd/server/main.go
```

### Running Tests

```bash
go test ./...
```