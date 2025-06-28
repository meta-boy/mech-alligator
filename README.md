# Mech-Alligator

![Mech-Alligator Logo](./frontend/public/head-logo.png)

Mech-Alligator is a full-stack application designed to...

## Project Structure

This project is composed of two main parts:

### Frontend

The frontend is a Next.js application. It provides the user interface for interacting with the backend services.

- **Framework**: Next.js
- **Language**: TypeScript
- **Styling**: Tailwind CSS

### Backend

The backend is a Go application that provides the API services and handles background tasks.

- **Language**: Go
- **Database**: PostgreSQL (implied by `lib/pq` and `migrate/v4` in `go.mod`)
- **Task Queue**: (Implied by `worker` service in `docker-compose.yaml`)

## Getting Started

To get the project up and running, follow these steps:

### Prerequisites

- Docker
- Docker Compose

### Setup

1.  **Clone the repository**:

    ```bash
    git clone <repository-url>
    cd mech-alligator
    ```

2.  **Environment Variables**: Create a `.env` file in the root directory based on `.env.example` and fill in the necessary environment variables.

    ```bash
    cp .env.example .env
    # Edit .env with your configurations
    ```

3.  **Run with Docker Compose**:

    ```bash
    docker-compose up --build
    ```

    This command will build the Docker images for both the frontend and backend, and then start all the services defined in `docker-compose.yaml`.

### Accessing the Application

- **Frontend**: Once the services are up, the frontend should be accessible at `http://localhost:3000` (or the port configured in your `.env` for the frontend).
- **Backend API**: The backend API will be running on the port configured in your `.env` for the backend (e.g., `http://localhost:8080`).

## Development

### Frontend Development

To run the frontend in development mode (without Docker Compose):

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

### Backend Development

To run the backend in development mode (without Docker Compose):

```bash
cd backend
go run cmd/api/main.go
# or for the worker
go run cmd/worker/main.go
```

## Images

Branding images are located in `frontend/public/`:
- `head-logo.png`
- `og-image.png`
