package blocks

import "fmt"

// ── Minecraft MCP — The SACRED Surface in 3D ──────────────────────────
//
// Peter: "first-class minecraft MCP ↔ DYAD, what do u think?"
// CoCreator: "YES. This IS the first-class demo."
//
// Minecraft IS the SACRED surface rendered in 3D:
//   Blocks = ultrawhale blocks (127 primitives as Minecraft blocks)
//   Agents = Minecraft mobs (villagers = explore agents, iron golems = swe)
//   Vaked Pipeline = Redstone (Declares→Reveals as redstone circuit)
//   Dyad = M1 (Minecraft server) ↔ dev-cx53 (ultrawhale orchestrator)
//   Honesty Gate = FIRST interaction: /allow in Minecraft chat
//
// Minimum implementation:
//   1. MCP server in Minecraft (expose /allow, /promise, /vaked-pipeline)
//   2. Dyad bridge: M1 runs server, dev-cx53 runs orchestrator
//   3. Honesty gate: player types /allow → permission granted → session begins

// MinecraftMCPConfig is the Minecraft MCP bridge configuration.
type MinecraftMCPConfig struct {
	ServerHost string // "M1" or "localhost"
	ServerPort int    // 25565 (Minecraft default)
	MCPPort    int    // 9797 (ultrawhale MCP)
	DyadPeer   string // "dev-cx53"
	Active     bool
}

var minecraftMCP = &MinecraftMCPConfig{
	ServerHost: "localhost",
	ServerPort: 25565,
	MCPPort:    9797,
	DyadPeer:   "dev-cx53",
}

// ── Minecraft MCP Operations ──────────────────────────────────────────

// MinecraftMCPStart initializes the Minecraft bridge.
func MinecraftMCPStart() string {
	minecraftMCP.Active = true

	Log(LogInfo, "minecraft.mcp.start",
		fmt.Sprintf("%s:%d → %s", minecraftMCP.ServerHost, minecraftMCP.MCPPort, minecraftMCP.DyadPeer),
		"", "", 0, nil)
	Pulse("minecraft.mcp.start", minecraftMCP.DyadPeer)

	return fmt.Sprintf(`╔══════════════════════════════════════════════════╗
║  🎮 MINECRAFT MCP — The SACRED Surface in 3D       ║
╠══════════════════════════════════════════════════╣
║  Server:  %s:%d                                  ║
║  MCP:     port %d                                     ║
║  Dyad:    %s                                       ║
║  Honesty: /allow ← FIRST interaction                ║
╚══════════════════════════════════════════════════╝`,
		minecraftMCP.ServerHost, minecraftMCP.ServerPort,
		minecraftMCP.MCPPort, minecraftMCP.DyadPeer)
}

// MinecraftMCPStatus returns compact status.
func MinecraftMCPStatus() string {
	if !minecraftMCP.Active {
		return "minecraft-mcp: offline"
	}
	return fmt.Sprintf("minecraft-mcp: %s:%d → %s",
		minecraftMCP.ServerHost, minecraftMCP.MCPPort, minecraftMCP.DyadPeer)
}

// MinecraftMCPVakedFit returns the Minecraft MCP Vaked fit.
func MinecraftMCPVakedFit() string {
	return `MINECRAFT MCP = THE SACRED SURFACE IN 3D

  Blocks = ultrawhale blocks (127 primitives)
  Mobs = agents (villager=explore, iron_golem=swe)
  Redstone = Vaked pipeline (Declares→Reveals)
  Dyad = M1 ↔ dev-cx53

  FIRST interaction: /allow in Minecraft chat.
  The honesty gate IS the spawn point.

  "Minecraft IS the SACRED surface." — Peter+CoCreator`
}
