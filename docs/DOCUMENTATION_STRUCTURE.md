# Documentation Structure

## Overview

All documentation for this project is organized in the `docs/` folder, following open-source best practices. Only `README.md` remains in the root directory as the main project entry point.

## Documentation Organization

### ✅ Root Directory

- `README.md` - Main project README (stays in root)

### ✅ Documentation Folder (`docs/`)

All project documentation is in `docs/` directory:

```
docs/
├── README.md                      # Documentation index
│
├── TESTING_QUICK_REFERENCE.md    # ⭐ Quick decision guide
├── TESTING_STRATEGY.md           # Complete testing strategy
├── E2E_TESTING_COMPARISON.md     # ⭐ KUTTL vs Golang E2E vs Cluster API Framework
├── KUTTL_EXECUTION_ORDER.md      # When test files and steps/ run
│
├── ENVTEST_EXPLAINED.md          # envtest detailed guide
├── README_ENVTEST.md             # Additional envtest info
│
├── KUTTL_FAQ.md                  # KUTTL frequently asked questions
├── KUTTL_USAGE.md                # KUTTL usage guide (includes Helm chart installation)
├── KUTTL_EXECUTION_ORDER.md      # KUTTL test execution order
├── KUTTL_GUIDE.md                # Complete KUTTL guide
├── KUTTL_SUMMARY.md              # KUTTL quick reference
├── KUTTL_ANSWER.md               # Direct answers to KUTTL questions
├── WHERE_TO_USE_KUTTL.md         # When and where to use KUTTL
│
└── SETUP.md                      # Setup instructions
```

**Total: 14 documentation files** (all in `docs/`)

## Documentation Standards

### 1. All Documentation in `docs/` Folder ✅

- ✅ All `.md` files (except `README.md`) are in `docs/`
- ✅ Standard open-source practice
- ✅ Easy to find and navigate
- ✅ Clear separation from code

### 2. Documentation Index

- `docs/README.md` serves as the documentation index
- Provides quick links to all documentation
- Organized by topic (Testing, KUTTL, Setup)
- Includes recommended reading order

### 3. Cross-References

All documentation files cross-reference each other using relative paths:
- `./FILENAME.md` for files in same directory
- `../docs/FILENAME.md` for references from root/tests

### 4. Naming Convention

- `UPPERCASE_WITH_UNDERSCORES.md` for documentation files
- Descriptive names that indicate content
- Consistent naming across all docs

## Accessing Documentation

### From Root Directory

```markdown
- [Testing Quick Reference](./docs/TESTING_QUICK_REFERENCE.md)
- [E2E Testing Comparison](./docs/E2E_TESTING_COMPARISON.md)
- [Testing Strategy](./docs/TESTING_STRATEGY.md)
```

### From `docs/` Directory

```markdown
- [Testing Quick Reference](./TESTING_QUICK_REFERENCE.md)
- [E2E Testing Comparison](./E2E_TESTING_COMPARISON.md)
- [Testing Strategy](./TESTING_STRATEGY.md)
```

### From Other Directories

```markdown
- [Testing Quick Reference](../docs/TESTING_QUICK_REFERENCE.md)
- [E2E Testing Comparison](../docs/E2E_TESTING_COMPARISON.md)
```

## Quick Links

### Start Here

1. **[Testing Quick Reference](./TESTING_QUICK_REFERENCE.md)** ⭐ - Quick decision guide
2. **[E2E Testing Comparison](./E2E_TESTING_COMPARISON.md)** ⭐ - KUTTL vs other approaches
3. **[Testing Strategy](./TESTING_STRATEGY.md)** - Complete overview

### KUTTL Documentation

1. **[KUTTL FAQ](./KUTTL_FAQ.md)** - Quick answers
2. **[KUTTL Execution Order](./KUTTL_EXECUTION_ORDER.md)** - When steps/ runs
3. **[KUTTL Usage](./KUTTL_USAGE.md)** - Detailed guide (Helm chart installation)
4. **[KUTTL Guide](./KUTTL_GUIDE.md)** - Complete guide

### Other Documentation

- **[envtest Explained](./ENVTEST_EXPLAINED.md)** - Integration testing
- **[Setup Guide](./SETUP.md)** - Project setup

## Adding New Documentation

When adding new documentation:

1. ✅ **Place in `docs/` folder** - Always use `docs/` directory
2. ✅ **Follow naming convention** - `UPPERCASE_WITH_UNDERSCORES.md`
3. ✅ **Update `docs/README.md`** - Add to documentation index
4. ✅ **Add cross-references** - Link from related documentation
5. ✅ **Update root `README.md`** - Add link if needed

### Example: Adding New Documentation

```bash
# 1. Create documentation file in docs/
touch docs/NEW_FEATURE_GUIDE.md

# 2. Update docs/README.md
# Add entry to documentation index

# 3. Update related documentation
# Add cross-references where appropriate

# 4. Update root README.md if needed
# Add link in Documentation section
```

## Verification

To verify all documentation is in `docs/`:

```bash
# Check root directory (should only have README.md)
ls -1 *.md | grep -v "^README.md$"

# List all docs in docs/ folder
ls -1 docs/*.md

# Count documentation files
find docs/ -name "*.md" -type f | wc -l
```

## References

- Main [README.md](../README.md) - Project overview
- [docs/README.md](./README.md) - Documentation index
- [Makefile](../Makefile) - Available commands
- [tests/](../tests/) - E2E test structure
