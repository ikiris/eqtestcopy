# Release Process

This project uses [release-please](https://github.com/googleapis/release-please) for automated releases.

## How it works

1. **Conventional Commits**: All commits should follow the [Conventional Commits](https://www.conventionalcommits.org/) specification
2. **Automatic PRs**: Release-please automatically creates PRs with version bumps and changelog updates
3. **Automatic Releases**: When PRs are merged to main, releases are automatically created

## Commit Message Format

Use the following format for your commit messages:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

### Examples

```
feat: add user authentication system
fix: resolve database connection timeout issue
docs: update API documentation
chore: update dependencies
```

## Release Types

- **Patch** (1.0.0 → 1.0.1): Bug fixes
- **Minor** (1.0.0 → 1.1.0): New features (backward compatible)
- **Major** (1.0.0 → 2.0.0): Breaking changes

## Manual Release

You can trigger a release manually by:

1. Going to the GitHub Actions tab
2. Selecting the "Release Please" workflow
3. Clicking "Run workflow"

## Release Assets

Each release automatically includes:

- Binary builds for Linux (amd64)
- Binary builds for Windows (amd64)
- Binary builds for macOS (amd64)
- Frontend distribution files
- Source code archives

## Configuration

The release process is configured in:
- `.github/workflows/release-please.yml` - GitHub Actions workflow
- `.github/release-please-config.json` - Release-please configuration
- `.github/release-please-manifest.json` - Package manifest
