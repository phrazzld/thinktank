---
id: dependency-injection-patterns
last_modified: '2025-06-03'
derived_from: orthogonality
enforced_by: 'Go interfaces, constructor functions, dependency injection frameworks, code review'
---

# Binding: Design Modular Systems Through Go's Interface-Based Dependency Injection

Use Go's interface system and constructor patterns to create loosely coupled, testable components that depend on abstractions rather than concrete implementations. Proper dependency injection enables component independence while maintaining clear boundaries and making testing straightforward.

## Rationale

This binding implements our orthogonality tenet by using Go's interface system to create truly independent components that can be developed, tested, and modified without affecting each other. Dependency injection is fundamental to building maintainable systems because it allows you to compose functionality from independent parts while keeping those parts loosely coupled.

Think of dependency injection like building with standardized electrical components. Each device (your struct) declares what kind of power supply it needs (interface requirements) without caring about the specific brand or model of power supply you connect to it. A LED light strip declares "I need 12V DC power" through its connector interface, but it doesn't care whether that power comes from a wall adapter, a battery pack, or a solar panel. This flexibility allows you to swap power sources without rewiring the light strip, test the light strip with a bench power supply, and combine different components in ways the original designers never imagined.

Without dependency injection, your components become like hardwired appliances where the microwave is permanently connected to one specific outlet in one specific wall. Moving the microwave requires rewiring the house, testing it requires access to that exact outlet, and if the outlet breaks, the microwave becomes useless. Go's interface system provides the standardized "plugs and sockets" that allow components to connect without hardwiring, enabling the flexibility and testability that orthogonal design requires.

## Rule Definition

Dependency injection patterns must establish these Go-specific practices:

- **Interface-Based Dependencies**: Define interfaces for all external dependencies and inject implementations through constructors. Components should depend on interfaces, not concrete types.

- **Constructor Injection**: Use constructor functions that accept dependencies as parameters, making all dependencies explicit and ensuring components are fully initialized when created.

- **Interface Segregation**: Design focused interfaces that represent single capabilities rather than large, monolithic interfaces. This allows components to depend only on the functionality they actually use.

- **Dependency Inversion**: High-level modules should not depend on low-level modules. Both should depend on abstractions (interfaces) that the high-level modules define.

- **Explicit Dependency Graphs**: Make dependency relationships clear through constructor signatures and avoid hidden dependencies like global variables or singleton access patterns.

- **Testable Component Design**: Structure components so that all external dependencies can be easily mocked or stubbed for testing, enabling isolation of the component under test.

**Injection Patterns:**
- Constructor injection (preferred for required dependencies)
- Method injection (for optional dependencies or strategy patterns)
- Interface injection (for dynamic dependency resolution)
- Factory patterns (for complex dependency creation)

**Component Boundaries:**
- Service layer interfaces (business logic abstraction)
- Repository interfaces (data access abstraction)
- External service interfaces (third-party integration)
- Infrastructure interfaces (logging, metrics, caching)

## Practical Implementation

1. **Design Service Layer with Interface Dependencies**: Create services that depend on abstractions:

   ```go
   // ✅ GOOD: Service with interface-based dependencies

   // Define focused interfaces for dependencies
   type UserRepository interface {
       FindByID(ctx context.Context, id string) (*User, error)
       Save(ctx context.Context, user *User) error
       Delete(ctx context.Context, id string) error
   }

   type EmailService interface {
       SendWelcomeEmail(ctx context.Context, user *User) error
       SendPasswordResetEmail(ctx context.Context, user *User, token string) error
   }

   type Logger interface {
       Info(msg string, fields ...interface{})
       Error(msg string, fields ...interface{})
       Debug(msg string, fields ...interface{})
   }

   // Service that depends on interfaces, not concrete types
   type UserService struct {
       userRepo     UserRepository
       emailService EmailService
       logger       Logger
   }

   // Constructor injection makes dependencies explicit
   func NewUserService(
       userRepo UserRepository,
       emailService EmailService,
       logger Logger,
   ) *UserService {
       return &UserService{
           userRepo:     userRepo,
           emailService: emailService,
           logger:       logger,
       }
   }

   func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
       s.logger.Info("creating user", "email", req.Email)

       user := &User{
           ID:    generateID(),
           Email: req.Email,
           Name:  req.Name,
       }

       if err := s.userRepo.Save(ctx, user); err != nil {
           s.logger.Error("failed to save user", "error", err)
           return nil, fmt.Errorf("failed to save user: %w", err)
       }

       if err := s.emailService.SendWelcomeEmail(ctx, user); err != nil {
           s.logger.Error("failed to send welcome email", "user_id", user.ID, "error", err)
           // Don't fail user creation if email fails
       }

       s.logger.Info("user created successfully", "user_id", user.ID)
       return user, nil
   }
   ```

