## ADDED Requirements

### Requirement: Framework components serve one API contract
Every framework component SHALL serve the same HTTP API for the
architecture's contract: the same routes, methods, status codes, response
media type, JSON error bodies, and request limits. Each framework component
SHALL ship the same framework-independent API test, and gobento SHALL fail
its own tests when the copies differ.

#### Scenario: Same requests, same responses
- **WHEN** two projects differ only in the selected framework and receive the same sequence of requests
- **THEN** they return the same status codes, an `application/json` media type, and equal JSON bodies

#### Scenario: Unsupported method on a known route
- **WHEN** a client sends `DELETE /items` to a project generated with any framework
- **THEN** the response status is `405`

#### Scenario: API test copies drift
- **WHEN** one framework component's API test differs from another's
- **THEN** gobento's test suite fails naming both components
