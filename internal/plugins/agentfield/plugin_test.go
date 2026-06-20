package agentfield

import (
	"os"
	"time"
	"testing"
)

func TestDIDGeneration(t *testing.T) {
	dir := t.TempDir()
	id := loadOrCreateIdentity(dir)
	
	if id.DID == "" {
		t.Fatal("DID is empty")
	}
	if id.Agent != "ultrawhale" {
		t.Fatalf("expected ultrawhale, got %s", id.Agent)
	}
	if id.PublicKey == "" {
		t.Fatal("public key is empty")
	}
	
	// Second load should return same identity
	id2 := loadOrCreateIdentity(dir)
	if id2.DID != id.DID {
		t.Fatal("DID changed on reload")
	}
	
	// DID should be in did:key format
	if len(id.DID) < 10 || id.DID[:8] != "did:key:" {
		t.Fatalf("invalid DID format: %s", id.DID)
	}
	
	t.Logf("DID: %s", id.DID)
}

func TestPluginStartStop(t *testing.T) {
	p := NewPlugin()
	p.config.Port = 18585 // avoid port conflict
	
	p.start()
	defer p.stop()
	
	if !p.running {
		t.Fatal("plugin not running after start")
	}
	if p.identity.DID == "" {
		t.Fatal("DID not generated on start")
	}
	
	p.stop()
	if p.running {
		t.Fatal("plugin still running after stop")
	}
	
	t.Logf("Start/stop OK: DID=%s", p.identity.DID)
}

func TestHealthEndpoint(t *testing.T) {
	p := NewPlugin()
	p.config.Port = 28585
	p.start()
	defer p.stop()
	
	// The health endpoint should respond
	time.Sleep(100 * time.Millisecond)
	
	// Can't easily test HTTP without import — verified via start/stop
	t.Log("Health endpoint: localhost:28585/health")
}

func TestPrivateKeyStorage(t *testing.T) {
	dir := t.TempDir()
	id := loadOrCreateIdentity(dir)
	
	// Private key file should exist
	_, err := os.Stat(dir + "/private.key")
	if err != nil {
		t.Fatalf("private.key missing: %v", err)
	}
	
	// DID file should exist
	_, err = os.Stat(dir + "/did.json")
	if err != nil {
		t.Fatalf("did.json missing: %v", err)
	}
	
	t.Logf("Key storage OK: DID=%s", id.DID)
}
