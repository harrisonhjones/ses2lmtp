---
inclusion: auto
description: Guidelines for keeping markdown documentation files in sync with code changes
---

# Documentation Update Guidelines

When making changes to this project, always review and update the relevant markdown documentation files to keep them in sync with the codebase.

## Documentation Files to Review

- **README.md**: Main project documentation - "what is this and how do I get started"
- **DEVELOPMENT.md**: Development guide covering local development, testing, building, and publishing
- **DOCKERHUB.md**: Docker Hub repository description with quick start guide, configuration, health checks, and available tags

**Note**: If new .md files are created in the project, update this steering document to include them in the list above.

## When to Update Documentation

- Adding new features or endpoints → Update README.md, DEVELOPMENT.md, and DOCKERHUB.md as appropriate
- Changing configuration or environment variables → Update README.md and DOCKERHUB.md
- Modifying Docker setup or deployment → Update README.md, DEVELOPMENT.md, and DOCKERHUB.md
- Adding new commands or workflows → Update DEVELOPMENT.md
- Changing health check behavior → Update README.md and DOCKERHUB.md
- Adding new tags or build processes → Update DEVELOPMENT.md and DOCKERHUB.md
- Changing testing procedures → Update DEVELOPMENT.md
- Creating new .md files → Update this steering document

## Documentation Consistency

Ensure that:
- README.md focuses on "what is this and how do I get started"
- DEVELOPMENT.md covers development workflows, testing, and publishing
- Configuration examples match actual code requirements
- Docker commands reflect current Dockerfile setup
- Environment variables are documented in README.md and DOCKERHUB.md
- Quick start instructions work with current implementation
- Tag descriptions match GitHub Actions workflow behavior