2. **Implement Interface Segregation for Focused Dependencies**: Break large interfaces into focused ones:

   ```go
   // ✅ GOOD: Segregated interfaces based on actual needs

   // Instead of one large interface, define focused capabilities
   type UserReader interface {
       FindByID(ctx context.Context, id string) (*User, error)
       FindByEmail(ctx context.Context, email string) (*User, error)
   }

   type UserWriter interface {
       Save(ctx context.Context, user *User) error
       Delete(ctx context.Context, id string) error
   }

   type UserSearcher interface {
       Search(ctx context.Context, query UserSearchQuery) ([]*User, error)
   }

   // Notification capabilities separated by channel
   type EmailNotifier interface {
       SendEmail(ctx context.Context, to string, subject, body string) error
   }

   type SMSNotifier interface {
       SendSMS(ctx context.Context, to, message string) error
   }

   type PushNotifier interface {
       SendPush(ctx context.Context, deviceID, message string) error
   }

   // Services depend only on what they actually use
   type UserQueryService struct {
       userReader UserReader
       logger     Logger
   }

   func NewUserQueryService(userReader UserReader, logger Logger) *UserQueryService {
       return &UserQueryService{
           userReader: userReader,
           logger:     logger,
       }
   }

   type UserNotificationService struct {
       userReader    UserReader
       emailNotifier EmailNotifier
       smsNotifier   SMSNotifier
       logger        Logger
   }

   func NewUserNotificationService(
       userReader UserReader,
       emailNotifier EmailNotifier,
       smsNotifier SMSNotifier,
       logger Logger,
   ) *UserNotificationService {
       return &UserNotificationService{
           userReader:    userReader,
           emailNotifier: emailNotifier,
           smsNotifier:   smsNotifier,
           logger:        logger,
       }
   }

   func (s *UserNotificationService) NotifyUser(ctx context.Context, userID, message string) error {
       user, err := s.userReader.FindByID(ctx, userID)
       if err != nil {
           return fmt.Errorf("failed to find user: %w", err)
       }

       // Send via preferred notification method
       if user.Preferences.NotificationMethod == "sms" {
           return s.smsNotifier.SendSMS(ctx, user.PhoneNumber, message)
       }

       return s.emailNotifier.SendEmail(ctx, user.Email, "Notification", message)
   }
   ```

