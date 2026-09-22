# Infrastructure Research & Strategy — Knowledge Graph

**Дата:** 2026-01 | **Статус:** Research Phase | **Назначение:** Discuss & Plan

---

## 📋 Обзор пункта 3: INFRASTRUCTURE

### Что входит

1. **Terraform/IaC scripts** — Infrastructure as Code
2. **CI/CD improvements** — GitHub Actions optimizations
3. **Monitoring stack** — Prometheus, Grafana, alerting
4. **Backup/DR procedures** — Disaster recovery

### Текущее состояние Knowledge Graph

```
❌ NO: Infrastructure as Code (manual cloud setup)
❌ NO: Terraform (no IaC)
⚠️  PARTIAL: CI/CD (GitHub Actions exists but has issues)
⚠️  PARTIAL: Monitoring (Prometheus rules created, no Grafana)
❌ NO: Backup automation (manual procedures only)
```

### Почему это важно

```
💰 Cost: Without IaC, easy to over-provision, hard to track
🚀 Speed: Manual setup = 2-3 hours per environment
🛡️  Safety: No backup automation = data loss risk
📊 Visibility: No monitoring = blind to issues until users report
```

---

## 🔍 RESEARCH: Cloud Platform Options

### Option A: AWS (Amazon Web Services)

**Pros:**
- ✅ Most popular, largest ecosystem
- ✅ Best Terraform support
- ✅ Managed PostgreSQL (RDS), Redis (ElastiCache), Kubernetes (EKS)
- ✅ Excellent backup/disaster recovery options
- ✅ Cost optimization tools built-in

**Cons:**
- ❌ Steep learning curve
- ❌ Pricing complex, easy to overspend
- ❌ Many services to choose from (paralysis)

**Typical Knowledge Graph Setup:**
```
- EC2 t3.medium (backend): $30-50/month
- RDS PostgreSQL db.t3.micro: $30-40/month
- ElastiCache Redis cache.t3.micro: $15-20/month
- ALB (load balancer): $15/month
- Data transfer: $5-10/month
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: ~$95-135/month (~$1,200-1,600/year)
```

**Terraform Complexity:** Medium (many services)

---

### Option B: Google Cloud Platform (GCP)

**Pros:**
- ✅ Better default pricing than AWS
- ✅ Excellent Kubernetes (GKE is best-in-class)
- ✅ Built-in monitoring (Stackdriver)
- ✅ Cleaner UI than AWS
- ✅ Good free tier

**Cons:**
- ❌ Smaller ecosystem than AWS
- ❌ Less documentation for advanced use cases
- ❌ Different naming conventions than AWS

**Typical Knowledge Graph Setup:**
```
- Compute Engine n1-standard-1: $25-35/month
- Cloud SQL PostgreSQL db-f1-micro: $25-35/month
- Cloud Memorystore Redis basic tier: $10-15/month
- Cloud Load Balancing: $15/month
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: ~$75-100/month (~$900-1,200/year)
```

**Terraform Complexity:** Medium (cleaner than AWS)

---

### Option C: Azure (Microsoft Azure)

**Pros:**
- ✅ Enterprise-friendly
- ✅ Good database options (Azure Database for PostgreSQL)
- ✅ Decent pricing for enterprises
- ✅ Good Kubernetes (AKS)

**Cons:**
- ❌ Confusing service names
- ❌ Documentation not as good as AWS/GCP
- ❌ Smaller community

**Typical Knowledge Graph Setup:**
```
- Virtual Machine B1s: $7-10/month
- Azure Database PostgreSQL: $20-30/month
- Azure Cache Redis: $12-18/month
- Load Balancer: $15/month
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: ~$54-73/month (~$650-900/year)
```

**Terraform Complexity:** Medium

---

### Option D: DigitalOcean (Simple & Cheap)

**Pros:**
- ✅ Simple pricing (no hidden charges)
- ✅ Cheap ($5/month per droplet)
- ✅ Easy to use
- ✅ Terraform support excellent
- ✅ Perfect for startups

