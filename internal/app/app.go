package app

import (
	"context"
	"embed"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	aiconges "github.com/kore/kore/internal/modules/ai/adapters/conges"
	aicra "github.com/kore/kore/internal/modules/ai/adapters/cra"
	aihttp "github.com/kore/kore/internal/modules/ai/adapters/http"
	aipostgres "github.com/kore/kore/internal/modules/ai/adapters/postgres"
	aistub "github.com/kore/kore/internal/modules/ai/adapters/stub"
	aitma "github.com/kore/kore/internal/modules/ai/adapters/tma"
	aiworkflow "github.com/kore/kore/internal/modules/ai/adapters/workflow"
	aiapp "github.com/kore/kore/internal/modules/ai/app"
	billinghttp "github.com/kore/kore/internal/modules/billing/adapters/http"
	billingpostgres "github.com/kore/kore/internal/modules/billing/adapters/postgres"
	billingapp "github.com/kore/kore/internal/modules/billing/app"
	budgetcra "github.com/kore/kore/internal/modules/budget/adapters/cra"
	budgethttp "github.com/kore/kore/internal/modules/budget/adapters/http"
	budgetpostgres "github.com/kore/kore/internal/modules/budget/adapters/postgres"
	budgetapp "github.com/kore/kore/internal/modules/budget/app"
	congescra "github.com/kore/kore/internal/modules/conges/adapters/cra"
	congeshttp "github.com/kore/kore/internal/modules/conges/adapters/http"
	congesnotif "github.com/kore/kore/internal/modules/conges/adapters/notifications"
	congesorg "github.com/kore/kore/internal/modules/conges/adapters/org"
	congespostgres "github.com/kore/kore/internal/modules/conges/adapters/postgres"
	congesworkflow "github.com/kore/kore/internal/modules/conges/adapters/workflow"
	congesapp "github.com/kore/kore/internal/modules/conges/app"
	crahttp "github.com/kore/kore/internal/modules/cra/adapters/http"
	crapdf "github.com/kore/kore/internal/modules/cra/adapters/pdf"
	crapostgres "github.com/kore/kore/internal/modules/cra/adapters/postgres"
	craapp "github.com/kore/kore/internal/modules/cra/app"
	notifhttp "github.com/kore/kore/internal/modules/notifications/adapters/http"
	notifpostgres "github.com/kore/kore/internal/modules/notifications/adapters/postgres"
	notifsmtp "github.com/kore/kore/internal/modules/notifications/adapters/smtp"
	notifapp "github.com/kore/kore/internal/modules/notifications/app"
	orghttp "github.com/kore/kore/internal/modules/org/adapters/http"
	orgpostgres "github.com/kore/kore/internal/modules/org/adapters/postgres"
	orgapp "github.com/kore/kore/internal/modules/org/app"
	publichttp "github.com/kore/kore/internal/modules/publicsite/adapters/http"
	publicnotif "github.com/kore/kore/internal/modules/publicsite/adapters/notifications"
	publicpostgres "github.com/kore/kore/internal/modules/publicsite/adapters/postgres"
	publicapp "github.com/kore/kore/internal/modules/publicsite/app"
	tmacra "github.com/kore/kore/internal/modules/tma/adapters/cra"
	tmahttp "github.com/kore/kore/internal/modules/tma/adapters/http"
	tmanotif "github.com/kore/kore/internal/modules/tma/adapters/notifications"
	tmapostgres "github.com/kore/kore/internal/modules/tma/adapters/postgres"
	tmaworkflow "github.com/kore/kore/internal/modules/tma/adapters/workflow"
	tmaapp "github.com/kore/kore/internal/modules/tma/app"
	wfhttp "github.com/kore/kore/internal/modules/workflow/adapters/http"
	wfnotif "github.com/kore/kore/internal/modules/workflow/adapters/notifications"
	wfpostgres "github.com/kore/kore/internal/modules/workflow/adapters/postgres"
	wfapp "github.com/kore/kore/internal/modules/workflow/app"
	"github.com/kore/kore/internal/platform/authx"
	"github.com/kore/kore/internal/platform/cache"
	"github.com/kore/kore/internal/platform/config"
	"github.com/kore/kore/internal/platform/db"
	"github.com/kore/kore/internal/platform/httpx"
	"github.com/kore/kore/internal/platform/logging"
	"github.com/kore/kore/internal/seed"
)

