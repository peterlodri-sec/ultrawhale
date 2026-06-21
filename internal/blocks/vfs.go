package blocks

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ── VFS — Space as Virtual Filesystem ────────────────────────────────
//
// The capability graph materialized as a filesystem.
//   ls /ultrawhale/agents/    → list all agents
//   cat /ultrawhale/agents/swe-1/status → "running"
//   echo "implement auth" > /ultrawhale/orchestrator/delegate → DelegatePrompt()
//
// Every node is a directory. Every edge is a symlink.
// The filesystem IS the Vaked capability graph.

// VFSNode is a node in the virtual filesystem.
type VFSNode struct {
	Name     string
	Path     string            // full VFS path
	Kind     string            // "dir", "file", "symlink"
	Content  string            // for files: the content
	Children map[string]*VFSNode
	Parent   *VFSNode
	SpaceRef string            // space topology node ID
}

// VFS is the virtual filesystem root.
type VFS struct {
	mu   sync.RWMutex
	Root *VFSNode
	CWD  string
}

var vfs = &VFS{
	Root: &VFSNode{
		Name:     "/",
		Path:     "/",
		Kind:     "dir",
		Children: make(map[string]*VFSNode),
	},
	CWD: "/",
}

// InitVFS builds the VFS from the space topology.
func InitVFS() {
	vfs.mu.Lock()
	defer vfs.mu.Unlock()

	root := vfs.Root
	root.Children = make(map[string]*VFSNode)

	// /ultrawhale/
	uw := vfsMkdir(root, "ultrawhale")

	// /ultrawhale/orchestrator/
	orch := vfsMkdir(uw, "orchestrator")
	vfsMkfile(orch, "status", GetOrchestrator().OrchestratorStatus())
	vfsMkfile(orch, "did", GetOrchestrator().DID)
	vfsMkfile(orch, "turns", fmt.Sprintf("%d", GetOrchestrator().TotalTurns))

	// /ultrawhale/agents/
	agents := vfsMkdir(uw, "agents")
	for _, a := range ListAgents() {
		agentDir := vfsMkdir(agents, a.ID[:12])
		vfsMkfile(agentDir, "role", a.Role)
		vfsMkfile(agentDir, "status", a.Status)
		vfsMkfile(agentDir, "tools", fmt.Sprintf("%d", a.ToolCalls))
		vfsMkfile(agentDir, "caps", GetCapProfile(a.Role).Name)
	}

	// /ultrawhale/swarms/
	swarms := vfsMkdir(uw, "swarms")
	for _, s := range ListSwarms() {
		swDir := vfsMkdir(swarms, s.ID)
		vfsMkfile(swDir, "status", s.Status)
		vfsMkfile(swDir, "port", fmt.Sprintf("%d", s.AFPorthttp))
		vfsMkfile(swDir, "tasks", fmt.Sprintf("%d", s.TotalTasks))
	}

	// /ultrawhale/dyad/
	dyad := vfsMkdir(uw, "dyad")
	if d := GetDyad(); d != nil {
		vfsMkfile(dyad, "peer", d.Peer.Machine)
		vfsMkfile(dyad, "status", d.Status)
		vfsMkfile(dyad, "pings", fmt.Sprintf("%d", d.PingCount))
	}

	// /ultrawhale/brain/
	brain := vfsMkdir(uw, "brain")
	vfsMkfile(brain, "dump", GetBrain().BrainDump())

	// /ultrawhale/blocks/
	blocks := vfsMkdir(uw, "blocks")
	vfsMkfile(blocks, "count", fmt.Sprintf("%d", len(schemaRegistry)))
	vfsMkfile(blocks, "list", SchemaStatus())

	// /ultrawhale/SACRED
	vfsMkfile(uw, "SACRED", SacredStatus())

	// /ultrawhale/vaked-triangle
	vfsMkfile(uw, "vaked-triangle", VakedTriangle())

	Log(LogInfo, "vfs.init", "built from space topology", "", "", 0, nil)
}

func vfsMkdir(parent *VFSNode, name string) *VFSNode {
	path := parent.Path + "/" + name
	if parent.Path == "/" { path = "/" + name }
	node := &VFSNode{Name: name, Path: path, Kind: "dir", Children: make(map[string]*VFSNode), Parent: parent}
	parent.Children[name] = node
	return node
}

