package cra

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	craports "github.com/kore/kore/internal/modules/cra/ports"
	reportdomain "github.com/kore/kore/internal/modules/reporting/domain"
	reportports "github.com/kore/kore/internal/modules/reporting/ports"
	"github.com/kore/kore/pkg/kernel"
)

type PlanningReader struct {
	cra craports.CRAService
}

func NewPlanningReader(cra craports.CRAService) reportports.CRAPlanningReader {
	return &PlanningReader{cra: cra}
}

func (r *PlanningReader) ListDailyActivity(ctx context.Context, tenant kernel.TenantID, period kernel.Period) ([]reportports.PlanningActivityRow, error) {
	if r.cra == nil {
		return nil, nil
	}
	rows, err := r.cra.ListDailyActivityInPeriod(ctx, tenant, period)
	if err != nil {
		return nil, err
	}
	out := make([]reportports.PlanningActivityRow, len(rows))
	for i, row := range rows {
		out[i] = reportports.PlanningActivityRow{
			UserID:       row.UserID,
			UserPrenom:   row.UserPrenom,
			UserNom:      row.UserNom,
			Day:          row.Day,
			Minutes:      row.Minutes,
			MissionID:    row.MissionID,
			MissionLabel: row.MissionLabel,
			ClientLabel:  row.ClientLabel,
		}
	}
	return out, nil
}

func BuildPlanningView(period kernel.Period, rows []reportports.PlanningActivityRow) reportdomain.PlanningView {
	byUser := make(map[string]*reportdomain.PlanningRow)
	order := make([]string, 0)
	for _, row := range rows {
		key := row.UserID.String()
		entry, ok := byUser[key]
		if !ok {
			name := strings.TrimSpace(row.UserPrenom + " " + row.UserNom)
			if name == "" {
				name = key[:8]
			}
			entry = &reportdomain.PlanningRow{UserID: row.UserID, UserName: name}
			byUser[key] = entry
			order = append(order, key)
		}
		hours := float64(row.Minutes) / 60
		label := row.MissionLabel
		if label == "" {
			label = fmt.Sprintf("%.1fh", hours)
		} else if row.Minutes > 0 {
			label = fmt.Sprintf("%s · %.1fh", label, hours)
		}
		entry.Slots = append(entry.Slots, reportdomain.PlanningSlot{
			Date:  row.Day,
			Label: label,
			Hours: hours,
		})
	}
	sort.Strings(order)
	out := make([]reportdomain.PlanningRow, 0, len(order))
	for _, key := range order {
		out = append(out, *byUser[key])
	}
	return reportdomain.PlanningView{Period: period, Rows: out}
}

func BuildGanttView(period kernel.Period, rows []reportports.PlanningActivityRow) reportdomain.GanttView {
	type missionAgg struct {
		id           string
		start        time.Time
		end          time.Time
		minutes      int
		missionLabel string
		clientLabel  string
	}
	agg := make(map[string]*missionAgg)
	for _, row := range rows {
		if row.MissionID == "" {
			continue
		}
		item, ok := agg[row.MissionID]
		if !ok {
			item = &missionAgg{
				id:           row.MissionID,
				start:        row.Day,
				end:          row.Day,
				missionLabel: row.MissionLabel,
				clientLabel:  row.ClientLabel,
			}
			agg[row.MissionID] = item
		}
		if row.Day.Before(item.start) {
			item.start = row.Day
		}
		if row.Day.After(item.end) {
			item.end = row.Day
		}
		item.minutes += row.Minutes
		if item.missionLabel == "" && row.MissionLabel != "" {
			item.missionLabel = row.MissionLabel
		}
		if item.clientLabel == "" && row.ClientLabel != "" {
			item.clientLabel = row.ClientLabel
		}
	}
	keys := make([]string, 0, len(agg))
	for id := range agg {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	items := make([]reportdomain.GanttItem, 0, len(keys))
	for _, id := range keys {
		item := agg[id]
		days := int(item.end.Sub(item.start).Hours()/24) + 1
		if days < 1 {
			days = 1
		}
		capacityMinutes := days * 480
		progress := float64(item.minutes) / float64(capacityMinutes)
		if progress > 1 {
			progress = 1
		}
		missionID, err := uuid.Parse(id)
		if err != nil {
			missionID = uuid.NewSHA1(uuid.NameSpaceURL, []byte(id))
		}
		label := ganttLabel(item.clientLabel, item.missionLabel, id)
		items = append(items, reportdomain.GanttItem{
			ID:        missionID,
			Label:     label,
			StartDate: item.start,
			EndDate:   item.end,
			Progress:  progress,
		})
	}
	return reportdomain.GanttView{Period: period, Items: items}
}

func ganttLabel(clientLabel, missionLabel, missionID string) string {
	clientLabel = strings.TrimSpace(clientLabel)
	missionLabel = strings.TrimSpace(missionLabel)
	switch {
	case clientLabel != "" && missionLabel != "":
		return clientLabel + " · " + missionLabel
	case missionLabel != "":
		return missionLabel
	case clientLabel != "":
		return clientLabel
	}
	if len(missionID) > 8 {
		return "Mission " + missionID[:8]
	}
	return "Mission " + missionID
}

var _ reportports.CRAPlanningReader = (*PlanningReader)(nil)
