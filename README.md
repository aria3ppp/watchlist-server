# Watchlist

Watchlist provides a convenient way for users to manage a list of movies and episodes they want to watch, and also create new movies and series.

[![Tests](https://github.com/aria3ppp/watchlist-server/actions/workflows/tests.yml/badge.svg)](https://github.com/aria3ppp/watchlist-server/actions/workflows/tests.yml)
[![Coverage Status](https://coveralls.io/repos/github/aria3ppp/watchlist-server/badge.svg?branch=master)](https://coveralls.io/github/aria3ppp/watchlist-server?branch=master)
[![Lint](https://github.com/aria3ppp/watchlist-server/actions/workflows/lint.yml/badge.svg)](https://github.com/aria3ppp/watchlist-server/actions/workflows/lint.yml)

## Code Architecture
The Watchlist API is developed in Go language and leverages the Echo router. It follows a modular, three-layer architecture with Transport, Application, and Repository layers. This design ensures single responsibility, better scalability and efficient data storage through the Repository pattern. The code is thoroughly tested with gomock and has comprehensive integration and end-to-end tests to guarantee seamless integration of third-party services and a fully functional API.

Users can sign up, log in, and authorize using JWT tokens. The API also enables token refresh to avoid repetitive logins. User security is prioritized with secure bcrypt hashing of passwords and refresh tokens.

The Watchlist API offers users a history of changes made by others to movies, series, and episodes. It has a robust search functionality powered by Elasticsearch and uses MinIO to store user avatars and movie and series posters.

## Installation
prerequisite:

- Docker and Docker compose
- Golang >=1.19
- Make

To clone project:

```bash
git clone https://github.com/aria3ppp/watchlist-server.git
```


To run the server in containers:

```bash
make server-up
```

To run tests:

```bash
make services-up && make test-all-cover
```

Now the local api documentation is available at [http://localhost:8080/v1/openapi](http://localhost:8080/v1/openapi).

The api documentation is also available to read for everyone at [http://aria3ppp.ir:8080/v1/openapi](http://aria3ppp.ir:8080/v1/openapi).

OAS3.0 openapi specs: [http://aria3ppp.ir:8080/v1/openapi/openapi.json](http://aria3ppp.ir:8080/v1/openapi/openapi.json).