# Security Policy

## Supported Versions

Mouse Mover is a small utility and security fixes are applied to the latest
release only.

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| older   | :x:                |

## Reporting a Vulnerability

Please report security vulnerabilities **privately** so they can be addressed
before public disclosure.

- **Preferred:** use GitHub's [private vulnerability reporting][gh-report]
  (Security tab → "Report a vulnerability").
- **Alternatively:** email **alvinjaison@outlook.com** with the details.

Please include:

- A description of the vulnerability and its impact
- Steps to reproduce (a proof of concept if possible)
- The affected version or commit

You can expect an acknowledgement within a few days. Once the issue is
confirmed and fixed, we will publish a release and credit you if you wish.

## Scope and threat model

Mouse Mover runs entirely as a standard, unprivileged user. It does not run as a
service, request elevation, open network connections, collect telemetry, or
write outside the current user's application-data directory. It only synthesizes
the configured mouse movement — never keyboard input.

Because it is unprivileged by design, it cannot (and does not attempt to) bypass
operating-system or enterprise security controls such as UIPI, the lock screen,
Secure Desktop, or group policy. Reports about these platform limitations are not
considered vulnerabilities.

[gh-report]: https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability
