# Mech-Alligator

<div align="center">
    <img src="./frontend/public/head-logo.png" alt="Mech-Alligator Logo" width="300">
</div>

A product aggregation platform that brings together inventory from trusted Indian resellers, making it easier to find and compare products without hopping between different websites.

## What does it do?

Ever wished you could search across multiple Indian resellers at once? That's exactly what Mech-Alligator does. Instead of opening 10 different tabs to check if a product is available, you can search once and see results from all our partner resellers.

**Core features:**
- Search products across multiple Indian resellers simultaneously
- Compare prices and availability in real-time
- Clean, unified interface that doesn't make your eyes bleed
- No more bookmark folders with 50+ reseller websites

## Tech Stack

**Frontend**: Next.js with TypeScript and Tailwind CSS - because we like our code typed and our styles utility-first.

**Backend**: Go with PostgreSQL - fast, reliable, and handles concurrent requests like a champ.

**Infrastructure**: Dockerized everything because "it works on my machine" isn't good enough.

## Quick Start

### What you'll need
- Docker and Docker Compose (if you don't have these, get them first)
- About 5 minutes of your time

### Getting it running

1. Clone this repo:
   ```bash
   git clone <repository-url>
   cd mech-alligator
   ```

2. Set up your environment variables:
   ```bash
   cp .env.example .env
   ```
   Open `.env` and fill in your configuration. Don't skip this step - the app won't work without proper environment variables.

3. Fire it up:
   ```bash
   docker-compose up --build
   ```

That's it. Docker will handle building everything and getting the services talking to each other.

### Where to find your app

- **Frontend**: http://localhost:3000 (or whatever port you configured)
- **Backend API**: http://localhost:8080 (or your configured backend port)

## Development Setup

### Frontend only
If you want to work on just the frontend:

```bash
cd frontend
npm install
npm run dev
```

The development server will start at http://localhost:3000 with hot reloading enabled.

### Backend only
For backend development:

```bash
cd backend
go run cmd/api/main.go        # Start the API server
go run cmd/worker/main.go     # Start the background worker
```

The API server handles web requests while the worker processes background tasks like updating product data from resellers.

## Project Structure

```
mech-alligator/
├── frontend/           # Next.js app
│   ├── public/        # Static assets (logos, images)
│   └── ...
├── backend/           # Go services
│   ├── cmd/
│   │   ├── api/       # Web API server
│   │   └── worker/    # Background task processor
│   └── ...
├── docker-compose.yaml
└── .env.example
```

## Contributing

Found a bug? Have an idea for improvement? PRs are welcome. Just make sure your code doesn't break existing functionality.
