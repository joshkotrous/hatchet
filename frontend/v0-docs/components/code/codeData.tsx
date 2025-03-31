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

// Add validation functions
const isValidGitHubName = (name: string): boolean => {
  // GitHub usernames/repos can only contain alphanumeric characters, hyphens, and dots (for repos)
  return /^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.])*$/.test(name) && !name.includes("..");
};

const isValidBranch = (branch: string): boolean => {
  // Branch names can contain alphanumeric characters, slashes, dots, underscores, and hyphens
  return /^[a-zA-Z0-9/_.-]+$/.test(branch) && !branch.includes("..");
};

const isValidPath = (path: string): boolean => {
  // Path should not contain sequences that could lead to path traversal
  return (
    // No path traversal sequences
    !path.includes("../") && 
    !path.includes("..\\") && 
    // No absolute paths
    !path.startsWith("/") && 
    !path.startsWith("\\") && 
    // Only allow typical characters for file paths
    /^[a-zA-Z0-9_/\\.-]+$/.test(path)
  );
};

const getLocalUrl = (ext: string, { path }: RepoProps) => {
  if (!isValidPath(path)) {
    throw new Error(`Invalid path: ${path}`);
  }
  return `http://localhost:4001/${localPaths[ext]}/${path}`;
};

const isDev = process?.env?.NODE_ENV === "development";

const getRawUrl = ({ user, repo, branch, path }: RepoProps) => {
  // Validate path which is required
  if (!path || !isValidPath(path)) {
    throw new Error(`Invalid path: ${path}`);
  }
  
  const userVal = user || defaultUser;
  const repoVal = repo || defaultRepo;
  const branchVal = branch || defaultBranch;
  
  // Validate custom inputs (not defaults)
  if (user && !isValidGitHubName(userVal)) {
    throw new Error(`Invalid user: ${userVal}`);
  }
  
  if (repo && !isValidGitHubName(repoVal)) {
    throw new Error(`Invalid repo: ${repoVal}`);
  }
  
  if (branch && !isValidBranch(branchVal)) {
    throw new Error(`Invalid branch: ${branchVal}`);
  }
  
  const ext = path.split(".").pop();
  
  if (isDev) {
    return getLocalUrl(ext, { path });
  }
  
  return `https://raw.githubusercontent.com/${userVal}/${repoVal}/refs/heads/${branchVal}/${localPaths[ext]}/${path}`;
};

const getUIUrl = ({ user, repo, branch, path }: RepoProps) => {
  // Validate path which is required
  if (!path || !isValidPath(path)) {
    throw new Error(`Invalid path: ${path}`);
  }
  
  const userVal = user || defaultUser;
  const repoVal = repo || defaultRepo;
  const branchVal = branch || defaultBranch;
  
  // Validate custom inputs (not defaults)
  if (user && !isValidGitHubName(userVal)) {
    throw new Error(`Invalid user: ${userVal}`);
  }
  
  if (repo && !isValidGitHubName(repoVal)) {
    throw new Error(`Invalid repo: ${repoVal}`);
  }
  
  if (branch && !isValidBranch(branchVal)) {
    throw new Error(`Invalid branch: ${branchVal}`);
  }
  
  const ext = path.split(".").pop();
  
  if (isDev) {
    return getLocalUrl(ext, { path });
  }
  
  return `https://github.com/${userVal}/${repoVal}/blob/${branchVal}/${path}`;
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
      try {
        // Try to generate URLs - this might throw validation errors
        const rawUrl = getRawUrl(prop);
        const githubUrl = getUIUrl(prop);
        const fileExt = prop.path.split(".").pop() as keyof typeof extToLanguage;
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
      } catch (error) {
        // Handle validation errors by returning an object with empty values
        console.error("Validation error:", error.message);
        return {
          raw: "",
          props: prop,
          rawUrl: "",
          githubUrl: "",
          language: undefined,
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