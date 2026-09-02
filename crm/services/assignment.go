package services

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

func assignTeamLeads(ctx context.Context, teamID int) error {
	team, err := orm.SearchOne(ctx, "crm.team", map[string]interface{}{"id": teamID})
	if err != nil {
		return err
	}
	if !asBool(team["assignment_enabled"]) {
		return fmt.Errorf("assignment not enabled")
	}
	domain := parseAssignmentDomain(orm.AsString(team["assignment_domain"]))
	teamMax, _ := orm.CoerceInt64(team["assignment_max"])
	if teamMax <= 0 {
		teamMax = 30
	}

	leads, err := orm.Search(ctx, "crm.lead", [][]interface{}{
		{"team_id", "=", teamID},
		{"user_id", "=", false},
		{"active", "=", true},
	})
	if err != nil {
		return err
	}
	members, err := orm.Search(ctx, "crm.team.member", [][]interface{}{
		{"team_id", "=", teamID},
		{"active", "=", true},
	})
	if err != nil || len(members) == 0 {
		return fmt.Errorf("no team members")
	}

	type candidate struct {
		userID int64
		load   int
		max    int64
	}
	candidates := make([]candidate, 0, len(members))
	for _, m := range members {
		if asBool(m["assignment_optout"]) {
			continue
		}
		uid, ok := orm.CoerceInt64(m["user_id"])
		if !ok || uid <= 0 {
			continue
		}
		maxLeads, _ := orm.CoerceInt64(m["assignment_max"])
		if maxLeads <= 0 {
			maxLeads = teamMax
		}
		load := countOpenLeadsForUser(ctx, uid)
		if int64(load) >= maxLeads {
			continue
		}
		candidates = append(candidates, candidate{userID: uid, load: load, max: maxLeads})
	}
	if len(candidates) == 0 {
		return fmt.Errorf("all members at capacity")
	}

	for _, lead := range leads {
		if len(domain) > 0 && !leadMatchesDomain(lead, domain) {
			continue
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].load != candidates[j].load {
				return candidates[i].load < candidates[j].load
			}
			return candidates[i].userID < candidates[j].userID
		})
		pick := candidates[0]
		lid, ok := orm.CoerceInt64(lead["id"])
		if !ok {
			continue
		}
		if err := orm.UpdateRecordByID(ctx, "crm.lead", int(lid), map[string]interface{}{
			"user_id": pick.userID,
		}); err != nil {
			return err
		}
		pick.load++
		candidates[0] = pick
		if int64(pick.load) >= pick.max {
			candidates = candidates[1:]
			if len(candidates) == 0 {
				break
			}
		}
	}
	return nil
}

func parseAssignmentDomain(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if strings.HasSuffix(key, "_id") {
			if n, err := strconv.ParseInt(val, 10, 64); err == nil {
				out[key] = n
				continue
			}
		}
		if val == "true" || val == "false" {
			out[key] = val == "true"
			continue
		}
		if n, err := strconv.ParseFloat(val, 64); err == nil {
			out[key] = n
			continue
		}
		out[key] = val
	}
	return out
}

func leadMatchesDomain(lead map[string]interface{}, domain map[string]interface{}) bool {
	for key, expected := range domain {
		got := lead[key]
		switch exp := expected.(type) {
		case float64:
			if numericFloat(got) < exp {
				return false
			}
		case int64:
			if g, ok := orm.CoerceInt64(got); !ok || g != exp {
				return false
			}
		case bool:
			if asBool(got) != exp {
				return false
			}
		default:
			if orm.AsString(got) != orm.AsString(exp) {
				return false
			}
		}
	}
	return true
}

func countOpenLeadsForUser(ctx context.Context, userID int64) int {
	rows, err := orm.Search(ctx, "crm.lead", [][]interface{}{
		{"user_id", "=", userID},
		{"active", "=", true},
		{"won_status", "!=", "lost"},
	})
	if err != nil {
		return 0
	}
	return len(rows)
}
