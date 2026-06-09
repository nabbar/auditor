## Go Code Auditor Application

## Overview

This document provides a detailed technical overview of the Go code auditing application that leverages LLM-based analysis to detect potential bugs in Go source code. The application uses Ollama for language model interactions, SQLite for data persistence, and implements a tree-based analysis approach for function dependency resolution.

---

## Table of Contents

1. [Usage & Execution Examples](#usage--execution-examples)
2. [Command-Line Flags and Options](#command-line-flags-and-options)
3. [Contextual Usage Recommendations & Operational Security](#contextual-usage-recommendations--operational-security)
4. [Architecture Components](#architecture-components)
5. [Application Flow](#application-flow)
6. [Key Features and Algorithms](#key-features-and-algorithms)
7. [Database Schema and Data Models](#database-schema-and-data-models)
8. [AI Transparency and License](#ai-transparency-and-license)

---

## Usage & Execution Examples

Since this tool is built for **on-premise, air-gapped corporate environments**, it must be executed locally on a machine with a running Ollama instance and access to a local clone of the target repositories.

### 1. Standard Interactive Execution

Run the audit with the default configuration (4 parallel workers, 10-level max depth) against your local repository. This will display a real-time progress bar:

```bash
./auditor --model "qwen3-code-expert" --db "my_project_vault.db" --report "vulnerabilities_report.md"
```

### 2. High-Performance Overnight Execution (Silent Mode)

For large repositories requiring a deep multi-hour audit (e.g., 12-hour comprehensive runs), redirect logs to a dedicated file. This automatically hides the progress bar and switches the application into a secure background processing mode:

```bash
./auditor --workers 8 --max-depth 12 --log-file "audit_execution.log" --report "deep_security_report.md"
```

### 3. Incremental Recovery after Interruption

If a long execution is interrupted, simply rerun the exact same command. Thanks to the integrated SQLite persistence layer, the application will skip already processed nodes instantly and resume exactly where it left off:

```bash
./auditor --db "my_project_vault.db" --log-file "audit_execution_resume.log"
```

---

## Command-Line Flags and Options

The application accepts the following flags to configure its behavior and control resource utilization during long audit windows.

### Core Configuration Flags
```
--model string          LLM model name (default "qwen3-code-expert")
--db string             SQLite database file (default "audit_vault.db")
--report string         Markdown report file (default "audit_bugs_report.md")
--url string            Ollama API URL (default "http://localhost:11434/api/generate")
--audit-passes int      Number of logical bug audit passes (default 1, 0 to disable)
--suspend-on-error      Stop execution immediately if a function fails definitively after retries
--log-file string       Path to log file for writing logs (enables silent mode if provided)
```

### Resource & Tree Control Flags
```
--workers int           Number of concurrent workers for tree analysis (default 4)
--max-depth int         Maximum recursion depth allowed for tree analysis to prevent stack overflow (default 10)
```
*Note: Connection pooling (max 10 open connections) and request timeouts (120 seconds per LLM call) are managed natively within the application code to guarantee stability during heavy parallel processing.*

---

## Contextual Usage Recommendations & Operational Security

This application is engineered exclusively as a **Deep Static Analysis & Logic Audit Tool** designed to run in fully sovereign, controlled, and isolated corporate environments. Because comprehensive dependency graph mapping and semantic logic analysis are highly compute-intensive (potentially requiring multi-hour executions depending on the codebase size), operation must strictly comply with the following deployment topology:

### 🔴 Architectural Constraints & Exclusions
1. **No CI/CD Pipelines Integration:** This tool is strictly **NOT** designed to be executed inside Continuous Integration/Continuous Deployment workflows (such as GitHub Actions, GitLab CI, or Jenkins). Its deep-reasoning execution model makes it incompatible with the fast feedback cycles required by deployment pipelines.
2. **Local and Air-Gapped Operations Only:** It must strictly run against a **local copy/clone** of the target repositories on private workstations or dedicated on-premise audit servers.
3. **Private Network Boundaries:** The underlying Ollama LLM service must reside within a secure, trusted private network or loopback interface (`localhost`). No source code data or metadata should ever cross the corporate network perimeter.
4. **Authorized Personnel Access:** Execution and report interpretation must be restricted to approved security auditors, code reviewers, or engineering leads.

### ⚙️ Performance Tuning for Extended Audits
For large enterprise repositories where the analysis window spans several hours:
- **Persistent Cache Dependency:** The SQLite database acts as a critical incremental cache. If the process is interrupted or needs to be chunked, existing summaries will be fetched instantly from the database without re-triggering LLM calls.
- **Resource Balancing:** Fine-tune the `--workers` flag based on your local host's processing power (GPU VRAM or CPU threads available for Ollama) to prevent system freezing during overnight runs.

---

## Architecture Components

### 1. Main Application Entry Point (`main.go`)
The primary entry point initializes the application with command-line flags, sets up logging configuration, manages database connections, and orchestrates the audit process through multiple passes.

### 2. Database Layer (`database/db.go`)
Handles SQLite database operations for storing:
- Function catalog information
- Dependency relationships
- Analysis status tracking
- Audit reports

The database manager encapsulates all database operations with methods to manage catalog data, function analysis status, and dependency tracking within a SQLite database. It provides optimized connection pooling and WAL mode configuration to prevent 'database is locked' errors during parallel execution.

### 3. File Processing Utilities (`fileutils/parser.go`)
Provides functionality for:
- Scanning project directories for Go source files
- Extracting specific function code from source files
- Parsing and processing Go code structures

### 4. LLM Integration (`llm/ollama.go`)
Implements communication with Ollama API for language model interactions:
- Model selection and configuration
- Prompt engineering for code analysis
- Response processing and validation
- Token management for optimized LLM calls

### 5. Data Models (`models/models.go`)
Defines the application's data structure including:
- Function definitions with metadata
- Dependency relationships
- Analysis status tracking
- LLM response formats

---

## Application Flow

1. **Initialization**: Command-line flags are parsed, configuration loaded, and dependencies injected
2. **Database Setup**: SQLite connection established with appropriate schema initialization
3. **Project Scanning**: Source files are scanned to catalog all functions in the project
4. **Catalog Synchronization**: Function catalog is saved and synchronized to pending queue
5. **Tree Analysis Pass 1**:
   - Recursive dependency tree analysis using LLM for function extraction
   - Dependency resolution through function call graph traversal
   - Summary generation with dependency context enrichment
6. **Logical Audit Pass 2**:
   - Code logic validation against dependencies
   - Bug detection through semantic analysis
   - Report generation in markdown format

---

## Key Features and Algorithms

### Tree-Based Analysis
The application implements a recursive tree traversal algorithm to understand function dependencies:
- **Dependency Resolution**: Identifies internal Go function dependencies
- **Cycle Detection**: Prevents infinite loops during recursive analysis
- **Summary Aggregation**: Builds comprehensive function summaries through dependency enrichment

### LLM Prompt Engineering
Two distinct LLM passes with specific prompt structures:
1. **Pass 1 (Discovery)**: Extracts function code and identifies direct dependencies with strict JSON formatting
2. **Pass 2 (Audit)**: Validates logic for potential bugs using external dependency context

### Progress Tracking
- Real-time progress visualization with animated progress bar
- Performance metrics tracking (functions per second)
- Detailed logging for debugging and monitoring

---

## Database Schema and Data Models

The SQLite database uses three main tables:

1. **catalog**: Stores original package function entries (`package`, `func_name`, `file_path`)

2. **functions**: Tracks analysis progress (`id`, `package`, `func_name`, `file_path`, `summary`, `status`)

3. **dependencies**: Manages dependency relationships (`function_id`, `dep_name`, `dep_type`)

### Main Data Structures

1. **Function**: Represents a function under audit with its current status lifecycle (`PENDING`, `IN_PROGRESS`, `EXTRACTED`, `AUDITED`, `FAILED_PARSE`).

2. **OllamaRequest**: Structures the request payload, injecting the native `"format": "json"` constraint when required.

3. **P1Response**: Target structure for decoding strict JSON responses from Pass 1, containing the functional summary and mapped dependencies.

---

## AI Transparency and License

### AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for optimizing, testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

### License

MIT License - See [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Nicolas JUHEL
