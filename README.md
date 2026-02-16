# Distributed TCP Chat with TUI

A high-performance, distributed TCP chat application written in **Go**. This project features a sleek Terminal User Interface (TUI) and uses **Redis Pub/Sub** to synchronize messages across multiple server nodes, allowing for horizontal scalability.

## Features

* **Distributed Architecture:** Multiple server instances can run simultaneously, synced via Redis.
* **Modern TUI:** Built with `Bubble Tea` and `Lip Gloss` for a responsive, stylish terminal experience.
* **Containerized:** Includes a `Dockerfile` for easy deployment and testing.
* **Concurrency:** Built using Go's lightweight goroutines for handling multiple TCP connections.

## Tech Stack

* **Language:** Go (Golang)
* **Backend:** TCP Sockets, Redis Pub/Sub
* **Frontend (TUI):** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
* **Containerization:** Docker

## Getting Started

### Prerequisites

* Go 1.20+
* Redis server running (locally or via Docker)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/your-repo-name.git
   cd your-repo-name
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

### Running the Distributed System

To test the distributed nature, you can run two server instances on different ports:

1. Start Redis (if not running):
   ```bash
   docker run -d --name redis -p 6379:6379 redis
   ```

2. Run Server Node 1:
   ```bash
   PORT=9001 go run main.go
   ```

3. Run Server Node 2:
   ```bash
   PORT=9002 go run main.go
   ```

4. Launch the TUI Client:
   ```bash
   go run tui.go -port 9001
   # In another terminal
   go run tui.go -port 9002
   ```

Messages sent to Node 1 will now appear on clients connected to Node 2!

## Docker Usage

You can build and run the chat server using Docker:

```bash
docker build -t my-chat-app .
docker run -p 9001:9001 my-chat-app
```

## 📂 Project Structure

* `main.go`: The server logic and Redis integration.
* `tui.go`: The Terminal User Interface client.
* `Dockerfile`: Container configuration.
* `go.mod`: Dependency management.

---

### 💡 Quick Tip on Your Dockerfile

Since you're using Redis, make sure your `Dockerfile` (or a `docker-compose.yml` if you want to be fancy) is configured to look for the Redis host via environment variables rather than `localhost`, as `localhost` inside a container refers to the container itself!
