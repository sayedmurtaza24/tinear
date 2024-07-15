CREATE VIRTUAL TABLE search USING fts5 (
  tokenize = "trigram",
  id UNINDEXED,
  title,
  description,
  state,
  project,
  team,
  assignee,
  labels
);
