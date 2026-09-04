package services

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("crm.lead", "action_send_sms", actionSendSMS)
}

func actionSendSMS(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != "crm.lead" || id <= 0 {
		return "", fmt.Errorf("invalid lead")
	}
	lead, err := orm.SearchOne(ctx, "crm.lead", map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	phone := orm.AsString(lead["phone"])
	if phone == "" {
		return "", fmt.Errorf("lead has no phone number")
	}
	// Stub: SMS composer would open here; no message is sent.
	_ = phone
	return vals["next"], nil
}