3. **Use Dependency Inversion with Application-Defined Interfaces**: Let high-level modules define their interface needs:

   ```go
   // ✅ GOOD: Application layer defines interfaces, infrastructure implements them

   // Application layer defines what it needs from infrastructure
   package application

   // The application defines this interface based on its needs
   type OrderEventPublisher interface {
       PublishOrderCreated(ctx context.Context, order *Order) error
       PublishOrderCancelled(ctx context.Context, orderID string) error
       PublishOrderShipped(ctx context.Context, orderID string, trackingInfo TrackingInfo) error
   }

   type PaymentProcessor interface {
       ChargeCard(ctx context.Context, amount Money, cardToken string) (*PaymentResult, error)
       RefundPayment(ctx context.Context, paymentID string, amount Money) (*RefundResult, error)
   }

   type InventoryManager interface {
       ReserveItems(ctx context.Context, items []OrderItem) (*Reservation, error)
       ReleaseReservation(ctx context.Context, reservationID string) error
       CommitReservation(ctx context.Context, reservationID string) error
   }

   // Order service defines its own interfaces
   type OrderService struct {
       orderRepo        OrderRepository
       eventPublisher   OrderEventPublisher
       paymentProcessor PaymentProcessor
       inventoryManager InventoryManager
       logger           Logger
   }

   func NewOrderService(
       orderRepo OrderRepository,
       eventPublisher OrderEventPublisher,
       paymentProcessor PaymentProcessor,
       inventoryManager InventoryManager,
       logger Logger,
   ) *OrderService {
       return &OrderService{
           orderRepo:        orderRepo,
           eventPublisher:   eventPublisher,
           paymentProcessor: paymentProcessor,
           inventoryManager: inventoryManager,
           logger:           logger,
       }
   }

   func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
       // Reserve inventory first
       reservation, err := s.inventoryManager.ReserveItems(ctx, req.Items)
       if err != nil {
           return nil, fmt.Errorf("failed to reserve inventory: %w", err)
       }

       // Process payment
       paymentResult, err := s.paymentProcessor.ChargeCard(ctx, req.Total, req.CardToken)
       if err != nil {
           // Release reservation on payment failure
           s.inventoryManager.ReleaseReservation(ctx, reservation.ID)
           return nil, fmt.Errorf("payment failed: %w", err)
       }

       // Create order
       order := &Order{
           ID:            generateOrderID(),
           CustomerID:    req.CustomerID,
           Items:         req.Items,
           Total:         req.Total,
           PaymentID:     paymentResult.PaymentID,
           ReservationID: reservation.ID,
           Status:        OrderStatusPending,
           CreatedAt:     time.Now(),
       }

       if err := s.orderRepo.Save(ctx, order); err != nil {
           // Rollback payment and reservation on save failure
           s.paymentProcessor.RefundPayment(ctx, paymentResult.PaymentID, req.Total)
           s.inventoryManager.ReleaseReservation(ctx, reservation.ID)
           return nil, fmt.Errorf("failed to save order: %w", err)
       }

       // Commit inventory reservation
       if err := s.inventoryManager.CommitReservation(ctx, reservation.ID); err != nil {
           s.logger.Error("failed to commit inventory reservation",
               "order_id", order.ID, "reservation_id", reservation.ID, "error", err)
       }

       // Publish event (don't fail order creation if this fails)
       if err := s.eventPublisher.PublishOrderCreated(ctx, order); err != nil {
           s.logger.Error("failed to publish order created event",
               "order_id", order.ID, "error", err)
       }

       return order, nil
   }
   ```

4. **Create Factory Functions for Complex Dependency Resolution**: Use factories when dependency creation is complex:

   ```go
   // ✅ GOOD: Factory pattern for complex dependency assembly

   type ServiceConfig struct {
       DatabaseURL      string
       RedisURL         string
       EmailAPIKey      string
       LogLevel         string
       PaymentGateway   string
   }

   type Services struct {
       UserService  *UserService
       OrderService *OrderService
       Logger       Logger
   }

   func NewServices(cfg ServiceConfig) (*Services, error) {
       // Create infrastructure dependencies
       logger := newLogger(cfg.LogLevel)

       db, err := sql.Open("postgres", cfg.DatabaseURL)
       if err != nil {
           return nil, fmt.Errorf("failed to connect to database: %w", err)
       }

       redisClient := redis.NewClient(&redis.Options{
           Addr: cfg.RedisURL,
       })

       // Create repositories
       userRepo := &PostgresUserRepository{db: db}
       orderRepo := &PostgresOrderRepository{db: db}

       // Create external services
       emailService := &SendGridEmailService{
           apiKey: cfg.EmailAPIKey,
           client: http.DefaultClient,
       }

       var paymentProcessor PaymentProcessor
       switch cfg.PaymentGateway {
       case "stripe":
           paymentProcessor = &StripePaymentProcessor{
               apiKey: cfg.StripeAPIKey,
           }
       case "paypal":
           paymentProcessor = &PayPalPaymentProcessor{
               clientID:     cfg.PayPalClientID,
               clientSecret: cfg.PayPalClientSecret,
           }
       default:
           return nil, fmt.Errorf("unsupported payment gateway: %s", cfg.PaymentGateway)
       }

       // Create cache
       cache := &RedisCache{client: redisClient}

       // Create event publisher
       eventPublisher := &RedisEventPublisher{
           client: redisClient,
           logger: logger,
       }

       // Assemble services with their dependencies
       userService := NewUserService(userRepo, emailService, logger)

       orderService := NewOrderService(
           orderRepo,
           eventPublisher,
           paymentProcessor,
           &InventoryService{cache: cache, logger: logger},
           logger,
       )

       return &Services{
           UserService:  userService,
           OrderService: orderService,
           Logger:       logger,
       }, nil
   }

   // Alternative: Use dependency injection framework like wire
   // //go:build wireinject
   // // +build wireinject

   func NewServicesWithWire(cfg ServiceConfig) (*Services, error) {
       wire.Build(
           // Infrastructure
           newLogger,
           newDatabase,
           newRedisClient,

           // Repositories
           wire.Bind(new(UserRepository), new(*PostgresUserRepository)),
           NewPostgresUserRepository,

           wire.Bind(new(OrderRepository), new(*PostgresOrderRepository)),
           NewPostgresOrderRepository,

           // External services
           wire.Bind(new(EmailService), new(*SendGridEmailService)),
           NewSendGridEmailService,

           wire.Bind(new(PaymentProcessor), new(*StripePaymentProcessor)),
           NewStripePaymentProcessor,

           // Application services
           NewUserService,
           NewOrderService,

           // Service collection
           wire.Struct(new(Services), "*"),
       )

       return &Services{}, nil
   }
   ```

