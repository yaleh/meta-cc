# Technical Debt Tracking System Design - Iteration 3

**Created**: 2025-10-17
**Purpose**: Track technical debt trends over time for data-driven decisions

---

## Tracking System Architecture

### 1. Data Collection

**Automated Metrics Collection**:
```yaml
collection_frequency: weekly
collection_trigger: "git push to main"
collection_tools:
  - gocyclo -over 10 ./... | tee debt/complexity-$(date +%Y-%m-%d).txt
  - dupl -threshold 50 ./... | tee debt/duplication-$(date +%Y-%m-%d).txt
  - go test -coverprofile=debt/coverage-$(date +%Y-%m-%d).out ./...
  - staticcheck ./... | tee debt/static-$(date +%Y-%m-%d).txt
```

**Storage Structure**:
```
/project-root/
  .debt-tracking/
    baselines/
      2025-10-17-baseline.json    # SQALE index, TD ratio, rating
    metrics/
      complexity/
        2025-10-17.json           # Weekly complexity data
      coverage/
        2025-10-17.json           # Weekly coverage data
    reports/
      monthly/
        2025-10-monthly.md        # Monthly trend report
```

---

### 2. Baseline Storage

**Baseline Schema** (JSON):
```json
{
  "date": "2025-10-17",
  "sqale": {
    "total_debt": 66.0,
    "development_cost": 425.3,
    "td_ratio": 15.52,
    "rating": "C"
  },
  "by_category": {
    "complexity_debt": 54.5,
    "duplication_debt": 1.0,
    "coverage_debt": 10.0,
    "static_analysis_debt": 0.5
  },
  "hotspots": [
    {"file": "cmd/mcp-server/executor.go", "debt": 8.0},
    {"file": "internal/mcp/tools_project.go", "debt": 4.0}
  ]
}
```

**Baseline Frequency**: Quarterly or after significant changes

---

### 3. Trend Tracking

**Time Series Metrics**:
```yaml
tracked_metrics:
  - metric: td_ratio
    frequency: weekly
    alert_threshold: 16.0  # Alert if >16%

  - metric: complexity_debt
    frequency: weekly
    trend: "decreasing"  # Expected trend

  - metric: test_coverage
    frequency: weekly
    alert_threshold: 75.0  # Alert if <75%

  - metric: hotspot_count
    frequency: monthly
    target: "<5"  # Target <5 critical hotspots
```

**Trend Analysis**:
- Calculate week-over-week change
- Identify debt accumulation rate
- Project when TD ratio hits thresholds

---

### 4. Visualization

**Dashboard Components**:

1. **TD Ratio Trend** (Line Chart):
   - X-axis: Time (weeks)
   - Y-axis: TD Ratio (%)
   - Threshold lines: 10% (B rating), 20% (C rating)

2. **Debt by Category** (Stacked Bar):
   - Complexity, Duplication, Coverage, Static Analysis
   - Shows debt composition over time

3. **Hotspot Heatmap**:
   - Files × Time matrix
   - Color intensity = debt hours

4. **Coverage Trend** (Line Chart):
   - Overall coverage %
   - Coverage by module

5. **Paydown Velocity** (Speedometer):
   - Hours of debt paid down per month
   - Target velocity: 10 hours/month

**Implementation**:
- Static site generator (Hugo, Jekyll)
- Charts: Plotly, D3.js
- Auto-generated from JSON data

---

### 5. Alerting

**Alert Rules**:
```yaml
alerts:
  - name: "TD Ratio Exceeded"
    condition: "td_ratio > 16%"
    action: "Email team lead, create tracking issue"

  - name: "Coverage Drop"
    condition: "coverage < 75%"
    action: "Block PR merge, require coverage fix"

  - name: "New Hotspot"
    condition: "file_debt > 5 hours && new"
    action: "Flag in code review"

  - name: "Debt Accumulation"
    condition: "Δtd_ratio > 1% in 2 weeks"
    action: "Schedule debt review meeting"
```

---

### 6. Reporting

