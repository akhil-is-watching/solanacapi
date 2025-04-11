# Solana Anchor Project

## Overview

This project combines a Go backend service with Solana blockchain development using the Anchor framework. The Go service provides an API for managing, compiling, and deploying Solana programs written with Anchor.

## Requirements

- **Go 1.23.x** - Required for running the backend service
- **Rust (latest version)** - Necessary for Solana program development
- **Anchor CLI (latest version)** - Framework for Solana program development
- **Solana CLI (latest version)** - Tools for interacting with the Solana blockchain

## Detailed Setup Instructions

### 1. Environment Setup

First, ensure all required dependencies are installed on your system:
```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env

# Install Solana CLI
sh -c "$(curl -sSfL https://release.solana.com/stable/install)"

# Install Anchor CLI
cargo install --git https://github.com/coral-xyz/anchor avm --locked
avm install latest
avm use latest
```

### 2. Project Setup

Clone the repository and set up the project:

```bash
# Clone the repository
git clone <repository-url>
cd <repository-name>

# Initialize Anchor workspace
anchor init anchor-workspace

# Build and run the Go service
go build
./project-name
```

## Project Structure

- `/controller` - Go API controllers for handling requests
- `/anchor-workspace` - Solana programs using the Anchor framework
- `/routes` - Routes for the application
- `/types` - Types for the application
- `/helpers` - Helper functions

## API Endpoints

The Go service provides several endpoints for interacting with Solana programs:

- `POST /compile` - Creates a new anchor program and compiles it

## Development Workflow

1. Write your Solana program in the Anchor workspace
2. Use the Go API to build and test your program


## Troubleshooting

If you encounter build errors, check the following:
- Ensure all dependencies are correctly installed
- Verify that your Solana program follows Anchor's conventions
- Check the build output for specific error messages