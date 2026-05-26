# Security Policy

## Supported Versions

Only the latest release of termimage is supported with security updates.

## Reporting a Vulnerability

If you discover a security vulnerability in termimage, please report it responsibly. **Do not open a public issue.**

Email us at [us@floatpane.com](mailto:us@floatpane.com) with:

- A description of the vulnerability
- Steps to reproduce the issue
- The potential impact
- Any suggested fixes (optional)

We will acknowledge your report within 48 hours and aim to provide a fix or mitigation plan within 7 days, depending on severity.

## Scope

This policy covers the termimage codebase and its official releases — including the bundled C decoder (`decode/stb_image.h`), the sandbox/worker boundary, and all rendering protocols. Issues in the bundled decoder are in-scope even though upstream is third-party, since untrusted bytes pass through it.

Of particular interest:

- Sandbox escapes from the worker subprocess on Linux (Landlock + seccomp bypass).
- Memory-safety issues in the C decoder reachable from supported image formats.
- Terminal escape-sequence injection through the rendering protocols.

Third-party dependencies outside the bundled decoder are outside our direct control, but we will work to address reported issues in them as quickly as possible.

## Disclosure

We ask that you give us reasonable time to address the issue before disclosing it publicly. We are committed to crediting reporters in release notes (unless you prefer to remain anonymous).
