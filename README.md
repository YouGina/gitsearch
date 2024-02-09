# gitsearch
This tool allows you to search through GitHub repositories to find files based on the provided query, iterating through up to 1000 search results

## Installation

### Prerequisites

- Go 1.20 or higher. You can download it from [the official Go website](https://golang.org/dl/).

### Setup
```bash
go install github.com/YouGina/gitsearch@latest
```

## Usage

To use this tool, run the binary with the path to your tokens.txt file and your search query as arguments. For example:

```bash
gitsearch /path/to/your/tokens.txt "your search query"
```

## Features

* Search GitHub repositories for specific queries in files named subdomains.txt.
* Automatically handle GitHub API rate limiting by rotating through multiple provided tokens.
* Decode and print base64 encoded file contents directly in the terminal.
