# Go Service AI Development Example

Licensed under the MIT License. See file [LICENSE](./LICENSE).

A simple Go microservice providing an HTTP REST API with standard library routing, graceful shutdown, and containerization support.

## API Endpoints

### `GET /hello`

Returns a greeting message in JSON format.

**Request:**
```http
GET /hello HTTP/1.1
Host: localhost:8080
```

**Response:**
- **Status:** `200 OK`
- **Content-Type:** `application/json`
- **Body:**
  ```json
  {"message":"hello"}
  ```

## Getting Started

### Prerequisites

- [Go](https://go.dev/) 1.22 or higher
- Optional: [Docker](https://www.docker.com/)

### Running Locally

```bash
# Run the server directly
go run .
```

By default, the server listens on port `8080`. You can configure the port using the `PORT` environment variable:

```bash
PORT=9000 go run .
```

### Testing the Endpoint

In another terminal, send a GET request:

```bash
curl -i http://localhost:8080/hello
```

Expected output:
```http
HTTP/1.1 200 OK
Content-Type: application/json
Date: ...
Content-Length: 20

{"message":"hello"}
```

### Running Tests

Execute the unit and integration test suite:

```bash
go test -v ./...
```

Run test coverage:

```bash
go test -cover ./...
```

### Building the Binary

```bash
go build -o server .
./server
```

### Docker

Build and run the container image:

```bash
docker build -t go-service-ai .
docker run -p 8080:8080 go-service-ai
```

## Links

* [Go Documentation](https://go.dev/doc/)
* [golang](https://hub.docker.com/_/golang/) - Official Image - Docker Hub
