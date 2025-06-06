---
derived_from: modularity
id: package-design
last_modified: '2025-05-14'
enforced_by: code review & style guides
---
# Binding: Organize Go Code Into Purpose-Driven Packages

Design Go packages as cohesive units with clear, focused responsibilities and
well-defined boundaries. Each package should have a single purpose, contain related
functionality, maintain high internal cohesion, and expose a minimal, well-documented
API that hides implementation details.

## Rationale

This binding directly implements our modularity tenet by establishing clear,
well-defined boundaries for Go code organization. Package design is the primary
mechanism for modularity in Go, serving as the foundational unit of code organization
that determines how components interact and compose together.

Think of Go packages like specialized departments in a well-run organization. Each
department has a clear purpose, its members work cohesively toward shared goals, and
there are established protocols for how departments communicate with each other. When a
department tries to do too much or lacks a coherent identity, the organization becomes
confused and inefficient. Similarly, when Go packages lack focus or appropriate
boundaries, codebases become tangled and difficult to maintain. Well-designed packages
create a map of your system's architecture that developers can navigate intuitively.

The impact of package design choices compounds over time. In the early days of a
project, poor package boundaries might seem like a minor inconvenience, but as the
system grows, these boundaries determine how easily developers can understand, test,
refactor, and extend the codebase. Packages with mixed responsibilities create hidden
dependencies and unexpected side effects when modified. In contrast, purpose-driven
packages allow developers to reason about one part of the system without holding the
entire codebase in their head—turning an overwhelmingly complex system into a collection
of manageable pieces.

## Rule Definition

This binding establishes clear requirements for how Go code should be organized into
packages:

- **Package Purpose and Identity**:

  - Each package MUST have a single, well-defined purpose that can be expressed in a
    short sentence
  - Package names MUST be concise, lower-case, single words without underscores that
    describe what the package contains
  - Packages SHOULD NOT be named after their patterns or implementation details (e.g.,
    avoid names like "factory", "manager", "util")
  - Package comments (`// Package foo ...`) at the top of doc.go or primary .go file
    MUST clearly explain the package's purpose

- **Package Structure and Organization**:

  - Projects MUST follow the standard Go project layout with `/cmd`, `/internal`, `/pkg`
    (when needed) directories
  - Group related functionality by domain concepts or features, not by technical roles
    (prefer `internal/user` over `internal/controllers`)
  - Place each Go package in its own directory with a name matching the package
  - Large packages SHOULD be split into more focused sub-packages when they exceed
    2000-3000 lines of code

- **Package Coupling and Cohesion**:

  - Packages MUST exhibit high internal cohesion (all code in the package works together
    for a unified purpose)
  - Packages MUST maintain low external coupling (minimal dependencies on other
    packages)
  - Circular dependencies between packages are STRICTLY PROHIBITED and should be
    detected in CI
  - Import graphs MUST form a directed acyclic graph (DAG) with clear hierarchical
    structure

- **Package API Design**:

  - Package APIs MUST be intentionally designed, not emergent from implementation needs
  - Only types, functions, and constants directly related to the package's purpose
    should be exported
  - Implementation details MUST be unexported (private to the package)
  - PREFER accepting interfaces and returning concrete types

- **Exceptions and Special Cases**:

  - Small utilities or helper functions SHOULD be kept in the package they serve rather
    than creating tiny utility packages
  - Utility code needed by multiple packages SHOULD be organized into purposeful shared
    packages (e.g., `internal/validation`) rather than generic catch-all utilities
  - Package main is an exception that often contains glue code connecting other
    packages; it should be kept minimal and focused on application bootstrapping

## Practical Implementation

1. **Establish a Standard Project Layout**: Start with the Go community's standard
   layout to provide familiar structure:

   ```
   project-root/
   ├── cmd/                 # Command-line applications
   │   └── myapp/
   │       └── main.go     # Application entry point
   ├── internal/            # Private code that cannot be imported
   │   ├── auth/           # Authentication-related code
   │   ├── models/         # Domain models
   │   └── server/         # HTTP server implementation
   ├── pkg/                 # Public library code (use sparingly)
   │   └── validator/      # Public validation library
   ├── api/                 # API definitions (OpenAPI/Protobuf)
   ├── web/                 # Web assets
   ├── configs/             # Configuration files
   ├── docs/                # Documentation
   ├── go.mod               # Module definition
   └── go.sum               # Module checksums
   ```

   Strictly adhere to the semantic meaning of these directories:

   - `/cmd`: Entry points for executables, with minimal code
   - `/internal`: Private code that cannot be imported by other modules
   - `/pkg`: Public library code that can be imported by other modules (use only when
     necessary)
   - `/api`: API definition files, schemas, protocol definitions

