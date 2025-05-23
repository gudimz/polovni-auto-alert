# Polovni Auto Alert

![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/gudimz/polovni-auto-alert/checker.yaml)
[![Coverage Status](https://coveralls.io/repos/github/gudimz/polovni-auto-alert/badge.svg?branch=main)](https://coveralls.io/github/gudimz/polovni-auto-alert?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/gudimz/polovni-auto-alert)](https://goreportcard.com/report/github.com/gudimz/polovni-auto-alert)
![GitHub License](https://img.shields.io/github/license/gudimz/polovni-auto-alert)
![GitHub Tag](https://img.shields.io/github/v/tag/gudimz/polovni-auto-alert)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gudimz/polovni-auto-alert)
![GitHub last commit](https://img.shields.io/github/last-commit/gudimz/polovni-auto-alert)
## Overview

Polovni Auto Alert is a Telegram bot specifically designed to help users stay updated with the latest car listings from [Polovni Automobili](https://www.polovniautomobili.com/). The bot allows users to subscribe to alerts for specific car brands, models, and other criteria available on the website.

## Features

- Subscribe to car listings alerts
- Unsubscribe from alerts
- List current subscriptions
- Set filters for brand, model, chassis, region, price, and year
- Receive notifications for new listings in Telegram

## Getting Started

### Setting Up Environment Variables

To configure the environment variables for the project, create a `.env` file in the root directory of the project. You can use the provided `.env.example` file as a template.

### Fetching data from Polovni Automobili
Use the provided Makefile to fetch the data from Polovni Automobili before running the project:

```sh
# Fetch the data from Polovni Automobili
make docker-compose-up-fetcher
```

### Building and Running the Project
Use the provided Makefile to build and run the project:

```sh
# Build the binaries
make all

# Build Docker images and run containers
make docker-compose-up

```

###  Stopping the Services
To stop and remove the running containers, use:

```sh
make docker-compose-down
```

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/gudimz/polovni-auto-alert/blob/main/LICENSE) file for details.
