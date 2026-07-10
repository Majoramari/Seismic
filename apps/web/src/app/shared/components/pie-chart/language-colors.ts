// Colors matching GitHub's linguist language colors, so
// charts look consistent with what developers already
// associate each language with. Source: github-linguist/linguist
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
  plain_text: '#6b7280',
  'objective-c': '#438eff',
  r: '#198CE7',
  perl: '#0298c3',
  haskell: '#5e5086',
  elixir: '#6e4a7e',
  clojure: '#db5855',
  zig: '#ec915c',
  nim: '#ffc200',
  jsx: '#f1e05a',
  tsx: '#3178c6',
  xml: '#0060ac',
  ini: '#d1dbe0',
  makefile: '#427819',
  vim: '#199f4b',
  'jupyter notebook': '#DA5B0B',
  svelte: '#ff3e00',
  astro: '#ff5a03',
  graphql: '#e10098',
  powershell: '#012456',
  julia: '#a270ba',
  matlab: '#e16737',
  fsharp: '#b845fc',
  erlang: '#B83998',
  crystal: '#000100',
  ocaml: '#3be133',
  groovy: '#4298b8',
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
// the same unknown language always gets the same color too,
// rather than depending on array position.
export function getLanguageColor(language: string, fallbackIndex: number): string {
  const key = language.toLowerCase().trim();
  const known = LANGUAGE_COLORS[key];
  if (known) return known;

  // Deterministic hash so the same unknown language always
  // maps to the same fallback color, regardless of array order.
  let hash = 0;
  for (let i = 0; i < key.length; i++) {
    hash = key.charCodeAt(i) + ((hash << 5) - hash);
  }
  const index = Math.abs(hash) % FALLBACK_COLORS.length;
  return FALLBACK_COLORS[index];
}
