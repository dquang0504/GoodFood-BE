# GoodFood Backend 🍔⚙️

<p align="center">
  <img src="https://raw.githubusercontent.com/dquang0504/GoodFood-BE/main/GoodFood-BE/assets/GoodFood-BE-cover.png" alt="GoodFood Banner" width="450" />
</p>

<h3 align="center">
  <i>Backend infrastructure for smarter, faster, and secure online food ordering</i>
</h3>

<p align="center">
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.24-blue" /></a>
  <a href="https://www.postgresql.org/"><img src="https://img.shields.io/badge/Postgres-15-blue" /></a>
  <a href="https://www.docker.com/"><img src="https://img.shields.io/badge/Docker-ready-blue" /></a>
  <a href="https://github.com/dquang0504/GoodFood-BE/actions">
    <img src="https://github.com/dquang0504/GoodFood-BE/actions/workflows/go-ci.yml/badge.svg" alt="CI Status"/>
  </a>
  <a href="https://codecov.io/gh/dquang0504/GoodFood-BE"><img src="https://img.shields.io/codecov/c/github/dquang0504/GoodFood-BE" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-green.svg" /></a>
</p>

Backend service for **GoodFood**, an e-commerce website tailored for individual restaurant use and specialized in food ordering.  
This repository handles the **core business logic, database management, and API services** for the GoodFood ecosystem.

---

## 🚀 Features
- User authentication (JWT-based, Google OAuth integration planned)
- Product management (CRUD, categories, reviews, ratings)
- Shopping cart and order management
- Integrated **VNPAY** and **PayPal** payment gateways
- Redis caching for:
  - Product listing & details
  - Review analytics
- Admin dashboard APIs for product and order analytics
- Real-time notifications with **WebSocket**
- TensorFlow integration for recommendation & analytics
- Dockerized deployment for portability & maintainability

---

## 🛠️ Tech Stack
- **Language:** Go (Golang)
- **Frameworks & Tools:** Gin, Fiber, Resty
- **Database:** PostgreSQL
- **ORM / DB Tools:** SQLBoiler
- **Cache:** Redis
- **Payments:** VNPAY, PayPal
- **Auth:** JWT, OAuth 2.0
- **Realtime:** WebSocket
- **Deployment:** Docker, AWS EC2
- **API:** REST API

---

## 📂 Project Structure
```bash
GoodFood-BE/
├── assets/           # Store media files
├── bin/worker        # Worker file         
├── cmd/              # Application entrypoints
├── internal/         # Core business logic
|   ├── auth/         # Middleware for authentication & authorization
|   ├── database/     # Database connection initialization and clean up
|   ├── dto/          # DTO
|   ├── jobs/         # Async concurrent jobs (sending mails, processing images)
|   ├── redis-database/      # Redis database connection
│   ├── server/       # HTTP handlers & routes
│   ├── models/       # SQLBoiler models
│   └── services/     # Business services
├── configs/          # Config & environment files
├── migrations/       # Database migrations
├── Dockerfile
├── docker-compose.yml
└── README.md
```
---

## ⚙️ Getting Started
## Prerequisites
* Go 1.22+
* PostgreSQL 15+
* Redis 7+
* Docker (optional)

## Installation
```bash
# Clone the repository
git clone https://github.com/dquang0504/GoodFood-BE.git
cd GoodFood-BE

# Install dependencies
go mod tidy
```
---

## Running Locally
```bash
# Start PostgreSQL & Redis via docker-compose
docker-compose up -d

# Run migrations
go run cmd/migrate/main.go

# Start the backend server
go run cmd/server/main.go
```

The server will be available at:
👉 http://localhost:8080

---

## 🔑 Environment Variables
Create a .env file in the root directory:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=goodfood

REDIS_HOST=localhost
REDIS_PORT=6379

VNPAY_SECRET=your_vnpay_secret
PAYPAL_CLIENT_ID=your_paypal_client_id
PAYPAL_SECRET=your_paypal_secret

JWT_SECRET=your_jwt_secret
```

---

## 📖 API Documentation
Swagger docs available at /swagger/index.html

---

## 🚀 Deployment
Dockerfile and docker-compose.yml included.

CI/CD via GitHub Actions.

AWS EC2 setup guide in docs/deployment.md.

---

## 🤝 Contributing
We welcome contributions!

Fork the repo

Create a new branch (feature/my-feature)

Commit changes (git commit -m 'Add feature')

Push branch & open a PR

Please follow Go best practices.

---

## 📜 License
Distributed under the MIT License. See LICENSE for more information.
