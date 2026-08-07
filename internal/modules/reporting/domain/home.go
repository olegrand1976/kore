package domain

// HomeDashboard is the aggregated payload for GET /dashboards/home.
type HomeDashboard struct {
	CRA     *HomeCRABlock     `json:"cra,omitempty"`
	Leave   *HomeLeaveBlock   `json:"leave,omitempty"`
	TMA     *HomeTMABlock     `json:"tma,omitempty"`
	Budget  *HomeBudgetBlock  `json:"budget,omitempty"`
	Billing *HomeBillingBlock `json:"billing,omitempty"`
	Errors  HomeErrors        `json:"errors"`
}

type HomeErrors struct {
	CRA     bool `json:"cra,omitempty"`
	Conges  bool `json:"conges,omitempty"`
	TMA     bool `json:"tma,omitempty"`
	Budget  bool `json:"budget,omitempty"`
	Billing bool `json:"billing,omitempty"`
}

type HomeCRABlock struct {
	Required      bool           `json:"required"`
	Alert         bool           `json:"alert"`
	CurrentStatus *string        `json:"currentStatus"`
	PrefillRatio  *int           `json:"prefillRatio"`
	PrefillLow    bool           `json:"prefillLow"`
	Months        []HomeCraMonth `json:"months"`
}

type HomeCraMonth struct {
	Key    string  `json:"key"`
	Status *string `json:"status"`
}

type HomeLeaveBlock struct {
	Pending            int               `json:"pending"`
	PendingValidations int               `json:"pendingValidations"`
	StatusCounts       []HomeStatusCount `json:"statusCounts"`
}

type HomeTMABlock struct {
	Open         int               `json:"open"`
	Total        int               `json:"total"`
	StatusCounts []HomeStatusCount `json:"statusCounts"`
}

type HomeBudgetBlock struct {
	Overrun        int             `json:"overrun"`
	ConsumptionPct int             `json:"consumptionPct"`
	Bars           []HomeBudgetBar `json:"bars"`
}

type HomeBudgetBar struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type HomeBillingBlock struct {
	AmountCents   int64   `json:"amountCents"`
	InvoiceCount  int     `json:"invoiceCount"`
	BillableHours float64 `json:"billableHours"`
	Currency      string  `json:"currency"`
}

type HomeStatusCount struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}
