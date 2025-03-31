import { Worker, WorkerOptions } from 'worker_threads';
import path from 'path';

export function runThreaded(scriptPath: string, options: WorkerOptions) {
  if (typeof scriptPath !== 'string' || !scriptPath) {
    throw new Error('Invalid script path provided');
  }
  
  try {
    // Resolve and normalize the script path
    const resolvedPath = path.normalize(require.resolve(scriptPath));
    
    // Check if the file is a TypeScript file
    const isTs = /\.ts$/.test(resolvedPath);
    
    if (isTs) {
      // For TypeScript files, create a sandboxed environment
      // Use JSON.stringify to properly escape the paths and prevent injection
      const rootDir = JSON.stringify(path.join(__dirname, '../../../'));
      const scriptPathJson = JSON.stringify(resolvedPath);
      
      const workerCode = `
      const wk = require('worker_threads');
      require('tsconfig-paths/register');
      require('ts-node').register({
        "include": ["src/**/*.ts"],
        "exclude": ["./dist"],
        "compilerOptions": {
          "types": ["node"],
          "target": "es2016",
          "esModuleInterop": true,
          "module": "commonjs",
          "rootDir": ${rootDir}
        }
      });
      
      // Use JSON.parse to safely get the file path
      const file = ${scriptPathJson};
      require(file);
      `;
      
      return new Worker(workerCode, {
        ...options,
        eval: true
      });
    } else {
      // For non-TypeScript files, use the file directly
      return new Worker(resolvedPath, options);
    }
  } catch (error) {
    throw new Error(`Failed to initialize worker thread: ${error.message}`);
  }
}

// execArgv:  ? ['--require', 'ts-node/register'] : undefined,