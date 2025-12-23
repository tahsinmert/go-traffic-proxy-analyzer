# Deployment Guide

This guide covers deploying RCIG in various environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Binary Deployment](#binary-deployment)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Production Considerations](#production-considerations)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Prerequisites

- Target backend service running and accessible
- Network connectivity between RCIG and backend
- Appropriate ports open (default: 8080)

## Binary Deployment

### Download Pre-built Binary

```bash
# Linux AMD64
wget https://github.com/tahsinmert/go-traffic-proxy-analyzer/releases/latest/download/rcig-linux-amd64
chmod +x rcig-linux-amd64
sudo mv rcig-linux-amd64 /usr/local/bin/rcig

# macOS AMD64
wget https://github.com/tahsinmert/go-traffic-proxy-analyzer/releases/latest/download/rcig-darwin-amd64
chmod +x rcig-darwin-amd64
sudo mv rcig-darwin-amd64 /usr/local/bin/rcig

# Verify installation
rcig --version
```

### Build from Source

```bash
git clone https://github.com/tahsinmert/go-traffic-proxy-analyzer.git
cd go-traffic-proxy-analyzer
make build
sudo cp bin/rcig /usr/local/bin/
```

### Configuration

Create configuration file `/etc/rcig/config.yml`:

```yaml
server:
  listen_addr: ":8080"
  target_url: "http://backend-service:9000"

metrics:
  enabled: true
  history_size: 1000

analyzer:
  tick_interval: "5s"
  drift_threshold: 0.40

alerts:
  - type: "console"
    enabled: true

logging:
  level: "info"
  format: "json"
```

### Systemd Service (Linux)

Create `/etc/systemd/system/rcig.service`:

```ini
[Unit]
Description=RCIG Reverse Proxy
After=network.target

[Service]
Type=simple
User=rcig
Group=rcig
Environment="RCIG_TARGET_URL=http://backend-service:9000"
Environment="RCIG_LISTEN_ADDR=:8080"
ExecStart=/usr/local/bin/rcig
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=rcig

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/rcig

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
# Create user
sudo useradd -r -s /bin/false rcig

# Enable service
sudo systemctl daemon-reload
sudo systemctl enable rcig
sudo systemctl start rcig

# Check status
sudo systemctl status rcig
sudo journalctl -u rcig -f
```

## Docker Deployment

### Using Docker Hub Image

```bash
docker run -d \
  --name rcig \
  -p 8080:8080 \
  -e RCIG_TARGET_URL=http://backend:9000 \
  -e RCIG_LISTEN_ADDR=:8080 \
  --restart unless-stopped \
  rcig:latest
```

### Building Custom Image

```bash
# Build
docker build -t rcig:custom .

# Run
docker run -d \
  --name rcig \
  -p 8080:8080 \
  -e RCIG_TARGET_URL=http://backend:9000 \
  --restart unless-stopped \
  rcig:custom
```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  rcig:
    image: rcig:latest
    ports:
      - "8080:8080"
    environment:
      - RCIG_TARGET_URL=http://backend:9000
      - RCIG_LISTEN_ADDR=:8080
    restart: unless-stopped
    depends_on:
      - backend
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/rcig/healthz"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s

  backend:
    image: your-backend:latest
    ports:
      - "9000:9000"
    restart: unless-stopped
```

Deploy:

```bash
docker-compose up -d
docker-compose logs -f rcig
```

## Kubernetes Deployment

### ConfigMap

`rcig-configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: rcig-config
  namespace: default
data:
  config.yml: |
    server:
      listen_addr: ":8080"
      target_url: "http://backend-service:9000"
    metrics:
      enabled: true
      history_size: 1000
    analyzer:
      tick_interval: "5s"
      drift_threshold: 0.40
    logging:
      level: "info"
      format: "json"
```

### Deployment

`rcig-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rcig
  namespace: default
  labels:
    app: rcig
spec:
  replicas: 2
  selector:
    matchLabels:
      app: rcig
  template:
    metadata:
      labels:
        app: rcig
    spec:
      containers:
      - name: rcig
        image: rcig:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: RCIG_TARGET_URL
          value: "http://backend-service:9000"
        - name: RCIG_LISTEN_ADDR
          value: ":8080"
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /rcig/healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
          timeoutSeconds: 3
        readinessProbe:
          httpGet:
            path: /rcig/healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 3
        volumeMounts:
        - name: config
          mountPath: /etc/rcig
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: rcig-config
```

### Service

`rcig-service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: rcig
  namespace: default
  labels:
    app: rcig
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: rcig
```

### Ingress (Optional)

`rcig-ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: rcig
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - rcig.example.com
    secretName: rcig-tls
  rules:
  - host: rcig.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: rcig
            port:
              number: 80
```

### Deploy to Kubernetes

```bash
kubectl apply -f rcig-configmap.yaml
kubectl apply -f rcig-deployment.yaml
kubectl apply -f rcig-service.yaml
kubectl apply -f rcig-ingress.yaml

# Check status
kubectl get pods -l app=rcig
kubectl logs -l app=rcig -f
kubectl get svc rcig
```

## Production Considerations

### Resource Sizing

**Small deployment** (< 1000 req/s):
- CPU: 100-500m
- Memory: 64-256Mi

**Medium deployment** (1000-10000 req/s):
- CPU: 500m-2
- Memory: 256Mi-1Gi

**Large deployment** (> 10000 req/s):
- CPU: 2-4
- Memory: 1-4Gi

### High Availability

#### Load Balancer Setup

```
        ┌──────────────┐
        │ Load Balancer│
        └──────┬───────┘
               │
        ┌──────┴───────┐
        │              │
   ┌────▼───┐    ┌────▼───┐
   │ RCIG 1 │    │ RCIG 2 │
   └────┬───┘    └────┬───┘
        │              │
        └──────┬───────┘
               │
        ┌──────▼───────┐
        │   Backend    │
        └──────────────┘
```

#### Multiple Instances

- Run at least 2 instances for HA
- Use health checks
- Configure automatic restarts
- Set up monitoring and alerting

### Security

#### TLS/HTTPS

```yaml
security:
  tls:
    enabled: true
    cert_file: "/etc/rcig/certs/tls.crt"
    key_file: "/etc/rcig/certs/tls.key"
```

#### Authentication

```yaml
auth:
  enabled: true
  api_keys:
    - key: "${API_KEY_ADMIN}"
      name: "admin"
```

#### Network Security

- Restrict inbound ports
- Use network policies (Kubernetes)
- Enable rate limiting
- Set up Web Application Firewall (WAF)

### Logging

#### Centralized Logging

**Fluentd/Fluent Bit**:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
data:
  fluent-bit.conf: |
    [INPUT]
        Name              tail
        Path              /var/log/containers/rcig*.log
        Parser            docker
        Tag               rcig.*
    [OUTPUT]
        Name              es
        Match             rcig.*
        Host              elasticsearch
        Port              9200
```

**Promtail + Loki**:

```yaml
clients:
  - url: http://loki:3100/loki/api/v1/push
scrape_configs:
  - job_name: rcig
    static_configs:
      - targets:
          - localhost
        labels:
          job: rcig
          __path__: /var/log/rcig/*.log
```

### Backup and Disaster Recovery

RCIG is stateless, so no backups needed. However:

1. **Configuration**: Version control your config files
2. **Metrics**: Export to external system (Prometheus)
3. **Alerts**: Configure multiple alert sinks

### Performance Tuning

#### Connection Pool

```go
// In proxy setup
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 100,
    IdleConnTimeout:     90 * time.Second,
}
```

#### Buffer Sizes

```yaml
metrics:
  history_size: 1000  # Increase for more historical data
```

#### Analysis Interval

```yaml
analyzer:
  tick_interval: "10s"  # Increase to reduce CPU usage
```

## Monitoring

### Health Check

```bash
curl http://localhost:8080/rcig/healthz
# Expected: 200 OK
```

### Metrics

```bash
curl http://localhost:8080/rcig/metrics | jq .
```

### Prometheus Integration

```yaml
scrape_configs:
  - job_name: 'rcig'
    static_configs:
      - targets: ['rcig:8080']
    metrics_path: '/metrics'
```

### Grafana Dashboard

Import dashboard ID: TBD (to be created)

Key metrics to monitor:
- Request rate
- Error rate
- Latency percentiles (p50, p95, p99)
- Alert frequency
- RCIG resource usage (CPU, memory)

## Troubleshooting

### RCIG Won't Start

**Check logs**:
```bash
# Systemd
sudo journalctl -u rcig -n 50

# Docker
docker logs rcig

# Kubernetes
kubectl logs -l app=rcig --tail=50
```

**Common issues**:
- Invalid `RCIG_TARGET_URL`
- Port already in use
- Permission denied (check user/group)

### Backend Connection Issues

**Test connectivity**:
```bash
# From RCIG host
curl http://backend-service:9000/health
```

**Check DNS**:
```bash
nslookup backend-service
```

**Verify network policies** (Kubernetes):
```bash
kubectl describe networkpolicy
```

### High Memory Usage

**Check metrics history size**:
```yaml
metrics:
  history_size: 100  # Reduce if needed
```

**Monitor with pprof**:
```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Drift Alerts Not Triggering

**Verify sufficient data**:
- Need `older_window_size + recent_window_size` samples
- Default: 80 + 20 = 100 samples

**Check thresholds**:
```yaml
analyzer:
  drift_threshold: 0.40  # Adjust if needed
```

**Enable debug logging**:
```yaml
logging:
  level: "debug"
```

### Performance Issues

**Check resource limits**:
```bash
# Docker
docker stats rcig

# Kubernetes
kubectl top pod -l app=rcig
```

**Profile the application**:
```bash
curl http://localhost:8080/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof
```

## Upgrading

### Rolling Upgrade (Kubernetes)

```bash
kubectl set image deployment/rcig rcig=rcig:v1.1.0
kubectl rollout status deployment/rcig
```

### Blue-Green Deployment

1. Deploy new version alongside old
2. Shift traffic gradually
3. Monitor for issues
4. Complete cutover or rollback

### Rollback

```bash
# Kubernetes
kubectl rollout undo deployment/rcig

# Docker Compose
docker-compose down
docker-compose up -d --force-recreate
```

## Support

For issues or questions:
- GitHub Issues: https://github.com/tahsinmert/go-traffic-proxy-analyzer/issues
- Documentation: https://github.com/tahsinmert/go-traffic-proxy-analyzer/tree/main/docs