**Cons:**
- ❌ Limited auto-scaling options
- ❌ Smaller than AWS/GCP
- ❌ Manual Kubernetes setup required

**Typical Knowledge Graph Setup:**
```
- App droplet (2GB RAM): $12/month
- Database droplet PostgreSQL (2GB): $15/month
- Redis droplet (1GB): $6/month
- Managed Database PostgreSQL: $15-30/month (alternative)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: ~$33-48/month (~$400-600/year) ✅ CHEAPEST
```

**Terraform Complexity:** Low (simplest)

---

### Option E: Self-Hosted (On-Premises or VPS)

**Pros:**
- ✅ Full control
- ✅ Can be very cheap ($5-20/month VPS)
- ✅ No vendor lock-in
- ✅ Good for learning

**Cons:**
- ❌ You're responsible for everything (backups, security, scaling)
- ❌ No managed services
- ❌ More operational overhead
- ❌ Hard to scale horizontally

**Typical Knowledge Graph Setup:**
```
- Single VPS (2GB RAM, 50GB SSD): $5-10/month
- Backup storage: $2-5/month
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: ~$7-15/month (~$85-180/year) ✅ CHEAPEST
```

**Terraform Complexity:** Low (but less to automate)

---

## 📊 Comparison Matrix

| Aspect | AWS | GCP | Azure | DO | VPS |
|--------|-----|-----|-------|----|----|
| **Cost/month** | $95-135 | $75-100 | $54-73 | $33-48 | $7-15 |
| **Ease of use** | Hard | Medium | Hard | Easy | Hard |
| **Terraform support** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Auto-scaling** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ❌ |
| **Managed services** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ❌ |
| **Documentation** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | - |
| **For startups** | ❌ | ⚠️ | ❌ | ✅ | ✅ |
| **For scale** | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |

---

## 🎯 RECOMMENDATION FOR KNOWLEDGE GRAPH

### **IF: Starting/MVP Phase**
→ **DigitalOcean** + Terraform
- Cost: $40/month
- Setup time: 2-3 hours
- Scaling: Manual (OK for MVP)
- Terraform: Simpler than AWS

### **IF: Growth Phase (1,000+ users)**
→ **GCP** + Kubernetes (GKE)
- Cost: $100-300/month
- Setup time: Full day
- Scaling: Automatic
- Terraform: Excellent support

### **IF: Enterprise**
→ **AWS** + Kubernetes (EKS)
- Cost: $200-500+/month
- Setup time: 2-3 days
- Scaling: Enterprise-grade
- Terraform: Most comprehensive

### **IF: Open Source / Community**
→ **Self-Hosted VPS** or **DigitalOcean**
- Cost: $10-50/month
- Runs on simple servers
- Community can contribute infrastructure

---

## 🛠️ TERRAFORM: What We'd Create

### Basic Infrastructure (DigitalOcean Example)

```hcl
# terraform/main.tf

terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.0"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

# App server
resource "digitalocean_droplet" "app" {
  name     = "kg-app-${var.environment}"
  region   = var.region
  size     = "s-2vcpu-2gb"  # 2GB RAM, 2 vCPU: $12/month
  image    = "docker-22-04"  # Pre-installed Docker
  
  tags = ["environment:${var.environment}", "app:knowledge-graph"]
}

# Database (PostgreSQL)
resource "digitalocean_database_cluster" "postgres" {
  name       = "kg-db-${var.environment}"
  engine     = "pg"
  version    = "15"
  region     = var.region
  node_count = 1
  size       = "db-s-1vcpu-1gb"  # $15/month
  
  tags = ["environment:${var.environment}"]
}

# Redis cache
resource "digitalocean_database_cluster" "redis" {
  name       = "kg-redis-${var.environment}"
  engine     = "redis"
  version    = "7"
  region     = var.region
  node_count = 1
  size       = "db-s-1vcpu-512mb-10gb"  # $6/month
  
  tags = ["environment:${var.environment}"]
}

# Firewall
resource "digitalocean_firewall" "app" {
  name = "kg-firewall-${var.environment}"
  
  inbound_rule {
    protocol         = "tcp"
    ports            = "80"
    sources {
      addresses = ["0.0.0.0/0"]
    }
  }
  
  inbound_rule {
    protocol         = "tcp"
    ports            = "443"
    sources {
      addresses = ["0.0.0.0/0"]
    }
  }
  
  outbound_rule {
    protocol              = "tcp"
    ports                 = "all"
    destinations {
      addresses = ["0.0.0.0/0"]
    }
  }
  
  droplet_ids = [digitalocean_droplet.app.id]
}

# Outputs
output "app_ip" {
  value = digitalocean_droplet.app.ipv4_address
}

output "db_host" {
  value = digitalocean_database_cluster.postgres.host
}

output "redis_host" {
  value = digitalocean_database_cluster.redis.host
}
```

