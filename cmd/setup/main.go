package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("\n⚡ ultrawhale setup — interactive infra configuration")
	fmt.Println(strings.Repeat("─", 52))
	reader := bufio.NewReader(os.Stdin)
	config := make(map[string]string)

	fmt.Print("\n[Cloudflare] API Token (enter to skip): ")
	cfToken, _ := reader.ReadString('\n')
	cfToken = strings.TrimSpace(cfToken)
	if cfToken != "" {
		config["CF_API_TOKEN"] = cfToken
		fmt.Print("[Cloudflare] Account ID: ")
		acctID, _ := reader.ReadString('\n')
		config["CF_ACCOUNT_ID"] = strings.TrimSpace(acctID)
		fmt.Println("  ✅ Cloudflare configured")
	} else { fmt.Println("  ⏭  Skipped") }

	fmt.Print("\n[Langfuse] Public Key (enter to skip): ")
	lfPub, _ := reader.ReadString('\n')
	if lfPub = strings.TrimSpace(lfPub); lfPub != "" {
		config["LANGFUSE_PUBLIC_KEY"] = lfPub
		fmt.Print("[Langfuse] Secret Key: ")
		lfSec, _ := reader.ReadString('\n')
		config["LANGFUSE_SECRET_KEY"] = strings.TrimSpace(lfSec)
		fmt.Print("[Langfuse] Host [https://langfuse.crabcc.app]: ")
		lfHost, _ := reader.ReadString('\n')
		if lfHost = strings.TrimSpace(lfHost); lfHost == "" { lfHost = "https://langfuse.crabcc.app" }
		config["LANGFUSE_HOST"] = lfHost
		fmt.Println("  ✅ Langfuse configured")
	} else { fmt.Println("  ⏭  Skipped") }

	fmt.Print("\n[NATS] URL [nats://crabcc-nats:4222]: ")
	natsURL, _ := reader.ReadString('\n')
	if natsURL = strings.TrimSpace(natsURL); natsURL == "" { natsURL = "nats://crabcc-nats:4222" }
	config["NATS_URL"] = natsURL
	fmt.Println("  ✅ NATS configured")

	fmt.Print("\n[bao] Token (enter to skip): ")
	baoTok, _ := reader.ReadString('\n')
	if baoTok = strings.TrimSpace(baoTok); baoTok != "" {
		config["VAULT_TOKEN"] = baoTok
		fmt.Println("  ✅ bao configured")
	} else { fmt.Println("  ⏭  Skipped") }

	fmt.Print("\n[Supabase] URL [http://localhost:8586]: ")
	supaURL, _ := reader.ReadString('\n')
	if supaURL = strings.TrimSpace(supaURL); supaURL == "" { supaURL = "http://localhost:8586" }
	config["SUPABASE_URL"] = supaURL
	fmt.Println("  ✅ Supabase configured")

	fmt.Println("\n" + strings.Repeat("─", 52))
	fmt.Print("Write config to ~/.whale/ultrawhale.toml? [Y/n]: ")
	confirm, _ := reader.ReadString('\n')
	if confirm = strings.TrimSpace(strings.ToLower(confirm)); confirm == "" || confirm == "y" || confirm == "yes" {
		home, _ := os.UserHomeDir()
		os.MkdirAll(home+"/.whale", 0o700)
		f, _ := os.Create(home + "/.whale/ultrawhale.toml")
		defer f.Close()
		fmt.Fprintln(f, "# ultrawhale config")
		for k, v := range config { fmt.Fprintf(f, "%s=%s\n", k, v) }
		fmt.Println("\n✅ Config written to ~/.whale/ultrawhale.toml")
	}
}
