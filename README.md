# Go‑Pub‑Sub

A lightweight publish‑subscribe service implemented in Go.  
This project allows you to set up topics, publish messages to those topics, and have subscribers receive them in real time. Built for simplicity, clarity, and ease of extension.

---

## 🧩 System Architecture Diagram

<img width="692" height="584" alt="Screenshot 2025-10-29 at 4 16 47 PM" src="https://github.com/user-attachments/assets/65221175-8a05-406f-846f-61e886d36134" />

---
## 🧠 Tech Stack Used

This project leverages a combination of modern tools and technologies to ensure performance, scalability, and real-time data handling.

| Technology | Purpose |
|-------------|----------|
| **Go (Golang)** | Core backend service, REST APIs, and WebSocket handling |
| **WebSocket** | Enables real-time bi-directional communication between client and server |
| **Redis** | Pub/Sub messaging system for instant data broadcast and event streaming |
| **MongoDB** | Database for storing user information, uploads, and message logs |
| **Cloudinary** | Cloud-based image upload and storage management |
| **Gemini API** | AI-powered caption generation or intelligent message processing |
| **Docker** | Containerization for consistent, scalable deployments across environments |

---

## 🚀 Features

- Topic creation and management (publish/subscribe)  
- Real-time message delivery to subscribed clients  
- Built-in Docker support for easy deployment  
- Environment configuration via `.env` sample  
- Modular codebase making it easy to extend (commands, internal logic)  
- Written in Go, optimized for performance and simplicity  


## 🔧 Getting Started

### Prerequisites

- Go (version 1.x)  
- Docker (optional, if you prefer container deployment)  
- A terminal or command line interface  

### Installation & Setup

1. Clone the repository  
   ```bash
   git clone https://github.com/sachinggsingh/Go‑Pub‑Sub.git
   cd Go‑Pub‑Sub
   ```

2. Copy the `.env.sample` file to `.env` and set your configuration values  
   ```bash
   cp .env.sample .env
   # Edit .env to adjust ports, topic defaults, etc.
   ```

3. Build and run with Go  
   ```bash
   go build -o pubsub ./cmd
   ./pubsub
   ```

   Or using Docker:  
   ```bash
   docker build -t go‑pubsub .
   docker run --env-file .env -p <your_port>:<container_port> go‑pubsub
   ```

4. Access the service (e.g., via HTTP endpoints or CLI commands) — see below for usage.

---

## 🛠 Usage

### Create a topic
```
curl -X POST http://localhost:<port>/topics      -H "Content-Type: application/json"      -d '{"name":"my-topic"}'
```

### Publish a message
```
curl -X POST http://localhost:<port>/topics/my-topic/publish      -H "Content-Type: application/json"      -d '{"message":"Hello, world!"}'
```

### Subscribe to a topic
```
curl http://localhost:<port>/topics/my-topic/subscribe
```
This will keep the connection open and stream new messages as they come in.

_(Adjust the URLs and ports per your `.env` configuration.)_

---

## 🔍 Architecture & Code Structure

- `cmd/` — Application entry point, command/HTTP routing logic  
- `internal/` — Core logic: topic management, message routing, subscription handling  
- `.env.sample` — Sample environment variables  
- `Dockerfile`, `.dockerignore` — Container setup files  
- `go.mod`, `go.sum` — Go module dependencies

---

## ✅ Why This Project?

- **Simplicity & clarity** — minimal dependencies, straightforward APIs  
- **Flexibility** — easily extended to persistent back‑end, message queues, clustering  
- **Go’s concurrency model** — leverages Go routines and channels for efficient message delivery  
- **Container‑ready** — deployable via Docker in minutes

---

## 🤝 Contributing

Contributions are very welcome! If you’d like to help:

1. Fork the repo  
2. Create a feature branch (`git checkout -b feature/my‑feature`)  
3. Commit your changes and push (`git push origin feature/my‑feature`)  
4. Open a pull request describing your change  

Please ensure code is well‑documented, maintains existing style, and includes tests (if applicable).

---

## 📬 Contact

Maintained by **Sachin Singh**  
- GitHub: [sachinggsingh](https://github.com/sachinggsingh)  
- Email: sachinggsingh@gmail.com
- Link [Sachin](https://sachinsingh.dev)
- LinkedIn / Instagram: feel free to connect!

---

Thank you for checking out Go‑Pub‑Sub. Happy coding! 🎉