5. **Design for Testability with Mock-Friendly Interfaces**: Create components that are easy to test in isolation:

   ```go
   // ✅ GOOD: Testable components with mock-friendly interfaces

   // Test file demonstrates how dependency injection enables easy testing
   package application_test

   import (
       "context"
       "errors"
       "testing"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/mock"
   )

   // Mock implementations for testing
   type MockUserRepository struct {
       mock.Mock
   }

   func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
       args := m.Called(ctx, id)
       if args.Get(0) == nil {
           return nil, args.Error(1)
       }
       return args.Get(0).(*User), args.Error(1)
   }

   func (m *MockUserRepository) Save(ctx context.Context, user *User) error {
       args := m.Called(ctx, user)
       return args.Error(0)
   }

   func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
       args := m.Called(ctx, id)
       return args.Error(0)
   }

   type MockEmailService struct {
       mock.Mock
   }

   func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, user *User) error {
       args := m.Called(ctx, user)
       return args.Error(0)
   }

   type MockLogger struct {
       mock.Mock
   }

   func (m *MockLogger) Info(msg string, fields ...interface{}) {
       m.Called(msg, fields)
   }

   func (m *MockLogger) Error(msg string, fields ...interface{}) {
       m.Called(msg, fields)
   }

   func TestUserService_CreateUser_Success(t *testing.T) {
       // Arrange
       ctx := context.Background()
       mockRepo := &MockUserRepository{}
       mockEmail := &MockEmailService{}
       mockLogger := &MockLogger{}

       service := NewUserService(mockRepo, mockEmail, mockLogger)

       req := CreateUserRequest{
           Email: "test@example.com",
           Name:  "Test User",
       }

       // Setup mock expectations
       mockLogger.On("Info", "creating user", mock.AnythingOfType("string"), req.Email).Once()
       mockRepo.On("Save", ctx, mock.MatchedBy(func(user *User) bool {
           return user.Email == req.Email && user.Name == req.Name
       })).Return(nil).Once()
       mockEmail.On("SendWelcomeEmail", ctx, mock.AnythingOfType("*User")).Return(nil).Once()
       mockLogger.On("Info", "user created successfully", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Once()

       // Act
       user, err := service.CreateUser(ctx, req)

       // Assert
       assert.NoError(t, err)
       assert.NotNil(t, user)
       assert.Equal(t, req.Email, user.Email)
       assert.Equal(t, req.Name, user.Name)
       assert.NotEmpty(t, user.ID)

       // Verify all expectations were met
       mockRepo.AssertExpectations(t)
       mockEmail.AssertExpectations(t)
       mockLogger.AssertExpectations(t)
   }

   func TestUserService_CreateUser_RepositoryError(t *testing.T) {
       // Arrange
       ctx := context.Background()
       mockRepo := &MockUserRepository{}
       mockEmail := &MockEmailService{}
       mockLogger := &MockLogger{}

       service := NewUserService(mockRepo, mockEmail, mockLogger)

       req := CreateUserRequest{
           Email: "test@example.com",
           Name:  "Test User",
       }

       expectedErr := errors.New("database connection failed")

       // Setup mock expectations
       mockLogger.On("Info", "creating user", mock.AnythingOfType("string"), req.Email).Once()
       mockRepo.On("Save", ctx, mock.AnythingOfType("*User")).Return(expectedErr).Once()
       mockLogger.On("Error", "failed to save user", "error", expectedErr).Once()

       // Act
       user, err := service.CreateUser(ctx, req)

       // Assert
       assert.Error(t, err)
       assert.Nil(t, user)
       assert.Contains(t, err.Error(), "failed to save user")
       assert.Contains(t, err.Error(), expectedErr.Error())

       // Email service should not be called if save fails
       mockEmail.AssertNotCalled(t, "SendWelcomeEmail")

       // Verify expectations
       mockRepo.AssertExpectations(t)
       mockLogger.AssertExpectations(t)
   }

   func TestUserService_CreateUser_EmailFailureDoesNotFailCreation(t *testing.T) {
       // Arrange
       ctx := context.Background()
       mockRepo := &MockUserRepository{}
       mockEmail := &MockEmailService{}
       mockLogger := &MockLogger{}

       service := NewUserService(mockRepo, mockEmail, mockLogger)

       req := CreateUserRequest{
           Email: "test@example.com",
           Name:  "Test User",
       }

       emailErr := errors.New("email service unavailable")

       // Setup mock expectations
       mockLogger.On("Info", "creating user", mock.AnythingOfType("string"), req.Email).Once()
       mockRepo.On("Save", ctx, mock.AnythingOfType("*User")).Return(nil).Once()
       mockEmail.On("SendWelcomeEmail", ctx, mock.AnythingOfType("*User")).Return(emailErr).Once()
       mockLogger.On("Error", "failed to send welcome email",
           mock.AnythingOfType("string"), mock.AnythingOfType("string"),
           "error", emailErr).Once()
       mockLogger.On("Info", "user created successfully",
           mock.AnythingOfType("string"), mock.AnythingOfType("string")).Once()

       // Act
       user, err := service.CreateUser(ctx, req)

       // Assert
       assert.NoError(t, err) // Email failure should not fail user creation
       assert.NotNil(t, user)
       assert.Equal(t, req.Email, user.Email)

       // Verify all expectations
       mockRepo.AssertExpectations(t)
       mockEmail.AssertExpectations(t)
       mockLogger.AssertExpectations(t)
   }
   ```

