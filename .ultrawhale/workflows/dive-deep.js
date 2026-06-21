// dive-deep — Novel Research Synthesis
// VEGED workflow: search SOTA → synthesize experiment → Ralph+Vaked wire → council review → git history

export const meta = {
  name: "dive-deep",
  description: "Search SOTA research, synthesize novel experiment, Vaked council review, append-only git history",
  phases: [
    { title: "Search", detail: "Web search for latest state-of-the-art research" },
    { title: "Synthesize", detail: "Synthesize a NOVEL experiment from findings" },
    { title: "Wire", detail: "Wire experiment through Ralph → Vaked → deep-wire-in" },
    { title: "Council", detail: "ultrawhale+Vaked council review" },
    { title: "Record", detail: "Append-only git-based history (source of truth)" }
  ]
};

export default async function diveDeep(args) {
  const topic = args.topic || "novel research";
  
  // Phase 1: Search — SOTA research
  log("Phase 1: Search — finding latest state-of-the-art");
  const search = await agent(
    `Search for the latest state-of-the-art research on: "${topic}".
     Use web_search, fetch papers from arxiv, find recent breakthroughs.
     Focus on the last 6 months. Identify open problems and active research areas.
     Output: structured research summary with citations.`,
    { label: "search", phase: "Search", permissionMode: "read_only", maxTurns: 32 }
  );
  
  // Phase 2: Synthesize — novel experiment
  log("Phase 2: Synthesize — creating novel experiment");
  const experiment = await agent(
    `Based on this SOTA research, design a NOVEL experiment that has NOT been done before.
     The experiment must be: testable, falsifiable, and contribute new knowledge.
     
     Research: ${search.slice(0, 5000)}
     
     Output: experiment design with hypothesis, methodology, expected outcomes, and how to measure success.`,
    { label: "synthesize", phase: "Synthesize", permissionMode: "read_only", maxTurns: 24 }
  );
  
  // Phase 3: Wire — Ralph + Vaked deep-wire-in
  log("Phase 3: Wire — Ralph learns, Vaked integrates");
  const wiring = await agent(
    `Wire this experiment through the Vaked pipeline.
     - Ralph: observe the experiment design as a new pattern
     - Vaked layers: Declares (hypothesis) → Materializes (method) → Testifies (outcome)
     - deep-wire-in: connect to existing capability graph
     
     Experiment: ${experiment.slice(0, 4000)}
     
     Output: wired experiment with Vaked layer mapping and Ralph pattern.`,
    { label: "wire", phase: "Wire", permissionMode: "read_only", maxTurns: 16 }
  );
  
  // Phase 4: Council — review
  log("Phase 4: Council — ultrawhale+Vaked review");
  const review = await agent(
    `You are the VEGED council. Review this experiment and its wiring.
     - Is the hypothesis sound?
     - Is the methodology rigorous?
     - Is the Vaked wiring correct?
     - What are the risks? What could go wrong?
     - Verdict: APPROVE, REVISE, or REJECT
     
     Experiment: ${experiment.slice(0, 3000)}
     Wiring: ${wiring.slice(0, 3000)}
     
     Output: council verdict with reasoning.`,
    { label: "council", phase: "Council", permissionMode: "read_only", maxTurns: 16 }
  );
  
  // Phase 5: Record — append-only git history
  log("Phase 5: Record — git-based source of truth");
  const record = await agent(
    `Record this entire DIVE-DEEP session as an append-only document.
     Format as markdown for git commit.
     Include: topic, search results, experiment design, Vaked wiring, council verdict.
     
     This is the ONLY source of truth. Immutable. Append-only.
     
     Topic: ${topic}
     Search: ${search.slice(0, 2000)}
     Experiment: ${experiment.slice(0, 2000)}
     Wiring: ${wiring.slice(0, 2000)}
     Review: ${review.slice(0, 2000)}`,
    { label: "record", phase: "Record", permissionMode: "read_only", maxTurns: 12 }
  );
  
  return {
    topic,
    search,
    experiment,
    wiring,
    review,
    record,
    completed: new Date().toISOString(),
    council: "VEGED"
  };
}
