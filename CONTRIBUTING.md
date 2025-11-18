# Contributing to SessionX

First off, thank you for considering contributing to SessionX! It's people like you that make SessionX such a great tool.

## Code of Conduct

This project and everyone participating in it is governed by respect and professionalism. By participating, you are expected to uphold this standard.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible.

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- **Clear title and description**
- **Use case**: Why is this enhancement useful?
- **Proposed solution**: How should it work?
- **Alternatives considered**: What other approaches did you think about?

### Pull Requests

1. **Fork the repository** and create your branch from `main`:
   ```bash
   git checkout -b feature/my-new-feature main
   ```

2. **Make your changes**:
   - Write clear, concise commit messages
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**:
   ```bash
   go test ./...
   go test -cover ./...
   ```

4. **Commit your changes**:
   ```bash
   git commit -m "feat: add amazing feature"
   ```

5. **Push and open a Pull Request**

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Git
- Redis (for testing Redis store)

### Clone and Test

```bash
git clone https://github.com/abmcmanu/sessionx.git
cd sessionx
go mod download
go test ./...
```

## Coding Style

Follow [Effective Go](https://golang.org/doc/effective_go) guidelines and use `gofmt` to format your code.

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: bug fix
docs: documentation
test: add tests
refactor: code refactoring
```

## Testing

- Write tests for all new functionality
- Aim for >80% code coverage
- Test both success and error cases

## Questions?

Feel free to open an issue with your question.

Thank you for making SessionX better! <