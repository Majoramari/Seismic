// Colors matching GitHub's linguist language colors, so
// charts look consistent with what developers already
// associate each language with.
export const LANGUAGE_COLORS: Record<string, string> = {
  javascript: '#f1e05a',
  typescript: '#3178c6',
  python: '#3572A5',
  go: '#00ADD8',
  rust: '#dea584',
  java: '#b07219',
  c: '#555555',
  'c++': '#f34b7d',
  'c#': '#178600',
  html: '#e34c26',
  css: '#563d7c',
  json: '#292929',
  markdown: '#083fa1',
  shell: '#89e051',
  ruby: '#701516',
  php: '#4F5D95',
  swift: '#F05138',
  kotlin: '#A97BFF',
  dart: '#00B4AB',
  vue: '#41b883',
  yaml: '#cb171e',
  sql: '#e38c00',
  lua: '#000080',
  scala: '#c22d40',
  toml: '#9c4221',
  dockerfile: '#384d54',
  plaintext: '#6b7280',
};

const FALLBACK_COLORS = [
  '#e8c547',
  '#3b82f6',
  '#ec4899',
  '#a855f7',
  '#ef4444',
  '#f97316',
  '#0ea5e9',
  '#22c55e',
  '#8b5cf6',
  '#14b8a6',
];

// Returns a consistent color for a given language. Known
// languages use their real linguist color. Unknown ones get
// a deterministic fallback based on the string itself, so
// the same unknown language always gets the same color too.
export function getLanguageColor(language: string, fallbackIndex: number): string {
  const known = LANGUAGE_COLORS[language.toLowerCase()];
  if (known) return known;
  return FALLBACK_COLORS[fallbackIndex % FALLBACK_COLORS.length];
}
