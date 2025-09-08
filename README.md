microservices-library-go (modular monorepo)

This repository provides modular building blocks for Go microservices: communication, database, storage, messaging, auth, logging, monitoring, tracing, rate limiting, utils, and more.

Structure

- Feature areas live in top-level directories (e.g. `communication`, `database`, `storage`, `messaging`, `auth`, `logging`, `graphql`, `monitoring`).
- Each area can be developed independently with its own `go.mod` for clean dependency boundaries.
- A root `go.work` ties modules together for local development.

Local development

- Ensure Go 1.21+.
- From repo root:

```bash
go work use -r .
```

Now any submodule can import siblings via standard module paths using local replacements from `go.work`.

Usage in your microservice

Import only the parts you need. Examples:

```go
import (
    // Communication gateway and providers
    commgw "github.com/anasamu/microservices-library-go/libs/communication/gateway"
    httpprov "github.com/anasamu/microservices-library-go/libs/communication/providers/http"

    // Database gateway and providers
    dbgw "github.com/anasamu/microservices-library-go/libs/database/gateway"
    pgprov "github.com/anasamu/microservices-library-go/libs/database/providers/postgresql"

    // Storage gateway and providers
    storegw "github.com/anasamu/microservices-library-go/libs/storage/gateway"
    s3prov "github.com/anasamu/microservices-library-go/libs/storage/providers/s3"
)
```

Notes

- Minimal `core` module added to satisfy existing requires; extend as needed.
- S3 provider AWS config import fixed to avoid name clash with local `config` package.

Repository

See: `https://github.com/anasamu/microservices-library-go`