//go:embed openapi.yaml
var openAPI embed.FS

type Application struct {
	cfg        config.Config
	log        *logging.Logger
	pool       *db.Pool
	cache      cache.Cache
	redisCache *cache.RedisCache
	router     *httpx.Router
	migrator   *db.MigrationRunner
	seed       *seed.Runner
	workerStop context.CancelFunc
}

func New(ctx context.Context, cfg config.Config) (*Application, error) {
	log := logging.New(cfg.LogLevel)
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	var redisCache *cache.RedisCache
	var appCache cache.Cache
	redisCache, err = cache.NewRedisCache(cfg.RedisAddr, cfg.RedisAuth, cfg.RedisDB, cfg.RedisTLS)
	if err != nil {
		return nil, err
	}
	if err := redisCache.Ping(ctx); err != nil {
		log.Logger.Warn("redis unavailable, falling back to in-memory cache", "error", err)
		appCache = cache.NewInMemoryCache()
		redisCache = nil
	} else {
		appCache = redisCache
	}

	keyBuilder := cache.NewKeyBuilder(cfg.RedisKeyPrefix)
	tokenIssuer := authx.NewTokenIssuer(cfg.JWTSigningKey, cfg.JWTTTL, cfg.JWTRefreshTTL)

	migrator := db.NewMigrationRunner(pool, AllModuleMigrations())

	orgRepo := orgpostgres.NewRepository(pool)
	notifRepo := notifpostgres.NewRepository(pool)
	wfRepo := wfpostgres.NewRepository(pool)
	craRepo := crapostgres.NewRepository(pool)
	congesRepo := congespostgres.NewRepository(pool)
	budgetRepo := budgetpostgres.NewRepository(pool)
	tmaRepo := tmapostgres.NewRepository(pool)
	billingRepo := billingpostgres.NewRepository(pool)
	publicRepo := publicpostgres.NewRepository(pool)

	billingService := billingapp.NewService(billingRepo, cfg.BillingTrialDays)

	orgService := orgapp.NewOrganizationService(orgRepo)
	platformService := orgapp.NewPlatformService(orgRepo)
	userService := orgapp.NewUserService(orgRepo, orgapp.NewArgon2Hasher(), tokenIssuer, billingService, appCache, keyBuilder, cfg.PlatformAdminLogins)
	clientService := orgapp.NewClientService(orgRepo)

	emailSender := notifsmtp.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
	notifService := notifapp.NewService(notifRepo, emailSender, orgRepo)
	wfService := wfapp.NewService(wfRepo, appCache, keyBuilder, wfnotif.NewTransitionPublisher(notifService))
	craService := craapp.NewService(craRepo, appCache, keyBuilder).
		WithPDFRenderer(crapdf.NewTenantRenderer(orgService))
	leaveTypeConfigRepo := congespostgres.NewLeaveTypeConfigRepoAdapter(congesRepo)
	societeReader := congesorg.NewSocieteReader(orgRepo)
	leaveTypeConfigService := congesapp.NewLeaveTypeConfigService(leaveTypeConfigRepo, societeReader)
	congesService := congesapp.NewService(
		congesRepo,
		congescra.NewFeederAdapter(craService),
		congesworkflow.NewAdapter(wfService),
		congesapp.WithNotifier(congesnotif.NewPublisherAdapter(notifService)),
		congesapp.WithTypeConfigs(leaveTypeConfigService),
	)
	budgetService := budgetapp.NewService(budgetRepo, budgetcra.NewReaderAdapter(craService))
	tmaService := tmaapp.NewService(
		tmaRepo,
		tmaworkflow.NewAdapter(wfService),
		tmacra.NewFeederAdapter(craService),
		budgetRepo,
		tmaapp.WithNotifier(tmanotif.NewPublisherAdapter(notifService)),
	)
	aiRepo := aipostgres.NewRepository(pool)
	aiService := aiapp.NewService(
		aiRepo,
		aistub.NewProvider(),
		aitma.NewReaderAdapter(tmaService),
		aicra.NewReaderAdapter(craService),
		aiconges.NewReaderAdapter(congesService),
		aiworkflow.NewReaderAdapter(wfService),
	)
	publicService := publicapp.NewServiceWithCache(publicRepo, billingService, publicnotif.NewNotifierAdapter(notifService), cfg.StripePublishableKey, appCache, keyBuilder)

	authorizer := authx.NewRBACAuthorizer(orgapp.DefaultPermissions())
	deps := httpx.Dependencies{
		Logger:            log,
		Pool:              pool,
		Cache:             appCache,
		TokenIssuer:       tokenIssuer,
		EntitlementReader: billingService,
		Authorizer:        authorizer,
	}
	router := httpx.NewRouter(deps)
	pingRedis := func(r *http.Request) error {
		if redisCache == nil {
			return nil
		}
		return redisCache.Ping(r.Context())
	}
	router.MountHealth(pool, pingRedis)

	router.Get("/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		data, err := openAPI.ReadFile("openapi.yaml")
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, "openapi unavailable")
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	router.Route("/api/v1", func(r chi.Router) {
		orghttp.RegisterRoutes(r, orgService, userService, clientService, tokenIssuer, authorizer, cfg.UploadsDir, billingService, leaveTypeConfigService)
		orghttp.RegisterPlatformRoutes(r, platformService, tokenIssuer, billingService)
		notifhttp.RegisterRoutes(r, notifService, tokenIssuer, authorizer, billingService)
		wfhttp.RegisterRoutes(r, wfService, tokenIssuer, authorizer, billingService)
		crahttp.RegisterRoutes(r, craService, tokenIssuer, authorizer, billingService)
		congeshttp.RegisterRoutes(r, congesService, leaveTypeConfigService, tokenIssuer, authorizer, billingService)
		budgethttp.RegisterRoutes(r, budgetService, tokenIssuer, authorizer, billingService)
		tmahttp.RegisterRoutes(r, tmaService, tokenIssuer, authorizer, billingService)
		aihttp.RegisterRoutes(r, aiService, tokenIssuer, authorizer, billingService)
		billinghttp.RegisterRoutes(r, billingService, tokenIssuer, authorizer, cfg.StripeWebhookSecret, billingService)
		publichttp.RegisterRoutes(r, publicService, appCache, keyBuilder)
	})

	seedRunner := seed.NewRunner(seed.Dependencies{
		Pool:          pool,
		OrgRepo:       orgRepo,
		Org:           orgService,
		Users:         userService,
		Clients:       clientService,
		Billing:       billingRepo,
		Workflow:      wfService,
		CRA:           craService,
		Leaves:        congesService,
		LeaveTypes:    leaveTypeConfigService,
		Budget:        budgetService,
		TMA:           tmaService,
		Notifications: notifService,
		Public:        publicService,
		PublicSlots:   publicRepo,
	})

	app := &Application{
		cfg:        cfg,
		log:        log,
		pool:       pool,
		cache:      appCache,
		redisCache: redisCache,
		router:     router,
		migrator:   migrator,
		seed:       seedRunner,
	}
	app.startNotificationWorker(notifService)
	return app, nil
}

func (a *Application) startNotificationWorker(notifService *notifapp.Service) {
	ctx, cancel := context.WithCancel(context.Background())
	a.workerStop = cancel
	notifapp.StartWorker(ctx, notifService, a.log, 60*time.Second)
}

func (a *Application) Migrate(ctx context.Context) error {
	return a.migrator.Up(ctx)
}

func (a *Application) Seed(ctx context.Context) error {
	return a.seed.Run(ctx)
}

func (a *Application) ResetSeed(ctx context.Context) error {
	if err := a.seed.ResetDemoTenant(ctx); err != nil {
		return err
	}
	return a.seed.Run(ctx)
}

func (a *Application) Handler() http.Handler {
	return a.router
}

func (a *Application) Close() {
	if a.workerStop != nil {
		a.workerStop()
	}
	if a.redisCache != nil {
		_ = a.redisCache.Close()
	}
	a.pool.Close()
}