**Size:** ~150 lines for complete DO infrastructure

---

## 🚀 CI/CD: GitHub Actions Issues & Fixes

### Current Issues

**Issue #1: NLP Service Timeout**
```yaml
# ❌ CURRENT: 600s timeout blocks entire pipeline
nlp:
  healthcheck:
    start_period: 600s    # ← Reduces CI speed
```

**Fix:** Already done (180s) — reduces CI by 7 minutes ✅

---

**Issue #2: No Caching for Dependencies**
```yaml
# ❌ Current: Downloads Go modules every run
- name: Build backend
  run: cd backend && go build ./cmd/server
  # No cache!
```

**Fix:**
```yaml
# ✅ With caching
- uses: actions/cache@v3
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('backend/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-

- name: Build backend
  run: cd backend && go build ./cmd/server
```

**Saves:** 3-5 minutes per build

---

**Issue #3: Docker Builds Not Cached**
```yaml
# ❌ Current: Rebuilds from scratch
- name: Build images
  run: docker build -f backend/Dockerfile -t kg-backend:latest .
  # No layer caching between runs
```

**Fix:**
```yaml
# ✅ With buildx and cache
- uses: docker/setup-buildx-action@v2

- uses: docker/build-push-action@v4
  with:
    context: .
    file: ./backend/Dockerfile
    tags: kg-backend:latest
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

**Saves:** 10-15 minutes on Docker builds

---

**Issue #4: Parallel Jobs Could Be Better**
```yaml
# Current: Some sequential steps
- name: Unit tests
- name: Integration tests  # Waits for unit to finish
- name: Build docker
```

**Fix:** Run in parallel where possible
```yaml
jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps: [...]
  
  integration-tests:
    runs-on: ubuntu-latest
    steps: [...]
  
  build-docker:
    needs: [unit-tests]  # Only depends on unit, not integration
    runs-on: ubuntu-latest
    steps: [...]
```

**Saves:** 10-20 minutes per run

---

## 📊 Monitoring Stack: Detailed Design

### Current State
```
✅ Prometheus rules created (16 alerts)
❌ Prometheus server not deployed
❌ Grafana not set up
❌ Alerting (Slack/email) not configured
❌ SLA tracking not implemented
```

### What We'd Create

**1. Prometheus Configuration**
```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'knowledge-graph'
    env: 'production'

alerting:
  alertmanagers:
  - static_configs:
    - targets:
      - alertmanager:9093