**Weekly Summary** (Automated Email):
```
Subject: Technical Debt Weekly Summary (2025-10-24)

TD Ratio: 15.2% (was 15.5%, -0.3% ↓)
SQALE Rating: C (Moderate)

Debt Paid Down This Week: 2.0 hours
- Refactored executor.go::buildCommand (1.5 hours)
- Increased githelper coverage (0.5 hours)

New Debt Added: 0.5 hours
- New feature in query_tools.go (+0.5 hours complexity)

Net Change: -1.5 hours (good progress!)

Top 3 Hotspots:
1. cmd/mcp-server/executor.go (6.5 hours, was 8.0)
2. internal/mcp/tools_project.go (4.0 hours, unchanged)
3. cmd/query_tools.go (5.0 hours, unchanged)

Action Items:
- Continue executor.go refactoring
- Review new query_tools.go feature for complexity
```

**Monthly Report** (Trend Analysis):
- TD ratio trend (chart)
- Debt accumulation vs paydown
- Hotspot evolution
- Recommendations for next month

**Quarterly Report** (Strategic Review):
- SQALE rating progression
- Architecture debt assessment
- Paydown ROI analysis
- Prevention effectiveness

---

## Implementation

### Step 1: Setup (1 day)
```bash
# Create tracking directory
mkdir -p .debt-tracking/{baselines,metrics,reports}

# Create baseline
python tools/calculate-sqale.py > .debt-tracking/baselines/$(date +%Y-%m-%d).json

# Add to .gitignore
echo ".debt-tracking/metrics/" >> .gitignore
echo ".debt-tracking/reports/" >> .gitignore
git add .debt-tracking/baselines/
```

### Step 2: Automate Collection (CI/CD)
```yaml
# .github/workflows/debt-tracking.yml
name: Track Technical Debt
on:
  push:
    branches: [main]
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday

jobs:
  track-debt:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Collect Metrics
        run: |
          gocyclo -over 10 ./... > .debt-tracking/metrics/complexity-$(date +%Y-%m-%d).txt
          go test -coverprofile=.debt-tracking/metrics/coverage-$(date +%Y-%m-%d).out ./...
      - name: Calculate SQALE
        run: python tools/calculate-sqale.py > .debt-tracking/baselines/$(date +%Y-%m-%d).json
      - name: Commit Results
        run: |
          git config user.name "Debt Tracker Bot"
          git add .debt-tracking/baselines/
          git commit -m "chore: update technical debt baseline [skip ci]"
          git push
```

### Step 3: Dashboard Setup (1 day)
```bash
# Install dashboard generator
npm install -g technical-debt-dashboard

# Generate dashboard
td-dashboard generate --input .debt-tracking/baselines/ --output public/debt-dashboard.html

# Serve locally
cd public && python -m http.server 8080
# Visit http://localhost:8080/debt-dashboard.html
```

### Step 4: Alerts (0.5 days)
```python
# tools/check-debt-alerts.py
import json
from datetime import datetime

with open('.debt-tracking/baselines/latest.json') as f:
    data = json.load(f)

if data['sqale']['td_ratio'] > 16.0:
    send_alert("TD ratio exceeded 16%: " + str(data['sqale']['td_ratio']))

if data['sqale']['td_ratio'] > data['previous']['td_ratio'] + 1.0:
    send_alert("TD ratio increased by >1%")
```

---

## Expected Impact

**Visibility Improvement**:
- Before: Point-in-time snapshots, no trends
- After: Weekly trends, monthly projections, quarterly reviews

**Decision Making**:
- Data-driven prioritization (highest accumulation hotspots)
- Early warning (alerts before debt spikes)
- ROI validation (paydown effectiveness measured)

**V_tracking Improvement**:
- Current: 0.15 (baseline only, no tracking)
- With system: 0.70 (historical tracking, trend analysis, forecasting)
- Expected ΔV_tracking: +0.55

**Time Investment**:
- Setup: 2.5 days (one-time)
- Maintenance: 1 hour/month (review reports)
- Payoff: Early debt detection saves 10+ hours/quarter
