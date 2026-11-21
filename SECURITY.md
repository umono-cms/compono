# Security Policy for Compono

Compono is developed by a single developer, Yunus Emre Bulut. Compono turns content written by people into HTML, so I take its security seriously.

## What Counts as a Vulnerability

Compono is a language, its parser and its HTML renderer. A vulnerability is a way for Compono source to do something it must not do. For example:

* **Escaping bypassed:** source that puts a script or raw HTML into the output. Compono's syntax never produces raw HTML: every syntax it does not cover stays escaped plain text.
* **Disallowed URL schemes:** a URL that reaches the output with a scheme the [URL scheme allowlist](README.md#url-scheme-allowlist) rejects, such as `javascript:`.
* **Crashes and resource exhaustion:** input that makes Compono panic or consume excessive CPU time or memory.

What an application adds to the output through Compono's API, such as the attributes returned by `WithAttributeHook`, is the application's responsibility.

## Security Fixes and the HTML Output

While Umono is in 0.x, fixes that change Compono's HTML output wait for Umono v1 (see [CONTRIBUTING.md](CONTRIBUTING.md#html-output-bugs)). Security fixes do not wait: they are released as a patch or a minor version even when they change the output.

## Supported Versions

Compono is currently in active development. **Only the latest minor release receives security updates.** Once a new minor version is released, the previous minor versions no longer receive security patches. Always upgrade to the latest version.

| Version | Supported |
| :--- | :--- |
| **Latest Minor Release (Current)** | ✅ Supported |
| **Older Versions** | ❌ Not Supported (Please upgrade) |

## Reporting a Vulnerability

If you discover a security vulnerability in Compono, I appreciate your help in disclosing it to me in a responsible manner.

**🚨 Please DO NOT report security vulnerabilities through public GitHub issues or discussions.**

Instead, please report them privately via email to: **yunusemrebulut123@gmail.com**

Please include the following information in your report to help me understand and resolve the issue quickly:
* **Description:** A brief description of the vulnerability.
* **Input:** The Compono source, and the options passed to `Convert` if any, that trigger it.
* **Impact:** The potential impact of the vulnerability (e.g., what an attacker could achieve).
* **Suggested Mitigation:** Any potential solutions or ideas on how to fix it (optional but highly appreciated).

### Response Timeline
* I will acknowledge receipt of your vulnerability report within **48 hours**.
* I will investigate the issue and determine if it is a valid vulnerability.
* I strive to keep you informed about my progress and will notify you when a patch is released.
