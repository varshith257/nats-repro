# 🧪 NATS JetStream Failure Reproduction 

This repository demonstrates **three high-severity issues** in NATS JetStream, reproduced using controlled environments with logs and terminal recordings.

## ✅ Reproduced Issues

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

