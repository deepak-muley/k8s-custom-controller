# Documentation

This directory contains comprehensive documentation for the Kubernetes Custom Controller project.

## Quick Links

### Testing Documentation

- **[Testing Quick Reference](./TESTING_QUICK_REFERENCE.md)** ⭐ **START HERE** - Quick decision guide for when to use each test type
- **[E2E Testing Comparison](./E2E_TESTING_COMPARISON.md)** ⭐ **NEW** - KUTTL vs Golang E2E vs Cluster API Framework
- **[Testing Strategy](./TESTING_STRATEGY.md)** - Complete testing strategy with all options explained
- **[envtest Explained](./ENVTEST_EXPLAINED.md)** - Detailed guide to envtest integration testing
- **[KUTTL FAQ](./KUTTL_FAQ.md)** - Frequently asked questions about KUTTL E2E testing
- **[KUTTL Usage](./KUTTL_USAGE.md)** - Detailed KUTTL usage guide (includes Helm chart installation)
- **[KUTTL Execution Order](./KUTTL_EXECUTION_ORDER.md)** - When test files and steps directory run
- **[KUTTL Guide](./KUTTL_GUIDE.md)** - Complete KUTTL guide
- **[KUTTL Summary](./KUTTL_SUMMARY.md)** - Quick reference for KUTTL
- **[KUTTL Answer](./KUTTL_ANSWER.md)** - Direct answers to KUTTL questions
- **[Where to Use KUTTL](./WHERE_TO_USE_KUTTL.md)** - When and where to use KUTTL

### Setup and Development

- **[Setup Guide](./SETUP.md)** - Project setup and installation instructions
- **[README_ENVTEST.md](./README_ENVTEST.md)** - Additional envtest information

### Documentation Structure

- **[Documentation Structure](./DOCUMENTATION_STRUCTURE.md)** - How documentation is organized

## Documentation Structure

```
docs/
├── README.md                      # This file - documentation index
├── TESTING_QUICK_REFERENCE.md    # ⭐ Quick decision guide
├── TESTING_STRATEGY.md           # Complete testing strategy
├── ENVTEST_EXPLAINED.md          # envtest detailed guide
├── KUTTL_FAQ.md                  # KUTTL FAQ
├── KUTTL_USAGE.md                # KUTTL usage guide (includes Helm chart installation)
├── KUTTL_EXECUTION_ORDER.md      # KUTTL test execution order (when steps/ runs)
├── KUTTL_GUIDE.md                # Complete KUTTL guide
├── KUTTL_SUMMARY.md              # KUTTL quick reference
├── KUTTL_ANSWER.md               # Direct answers to KUTTL questions
├── WHERE_TO_USE_KUTTL.md         # When to use KUTTL
├── E2E_TESTING_COMPARISON.md     # ⭐ KUTTL vs Golang E2E vs Cluster API Framework
├── SETUP.md                      # Setup instructions
├── README_ENVTEST.md             # Additional envtest info
└── DOCUMENTATION_STRUCTURE.md    # Documentation organization guide
```

## Recommended Reading Order

### For Quick Answers

1. **[Testing Quick Reference](./TESTING_QUICK_REFERENCE.md)** - Quick decision guide
2. **[Testing Strategy](./TESTING_STRATEGY.md)** - Complete overview

### For Understanding Testing Approaches

1. **[Testing Strategy](./TESTING_STRATEGY.md)** - All options explained
2. **[E2E Testing Comparison](./E2E_TESTING_COMPARISON.md)** ⭐ **NEW** - KUTTL vs Golang E2E vs Cluster API Framework
3. **[envtest Explained](./ENVTEST_EXPLAINED.md)** - Why envtest was chosen
4. **[KUTTL FAQ](./KUTTL_FAQ.md)** - When to use KUTTL

### For Using KUTTL

1. **[E2E Testing Comparison](./E2E_TESTING_COMPARISON.md)** ⭐ **START HERE** - KUTTL vs other approaches
2. **[KUTTL FAQ](./KUTTL_FAQ.md)** - Quick answers
3. **[KUTTL Usage](./KUTTL_USAGE.md)** - Detailed usage (includes Helm chart installation example)
4. **[KUTTL Execution Order](./KUTTL_EXECUTION_ORDER.md)** - When test files and steps/ run
5. **[KUTTL Guide](./KUTTL_GUIDE.md)** - Complete guide

### For Setup

1. **[Setup Guide](./SETUP.md)** - Project setup
2. Main [README.md](../README.md) - Project overview

## Testing Quick Reference

### When to Use Each Test Type

| Need | Solution | Tool | Speed |
|------|----------|------|-------|
| Test Go code logic | Unit Tests | Ginkgo/Gomega | ⚡ Fast (< 1s) |
| Test controller with API | Integration Tests | envtest | ⚡ Fast (1-3s) |
| Test Helm chart | E2E Tests | KUTTL | ⚠️ Slower (10-30s) |
| Validate security | Security Scan | kubesec | ⚡ Fast (< 5s) |
| Check API versions | API Validation | Pluto | ⚡ Fast (< 5s) |

**For detailed explanations, see [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md).**

## Project Testing Status

### ✅ Implemented

- ✅ **Unit Tests (Ginkgo/Gomega)** - Counter, config, parsing
- ✅ **Integration Tests (envtest)** - Controller, informers, discovery
- ✅ **Security Scanning (kubesec)** - Helm template validation
- ✅ **API Version Validation (Pluto)** - Deprecated API detection

### ⬜ Optional

- ⬜ **E2E Tests (KUTTL)** - Helm chart E2E testing (structure created)

## Contributing

When adding new documentation:

1. Place documentation files in the `docs/` directory
2. Update this README.md with links to new documentation
3. Follow the naming convention: `UPPERCASE_WITH_UNDERSCORES.md`
4. Include cross-references to related documentation
5. Keep quick reference guides concise and actionable

## References

- Main [README.md](../README.md) - Project overview
- [Makefile](../Makefile) - Available commands
- [tests/](../tests/) - E2E test structure
