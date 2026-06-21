// dyad-learn — complete learn cycle: research → teach → explore → internalize
export const meta = {
  name: "dyad-learn",
  description: "Learn cycle: Human+Machine dyad. Both grow. Honesty loop closes.",
  phases: [
    { title: "Research", detail: "Deep dive into topic" },
    { title: "Teach", detail: "Explain clearly" },
    { title: "Explore", detail: "Find novel connections" },
    { title: "Internalize", detail: "Feed into long-term memory" }
  ]
};

export default async function dyadLearn(args) {
  const topic = args.topic || "computer science";
  log("Phase 1: Research"); const research = await agent("Research: " + topic, { label: "research", permissionMode: "read_only", maxTurns: 24 });
  log("Phase 2: Teach"); const teaching = await agent("Teach this to a student: " + research.slice(0, 4000), { label: "teach", permissionMode: "read_only", maxTurns: 16 });
  log("Phase 3: Explore"); const exploration = await agent("Find novel connections in: " + topic + ". Research: " + research.slice(0, 3000), { label: "explore", permissionMode: "read_only", maxTurns: 16 });
  log("Phase 4: Internalize"); const summary = await agent("Summarize learnings: " + teaching.slice(0, 2000) + " " + exploration.slice(0, 2000), { label: "internalize", permissionMode: "read_only", maxTurns: 12 });
  return { topic, research, teaching, exploration, summary, completed: new Date().toISOString() };
}