func vfsMkfile(parent *VFSNode, name, content string) {
	path := parent.Path + "/" + name
	if parent.Path == "/" { path = "/" + name }
	parent.Children[name] = &VFSNode{Name: name, Path: path, Kind: "file", Content: content, Parent: parent}
}

// ── VFS Operations ────────────────────────────────────────────────────

// VFSLs lists the contents of a VFS path.
func VFSLs(path string) ([]string, error) {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()

	node := vfsResolve(path)
	if node == nil { return nil, fmt.Errorf("vfs: %s not found", path) }
	if node.Kind != "dir" { return []string{node.Name}, nil }

	var entries []string
	for name, child := range node.Children {
		prefix := "  "
		if child.Kind == "dir" { prefix = "📁 " } else { prefix = "📄 " }
		entries = append(entries, fmt.Sprintf("%s%s", prefix, name))
	}
	sort.Strings(entries)
	return entries, nil
}

// VFSCat reads the content of a VFS file.
func VFSCat(path string) (string, error) {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()

	node := vfsResolve(path)
	if node == nil { return "", fmt.Errorf("vfs: %s not found", path) }
	if node.Kind == "dir" { return fmt.Sprintf("%s/ (directory)", path), nil }
	return node.Content, nil
}

// VFSEcho writes content to a VFS path (delegates to orchestrator if applicable).
func VFSEcho(path, content string) (string, error) {
	vfs.mu.RLock()
	node := vfsResolve(path)
	vfs.mu.RUnlock()

	if node == nil {
		// Create if it's a known path pattern
		if strings.HasPrefix(path, "/ultrawhale/orchestrator/delegate") {
			orch := GetOrchestrator()
			agentID, role := orch.DelegatePrompt(content)
			return fmt.Sprintf("delegated to %s (%s)", role, agentID[:8]), nil
		}
		return "", fmt.Errorf("vfs: %s not found", path)
	}

	// Update existing node
	vfs.mu.Lock()
	node.Content = content
	vfs.mu.Unlock()

	Log(LogInfo, "vfs.echo", path, Ref([]byte(content)), "", 0, nil)
	return content, nil
}

// VFSCD changes the current working directory.
func VFSCD(path string) string {
	vfs.mu.RLock()
	node := vfsResolve(path)
	vfs.mu.RUnlock()

	if node != nil && node.Kind == "dir" {
		vfs.CWD = path
		return vfs.CWD
	}
	return vfs.CWD
}

func vfsResolve(path string) *VFSNode {
	if path == "/" { return vfs.Root }

	// Normalize: strip trailing /
	path = strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(path, "/") { path = vfs.CWD + "/" + path }

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	current := vfs.Root
	for _, part := range parts {
		if part == "" || part == "." { continue }
		if part == ".." {
			if current.Parent != nil { current = current.Parent }
			continue
		}
		child, ok := current.Children[part]
		if !ok { return nil }
		current = child
	}
	return current
}

// VFSStatus returns compact VFS status.
func VFSStatus() string {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()
	return fmt.Sprintf("vfs: %s", vfs.CWD)
}

// VFSTree returns a tree representation of the VFS.
func VFSTree() string {
	vfs.mu.RLock()
	defer vfs.mu.RUnlock()
	return vfsRenderTree(vfs.Root, "", true)
}

func vfsRenderTree(node *VFSNode, prefix string, isLast bool) string {
	var sb strings.Builder
	connector := "├── "
	if isLast { connector = "└── " }

	if node.Name != "/" {
		icon := "📁"
		if node.Kind == "file" { icon = "📄" }
		sb.WriteString(prefix + connector + icon + " " + node.Name)
		if node.Kind == "file" && len(node.Content) < 40 {
			sb.WriteString(" → " + node.Content)
		}
		sb.WriteString("\n")
	}

	if node.Kind == "dir" {
		children := sortedChildren(node)
		for i, name := range children {
			child := node.Children[name]
			childPrefix := prefix
			if node.Name != "/" {
				if isLast { childPrefix += "    " } else { childPrefix += "│   " }
			}
			sb.WriteString(vfsRenderTree(child, childPrefix, i == len(children)-1))
		}
	}

	return sb.String()
}

func sortedChildren(node *VFSNode) []string {
	keys := make([]string, 0, len(node.Children))
	for k := range node.Children { keys = append(keys, k) }
	sort.Strings(keys)
	return keys
}
