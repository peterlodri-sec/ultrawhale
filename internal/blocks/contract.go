package blocks

import (
	"fmt"
	"sync"
)

// ── Contract Primitive — Formal Verification ──────────────────────────
// Vaked layer: Declares. Validates that a block satisfies a formal spec.
// Pre-condition → Operation → Post-condition → Invariant.

// Contract is a formal specification for a block operation.
type Contract struct {
	Name       string   // "WriteMustBeJournaled", "SedMustPreserveRef"
	PreCond    string   // what must be true before
	Operation  string   // the operation being verified
	PostCond   string   // what must be true after
	Invariant  string   // what must always be true
	Verified   bool
	Violations int64
}

// ContractRegistry holds all formal contracts.
type ContractRegistry struct {
	mu        sync.Mutex
	contracts map[string]*Contract
}

var contractRegistry = &ContractRegistry{contracts: make(map[string]*Contract)}

// DefineContract registers a formal contract.
func DefineContract(name, pre, op, post, invariant string) *Contract {
	c := &Contract{Name: name, PreCond: pre, Operation: op, PostCond: post, Invariant: invariant}
	contractRegistry.mu.Lock()
	contractRegistry.contracts[name] = c
	contractRegistry.mu.Unlock()
	Log(LogInfo, "contract.define", name, "", "", 0, nil)
	return c
}

// VerifyContract checks if a contract holds after an operation.
func VerifyContract(name string, preOk, postOk, invariantOk bool) {
	contractRegistry.mu.Lock()
	defer contractRegistry.mu.Unlock()
	if c, ok := contractRegistry.contracts[name]; ok {
		c.Verified = preOk && postOk && invariantOk
		if !c.Verified { c.Violations++ }
	}
}

// ContractStatus returns compact contract status.
func ContractStatus() string {
	contractRegistry.mu.Lock()
	defer contractRegistry.mu.Unlock()
	verified := 0
	for _, c := range contractRegistry.contracts {
		if c.Verified { verified++ }
	}
	return fmt.Sprintf("contract: %d defined (%d verified, %d violations)",
		len(contractRegistry.contracts), verified, totalViolations(contractRegistry.contracts))
}

func totalViolations(contracts map[string]*Contract) int64 {
	var total int64
	for _, c := range contracts { total += c.Violations }
	return total
}

// Built-in contracts
func init() {
	DefineContract("WriteMustBeJournaled",
		"file exists", "Write()", "journal depth increased by 1",
		"journal depth >= 0")

	DefineContract("SedMustPreserveSize",
		"file size S", "SedAll()", "file size >= S - len(find)*count",
		"content integrity preserved")

	DefineContract("SpaceEdgeMustBeGated",
		"two nodes exist", "ConnectNodes()", "edge only if context permits",
		"capability graph integrity")
}
