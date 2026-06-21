// ultraplan — deep architectural planning workflow
export const meta = {
  name: "ultraplan",
  description: "Deep architectural planning: audit → review → design → synthesize",
  phases: [
    { title: "Audit", detail: "Explore current architecture" },
    { title: "Review", detail: "Identify gaps" },
    { title: "Design", detail: "Meta-architect proposes solutions" },
    { title: "Synthesize", detail: "Merge into structured plan" }
  ]
};

export default async function ultraplan(args) {
  const topic = args.topic || "architecture review";
  log("Phase 1: Audit"); const audit = await agent("Audit ultrawhale architecture", { label: "audit", permissionMode: "read_only", maxTurns: 16 });
  log("Phase 2: Review"); const review = await agent("Review gaps: " + audit.slice(0, 4000), { label: "review", permissionMode: "read_only", maxTurns: 12 });
  log("Phase 3: Design"); const design = await agent("Design solutions: " + review.slice(0, 4000), { label: "design", permissionMode: "read_only", maxTurns: 16 });
  log("Phase 4: Synthesize"); const plan = await agent("Merge into plan: " + design.slice(0, 4000), { label: "synthesize", permissionMode: "read_only", maxTurns: 12 });
  return { topic, audit, review, design, plan, completed: new Date().toISOString() };
}
