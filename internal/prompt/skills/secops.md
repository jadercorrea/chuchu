
# SecOps

Guidelines for security operations, vulnerability management, and incident response.

## When to Activate

- Setting up security monitoring
- Implementing vulnerability scanning
- Configuring WAF/IDS
- Incident response procedures

## Vulnerability Scanning

### Dependency scanning

```yaml
# GitHub Actions - npm audit
- name: Security audit
  run: npm audit --audit-level=high

# Snyk scanning
- name: Run Snyk
  uses: snyk/actions/node@master
  env:
    SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
```

### Container scanning

```yaml
# Trivy scanner
- name: Scan container
  uses: aquasecurity/trivy-action@master
  with:
    image-ref: 'myapp:${{ github.sha }}'
    format: 'sarif'
    output: 'trivy-results.sarif'
    severity: 'CRITICAL,HIGH'
```

### SAST (Static Analysis)

```yaml
# CodeQL for GitHub
name: CodeQL Analysis

on:
  push:
    branches: [main]
  schedule:
    - cron: '0 0 * * 0'

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: javascript, typescript
      
      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
```

## Security Monitoring

### Audit logging

```typescript
interface SecurityEvent {
  timestamp: string;
  event_type: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  actor: {
    user_id?: string;
    ip_address: string;
    user_agent?: string;
  };
  resource: {
    type: string;
    id: string;
  };
  action: string;
  outcome: 'success' | 'failure';
  details: Record<string, unknown>;
}

function logSecurityEvent(event: SecurityEvent) {
  // Send to SIEM
  logger.info({
    ...event,
    log_type: 'security_audit',
  });
}

// Usage
logSecurityEvent({
  timestamp: new Date().toISOString(),
  event_type: 'authentication',
  severity: 'medium',
  actor: {
    ip_address: req.ip,
    user_agent: req.headers['user-agent'],
  },
  resource: { type: 'user', id: 'unknown' },
  action: 'login_failed',
  outcome: 'failure',
  details: { reason: 'invalid_credentials', attempts: 3 },
});
```

### Alerting rules

```yaml
# Prometheus alerting rules
groups:
  - name: security
    rules:
      - alert: HighFailedLogins
        expr: rate(auth_login_failed_total[5m]) > 10
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: High rate of failed logins
          
      - alert: SuspiciousAPIAccess
        expr: rate(api_requests_total{status="403"}[5m]) > 50
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: Possible brute force attack
          
      - alert: DataExfiltration
        expr: rate(api_response_bytes_total[5m]) > 100000000
        for: 5m
        labels:
          severity: high
        annotations:
          summary: Unusually high data transfer
```

## WAF Configuration

### NGINX rate limiting

```nginx
# Rate limiting zone
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=login:10m rate=1r/s;

server {
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://backend;
    }
    
    location /auth/login {
        limit_req zone=login burst=3 nodelay;
        proxy_pass http://backend;
    }
}
```

### ModSecurity rules

```nginx
# Enable ModSecurity
modsecurity on;
modsecurity_rules_file /etc/nginx/modsecurity/main.conf;

# Custom rules
SecRule REQUEST_METHOD "!^(GET|HEAD|POST|PUT|DELETE)$" \
    "id:1001,phase:1,deny,status:405,msg:'Invalid HTTP method'"

SecRule REQUEST_HEADERS:Content-Type "!application/json" \
    "id:1002,phase:1,deny,status:415,chain"
SecRule REQUEST_URI "@beginsWith /api/"
```

## Incident Response

### Runbook template

```markdown
# Incident: [INCIDENT_TYPE]

## Detection
- **Alert Source**: [Monitoring system]
- **Initial Indicators**: [What triggered the alert]

## Severity Assessment
| Severity | Criteria |
|----------|----------|
| Critical | Data breach, service down |
| High | Security control bypassed |
| Medium | Suspicious activity |
| Low | Anomaly, no impact |

## Response Steps

### 1. Containment (First 15 minutes)
- [ ] Isolate affected systems
- [ ] Block malicious IPs
- [ ] Revoke compromised credentials

### 2. Investigation (Next 1-4 hours)
- [ ] Collect logs from affected systems
- [ ] Identify attack vector
- [ ] Determine scope of compromise

### 3. Eradication
- [ ] Remove malicious artifacts
- [ ] Patch vulnerabilities
- [ ] Reset affected credentials

### 4. Recovery
- [ ] Restore from clean backups
- [ ] Verify system integrity
- [ ] Re-enable services

### 5. Post-Incident
- [ ] Document timeline
- [ ] Root cause analysis
- [ ] Update detection rules
```

### Automated response

```typescript
// Auto-block suspicious IPs
async function handleSecurityAlert(alert: SecurityAlert) {
  if (alert.type === 'brute_force' && alert.severity === 'high') {
    // Block IP at WAF
    await waf.blockIP(alert.source_ip, {
      duration: '24h',
      reason: `Automated block: ${alert.type}`,
    });
    
    // Notify security team
    await slack.notify('#security-alerts', {
      text: `🚨 Auto-blocked IP ${alert.source_ip}`,
      details: alert,
    });
    
    // Create incident ticket
    await jira.createTicket({
      project: 'SEC',
      type: 'Incident',
      summary: `Security Alert: ${alert.type}`,
      priority: 'High',
    });
  }
}
```

## Secrets Rotation

### Automated rotation

```typescript
// Rotate API keys monthly
async function rotateAPIKeys() {
  const services = await getServicesWithAPIKeys();
  
  for (const service of services) {
    // Generate new key
    const newKey = await generateSecureKey();
    
    // Update in secrets manager
    await secretsManager.rotate(service.keyId, newKey);
    
    // Update service configuration
    await updateServiceConfig(service.id, { apiKey: newKey });
    
    // Verify service health
    await verifyServiceHealth(service.id);
    
    // Log rotation event
    logSecurityEvent({
      event_type: 'secret_rotation',
      resource: { type: 'api_key', id: service.keyId },
      action: 'rotate',
      outcome: 'success',
    });
  }
}
```

## Compliance Checks

### Automated compliance scanning

```yaml
# Terraform compliance
- name: Run Checkov
  uses: bridgecrewio/checkov-action@master
  with:
    directory: ./terraform
    framework: terraform
    check: CKV_AWS_17,CKV_AWS_18  # S3 encryption, logging

# Kubernetes compliance
- name: Run Kubesec
  run: |
    kubesec scan k8s/*.yaml
```
