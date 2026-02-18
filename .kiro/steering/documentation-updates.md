---
inclusion: auto
description: Guidelines for keeping markdown documentation files in sync with code changes
---

# Documentation Update Guidelines

When making changes to this project, always review and update the relevant markdown documentation files to keep them in sync with the codebase.

## Documentation Files to Review

- **README.md**: Main project documentation including features, configuration, local development, Docker usage, and testing instructions
- **Dockerhub.md**: Docker Hub repository description with quick start guide, configuration, health checks, and available tags

**Note**: If new .md files are created in the project, update this steering document to include them in the list above.

## When to Update Documentation

- Adding new features or endpoints → Update both README.md and Dockerhub.md
- Changing configuration or environment variables → Update both files
- Modifying Docker setup or deployment → Update both files
- Adding new commands or workflows → Update README.md
- Changing health check behavior → Update both files
- Adding new tags or build processes → Update Dockerhub.md
- Creating new .md files → Update this steering document

## Documentation Consistency

Ensure that:
- Configuration examples match actual code requirements
- Docker commands reflect current Dockerfile setup
- Environment variables are documented in both files
- Quick start instructions work with current implementation
- Tag descriptions match GitHub Actions workflow behavior
