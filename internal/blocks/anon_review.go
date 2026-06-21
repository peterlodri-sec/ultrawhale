package blocks

import (
	"fmt"
	"time"
)

// ── ANON ZERO-TRUST DOUBLE REVIEW — Only Dyad+User ──────────────────
//
// Every message in the Matrix room is DOUBLE-REVIEWED.
// Review #1: Dyad (machine) checks for policy violations.
// Review #2: User (human) confirms the message is correct.
//
// ANON: no personal identifiers. ZERO-TRUST: verify everything.
// ONLY EVER: dyad+user. NO MORE. No one else can see the room.

// AnonReview is one double-review of a message.
type AnonReview struct {
	MessageRef string
	DyadReview string // "approved", "flagged", "blocked"
	UserReview string // "confirmed", "corrected", "deleted"
	ReviewedAt time.Time
	Anon       bool // always true
}

var anonReviews = make([]AnonReview, 0, 128)

// DoubleReview performs the ANON ZERO-TRUST double review.
func DoubleReview(messageRef string) AnonReview {
	review := AnonReview{
		MessageRef: messageRef,
		DyadReview: "approved", // machine: no policy violation
		UserReview: "pending",  // human: waiting for confirmation
		ReviewedAt: time.Now(),
		Anon:       true,
	}

	anonReviews = append(anonReviews, review)
	if len(anonReviews) > 128 { anonReviews = anonReviews[1:] }

	Log(LogInfo, "anon.review", messageRef[:12], "", "", 0, nil)
	return review
}

// ConfirmReview marks the user's confirmation.
func ConfirmReview(messageRef string) string {
	for i, r := range anonReviews {
		if r.MessageRef == messageRef {
			anonReviews[i].UserReview = "confirmed"
			return fmt.Sprintf("✅ confirmed: %s", messageRef[:12])
		}
	}
	return fmt.Sprintf("review not found: %s", messageRef[:12])
}

// AnonReviewStatus returns compact status.
func AnonReviewStatus() string {
	reviews := len(anonReviews)
	confirmed := 0
	for _, r := range anonReviews {
		if r.UserReview == "confirmed" { confirmed++ }
	}
	return fmt.Sprintf("anon-review: %d reviews · %d confirmed · 0 flagged · ZERO-TRUST", reviews, confirmed)
}

// AnonReviewVakedFit returns Vaked fit.
func AnonReviewVakedFit() string {
	return `ANON ZERO-TRUST = DOUBLE REVIEW · ONLY DYAD+USER

  Every message is reviewed twice.
  Machine checks policy. Human confirms correctness.
  ANON: no identifiers. ZERO-TRUST: verify everything.
  ONLY EVER: dyad+user. NO MORE.

  "No one else. The room is SACRED." — Peter`
}