1. **Organize by Domain Concepts**: Structure packages around business domains and
   features rather than technical layers:

   ```go
   // ❌ BAD: Technical/layer-based organization
   internal/
   ├── controllers/     // All controllers across features
   ├── services/        // All business logic
   ├── repositories/    // All data access
   └── models/          // All data structures

   // ✅ GOOD: Domain/feature-based organization
   internal/
   ├── user/            // Everything related to users
   │   ├── user.go      // Core domain types
   │   ├── service.go   // Business logic
   │   ├── repository.go // Data access
   │   └── handler.go   // HTTP handlers
   ├── order/           // Everything related to orders
   │   ├── order.go
   │   ├── service.go
   │   ├── repository.go
   │   └── handler.go
   └── payment/         // Everything related to payments
       ├── payment.go
       ├── service.go
       ├── repository.go
       └── handler.go
   ```

   This approach promotes:

   - Higher cohesion as related code stays together
   - Clearer ownership of features
   - Easier navigation for new developers
   - Natural boundaries for changes and testing

1. **Design Clean Package APIs**: Explicitly define what's exported and what remains
   internal:

   For each package, create a clear contract with the rest of the system:

   ```go
   // package: internal/order/order.go

   // Package order manages order processing and storage.
   package order

   // Order represents a customer order in the system.
   // This type is exported as part of the package's public API.
   type Order struct {
       ID        string
       CustomerID string
       Items     []Item
       Status    OrderStatus
       CreatedAt time.Time
   }

   // Item represents an individual line item within an order.
   // This type is exported as it's part of Order.
   type Item struct {
       ProductID string
       Quantity  int
       Price     decimal.Decimal
   }

   // OrderStatus represents the current state of an order.
   type OrderStatus string

   // Define valid order statuses as constants
   const (
       StatusPending   OrderStatus = "pending"
       StatusConfirmed OrderStatus = "confirmed"
       StatusShipped   OrderStatus = "shipped"
       StatusDelivered OrderStatus = "delivered"
       StatusCancelled OrderStatus = "cancelled"
   )

   // internal type not exposed outside the package
   type orderValidator struct {
       // implementation details
   }

   // unexported function used internally
   func validateOrderItems(items []Item) error {
       // implementation
   }
   ```

   Then define service interfaces that express dependencies:

   ```go
   // package: internal/order/service.go

   // Service defines the operations available on orders.
   // This interface is exported to allow other packages to use it.
   type Service interface {
       Create(ctx context.Context, customerID string, items []Item) (*Order, error)
       Get(ctx context.Context, id string) (*Order, error)
       Update(ctx context.Context, order *Order) error
       Cancel(ctx context.Context, id string) error
   }

   // Repository defines the storage operations required by the order service.
   // This is an internal dependency the service needs.
   type Repository interface {
       Store(ctx context.Context, order *Order) error
       FindByID(ctx context.Context, id string) (*Order, error)
       Update(ctx context.Context, order *Order) error
   }

   // implementation of the service
   type service struct {
       repo Repository
       // other dependencies
   }

   // NewService creates a new order service.
   // This factory function is exported to allow creating the service.
   func NewService(repo Repository) Service {
       return &service{
           repo: repo,
       }
   }

   // Implementation methods for the service
   func (s *service) Create(ctx context.Context, customerID string, items []Item) (*Order, error) {
       // implementation
   }
   ```

1. **Set up Dependency Management**: Use interfaces and dependency injection to manage
   and limit coupling:

   ```go
   // package: internal/app/app.go

   // App represents the application and its dependencies.
   type App struct {
       UserService    user.Service
       OrderService   order.Service
       PaymentService payment.Service
       // other dependencies
   }

   // NewApp creates a new application instance with all dependencies wired up.
   func NewApp(cfg Config) (*App, error) {
       // Set up database
       db, err := database.Connect(cfg.DatabaseURL)
       if err != nil {
           return nil, fmt.Errorf("connecting to database: %w", err)
       }

       // Create repositories
       userRepo := user.NewRepository(db)
       orderRepo := order.NewRepository(db)
       paymentRepo := payment.NewRepository(db)

       // Create services with their dependencies
       userService := user.NewService(userRepo, cfg.UserConfig)
       orderService := order.NewService(orderRepo, userService)
       paymentService := payment.NewService(paymentRepo, orderService, cfg.PaymentConfig)

       return &App{
           UserService:    userService,
           OrderService:   orderService,
           PaymentService: paymentService,
       }, nil
   }
   ```

