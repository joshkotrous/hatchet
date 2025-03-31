/* eslint-disable no-console */
import { failureWorkflow } from './workflow';

async function main() {
  try {
    const res = await failureWorkflow.run({});
    console.log(res);
  } catch (e) {
    // Log only the error message and not the entire error object
    console.log('error', e instanceof Error ? e.message : 'An unknown error occurred');
  }
}

if (require.main === module) {
  main()
    .catch(console.error)
    .finally(() => process.exit(0));
}