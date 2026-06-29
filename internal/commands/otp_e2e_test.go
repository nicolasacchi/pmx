package commands

import (
	"testing"
)

// otp is a composite SERVER-scope command: it fans out concurrent requests to
// recent outbound messages and to bounces for the recipient. With a respBody of
// `{"Messages":[],"Bounces":[]}` both lists are empty, so firstMessageID is empty
// and the per-message details follow-up never fires.

func TestE2EOtpFansOut(t *testing.T) {
	cr, err := runCmd(t, 200, `{"Messages":[],"Bounces":[]}`, "otp", "x@y.com")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !cr.hitPath("/messages/outbound") {
		t.Errorf("otp did not hit /messages/outbound: %v", cr.Paths)
	}
	if !cr.hitPath("/bounces") {
		t.Errorf("otp did not hit /bounces: %v", cr.Paths)
	}
	if cr.ServerToken != "srv-tok" {
		t.Errorf("server token not sent: %q", cr.ServerToken)
	}
	if cr.AccountToken != "" {
		t.Errorf("account token must not be sent on server scope: %q", cr.AccountToken)
	}
}
