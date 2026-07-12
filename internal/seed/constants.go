package seed

import "github.com/google/uuid"

const (
	TenantName = "Kore Demo"

	AdminLogin      = "ADM_admin"
	AdminPassword   = "Admin123!"
	ManagerLogin    = "MGR_manager"
	ManagerPassword = "Manager123!"
	CollabLogin     = "COL_collab"
	CollabPassword  = "Collab123!"
	CommercialLogin = "COM_commercial"
	CommercialPass  = "Commercial123!"

	DemoSocieteName    = "Kore Demo SAS"
	DemoSiteLabel      = "Paris HQ"
	DemoAppLabel       = "Portail Client ACME"
	DemoEquipeLabel    = "Équipe Dev"
	DemoClientName     = "ACME Corp"
	DemoClientTVA      = "FR12345678901"
	MarkerLogin        = CollabLogin
	TrialSeats         = 50
)

var (
	DemoTenantID = uuid.MustParse("00000000-0000-4000-8000-000000000001")
	DemoSocieteID = uuid.MustParse("00000000-0000-4000-8000-000000000010")
	DemoSiteID    = uuid.MustParse("00000000-0000-4000-8000-000000000011")
	DemoServiceID = uuid.MustParse("00000000-0000-4000-8000-000000000012")
	DemoAppID     = uuid.MustParse("00000000-0000-4000-8000-000000000013")
	DemoEquipeID  = uuid.MustParse("00000000-0000-4000-8000-000000000014")
)
