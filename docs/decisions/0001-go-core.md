# ADR 0001: Use Go for the analysis core

Status: Accepted

Go provides a small portable binary, efficient streaming I/O, explicit cancellation, and mature standard-library archive support. The core therefore owns archive safety, hashing, classification, orchestration, and workspace output. Language-specific analysis remains out of process.

