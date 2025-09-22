# GoodFood-BE 🍔🗄️
[![Go Version](https://img.shields.io/badge/Go-1.24-blue)](https://golang.org/)
[![Postgres](https://img.shields.io/badge/Postgres-15-blue)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-ready-blue)](https://www.docker.com/)

Backend service for **GoodFood**, an e-commerce platform specialized in food ordering.  
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
- Dockerized deployment for portability & maintainability

---

## 🛠️ Tech Stack
- **Language:** Go (Golang)
- **Frameworks & Tools:** Gin, SQLBoiler, GORM (if any)
- **Database:** PostgreSQL
- **Cache:** Redis
- **Payments:** VNPAY, PayPal
- **Deployment:** Docker, AWS EC2

---

## 📂 Project Structure
GoodFood-BE/
│── cmd/ # Application entrypoints
│── internal/ # Core business logic
│ ├── server/ # HTTP handlers & routes
│ ├── models/ # SQLBoiler models
│ └── services/ # Business services
│── configs/ # Config & environment files
│── migrations/ # Database migrations
│── Dockerfile
│── docker-compose.yml
│── README.md

yaml
Copy code

---

## ⚙️ Getting Started

### Prerequisites
- Go 1.22+
- PostgreSQL 15+
- Redis 7+
- Docker (optional)

### Installation
```bash
# Clone the repository
git clone https://github.com/dquang0504/GoodFood-BE.git
cd GoodFood-BE

# Install dependencies
go mod tidy
Running Locally
bash
Copy code
# Start PostgreSQL & Redis via docker-compose
docker-compose up -d

# Run migrations
go run cmd/migrate/main.go

# Start the backend server
go run cmd/server/main.go
The server will be available at:
👉 http://localhost:8080

🔑 Environment Variables
Create a .env file in the root directory:

env
Copy code
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
📖 API Documentation
Swagger docs available at /swagger/index.html

Example Postman collection: GoodFood API Docs

🚀 Deployment
Dockerfile and docker-compose.yml included.

CI/CD via GitHub Actions (to be added).

AWS EC2 setup guide in docs/deployment.md.

🤝 Contributing
We welcome contributions!

Fork the repo

Create a new branch (feature/my-feature)

Commit changes (git commit -m 'Add feature')

Push branch & open a PR

Please follow Go best practices.

📜 License
Distributed under the MIT License. See LICENSE for more information.

yaml
Copy code

---