1. **Visualize and Enforce Package Relationships**: Regularly analyze and optimize your
   dependency graph:

   ```bash
   # Install go-tools to analyze dependencies
   go install golang.org/x/tools/cmd/godepgraph@latest

   # Generate dependency graph for your project
   godepgraph -s github.com/your-org/your-project > deps.dot

   # Convert to a viewable format
   dot -Tpng deps.dot -o deps.png
   ```

   Design your CI pipeline to enforce healthy package structures:

   ```yaml
   # In your CI configuration
   checks:
     - name: Detect circular dependencies
       command: go-cyclo
       args: ["./..."]

     - name: Check package sizes
       command: go-package-size
       args: ["--max-lines=3000", "./..."]
   ```

## Examples

```go
// ❌ BAD: Unfocused package with mixed responsibilities
// Package utils provides various helper functions for the application.
package utils

// User-related functions
func ValidateUserEmail(email string) bool { /* ... */ }
func HashPassword(password string) (string, error) { /* ... */ }

// Order-related functions
func CalculateOrderTotal(items []Item) decimal.Decimal { /* ... */ }
func GenerateOrderID() string { /* ... */ }

// Generic utilities
func FormatTimestamp(t time.Time) string { /* ... */ }
func ToSnakeCase(s string) string { /* ... */ }
```

```go
// ✅ GOOD: Focused packages with clear responsibilities
// Package user handles user-related operations and model.
package user

// ValidateEmail checks if an email address is valid.
func ValidateEmail(email string) bool { /* ... */ }

// HashPassword securely hashes a user password.
func HashPassword(password string) (string, error) { /* ... */ }
```

```go
// Package order manages customer orders and order processing.
package order

// CalculateTotal computes the total price for an order's items.
func CalculateTotal(items []Item) decimal.Decimal { /* ... */ }

// GenerateID creates a new unique order identifier.
func GenerateID() string { /* ... */ }
```

```go
// Package timeutil provides time formatting and parsing utilities.
package timeutil

// Format returns a formatted timestamp string.
func Format(t time.Time) string { /* ... */ }
```

```go
// ❌ BAD: Circular dependency between packages
// package user imports payment
package user

import (
    "github.com/your-org/your-project/internal/payment"
)

type User struct {
    // ...
}

func (u *User) CanMakePayment(amount decimal.Decimal) bool {
    return payment.CheckUserCredit(u.ID, amount)
}
```

```go
// package payment imports user
package payment

import (
    "github.com/your-org/your-project/internal/user"
)

func CheckUserCredit(userID string, amount decimal.Decimal) bool {
    u, err := user.GetByID(userID)
    if err != nil {
        return false
    }
    return u.CreditLimit >= amount
}
```

```go
// ✅ GOOD: Break circular dependency with interfaces
// package user defines an interface for payment operations
package user

type PaymentChecker interface {
    CheckCredit(userID string, amount decimal.Decimal) bool
}

type User struct {
    // ...
}

// Service handles user operations.
type Service struct {
    paymentChecker PaymentChecker
}

func NewService(paymentChecker PaymentChecker) *Service {
    return &Service{
        paymentChecker: paymentChecker,
    }
}

func (s *Service) CanMakePayment(userID string, amount decimal.Decimal) bool {
    return s.paymentChecker.CheckCredit(userID, amount)
}
```

```go
// package payment implements user.PaymentChecker
package payment

import (
    "github.com/your-org/your-project/internal/user"
)

// Service handles payment operations.
type Service struct {
    // dependencies without importing user
    // db, etc.
}

// Make sure Service implements user.PaymentChecker
var _ user.PaymentChecker = (*Service)(nil)

func (s *Service) CheckCredit(userID string, amount decimal.Decimal) bool {
    // Implementation without importing user package
    creditLimit, err := s.fetchCreditLimitFromDB(userID)
    if err != nil {
        return false
    }
    return creditLimit >= amount
}
```

## Related Bindings

- [modularity](../../docs/tenets/modularity.md): This binding is a Go-specific application of
  general modularity principles. While modularity establishes the conceptual foundation,
  go-package-design provides concrete rules for implementing modularity in Go codebases
  through package design.

- [dependency-inversion](../../docs/bindings/core/dependency-inversion.md): Well-designed Go packages often use
  dependency inversion to manage coupling between packages. By defining interfaces in
  consumer packages rather than implementation packages, you create a more flexible,
  testable system with dependencies pointing in the right direction.

- [code-size](../../docs/bindings/core/code-size.md): Package design and code size work together to maintain
  manageable units of code. When packages grow too large, they often need to be broken
  down into smaller, more focused packages.

- [go-interface-design](../../docs/bindings/categories/go/interface-design.md): Package boundaries and interface design
  are closely related. Well-designed interfaces help define clean package boundaries and
  APIs, while good package organization provides natural places to define
  domain-specific interfaces.

- [hex-domain-purity](../../docs/bindings/core/hex-domain-purity.md): Package design in Go can directly support
  hexagonal architecture by organizing packages to reflect domain boundaries and
  infrastructure adapters, keeping domain logic pure from external concerns.
