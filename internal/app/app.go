package app

import (
	"context"
	"embed"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	adminhttp "github.com/kore/kore/internal/modules/admin/adapters/http"
	adminpostgres "github.com/kore/kore/internal/modules/admin/adapters/postgres"
	adminapp "github.com/kore/kore/internal/modules/admin/app"
	aiconges "github.com/kore/kore/internal/modules/ai/adapters/conges"
	aicra "github.com/kore/kore/internal/modules/ai/adapters/cra"
	aihttp "github.com/kore/kore/internal/modules/ai/adapters/http"
	aipostgres "github.com/kore/kore/internal/modules/ai/adapters/postgres"
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
	crainvoicing "github.com/kore/kore/internal/modules/cra/adapters/invoicing"
	cramaintenance "github.com/kore/kore/internal/modules/cra/adapters/maintenance"
	craorg "github.com/kore/kore/internal/modules/cra/adapters/org"
	crapdf "github.com/kore/kore/internal/modules/cra/adapters/pdf"
	crapostgres "github.com/kore/kore/internal/modules/cra/adapters/postgres"
	crassii "github.com/kore/kore/internal/modules/cra/adapters/ssii"
	crasupport "github.com/kore/kore/internal/modules/cra/adapters/support"
	cratma "github.com/kore/kore/internal/modules/cra/adapters/tma"
	craapp "github.com/kore/kore/internal/modules/cra/app"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	ettcra "github.com/kore/kore/internal/modules/ett/adapters/cra"
	etthttp "github.com/kore/kore/internal/modules/ett/adapters/http"
	ettpostgres "github.com/kore/kore/internal/modules/ett/adapters/postgres"
	ettapp "github.com/kore/kore/internal/modules/ett/app"
	integrationshttp "github.com/kore/kore/internal/modules/integrations/adapters/http"
	integrationorg "github.com/kore/kore/internal/modules/integrations/adapters/org"
	integrationspostgres "github.com/kore/kore/internal/modules/integrations/adapters/postgres"
	integrationstaiga "github.com/kore/kore/internal/modules/integrations/adapters/taiga"
	integrationstma "github.com/kore/kore/internal/modules/integrations/adapters/tma"
	integrationsapp "github.com/kore/kore/internal/modules/integrations/app"
	integrationsdomain "github.com/kore/kore/internal/modules/integrations/domain"
	integrationports "github.com/kore/kore/internal/modules/integrations/ports"
	invoicinghttp "github.com/kore/kore/internal/modules/invoicing/adapters/http"
	invoicingnotif "github.com/kore/kore/internal/modules/invoicing/adapters/notifications"
	invoicingorg "github.com/kore/kore/internal/modules/invoicing/adapters/org"
	invoicingpostgres "github.com/kore/kore/internal/modules/invoicing/adapters/postgres"
	invoicingworkflow "github.com/kore/kore/internal/modules/invoicing/adapters/workflow"
	invoicingapp "github.com/kore/kore/internal/modules/invoicing/app"
	invoicingdomain "github.com/kore/kore/internal/modules/invoicing/domain"
	maintenancecra "github.com/kore/kore/internal/modules/maintenance/adapters/cra"
	maintenancehttp "github.com/kore/kore/internal/modules/maintenance/adapters/http"
	maintenancepostgres "github.com/kore/kore/internal/modules/maintenance/adapters/postgres"
	maintenanceapp "github.com/kore/kore/internal/modules/maintenance/app"
	notifhttp "github.com/kore/kore/internal/modules/notifications/adapters/http"
	notifpostgres "github.com/kore/kore/internal/modules/notifications/adapters/postgres"
	notifsmtp "github.com/kore/kore/internal/modules/notifications/adapters/smtp"
	notifapp "github.com/kore/kore/internal/modules/notifications/app"
	notifports "github.com/kore/kore/internal/modules/notifications/ports"
	orgbilling "github.com/kore/kore/internal/modules/org/adapters/billing"
	orghttp "github.com/kore/kore/internal/modules/org/adapters/http"
	orgintegrations "github.com/kore/kore/internal/modules/org/adapters/integrations"
	orgpostgres "github.com/kore/kore/internal/modules/org/adapters/postgres"
	orgapp "github.com/kore/kore/internal/modules/org/app"
	projecthttp "github.com/kore/kore/internal/modules/project/adapters/http"
	projectorg "github.com/kore/kore/internal/modules/project/adapters/org"
	projectpostgres "github.com/kore/kore/internal/modules/project/adapters/postgres"
	projectapp "github.com/kore/kore/internal/modules/project/app"
	publichttp "github.com/kore/kore/internal/modules/publicsite/adapters/http"
	publicnotif "github.com/kore/kore/internal/modules/publicsite/adapters/notifications"
	publicpostgres "github.com/kore/kore/internal/modules/publicsite/adapters/postgres"
	publicapp "github.com/kore/kore/internal/modules/publicsite/app"
	reportingbudget "github.com/kore/kore/internal/modules/reporting/adapters/budget"
	reportingconges "github.com/kore/kore/internal/modules/reporting/adapters/conges"
	reportingcra "github.com/kore/kore/internal/modules/reporting/adapters/cra"
	reportinghttp "github.com/kore/kore/internal/modules/reporting/adapters/http"
	reportinginvoicing "github.com/kore/kore/internal/modules/reporting/adapters/invoicing"
	reportingorg "github.com/kore/kore/internal/modules/reporting/adapters/org"
	reportingpostgres "github.com/kore/kore/internal/modules/reporting/adapters/postgres"
	reportingtma "github.com/kore/kore/internal/modules/reporting/adapters/tma"
	reportingapp "github.com/kore/kore/internal/modules/reporting/app"
	reportingports "github.com/kore/kore/internal/modules/reporting/ports"
	ssiicalendar "github.com/kore/kore/internal/modules/ssii/adapters/calendar"
	ssiicra "github.com/kore/kore/internal/modules/ssii/adapters/cra"
	ssiihttp "github.com/kore/kore/internal/modules/ssii/adapters/http"
	ssiiorg "github.com/kore/kore/internal/modules/ssii/adapters/org"
	ssiipostgres "github.com/kore/kore/internal/modules/ssii/adapters/postgres"
	ssiiapp "github.com/kore/kore/internal/modules/ssii/app"
	supportcra "github.com/kore/kore/internal/modules/support/adapters/cra"
	supporthttp "github.com/kore/kore/internal/modules/support/adapters/http"
	supportpostgres "github.com/kore/kore/internal/modules/support/adapters/postgres"
	supportapp "github.com/kore/kore/internal/modules/support/app"
	tmacra "github.com/kore/kore/internal/modules/tma/adapters/cra"
	tmahttp "github.com/kore/kore/internal/modules/tma/adapters/http"
	tmanotif "github.com/kore/kore/internal/modules/tma/adapters/notifications"
	tmapostgres "github.com/kore/kore/internal/modules/tma/adapters/postgres"
	tmaproject "github.com/kore/kore/internal/modules/tma/adapters/project"
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

type tenantAccessEmailAdapter struct {
	notifier notifports.TransactionalNotifier
}

func (a tenantAccessEmailAdapter) SendTenantAccessEmail(ctx context.Context, to string, subject string, body string) error {
	return a.notifier.NotifyTransactional(ctx, notifports.TransactionalMessage{
		Recipients:    []string{to},
		Subject:       subject,
		Body:          body,
		SkipSignature: true,
	})
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
	attachmentRepo := orgpostgres.NewAttachmentRepository(pool)
	notifRepo := notifpostgres.NewRepository(pool)
	wfRepo := wfpostgres.NewRepository(pool)
	craRepo := crapostgres.NewRepository(pool)
	congesRepo := congespostgres.NewRepository(pool)
	budgetRepo := budgetpostgres.NewRepository(pool)
	tmaRepo := tmapostgres.NewRepository(pool)
	billingRepo := billingpostgres.NewRepository(pool)
	publicRepo := publicpostgres.NewRepository(pool)
	integrationsRepo := integrationspostgres.NewRepository(pool)
	invoicingRepo := invoicingpostgres.NewRepository(pool)
	adminRepo := adminpostgres.NewRepository(pool)
	reportingRepo := reportingpostgres.NewRepository(pool)
	ssiiRepo := ssiipostgres.NewRepository(pool)
	ettRepo := ettpostgres.NewRepository(pool)
	supportRepo := supportpostgres.NewRepository(pool)
	maintenanceRepo := maintenancepostgres.NewRepository(pool)

	billingService := billingapp.NewService(billingRepo, cfg.StripeSecretKey, cfg.StripeWebhookSecret, cfg.BillingTrialDays)

	totpKey, err := orgapp.NewTotpEncryptionKey(cfg.TOTPEncryptionKey, cfg.JWTSigningKey, cfg.DevSeedEnabled)
	if err != nil {
		return nil, err
	}

	orgService := orgapp.NewOrganizationService(orgRepo, orgintegrations.NewTaigaLinkReader(integrationsRepo))
	attachmentChecker := NewAttachmentResourceChecker(tmaRepo, supportRepo, maintenanceRepo)
	attachmentService := orgapp.NewAttachmentService(attachmentRepo, attachmentChecker)
	userService := orgapp.NewUserService(orgRepo, orgapp.NewArgon2Hasher(), tokenIssuer, billingService, appCache, keyBuilder, cfg.PlatformAdminLogins, totpKey)
	clientService := orgapp.NewClientService(orgRepo, orgapp.WithMissionContactCleaner(ssiiorg.MissionContactCleaner{Repo: ssiiRepo}))

	emailSender := notifsmtp.NewSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)
	pushSender := notifapp.NewPushSender(cfg)
	notifService := notifapp.NewService(notifRepo, emailSender, orgRepo,
		notifapp.WithPush(notifRepo, pushSender, cfg.PushEnabled),
	)
	deviceService := notifapp.NewDeviceService(notifRepo)
	tenantAccessService := orgapp.NewTenantAccessService(orgRepo, tenantAccessEmailAdapter{notifier: notifService})
	wfEffects := wfnotif.NewEffectsExecutor(orgRepo, notifService)
	wfService := wfapp.NewService(wfRepo, appCache, keyBuilder, wfnotif.NewTransitionPublisher(notifService), wfEffects)
	craService := craapp.NewService(craRepo, appCache, keyBuilder).
		WithPDFRenderer(crapdf.NewChromedpRenderer(crapdf.NewTenantRenderer(orgService,
			crapdf.WithUserIdentityResolver(craorg.NewUserIdentityResolver(orgRepo)),
			crapdf.WithWorkRefLabelReaders(map[string]craports.WorkRefLabelReader{
				"tma":          cratma.NewWorkRefLabelReader(tmaRepo),
				"ticket":       crasupport.NewWorkRefLabelReader(supportRepo),
				"work_request": cramaintenance.NewWorkRefLabelReader(maintenanceRepo),
			}),
		))).
		WithCalendarReader(craorg.NewSocieteReader(orgRepo)).
		WithRejectNotifier(notifService, craorg.NewEmailResolver(orgRepo)).
		WithMissionRateReader(crassii.NewMissionRateReader(ssiiRepo)).
		WithApplicationBillingReader(craorg.NewApplicationBillingReader(orgRepo)).
		WithETTRecordReader(ettcra.NewRecordReader(ettRepo))
	requestSettingsService := orgapp.NewRequestSettingsService(orgRepo)
	invoicingService := invoicingapp.NewService(
		invoicingRepo,
		invoicingapp.WithPDPGateway(invoicingapp.NewPDPGateway(cfg)),
		invoicingapp.WithMissionReader(ssiicra.NewMissionReader(ssiiRepo, craService)),
		invoicingapp.WithEnabledReader(requestSettingsService),
		invoicingapp.WithClientContactReader(invoicingorg.NewClientContactReader(clientService)),
		invoicingapp.WithMailSender(invoicingnotif.NewMailSender(notifService)),
		invoicingapp.WithWorkflow(invoicingworkflow.NewAdapter(wfService)),
	)
	craService.WithInvoicePublisher(crainvoicing.NewDraftPublisher(invoicingService))
	leaveTypeConfigRepo := congespostgres.NewLeaveTypeConfigRepoAdapter(congesRepo)
	societeReader := congesorg.NewSocieteReader(orgRepo)
	leaveTypeConfigService := congesapp.NewLeaveTypeConfigService(leaveTypeConfigRepo, societeReader)
	platformService := orgapp.NewPlatformServiceFull(
		orgRepo,
		orgRepo,
		orgapp.NewArgon2Hasher(),
		orgbilling.TrialAdapter{Ensure: billingRepo.EnsureTrial},
		leaveTypeConfigService,
		cfg.GeminiModel,
	)
	congesService := congesapp.NewService(
		congesRepo,
		congescra.NewFeederAdapter(craService),
		congesworkflow.NewAdapter(wfService),
		congesapp.WithNotifier(congesnotif.NewPublisherAdapter(notifService)),
		congesapp.WithTypeConfigs(leaveTypeConfigService),
	)
	budgetService := budgetapp.NewService(budgetRepo, budgetcra.NewReaderAdapter(craService), budgetapp.WithCache(appCache, keyBuilder))
	tmaService := tmaapp.NewService(
		tmaRepo,
		tmaworkflow.NewAdapter(wfService),
		tmacra.NewFeederAdapter(craService),
		tmaapp.WithNotifier(tmanotif.NewPublisherAdapter(notifService)),
		tmaapp.WithAgileValidator(tmaproject.NewArtifactValidator(pool)),
		tmaapp.WithWipChecker(tmaproject.NewWipChecker(pool)),
	)
	projectRepo := projectpostgres.NewRepository(pool)
	projectService := projectapp.NewService(projectRepo, projectorg.NewApplicationReader(orgService))
	aiRepo := aipostgres.NewRepository(pool)
	aiService := aiapp.NewService(
		aiRepo,
		aiapp.NewLLMProvider(cfg, platformService),
		aitma.NewReaderAdapter(tmaService),
		aicra.NewReaderAdapter(craService),
		aiconges.NewReaderAdapter(congesService),
		aiworkflow.NewReaderAdapter(wfService),
	)
	publicService := publicapp.NewServiceWithCache(publicRepo, billingService, publicnotif.NewNotifierAdapter(notifService), cfg.StripePublishableKey, appCache, keyBuilder)

	integrationOpts := []integrationsapp.ServiceOption{
		integrationsapp.WithWebhookDispatcher(integrationsapp.NewWebhookDispatcher(integrationsRepo)),
	}
	if integrationsapp.PennylaneEnabled(cfg) {
		integrationOpts = append(integrationOpts, integrationsapp.WithPennylaneClient(integrationsapp.NewPennylaneClient(cfg)))
	}
	integrationsService := integrationsapp.NewService(integrationsRepo, integrationOpts...)
	integrationsKeyService := integrationsapp.NewApiKeyService(integrationsRepo)
	var taigaGateway integrationports.TaigaGateway
	if integrationstaiga.Enabled(integrationstaiga.Config{
		BaseURL:  cfg.TaigaBaseURL,
		Username: cfg.TaigaServiceUsername,
		Password: cfg.TaigaServicePassword,
	}) {
		taigaGateway = integrationstaiga.NewClient(integrationstaiga.Config{
			BaseURL:  cfg.TaigaBaseURL,
			Username: cfg.TaigaServiceUsername,
			Password: cfg.TaigaServicePassword,
		})
	}
	taigaAppCreator := integrationorg.NewApplicationCreator(orgService)
	taigaIntegrationService := integrationsapp.NewTaigaService(integrationsRepo, integrationsapp.TaigaConfig{
		BaseURL:     cfg.TaigaBaseURL,
		ProjectSlug: cfg.TaigaProjectSlug,
	}, integrationstma.NewDemandGate(tmaService), taigaGateway, taigaAppCreator)
	adminService := adminapp.NewService(adminRepo)
	reportingLeaveReader := reportingconges.NewLeaveReader(congesService)
	reportingService := reportingapp.NewService(
		reportingRepo,
		reportingcra.NewBillableReader(craService),
		reportingcra.NewPlanningReader(craService),
		reportinginvoicing.NewBillingReader(invoicingRepo),
		reportingLeaveReader,
		reportingtma.NewDemandReader(tmaService),
		reportingapp.WithHomeReaders(
			reportingorg.NewHomeUserReader(userService),
			reportingcra.NewHomeReader(craService),
			reportingLeaveReader,
			reportingbudget.NewHomeReader(budgetService, orgService),
		),
		reportingapp.WithCache(appCache, keyBuilder),
	)
	ssiiService := ssiiapp.NewService(
		ssiiRepo,
		ssiicra.NewFeederAdapter(craService),
		ssiicra.NewCleanerAdapter(craService),
		ssiicalendar.NewGateway(congesRepo),
	)
	ettService := ettapp.NewService(ettRepo, craService, craService, orgRepo)
	supportService := supportapp.NewService(supportRepo, supportcra.NewFeederAdapter(craService), nil)
	maintenanceService := maintenanceapp.NewService(maintenanceRepo, maintenancecra.NewFeederAdapter(craService))

	taigaProjectMapping, err := integrationsapp.ParseTaigaKoreMapping(cfg.TaigaKoreMapping)
	if err != nil {
		return nil, err
	}

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
		oidcService := orgapp.NewOIDCService(orgRepo, tokenIssuer, billingService, orgapp.NewArgon2Hasher(), appCache, keyBuilder)
		idpService := orgapp.NewIdentityProviderService(orgRepo)
		orghttp.RegisterRoutes(r, orgService, userService, clientService, tenantAccessService, tokenIssuer, authorizer, cfg.UploadsDir, attachmentService, billingService, leaveTypeConfigService, requestSettingsService, taigaIntegrationService)
		orghttp.RegisterOIDCRoutes(r, oidcService, idpService, authorizer)
		orghttp.RegisterPlatformRoutes(r, platformService, tokenIssuer, billingService)
		orghttp.RegisterPublicSignupRoutes(r, platformService, appCache, keyBuilder, cfg.PublicSignupEnabled)
		notifhttp.RegisterRoutes(r, notifService, deviceService, tokenIssuer, authorizer, billingService)
		wfhttp.RegisterRoutes(r, wfService, tokenIssuer, authorizer, billingService)
		crahttp.RegisterRoutes(r, craService, tokenIssuer, authorizer, billingService, requestSettingsService)
		congeshttp.RegisterRoutes(r, congesService, leaveTypeConfigService, tokenIssuer, authorizer, billingService)
		budgethttp.RegisterRoutes(r, budgetService, tokenIssuer, authorizer, billingService)
		tmahttp.RegisterRoutes(r, tmaService, tokenIssuer, authorizer, billingService, requestSettingsService)
		projecthttp.RegisterRoutes(r, projectService, tokenIssuer, authorizer, billingService)
		aihttp.RegisterRoutes(r, aiService, tokenIssuer, authorizer, billingService)
		billinghttp.RegisterRoutes(r, billingService, tokenIssuer, authorizer, cfg.StripeWebhookSecret, billingService)
		integrationshttp.RegisterRoutes(r, integrationsService, integrationsKeyService, tokenIssuer, authorizer, billingService)
		integrationshttp.RegisterTaigaRoutes(r, taigaIntegrationService, tokenIssuer, authorizer, billingService, appCache, keyBuilder, cfg.TaigaWebhookSecret, cfg.TaigaDefaultTenantID, taigaProjectMapping)
		invoicinghttp.RegisterRoutes(r, invoicingService, tokenIssuer, authorizer, billingService, requestSettingsService, cfg.PDPWebhookSecret, cfg.PublicBaseURL, appCache, keyBuilder)
		adminhttp.RegisterRoutes(r, adminService, tokenIssuer, authorizer, billingService)
		reportinghttp.RegisterRoutes(r, reportingService, tokenIssuer, authorizer, billingService)
		ssiihttp.RegisterRoutes(r, ssiiService, tokenIssuer, authorizer, billingService)
		etthttp.RegisterRoutes(r, ettService, tokenIssuer, authorizer, billingService)
		supporthttp.RegisterRoutes(r, supportService, tokenIssuer, authorizer, billingService, requestSettingsService)
		maintenancehttp.RegisterRoutes(r, maintenanceService, tokenIssuer, authorizer, billingService, requestSettingsService)
		publichttp.RegisterRoutes(r, publicService, appCache, keyBuilder)

		apiKeyLookup := httpx.NewApiKeyLookup(
			integrationsRepo.GetApiKeyByHash,
			func(ctx context.Context, key integrationsdomain.ApiKey) error {
				now := time.Now().UTC()
				key.LastUsedAt = &now
				return integrationsRepo.SaveApiKey(ctx, key)
			},
		)
		r.Route("/open", func(pr chi.Router) {
			pr.Use(httpx.PublicAPIStack(apiKeyLookup, appCache, keyBuilder))
			pr.Get("/invoices", func(w http.ResponseWriter, req *http.Request) {
				identity, _ := authx.FromContext(req.Context())
				items, err := invoicingService.List(req.Context(), identity.TenantID)
				if err != nil {
					if errors.Is(err, invoicingdomain.ErrInvoicingDisabled) {
						httpx.WriteError(w, http.StatusForbidden, httpx.ErrCodeForbidden, err.Error())
						return
					}
					httpx.WriteError(w, http.StatusInternalServerError, httpx.ErrCodeInternal, err.Error())
					return
				}
				httpx.WriteData(w, http.StatusOK, items)
			})
		})
	})

	seedRunner := seed.NewRunner(seed.Dependencies{
		Pool:          pool,
		OrgRepo:       orgRepo,
		Org:           orgService,
		Users:         userService,
		Clients:       clientService,
		Billing:       billingRepo,
		Workflow:      wfService,
		Cache:         appCache,
		Keys:          keyBuilder,
		CRA:           craService,
		Leaves:        congesService,
		LeaveTypes:    leaveTypeConfigService,
		Budget:        budgetService,
		TMA:           tmaService,
		Project:       projectService,
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
	app.startBackgroundWorkers(notifService, craService, orgRepo, reportingService, reportingRepo, appCache, keyBuilder, log)
	return app, nil
}

func (a *Application) startBackgroundWorkers(
	notifService *notifapp.Service,
	craService *craapp.Service,
	orgRepo *orgpostgres.Repository,
	reportingService reportingports.ReportingService,
	reportingRepo *reportingpostgres.Repository,
	appCache cache.Cache,
	keyBuilder cache.KeyBuilder,
	log *logging.Logger,
) {
	ctx, cancel := context.WithCancel(context.Background())
	a.workerStop = cancel
	notifapp.StartWorker(ctx, notifService, log, 60*time.Second)
	craapp.StartReminderWorker(
		ctx,
		craapp.NewReminderWorker(craService, orgRepo, notifService, appCache, keyBuilder, log),
		craapp.ReminderWorkerInterval,
	)
	reportingapp.StartSnapshotWorker(
		ctx,
		reportingapp.NewSnapshotWorker(reportingService, reportingRepo, appCache, keyBuilder, log),
		reportingapp.SnapshotWorkerInterval,
	)
}

func (a *Application) Migrate(ctx context.Context) error {
	return a.migrator.Up(ctx)
}

func (a *Application) Seed(ctx context.Context) error {
	return a.seed.Run(ctx)
}

func (a *Application) BootstrapLLIT(ctx context.Context) error {
	return a.seed.BootstrapLLIT(ctx)
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
