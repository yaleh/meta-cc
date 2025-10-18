# Principle: Security-First Priority

**Category**: Principle
**Domain**: dependency-management, security
**Source**: Iteration 1
**Status**: Validated
**Tags**: [security, prioritization, risk-management, vulnerabilities]

---

## Statement

**Patch HIGH and CRITICAL severity vulnerabilities immediately, before all other dependency work.**

---

## Rationale

Security vulnerabilities represent exploitable risks that can lead to:
- Data breaches (unauthorized access to sensitive information)
- System compromises (attacker control of infrastructure)
- Service disruptions (denial of service attacks)
- Compliance violations (regulatory penalties)

High and critical severity vulnerabilities are actively exploited in the wild. Delaying patches increases exposure window and incident probability.

**Time-to-patch is the primary security metric.** Fast patching minimizes:
- Exposure window (time attackers have to exploit)
- Incident probability (fewer days = fewer opportunities)
- Blast radius (if exploited, faster containment)

---

## Evidence from Iterations

### Iteration 1: Vulnerability Assessment

**Context**: 7 vulnerabilities found (2 HIGH, 5 MEDIUM)

**Security-First Decision**:
- **Action**: Immediately prioritized Go upgrade to fix HIGH severity vulns
- **Timeline**: Addressed in same iteration as discovery
- **Result**: All 7 vulnerabilities fixed via single Go 1.23.1 → 1.24.9 upgrade

**Alternative (deprioritization)**:
- **Risk**: HIGH severity vulnerabilities remain exploitable for weeks/months
- **Impact**: RCE potential (GO-2025-3956), information disclosure (GO-2025-3751)

**Validation**: Security-first approach eliminated all vulnerabilities in 1 iteration.

---

## Applications

### Application 1: Triage Workflow

When vulnerabilities discovered:

1. **Classify by severity** (CRITICAL > HIGH > MEDIUM > LOW)
2. **Prioritize CRITICAL and HIGH immediately**
3. **Defer MEDIUM and LOW** (batch with regular updates)

Example:
- CRITICAL: Drop everything, patch now
- HIGH: Patch within 1 week
- MEDIUM: Patch within 1 month
- LOW: Patch opportunistically

### Application 2: Update Ordering

When multiple dependency updates available:

**Security-first ordering**:
1. Security patches (vulnerabilities fixed)
2. Bug fixes (stability improvements)
3. Feature updates (new capabilities)
4. Performance optimizations

### Application 3: Resource Allocation

When team capacity limited:

**Security-first allocation**:
- Assign best engineers to security patches
- Allocate testing resources to security updates first
- Fast-track security fixes through review process

---

## Cross-Ecosystem Validation

### Go Ecosystem

**Tool**: govulncheck
**Severity**: CRITICAL, HIGH, MEDIUM, LOW
**Application**: Prioritize HIGH severity Go standard library upgrades

**Example**: Iteration 1 prioritized Go upgrade for 2 HIGH vulns

### npm Ecosystem

**Tool**: npm audit
**Severity**: CRITICAL, HIGH, MODERATE, LOW, INFO
**Application**: `npm audit fix` for CRITICAL/HIGH, manual review for others

**Transfer**: 100% applicable

### pip Ecosystem

**Tool**: pip-audit
**Severity**: HIGH, MEDIUM, LOW
**Application**: Prioritize HIGH severity package upgrades

**Transfer**: 100% applicable

### cargo Ecosystem

**Tool**: cargo-audit
**Severity**: CRITICAL, HIGH, MEDIUM, LOW, INFORMATIONAL
**Application**: Prioritize CRITICAL/HIGH crate updates

**Transfer**: 100% applicable

**Transferability**: 100% (all ecosystems support severity-based prioritization)

---

## Trade-offs

### Benefits

- **Risk minimization**: Fastest vulnerability remediation
- **Compliance**: Demonstrates security diligence
- **Incident prevention**: Fewer security-related outages

### Costs

- **Disruption**: May interrupt feature development
- **Urgency stress**: Pressure to patch quickly
- **Testing burden**: Accelerated testing timelines

### When to Deviate

**Platform-specific vulnerabilities** (e.g., ppc64le-specific timing side-channel):
- Lower priority if platform not used in deployment
- Still patch, but not urgently

**False positives**:
- Verify exploitability before treating as urgent
- Document why deprioritized (audit trail)

**Mitigation exists**:
- If WAF/firewall blocks exploit, reduce urgency slightly
- Still patch, but less time-critical

---

## Related Principles

- **Test-Before-Update**: Security patches still need testing (but faster timeline)
- **Batch-Remediation**: Batch MEDIUM/LOW severity, not HIGH/CRITICAL
- **Platform-Context**: Consider deployment context when prioritizing

---

## Validation Status

**Tested In**: Iteration 1 (Go ecosystem)
**Transferred To**: npm, pip, cargo (research validation)
**Success Rate**: 100% (all vulnerabilities fixed in 1 iteration)
**Reusability**: Universal (applies to all ecosystems)

---

**Created**: 2025-10-17 (Iteration 2)
**Last Updated**: 2025-10-17
**Version**: 1.0
**Status**: Validated
