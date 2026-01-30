---
layout: default
title: Guides
nav_order: 4
has_children: true
---

# Guides
{: .no_toc }

Practical guides for common SAML debugging scenarios.
{: .fs-6 .fw-300 }

---

## Quick Start: Debug an SSO Issue

The fastest way to debug SSO is with a HAR file:

```bash
# 1. Ask user to capture HAR file from browser DevTools
# 2. Inspect the complete SSO flow
samlurai inspect -f capture.har

# 3. If encrypted, add your private key
samlurai inspect -f capture.har -k private.pem
```

## Available Guides

- [Debugging SSO]({% link guides/debugging-sso.md %}) - Step-by-step guide to debugging SSO issues, including how to capture and analyze HAR files
- [Working with Encrypted Assertions]({% link guides/encrypted-assertions.md %}) - Handling encrypted SAML data from HAR files and other sources
- [SAML Basics]({% link guides/saml-basics.md %}) - Understanding SAML concepts
