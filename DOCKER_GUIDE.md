# 🐳 Docker for Apex: The Production Standard

Since we're moving to "Serious Coding," Docker is the industry standard for ensuring your code runs exactly the same way on your laptop as it does in the cloud.

## 1. What is Docker? (The "Shipping Container" Analogy)
Before shipping containers, loading a ship was a nightmare. You had loose sacks of flour, barrels of oil, and crates of electronics. They didn't stack, they leaked, and they were hard to move.

**Docker is a Shipping Container for Code.**
You pack your Go app, the Linux OS it needs, and all its settings into one "Container." Once it's in the container, it doesn't matter if the "Ship" (the computer) is Windows, Mac, or Linux—the container works the same way.

## 2. Key Concepts

### A. The Image (The Blueprint)
An **Image** is a read-only template with instructions for creating a Docker container. Think of it like a `class` in programming.

### B. The Container (The Instance)
A **Container** is a runnable instance of an image. It's isolated from your computer. Think of it like an `object` instantiated from a class.

### C. Docker Compose (The Orchestrator)
Our Apex system needs two things to talk to each other:
1. The **Receiver** (Go App)
2. The **Database** (PostgreSQL)

`docker-compose` allows us to define and run these multi-container applications with a single command.

---

## 3. Installation Guide (Windows)

To run the commands I gave you, you need **Docker Desktop**.

1. **Download**: Go to [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/).
2. **Install**: Follow the installer. It will likely ask you to enable **WSL 2** (Windows Subsystem for Linux). Say **YES**.
3. **Reboot**: You'll need to restart your computer.
4. **Verify**: Open your terminal and run:
   ```cmd
   docker compose version
   ```

---

## 4. The Apex Docker Setup

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine  # Use a lightweight Linux version of Postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: apex
    ports:
      - "5432:5432"            # Map port 5432 inside the container to your 5432
```

When you run `docker compose up -d`, Docker will:
1. Download the Postgres image.
2. Start it in an isolated sandbox.
3. Open a "tunnel" (Port 5432) so your Go code can talk to it.

---

## 5. Recommended Reading
- [Docker Curriculum (Beginner Friendly)](https://docker-curriculum.com/)
- [Official Docker "Get Started" Guide](https://docs.docker.com/get-started/)
- [Docker Compose Overview](https://docs.docker.com/compose/)

Had to use AI on this one since I was a bit new on Docker, it wasn't hard but I asked it to show me the best way to do it and so even asked it to recommend me books. Thankfully now I am well equipped with full knowledge on it. If there is anyone reading this, don't just use AI for coding. Don't just ask and then ignore the rest. You need to read the books/documentation and understand the concepts and actually code it(even if not everything) let it be your assistant(you are the master). There is fun in learning new concepts in our field, that's what even makes it so joyful at times.  
