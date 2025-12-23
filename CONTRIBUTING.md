## Contributing to RCIG

Thank you for considering contributing to **Real Time Code Intelligence Gateway (RCIG)**!

This document describes how to propose changes and what we expect from contributors to keep the project healthy and maintainable.

---

## Code of Conduct

By participating in this project, you agree to follow the guidelines described in the `CODE_OF_CONDUCT.md`.  
Please read it before opening issues or submitting pull requests.

---

## Ways to Contribute

- **Bug reports**
  - Report reproducible issues.
  - Include environment details (Go version, OS, configuration).
  - Provide logs, stack traces, or sample requests if possible.

- **Feature requests**
  - Explain the problem you are trying to solve.
  - Describe your ideal solution.
  - Mention alternatives you have considered.

- **Documentation improvements**
  - Clarify existing docs.
  - Add missing sections, examples, or diagrams.

- **Code contributions**
  - Implement new features.
  - Improve performance or reliability.
  - Add tests or refactor for readability.

---

## Development Setup

1. **Fork** the repository on GitHub.
2. **Clone** your fork:

   ```bash
   git clone https://github.com/<your-username>/rcig.git
   cd rcig
   ```

3. Ensure you have **Go 1.22+** installed.
4. Install dependencies:

   ```bash
   go mod tidy
   ```

5. Run the project locally:

   ```bash
   export RCIG_TARGET_URL=http://localhost:9000
   export RCIG_LISTEN_ADDR=:8080

   go run ./cmd/proxy
   ```

6. (Optional) Run tests:

   ```bash
   go test ./...
   ```

---

## Branching & Workflow

- Use a **feature branch** for your work:

  ```bash
  git checkout -b feature/my-awesome-change
  ```

- Keep your branch up to date with the main branch:

  ```bash
  git fetch origin
  git rebase origin/main
  ```

---

## Coding Guidelines

- **Language**: Go (1.22+).
- **Style**:
  - Follow the standard Go style (`gofmt`, `go vet`).
  - Prefer small, focused functions.
  - Keep packages cohesive and well named.
- **Error handling**:
  - Return descriptive errors.
  - Log context where appropriate.
- **Logging**:
  - Use the standard `log` package.
  - Avoid excessive or noisy logs.

---

## Testing

- Add tests for new functionality when reasonable.
- Ensure all tests pass:

  ```bash
  go test ./...
  ```

- For changes that affect behavior, consider adding integration-style tests or at least documenting manual test steps in the PR description.

---

## Commit Messages

- Use clear and descriptive commit messages.
- Example format:
  - `fix: handle nil registry in metrics handler`
  - `feat: add Slack alert sink`
  - `docs: improve README quick start section`

---

## Pull Request Guidelines

Before opening a pull request:

1. Make sure the code builds and tests pass.
2. Update or add documentation if behavior changes.
3. Link any related issues in the PR description.
4. Provide a short **summary** of the change and **how to test it**.

We aim to review contributions as soon as possible, but response times may vary.  
Thank you for your patience and for helping improve RCIG!