## Examples

```go
// ❌ BAD: Tightly coupled service with concrete dependencies
type UserService struct {
    db           *sql.DB  // Concrete database dependency
    emailClient  *smtp.Client  // Concrete email client
    logger       *log.Logger  // Concrete logger
}

func NewUserService() *UserService {
    // Hardcoded database connection
    db, _ := sql.Open("postgres", "postgres://localhost/mydb")

    // Hardcoded email configuration
    emailClient, _ := smtp.Dial("smtp.gmail.com:587")

    // Hardcoded logger
    logger := log.New(os.Stdout, "[USER] ", log.LstdFlags)

    return &UserService{
        db:          db,
        emailClient: emailClient,
        logger:      logger,
    }
}

func (s *UserService) CreateUser(email, name string) error {
    // Direct database interaction
    _, err := s.db.Exec("INSERT INTO users (email, name) VALUES ($1, $2)", email, name)
    if err != nil {
        s.logger.Printf("Failed to save user: %v", err)
        return err
    }

    // Direct email sending
    err = s.emailClient.Mail("welcome@company.com")
    if err != nil {
        // No way to test this without actual email server
        return err
    }

    return nil
}

// Problems:
// - Impossible to test without real database and email server
// - Cannot swap implementations (e.g., for different environments)
// - Hidden dependencies make component behavior unpredictable
// - Changes to infrastructure require changes to business logic
```

