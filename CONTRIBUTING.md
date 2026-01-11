# Contributing to Woodpecker Config Provider

Thank you for your interest in contributing to Woodpecker Config Provider! 🎉

This document provides guidelines and instructions for contributing to this project.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Code Contributions](#code-contributions)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)

## 📜 Code of Conduct

This project follows a code of conduct that all contributors are expected to adhere to:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards other community members

## 🤝 How Can I Contribute?

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

**When reporting a bug, include:**

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs **actual behavior**
- **Environment details:**
  - Go version (`go version`)
  - Operating system
  - Docker version (if applicable)
  - Git platform (Gitea/GitHub/GitLab) and version
- **Logs** with `PLUGIN_DEBUG=true` enabled
- **Configuration** (sanitize sensitive data like tokens)

**Example bug report:**

```markdown
### Bug Description
Config provider returns 404 when accessing GitLab repository.

### Steps to Reproduce
1. Set `SERVERTYPE=gitlab`
2. Configure `SERVER_URL=https://gitlab.com`
3. Trigger pipeline build
4. Check logs

### Expected Behavior
Should fetch files from dronefiles repository.

### Actual Behavior
Returns 404 error.

### Environment
- Go: 1.24.0
- OS: macOS ARM64
- GitLab: gitlab.com (SaaS)
- Config Provider: v2.0.0

### Logs
```
[DEBUG] Response status: 404
```

### Configuration
```yaml
SERVERTYPE=gitlab
SERVER_URL=https://gitlab.com
TOKEN=glpat-***
```
```

### Suggesting Enhancements

We welcome enhancement suggestions! Please include:

- **Clear use case** - Why is this enhancement needed?
- **Proposed solution** - How would you implement it?
- **Alternatives considered** - What other approaches did you think about?
- **Additional context** - Screenshots, examples, references

**Example enhancement request:**

```markdown
### Enhancement Description
Add support for Bitbucket Server

### Use Case
Many organizations use Bitbucket Server for source control and would benefit
from centralized pipeline configuration.

### Proposed Solution
Implement `fetchFilesFromBitbucket()` using the Bitbucket REST API similar to
existing GitHub/GitLab implementations.

### Alternatives Considered
- Using Bitbucket Pipes (too limited)
- Repository webhooks (requires per-repo setup)

### Additional Context
- Bitbucket REST API: https://developer.atlassian.com/server/bitbucket/rest/
- Similar to GitLab implementation pattern
```

### Code Contributions

We love code contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Write or update tests**
5. **Update documentation**
6. **Commit your changes** (follow commit guidelines)
7. **Push to your fork** (`git push origin feature/amazing-feature`)
8. **Open a Pull Request**

## 🛠️ Development Setup

### Prerequisites

- Go 1.24 or higher
- Git
- Docker (optional, for testing)
- Access to Gitea/GitHub/GitLab for testing

### Setup Steps

```bash
# 1. Fork and clone the repository
git clone https://github.com/YahuiWong/woodpecker-config-provider.git
cd woodpecker-config-provider

# 2. Install dependencies
go mod download

# 3. Run tests to verify setup
go test -v ./...

# 4. Build the binary
go build -o woodpecker-config-provider .

# 5. Run locally for testing
export PLUGIN_DEBUG=true
export SERVERTYPE=gitea
export SERVER_URL=https://your-git-server.com
export TOKEN=your_test_token
./woodpecker-config-provider
```

### Running with Docker

```bash
# Build Docker image
docker build -t woodpecker-config-provider:dev .

# Run with Docker
docker run -p 8000:8000 \
  -e PLUGIN_DEBUG=true \
  -e SERVERTYPE=gitea \
  -e SERVER_URL=https://git.example.com \
  -e TOKEN=your_token \
  woodpecker-config-provider:dev
```

## 💻 Coding Standards

### Go Code Style

Follow standard Go conventions:

- **Use `gofmt`** - Format all code with `gofmt -s`
- **Use `golint`** - Check with `golint ./...`
- **Use `go vet`** - Validate with `go vet ./...`

```bash
# Format code
gofmt -s -w .

# Lint code
golint ./...

# Vet code
go vet ./...
```

### Code Organization

- **Keep functions small** - Single responsibility principle
- **Add comments** - Especially for exported functions and complex logic
- **Handle errors** - Never ignore errors, log them appropriately
- **Use meaningful names** - Clear variable and function names

**Good example:**

```go
// fetchFilesFromGitea retrieves all YAML configuration files from the specified
// Gitea repository path using the official Gitea SDK.
func fetchFilesFromGitea(namespace, repo, branch, path string) ([]GiteaFile, error) {
    debugLog("fetchFilesFromGitea - namespace: %s, repo: %s, branch: %s, path: %s",
        namespace, repo, branch, path)

    client, err := gitea.NewClient(GiteaURL, gitea.SetToken(GiteaToken))
    if err != nil {
        debugLog("ERROR: Failed to create Gitea client: %v", err)
        return nil, fmt.Errorf("create gitea client: %w", err)
    }

    // ... rest of implementation
}
```

### Error Handling

- **Wrap errors** with context using `fmt.Errorf("context: %w", err)`
- **Log errors** before returning them
- **Use custom error types** for specific error cases when needed

```go
// Good error handling
content, err := client.GetFile(namespace, repo, branch, filePath)
if err != nil {
    debugLog("ERROR: Failed to fetch file %s: %v", filePath, err)
    return nil, fmt.Errorf("fetch file %s: %w", filePath, err)
}
```

### Testing Standards

- **Write tests** for all new functions
- **Test edge cases** - Empty inputs, nil values, errors
- **Use table-driven tests** when appropriate
- **Aim for >80% coverage**

```go
func TestFetchFilesFromGitea(t *testing.T) {
    tests := []struct {
        name      string
        namespace string
        repo      string
        branch    string
        path      string
        wantErr   bool
    }{
        {
            name:      "valid path",
            namespace: "admin",
            repo:      "dronefiles",
            branch:    "main",
            path:      "myproject/main",
            wantErr:   false,
        },
        {
            name:      "invalid namespace",
            namespace: "",
            repo:      "dronefiles",
            branch:    "main",
            path:      "myproject/main",
            wantErr:   true,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := fetchFilesFromGitea(tt.namespace, tt.repo, tt.branch, tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("fetchFilesFromGitea() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 📝 Commit Message Guidelines

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semi-colons, etc)
- `refactor`: Code refactoring without changing functionality
- `test`: Adding or updating tests
- `chore`: Maintenance tasks (dependencies, build config, etc)
- `perf`: Performance improvements

### Examples

**Feature:**
```
feat(gitlab): add GitLab API support

Implement fetchFilesFromGitLab() using official GitLab SDK.
Supports both GitLab.com and self-hosted instances.

Closes #42
```

**Bug fix:**
```
fix(gitea): correct base64 decoding issue

Gitea SDK GetFile() returns raw bytes, not base64 encoded.
Remove unnecessary decoding step.

Fixes #58
```

**Documentation:**
```
docs(readme): add troubleshooting section

Add common issues and solutions:
- 404 errors
- Authentication failures
- YAML parsing errors
```

**Chore:**
```
chore(deps): update go-github to v58

Update github.com/google/go-github from v57 to v58.
Includes bug fixes and new API features.
```

### Commit Message Rules

- Use imperative mood ("Add feature" not "Added feature")
- Keep subject line under 72 characters
- Capitalize first letter of subject
- Don't end subject with a period
- Separate subject from body with blank line
- Wrap body at 72 characters
- Reference issues and PRs in footer

## 🔄 Pull Request Process

### Before Submitting

- [ ] **Code compiles** - `go build` succeeds
- [ ] **Tests pass** - `go test -v ./...` passes
- [ ] **Code formatted** - Run `gofmt -s -w .`
- [ ] **No lint errors** - Run `golint ./...`
- [ ] **Documentation updated** - README, comments, CHANGELOG
- [ ] **Commits are clean** - Follow commit message guidelines

### PR Checklist

Create a PR with:

**1. Clear title** following commit message format:
```
feat(github): add GitHub Enterprise support
```

**2. Description** including:
```markdown
## What does this PR do?
Adds support for GitHub Enterprise by allowing custom base URLs.

## Why is this needed?
Many organizations use GitHub Enterprise and need config provider support.

## Changes
- Add `WithEnterpriseURLs()` configuration
- Update README with GHE examples
- Add tests for custom URLs

## Testing
- [ ] Tested with GitHub.com
- [ ] Tested with GitHub Enterprise
- [ ] All tests pass
- [ ] Documentation updated

## Related Issues
Closes #45
```

**3. Small, focused changes** - Keep PRs small and focused on one feature/fix

**4. Request review** from maintainers

### Review Process

1. **Automated checks** must pass (tests, linting)
2. **At least one approval** from maintainers required
3. **Address feedback** - Respond to review comments
4. **Squash commits** if requested
5. **Maintainer will merge** once approved

### After Merge

- Your contribution will be in the next release
- You'll be added to the contributors list
- Thank you! 🎉

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestTemplateRendering

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Writing Tests

**Unit tests** for individual functions:

```go
// main_test.go
func TestRenderTemplate(t *testing.T) {
    data := TemplateData{
        Repo: RepoInfo{
            Name:  "myproject",
            Owner: "admin",
        },
        Pipeline: PipelineInfo{
            Branch: "main",
        },
    }

    result, err := renderTemplate("{{ .Repo.Name }}", data)
    if err != nil {
        t.Fatalf("renderTemplate() error = %v", err)
    }

    if result != "myproject" {
        t.Errorf("renderTemplate() = %v, want %v", result, "myproject")
    }
}
```

**Integration tests** for API interactions:

```go
// github_test.go
func TestFetchFilesFromGitHub(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Test with real GitHub API
    files, err := fetchFilesFromGitHub("octocat", "Hello-World", "master", "README")
    if err != nil {
        t.Fatalf("fetchFilesFromGitHub() error = %v", err)
    }

    if len(files) == 0 {
        t.Error("Expected at least one file")
    }
}
```

### Test Coverage Goals

- **Overall coverage**: >80%
- **Critical paths**: 100% (SDK integration, template rendering)
- **Error handling**: All error paths tested

## 📚 Documentation

### Code Documentation

- **Add godoc comments** for all exported functions, types, and constants
- **Include examples** in comments when helpful
- **Document parameters and return values**

```go
// renderTemplate processes a Go template string with the provided data.
// It returns the rendered result or an error if template parsing or execution fails.
//
// Template variables:
//   - .Repo.Name: Repository name
//   - .Repo.Owner: Repository owner
//   - .Pipeline.Branch: Current branch
//
// Example:
//   result, err := renderTemplate("{{ .Repo.Name }}/{{ .Pipeline.Branch }}", data)
//   // result: "myproject/main"
func renderTemplate(tmplStr string, data TemplateData) (string, error) {
    // Implementation...
}
```

### README Updates

When adding features:

1. Update **Features** section
2. Add **Configuration example**
3. Include **Usage example**
4. Update **Environment variables** table if needed
5. Add to **Changelog**

### CHANGELOG

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [Unreleased]

### Added
- GitHub Enterprise support (#45)
- New environment variable `GITHUB_ENTERPRISE_URL`

### Fixed
- Base64 decoding issue with Gitea SDK (#58)

### Changed
- Updated go-github to v58

## [2.0.0] - 2026-01-11

### Added
- Multi-platform support (Gitea, GitHub, GitLab)
- Official SDK integration
...
```

## 🎯 Areas Looking for Contributions

We especially welcome contributions in these areas:

### High Priority

- [ ] **Additional platform support** (Bitbucket, Azure DevOps)
- [ ] **Enhanced error messages** with actionable suggestions
- [ ] **Performance optimizations** (caching, connection pooling)
- [ ] **Metrics and monitoring** (Prometheus metrics)

### Medium Priority

- [ ] **GitHub Actions** CI/CD workflow
- [ ] **Integration tests** for all platforms
- [ ] **Configuration validation** (validate templates before use)
- [ ] **Multi-language README** (Chinese, Japanese, etc.)

### Good First Issues

Looking for your first contribution? Check issues labeled:
- `good first issue` - Simple, well-defined tasks
- `help wanted` - We'd love help on these
- `documentation` - Documentation improvements

## 💡 Tips for Success

1. **Start small** - Fix a typo, improve error message, add test
2. **Ask questions** - Open an issue to discuss before big changes
3. **Be patient** - Reviews take time, we appreciate your contribution
4. **Learn from feedback** - Reviews help you grow as a developer
5. **Have fun!** - Enjoy contributing to open source

## 📞 Getting Help

- **Questions?** Open a [Discussion](https://github.com/YahuiWong/woodpecker-config-provider/discussions)
- **Bug?** Open an [Issue](https://github.com/YahuiWong/woodpecker-config-provider/issues)
- **Chat?** Check if there's a Discord/Slack (if applicable)

## 🙏 Thank You

Every contribution, no matter how small, is valuable and appreciated. Thank you for making Woodpecker Config Provider better!

---

**Happy Contributing!** 🚀

*Last updated: 2026-01-11*