rule_files:
- /etc/prometheus/rules/*.yml

scrape_configs:
- job_name: 'kg-backend'
  kubernetes_sd_configs:
  - role: pod
  relabel_configs: [...]

- job_name: 'kg-postgres'
  static_configs:
  - targets: ['postgres-exporter:9187']

- job_name: 'kg-redis'
  static_configs:
  - targets: ['redis-exporter:9121']
```

**2. Grafana Dashboard**
```json
// dashboard.json
{
  "dashboard": {
    "title": "Knowledge Graph - Production",
    "panels": [
      {
        "title": "Request Rate (req/s)",
        "targets": [
          {"expr": "rate(http_requests_total[5m])"}
        ]
      },
      {
        "title": "Error Rate (%)",
        "targets": [
          {"expr": "rate(http_requests_total{status=~\"5..\"}[5m]) * 100"}
        ]
      },
      {
        "title": "P95 Latency (ms)",
        "targets": [
          {"expr": "histogram_quantile(0.95, http_request_duration_seconds) * 1000"}
        ]
      },
      {
        "title": "Database Pool Utilization (%)",
        "targets": [
          {"expr": "db_pool_utilization_percent"}
        ]
      },
      {
        "title": "PostgreSQL Connections",
        "targets": [
          {"expr": "postgresql_stat_activity_count"}
        ]
      }
    ]
  }
}
```

**3. Alertmanager Configuration**
```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m

route:
  receiver: 'default'
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 24h
  routes:
  - match:
      severity: critical
    receiver: 'critical'
    continue: true
    repeat_interval: 5m

receivers:
- name: 'default'
  slack_configs:
  - api_url: ${SLACK_WEBHOOK_URL}
    channel: '#alerts'
    title: 'Alert: {{ .GroupLabels.alertname }}'

- name: 'critical'
  pagerduty_configs:
  - service_key: ${PAGERDUTY_KEY}
  email_configs:
  - to: 'ops-team@example.com'
    from: 'alertmanager@example.com'
```

---

## 💾 Backup & Disaster Recovery

### Current State
```
❌ NO automated backups
❌ NO disaster recovery plan
❌ NO backup verification
❌ NO RTO/RPO defined
```

### What We'd Create

**1. PostgreSQL Backups (AWS S3)**
```bash
#!/bin/bash
# backup-db.sh

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="kg-db-backup-${TIMESTAMP}.sql.gz"

# Full backup
pg_dump \
  -h $DB_HOST \
  -U $DB_USER \
  -d $DB_NAME \
  --format=custom \
  --compress=9 \
  | aws s3 cp - "s3://kg-backups/postgres/${BACKUP_FILE}"

# Keep only last 30 days
aws s3 rm "s3://kg-backups/postgres" \
  --recursive \
  --exclude "*" \
  --include "kg-db-backup-*" \
  --older-than-days 30
```

**2. Backup Schedule (cron or K8s CronJob)**
```yaml
# kubernetes/backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: kg-db-backup
spec:
  schedule: "0 2 * * *"  # 2 AM daily
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15-alpine
            command:
            - /bin/sh
            - -c
            - |
              pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME \
                | gzip \
                | aws s3 cp - s3://kg-backups/$(date +%Y%m%d_%H%M%S).sql.gz
            env:
            - name: DB_HOST
              valueFrom:
                secretKeyRef:
                  name: db-credentials
                  key: host
          restartPolicy: OnFailure
```

**3. Recovery Testing (Monthly)**
```bash
#!/bin/bash
# restore-and-test.sh

# Download latest backup
LATEST=$(aws s3api list-objects-v2 \
  --bucket kg-backups \
  --prefix postgres \
  --query 'Contents | sort_by(@, &LastModified) | [-1].[Key]' \
  --output text)

aws s3 cp "s3://kg-backups/${LATEST}" ./backup.sql.gz

# Restore to test DB
createdb kg_restore_test
gunzip -c backup.sql.gz | psql -d kg_restore_test

# Verify
psql -d kg_restore_test -c "SELECT COUNT(*) FROM notes;"

# Cleanup
dropdb kg_restore_test
rm backup.sql.gz
```

---

## 📈 Implementation Roadmap (Proposed)

### Phase 1: Week 1 (Foundation)
```
- Choose cloud platform (recommend: DigitalOcean for MVP)
- Create basic Terraform configuration
- Test manual provisioning
- Document infrastructure setup
⏱️ Effort: 8-10 hours
```

### Phase 2: Weeks 2-3 (CI/CD)
```
- Add GitHub Actions caching (deps + Docker layers)
- Optimize parallel jobs
- Reduce CI time from 25min → 15min
- Document CI/CD flow
⏱️ Effort: 6-8 hours
```

### Phase 3: Weeks 4-5 (Monitoring)
```
- Deploy Prometheus server
- Set up Grafana dashboards
- Configure Alertmanager
- Test alerts (Slack/email)
⏱️ Effort: 10-12 hours
```

### Phase 4: Weeks 6-7 (Backup/DR)
```
- Implement automated backups (daily)
- Create backup verification script
- Set up restore procedures
- Document RTO/RPO
- Monthly restore testing
⏱️ Effort: 8-10 hours
```

### Phase 5: Week 8 (Documentation)
```
- Document complete infrastructure setup
- Create runbooks (deploy, backup, recovery)
- Team training on monitoring
- Create alert response procedures
⏱️ Effort: 6-8 hours
```

**Total:** ~40-50 hours over 8 weeks

---

## 💰 Cost Comparison

### MVP (DigitalOcean)
```
App server (s-2vcpu-2gb):  $12/month
DB (db-s-1vcpu-1gb):      $15/month
Redis (db-s-1vcpu-512mb): $6/month
Backup storage (100GB):   $5/month
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: $38/month = $456/year
```

### Growth (GCP)
```
Compute (n1-standard-1):   $30/month
Cloud SQL PostgreSQL:      $30/month
Cloud Memorystore Redis:   $12/month
Load Balancer:             $15/month
Backup storage:            $10/month
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: $97/month = $1,164/year
```

### Enterprise (AWS)
```
EC2 (t3.medium):           $35/month
RDS PostgreSQL:            $35/month
ElastiCache Redis:         $18/month
ALB:                       $15/month
Data transfer:             $10/month
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TOTAL: $113/month = $1,356/year
```

---

## 🎯 Questions for Discussion

### Platform Choice
1. **Where is your team/users located?**
   - AWS: US, EU, most regions
   - GCP: US, EU, Asia-Pacific
   - DigitalOcean: US, EU, Asia-Pacific
   - Azure: Enterprise/Microsoft-centric

2. **How many users initially?**
   - < 1,000: DigitalOcean ($40/month)
   - 1,000-10,000: GCP ($100/month)
   - 10,000+: AWS ($200+/month)

3. **Any cloud vendor preference?**
   - If none → DigitalOcean is simplest
   - If AWS → we'll use AWS
   - If GCP → we'll use GCP

### Implementation Priority
1. **What's blocking you most right now?**
   - Lack of monitoring? (start with Prometheus)
   - Manual deployments? (start with Terraform)
   - No backups? (start with backup automation)

2. **How much time can you invest?**
   - 10 hours/week → 4-5 weeks complete
   - 5 hours/week → 8-10 weeks complete
   - < 5 hours/week → focus on most critical

---

## 📚 Deliverables We'd Create

### If you choose DigitalOcean:
```
✅ Terraform code (complete infrastructure)
✅ GitHub Actions improvements (caching, parallelization)
✅ Prometheus + Grafana setup
✅ Backup scripts + restore procedures
✅ Monitoring dashboard
✅ Alert configurations
✅ Runbook documentation
✅ Team training materials
```

### If you choose AWS:
```
✅ Same as above, but with AWS services
✅ Terraform for ECS, RDS, ElastiCache
✅ Additional: AWS Backup, AWS Monitoring
✅ IAM policies and security groups
```

### If you choose GCP:
```
✅ Same as above, but with GCP services
✅ Terraform for Compute, Cloud SQL, Memorystore
✅ Additional: Cloud Monitoring integration
```

---

## 🚀 Next Steps

1. **Choose platform** — AWS/GCP/DigitalOcean/Self-hosted?
2. **Prioritize phases** — All or specific ones?
3. **Timeline** — When do you need this?
4. **Team involvement** — Who implements?
5. **Budget** — Any constraints?

**Recommendation:** Start with DigitalOcean + Terraform (simplest, cheapest). Can migrate to AWS/GCP later.

---

**Ждём вашего feedback перед началом реализации!**

Какие вопросы/уточнения по infrastructure?
