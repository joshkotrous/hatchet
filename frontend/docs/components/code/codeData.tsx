const defaultUser = "hatchet-dev";

const defaultRepo = "hatchet";

const localPaths = {
  ts: "sdks/typescript",
  py: "sdks/python",
  go: "",
};

export const extToLanguage = {
  ts: "typescript",
  py: "python",
  go: "go",
};

const defaultBranch = "main";

export type RepoProps = {
  user?: string;
  repo?: string;
  branch?: string;
  path: string;
};

// Validate path to prevent traversal
const validatePath = (path: string): string => {
  if (!path) return '';
  // Prevent path traversal by removing '..' segments
  return path.split('/').filter(part => part !== '..').join('/');
};

// Sanitize URL parameters
const sanitizeUrlParam = (param: string | undefined, defaultValue: string): string => {
  if (!param) {
    return defaultValue;
  }
  // Only allow alphanumeric characters, hyphens, and underscores
  return param.replace(/[^a-zA-Z0-9_\-]/g, '');
};

const getLocalUrl = (ext: string, { path }: RepoProps) => {
  const safePath = validatePath(path);
  return `http://localhost:4001/${localPaths[ext]}/${safePath}`;
};

const isDev = process?.env?.NODE_ENV === "development";

const getRawUrl = ({ user, repo, branch, path }: RepoProps) => {
  const safePath = validatePath(path);
  const safeUser = sanitizeUrlParam(user, defaultUser);
  const safeRepo = sanitizeUrlParam(repo, defaultRepo);
  const safeBranch = sanitizeUrlParam(branch, defaultBranch);
  
  const ext = safePath.split(".").pop() || '';
  
  if (isDev) {
    return getLocalUrl(ext, { path: safePath });
  }
  
  return `https://raw.githubusercontent.com/${safeUser}/${safeRepo}/refs/heads/${safeBranch}/${localPaths[ext]}/${safePath}`;
};

const getUIUrl = ({ user, repo, branch, path }: RepoProps) => {
  const safePath = validatePath(path);
  const safeUser = sanitizeUrlParam(user, defaultUser);
  const safeRepo = sanitizeUrlParam(repo, defaultRepo);
  const safeBranch = sanitizeUrlParam(branch, defaultBranch);
  
  const ext = safePath.split(".").pop() || '';
  
  if (isDev) {
    return getLocalUrl(ext, { path: safePath });
  }
  
  return `https://github.com/${safeUser}/${safeRepo}/blob/${safeBranch}/${localPaths[ext]}/${safePath}`;
};

export type Src = {
  raw: string;
  props: RepoProps;
  rawUrl: string;
  githubUrl: string;
  language?: string;
};
export const getSnippets = (
  props: RepoProps[]
): Promise<{ props: { ssg: { contents: Src[] } } }> => {
  return Promise.all(
    props.map(async (prop) => {
      const rawUrl = getRawUrl(prop);
      const githubUrl = getUIUrl(prop);
      
      // Extract extension from validated path for consistency
      const safePath = validatePath(prop.path);
      const fileExt = (safePath.split(".").pop() || '') as keyof typeof extToLanguage;
      const language = extToLanguage[fileExt];

      try {
        const response = await fetch(rawUrl);
        const raw = await response.text();

        return {
          raw,
          props: prop,
          rawUrl,
          githubUrl,
          language,
        };
      } catch (error) {
        // Return object with empty raw content but preserve URLs on failure
        return {
          raw: "",
          props: prop,
          rawUrl,
          githubUrl,
          language,
        };
      }
    })
  ).then((results) => ({
    props: {
      ssg: {
        contents: results,
      },
      // revalidate every 60 seconds
      revalidate: 60,
    },
  }));
};