```go
// ✅ GOOD: Loosely coupled service with interface dependencies
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    FindByEmail(ctx context.Context, email string) (*User, error)
}

type EmailService interface {
    SendWelcomeEmail(ctx context.Context, user *User) error
}

type Logger interface {
    Info(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
}

type UserService struct {
    userRepo     UserRepository
    emailService EmailService
    logger       Logger
}

func NewUserService(
    userRepo UserRepository,
    emailService EmailService,
    logger Logger,
) *UserService {
    return &UserService{
        userRepo:     userRepo,
        emailService: emailService,
        logger:       logger,
    }
}

func (s *UserService) CreateUser(ctx context.Context, email, name string) (*User, error) {
    user := &User{
        ID:    generateID(),
        Email: email,
        Name:  name,
    }

    if err := s.userRepo.Save(ctx, user); err != nil {
        s.logger.Error("failed to save user", "email", email, "error", err)
        return nil, fmt.Errorf("failed to save user: %w", err)
    }

    // Email failure doesn't fail user creation
    if err := s.emailService.SendWelcomeEmail(ctx, user); err != nil {
        s.logger.Error("failed to send welcome email", "user_id", user.ID, "error", err)
    }

    s.logger.Info("user created successfully", "user_id", user.ID)
    return user, nil
}

// Benefits:
// - Easy to test with mocks
// - Can swap implementations (in-memory for tests, postgres for production)
// - All dependencies are explicit in constructor
// - Business logic is isolated from infrastructure concerns
```

```go
// ❌ BAD: God interface that violates interface segregation
type DatabaseManager interface {
    // User operations
    SaveUser(user *User) error
    FindUser(id string) (*User, error)
    DeleteUser(id string) error

    // Order operations
    SaveOrder(order *Order) error
    FindOrder(id string) (*Order, error)
    UpdateOrderStatus(id string, status Status) error

    // Product operations
    SaveProduct(product *Product) error
    FindProduct(id string) (*Product, error)
    UpdateProductPrice(id string, price Money) error

    // Analytics operations
    GetUserStats() (*UserStats, error)
    GetOrderReport(from, to time.Time) (*OrderReport, error)

    // Administrative operations
    BackupDatabase() error
    VacuumTables() error
    CheckHealth() error
}

type OrderService struct {
    database DatabaseManager  // Depends on entire interface even though it only needs order operations
}

// Problems:
// - OrderService depends on user, product, and analytics operations it doesn't use
// - Changes to user operations can break OrderService compilation
// - Testing requires mocking the entire interface
// - Interface is hard to implement and maintain
```

```go
// ✅ GOOD: Segregated interfaces based on actual usage
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id string) (*Order, error)
    UpdateStatus(ctx context.Context, id string, status Status) error
}

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
}

type ProductRepository interface {
    FindByID(ctx context.Context, id string) (*Product, error)
    UpdatePrice(ctx context.Context, id string, price Money) error
}

type InventoryService interface {
    ReserveItems(ctx context.Context, items []OrderItem) (*Reservation, error)
    CommitReservation(ctx context.Context, reservationID string) error
}

type OrderService struct {
    orderRepo   OrderRepository   // Only depends on order operations
    userRepo    UserRepository    // Only needs user lookup
    productRepo ProductRepository // Only needs product operations
    inventory   InventoryService  // Only needs inventory operations
    logger      Logger
}

func NewOrderService(
    orderRepo OrderRepository,
    userRepo UserRepository,
    productRepo ProductRepository,
    inventory InventoryService,
    logger Logger,
) *OrderService {
    return &OrderService{
        orderRepo:   orderRepo,
        userRepo:    userRepo,
        productRepo: productRepo,
        inventory:   inventory,
        logger:      logger,
    }
}

// Benefits:
// - Each dependency interface is focused and minimal
// - Changes to user operations don't affect OrderService
// - Easy to test by mocking only required interfaces
// - Clear separation of concerns
// - Components can evolve independently
```

