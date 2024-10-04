#!/bin/bash

# Set the output file names
COVERAGE_FILE="coverage.out"
HTML_REPORT="coverage.html"

# Define color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
RESET='\033[0m'  # No color

# Function to display usage information
usage() {
    echo -e "${YELLOW}Usage: $0${RESET}"
    echo -e "${YELLOW}Run tests with coverage and generate reports.${RESET}"
    exit 1
}

# Run tests and collect coverage
echo -e "${BLUE}Running tests with coverage...${RESET}"
if go test -coverprofile=$COVERAGE_FILE -covermode=atomic ./...; then
    echo -e "${GREEN}Tests completed successfully.${RESET}"
else
    echo -e "${RED}Tests failed. Please fix the issues before checking coverage.${RESET}"
    exit 1
fi

# Generate the coverage report
echo -e "${BLUE}Generating coverage report...${RESET}"
go tool cover -func=$COVERAGE_FILE
go tool cover -html=$COVERAGE_FILE -o $HTML_REPORT

echo -e "${GREEN}Coverage report generated: $HTML_REPORT${RESET}"
