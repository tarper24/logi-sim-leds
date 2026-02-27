# Contributing to logi-sim-leds

Thank you for your interest in contributing to logi-sim-leds! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive criticism
- Assume good intentions

## How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:

1. **Clear title**: Summarize the problem
2. **Description**: Detailed explanation of the issue
3. **Steps to reproduce**: How to trigger the bug
4. **Expected behavior**: What should happen
5. **Actual behavior**: What actually happens
6. **Environment**:
   - OS and version
   - Go version
   - Device model (G29/G920/G923)
   - Game and version

### Suggesting Features

Feature suggestions are welcome! Please create an issue with:

1. **Clear title**: Summarize the feature
2. **Use case**: Why is this feature needed?
3. **Proposed solution**: How should it work?
4. **Alternatives**: Other approaches considered

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes**:
   - Follow the existing code style
   - Add tests if applicable
   - Update documentation
4. **Test your changes**:
   ```bash
   make test
   make build
   ```
5. **Commit with clear messages**:
   ```bash
   git commit -m "Add feature: description"
   ```
6. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```
7. **Open a Pull Request**

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make
- HID library (see README for platform-specific instructions)

### Building

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/logi-sim-leds.git
cd logi-sim-leds

# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Run the application
make run
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` to format code
- Write clear, descriptive variable names
- Add comments for complex logic
- Keep functions focused and small
- Use interfaces for modularity

## Adding New Features

### Adding a New Game

1. Create package: `pkg/games/yourgame/`
2. Implement `core.GameInterface`
3. Parse game's telemetry protocol
4. Add to manager in `pkg/manager/manager.go`
5. Update README.md with setup instructions
6. Add tests

### Adding a New Device

1. Create package: `pkg/devices/yourdevice/`
2. Implement `core.DeviceInterface`
3. Add device detection logic
4. Add to detector in manager
5. Update README.md with supported devices
6. Add tests

## Testing

- Write unit tests for new functionality
- Test on actual hardware when possible
- Test on different platforms if available
- Ensure existing tests still pass

## Documentation

- Update README.md for user-facing changes
- Update CHANGELOG.md with your changes
- Add comments to exported functions
- Include examples for new features

## Commit Message Guidelines

Format: `<type>: <description>`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
- `feat: add support for Thrustmaster T300`
- `fix: handle device disconnection gracefully`
- `docs: update installation instructions`

## Questions?

Feel free to open an issue for any questions about contributing!
