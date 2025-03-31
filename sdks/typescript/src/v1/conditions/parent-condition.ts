import { Condition, Action } from './base';
import { CreateWorkflowTaskOpts } from '../task';

export interface Parent {
  /**
   * The parent workflow task this condition is associated with.
   * This establishes a relationship between this condition and a specific workflow task.
   */
  parent: CreateWorkflowTaskOpts<any, any>;

  /**
   * Optional CEL expression to evaluate against the parent task's data.
   * When provided, the condition will only trigger if this expression evaluates to true.
   * @example "input.status == 'completed'", "input.result > 0"
   */
  expression?: string;
}

/**
 * Validates a CEL (Common Expression Language) expression to ensure it doesn't contain potentially
 * unsafe patterns. This is a basic validation that checks for common unsafe patterns.
 * 
 * @param expression The CEL expression to validate
 * @returns true if the expression is considered safe, false otherwise
 */
function validateCELExpression(expression?: string): boolean {
  if (!expression) return true;
  
  // Patterns that are considered unsafe in CEL expressions
  const unsafePatterns = [
    /\beval\s*\(/,                      // Prevent eval()
    /\bFunction\s*\(/,                  // Prevent Function constructor
    /\bnew\s+Function\s*\(/,            // Prevent new Function()
    /\(.*\)\s*=>\s*{/,                  // Prevent arrow function with block body
    /function\s*\(/,                    // Prevent function declarations
    /\bimport\s*\(/,                    // Prevent dynamic imports
    /\brequire\s*\(/,                   // Prevent require()
    /__proto__/,                        // Prevent prototype manipulation
    /\bprocess\b/,                      // Prevent Node.js process object
    /\bdocument\b/,                     // Prevent DOM access
    /\bwindow\b/,                       // Prevent window access
    /\bglobal\b/,                       // Prevent global object access
  ];
  
  // Check if the expression contains any unsafe patterns
  return !unsafePatterns.some(pattern => pattern.test(expression));
}

/**
 * Represents a condition that is triggered based on a parent workflow task.
 * This condition monitors the specified parent task and evaluates
 * any provided expression against the task's data.
 *
 * @example
 * // Create a condition that triggers when a parent task completes successfully
 * const parentCondition = new ParentCondition(
 *   parentTaskOpts,
 *   "input.status == 'success'",
 *   () => console.log("Parent task completed successfully!")
 * );
 */
export class ParentCondition extends Condition {
  /** The parent workflow task this condition is associated with */
  parent: CreateWorkflowTaskOpts<any, any>;

  /**
   * Creates a new condition that is triggered based on a parent workflow task.
   *
   * @param parent The parent workflow task this condition is associated with
   * @param expression Optional CEL expression to evaluate against the parent task's data
   * @param readableDataKey Optional key to access data from the parent task
   * @param action Optional action to execute when the condition is met
   */
  constructor(
    parent: CreateWorkflowTaskOpts<any, any>,
    expression?: string,
    readableDataKey?: string,
    action?: Action
  ) {
    // Validate the CEL expression for potentially unsafe patterns
    if (expression && !validateCELExpression(expression)) {
      throw new Error('Invalid CEL expression: The expression contains potentially unsafe patterns');
    }

    super({
      readableDataKey: readableDataKey || `parent-${parent.name || Date.now().toString()}`,
      action,
      orGroupId: '',
      expression: expression || '',
    });
    this.parent = parent;
  }
}