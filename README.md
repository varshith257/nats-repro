# 🧪 NATS JetStream Failure Reproduction 

This repository demonstrates **three high-severity issues** in NATS JetStream, reproduced using controlled environments with logs and terminal recordings.

## Reproduced Issues

### 1. `monitor goroutine not running`
- **Symptom:** Stream fails to recover leadership after restart.
- **Error:** `JetStream stream 'app > some-stream' is not current: monitor goroutine not running`
- **Trigger:** Multi-node NATS cluster with JetStream enabled and manual restarts.

### 2. Pull Consumer Deadlock with `LastPerSubject` + `MaxAckPending = 1`
- **Symptom:** Consumer fetches 2 messages and then stalls with `context deadline exceeded`.
- **Configuration:**
```bash
Deliver Policy: LastPerSubject
Ack Policy: Explicit
MaxAckPending: 1
```
- **Trigger:** Sending >2 messages with overlapping subjects and low ack window

### 3. NATS Docker Port Conflicts + Residual Container Failures
- **Symptom:** Docker container fails to start due to `port already in use` or container name conflict.
- **Fix Steps:**
```bash
sudo lsof -t -i :4222 | xargs sudo kill -9
docker rm -f natsjs
```

## How to Reproduce

### 1. Clone the Repo
```bash
git clone https://github.com/varshith257/nats-repro.git
cd nats-repro
```

### 2. Start JetStream Server
```bash
docker run -d --name natsjs -p 4222:4222 -p 8222:8222 nats:2.11.4 -js
```

### 3. Set Env and Build
```bash
export NATS_URL=nats://localhost:4222
go build -o fetcher
```

### 4. Run Client
```bash
./fetcher > logs/fetcher.log 2>&1
```

5. Inspect Logs
```bash
cat logs/fetcher.log
cat logs/consumer_info.log
```

### Expected Output
Initial messages are received and acknowledged:

```log
Received five.1, acking…
Received five.2, acking…
````

Then the client stalls with:
```log
Fetch error: context deadline exceeded
```

The consumer info shows:
```lof
Deliver Policy: Last Per Subject
Max Ack Pending: 1
```