```go
// ❌ BAD: Factory that creates concrete dependencies, making testing difficult
func NewOrderService() *OrderService {
    // Hardcoded concrete implementations
    db := connectToDatabase("postgres://prod-db/orders")
    emailClient := &SMTPEmailClient{
        host: "smtp.company.com",
        port: 587,
    }
    paymentGateway := &StripePaymentGateway{
        apiKey: os.Getenv("STRIPE_API_KEY"),
    }
    logger := log.New(os.Stdout, "[ORDER] ", log.LstdFlags)

    return &OrderService{
        db:             db,
        emailClient:    emailClient,
        paymentGateway: paymentGateway,
        logger:         logger,
    }
}

// Problems:
// - Cannot test without real database, email server, and Stripe
// - Cannot swap implementations for different environments
// - Hidden environment dependencies
// - Hard to isolate component behavior in tests
```

```go
// ✅ GOOD: Factory that accepts configurations and creates proper abstractions
type ServiceDependencies struct {
    OrderRepo        OrderRepository
    PaymentProcessor PaymentProcessor
    EmailService     EmailService
    Logger           Logger
}

func NewOrderServiceDependencies(cfg Config) (*ServiceDependencies, error) {
    // Create database connection
    db, err := sql.Open("postgres", cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Create concrete implementations that implement the interfaces
    orderRepo := &PostgresOrderRepository{db: db}

    var paymentProcessor PaymentProcessor
    switch cfg.PaymentProvider {
    case "stripe":
        paymentProcessor = &StripePaymentProcessor{apiKey: cfg.StripeAPIKey}
    case "paypal":
        paymentProcessor = &PayPalPaymentProcessor{
            clientID:     cfg.PayPalClientID,
            clientSecret: cfg.PayPalClientSecret,
        }
    default:
        return nil, fmt.Errorf("unsupported payment provider: %s", cfg.PaymentProvider)
    }

    emailService := &SendGridEmailService{
        apiKey: cfg.SendGridAPIKey,
        client: http.DefaultClient,
    }

    logger := &StructuredLogger{
        level:  cfg.LogLevel,
        output: os.Stdout,
    }

    return &ServiceDependencies{
        OrderRepo:        orderRepo,
        PaymentProcessor: paymentProcessor,
        EmailService:     emailService,
        Logger:           logger,
    }, nil
}

func NewOrderService(deps *ServiceDependencies) *OrderService {
    return &OrderService{
        orderRepo:        deps.OrderRepo,
        paymentProcessor: deps.PaymentProcessor,
        emailService:     deps.EmailService,
        logger:           deps.Logger,
    }
}

// For testing, easily create with mocks
func NewOrderServiceForTesting(
    orderRepo OrderRepository,
    paymentProcessor PaymentProcessor,
    emailService EmailService,
    logger Logger,
) *OrderService {
    return &OrderService{
        orderRepo:        orderRepo,
        paymentProcessor: paymentProcessor,
        emailService:     emailService,
        logger:           logger,
    }
}

// Benefits:
// - Production factory creates real implementations
// - Test factory accepts mocks
// - Configuration-driven implementation selection
// - Clear separation between dependency creation and service creation
// - Easy to test and maintain
```

## Related Bindings

- [component-isolation.md](../../core/component-isolation.md): Dependency injection is the primary mechanism for implementing component isolation in Go. Interface-based injection creates the boundaries that prevent components from becoming coupled, enabling independent development and testing.

- [interface-design.md](../../docs/bindings/categories/go/interface-design.md): Go's interface design patterns provide the foundation for effective dependency injection. Well-designed interfaces create clean contracts between components, while dependency injection makes those contracts explicit and testable.

- [extract-common-logic.md](../../core/extract-common-logic.md): Dependency injection enables the extraction and reuse of common logic through shared interfaces and implementations. Components can depend on common abstractions without duplicating functionality or creating tight coupling.

- [orthogonality.md](../../tenets/orthogonality.md): This binding directly implements the orthogonality tenet by using Go's interface system to create independent, composable components. Dependency injection ensures that changes to one component don't ripple through the system, maintaining the independence that orthogonal design requires.
