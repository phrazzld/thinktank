# Project-Specific Claude Commands

This directory contains Claude commands tailored for the thinktank project.

## Backlog Management Commands

- **backlog-align**: Analyze project against leyline philosophy and add alignment tasks
- **backlog-groom**: Organize and prioritize the local BACKLOG.md file
- **backlog-gordian**: Identify and solve complex project issues
- **backlog-ideate**: Generate new ideas and tasks for the project

## Usage

All commands work with the local `BACKLOG.md` file in the project root and are project-aware, using local leyline documents from `./docs/leyline/`.

Example:
```
/backlog-groom
```

## Task Format

Tasks in BACKLOG.md follow this format:
```
- [ ] [HIGH/MED/LOW] [TYPE] Description
```

Types: ALIGN, REFACTOR, FEATURE, BUG, DOCS, TEST, CHORE
