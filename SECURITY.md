# Security Policy

## Supported Versions

This project is currently in active development. We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue in meta-cc, please report it responsibly.

### How to Report

Please report security vulnerabilities by:

1. **Email**: Create a GitHub issue with the label "security" (for non-critical issues)
2. **Private Disclosure**: For critical vulnerabilities, please use GitHub's private vulnerability reporting feature

### What to Include

When reporting a vulnerability, please include:

- Description of the vulnerability
- Steps to reproduce the issue
- Affected versions
- Potential impact
- Suggested fix (if available)

### Response Timeline

- **Initial Response**: Within 48 hours of report
- **Status Update**: Within 7 days with assessment and planned fix timeline
- **Resolution**: Timeframe depends on severity and complexity

### Security Update Process

1. Vulnerability is confirmed and assessed
2. Fix is developed and tested
3. Security advisory is published
4. Patch is released
5. Users are notified through GitHub releases and CHANGELOG

## Security Best Practices

When using meta-cc:

- Always use the latest version
- Review session data access permissions when configuring MCP
- Be cautious when sharing session history outputs (may contain sensitive project information)
- Validate file paths when using CLI commands to prevent path traversal

## Scope

This security policy applies to:

- meta-cc CLI tool
- meta-cc MCP server
- Associated documentation and examples

Thank you for helping keep meta-cc and its users secure.
