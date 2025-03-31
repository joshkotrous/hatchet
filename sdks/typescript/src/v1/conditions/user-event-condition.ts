import { Condition, Action } from './base';

export interface UserEvent {
  /**
   * The unique key identifying the specific user event to monitor.
   * This should match the event key that will be emitted from your application.
   * @example "button:clicked", "page:viewed", "form:submitted"
   */
  eventKey: string;

  /**
   * Optional CEL expression to evaluate against the event data.
   * When provided, the condition will only trigger if this expression evaluates to true.
   * Expressions are validated for security before execution.
   * @example "input.quantity > 5", "input.status == 'completed'"
   */
  expression?: string;

  /**
   * Optional unique identifier for the event data in the readable stream.
   * When multiple conditions are listening to the same event type,
   * using a custom readableDataKey prevents duplicate data processing
   * by differentiating between the conditions in the data store.
   * If not specified, the eventKey will be used as the default identifier.
   */
  readableDataKey?: string;
}

/**
 * Validates a CEL expression for security concerns.
 * Rejects expressions with potentially unsafe patterns.
 * 
 * @param expression The CEL expression to validate
 * @throws Error if the expression contains potentially unsafe patterns
 */
function validateCelExpression(expression: string): void {
  if (!expression || expression.trim() === '') {
    return;
  }

  // Check for potentially unsafe patterns in CEL expressions
  const unsafePatterns = [
    /\beval\b/i,                 // eval
    /\bFunction\b/i,             // Function constructor
    /\bprocess\b/i,              // process access
    /\bglobal\b/i,               // global object
    /\b__(proto|dirname|filename)__\b/i, // special properties
    /\b(require|import)\b/i,     // module imports
    /\bexec\b/i,                 // execution functions
    /\;/,                        // semicolons (not valid in CEL)
    /\{.*\}/                     // code blocks (not valid in CEL)
  ];

  for (const pattern of unsafePatterns) {
    if (pattern.test(expression)) {
      throw new Error(`Security validation failed: Potentially unsafe pattern detected in expression`);
    }
  }
}

/**
 * Represents a condition that is triggered based on a specific user event.
 * This condition monitors for events with the specified key and evaluates
 * any provided expression against the event data.
 *
 * @param eventKey The key identifying the specific user event to monitor
 * @param expression The CEL expression to evaluate against the event data
 * @param readableDataKey Optional parameter that provides a unique identifier for the data.
 *                        When multiple conditions are listening to the same event, using a custom
 *                        readableDataKey prevents duplicate data by differentiating between the conditions.
 *                        If not provided, defaults to the eventKey.
 * @param action Optional action to execute when the condition is met
 *
 * @example
 * // Create a condition that triggers when a "purchase" event occurs with amount > 100
 * const purchaseCondition = new UserEventCondition(
 *   "purchase",
 *   "data.amount > 100",
 *   "high_value_purchase",
 *   () => console.log("High value purchase detected!")
 * );
 * 
 * @throws Error if the provided expression fails security validation
 */
export class UserEventCondition extends Condition {
  eventKey: string;
  expression: string;

  constructor(eventKey: string, expression: string, readableDataKey?: string, action?: Action) {
    // Validate the expression for security before using it
    validateCelExpression(expression);
    
    super({
      readableDataKey: readableDataKey || eventKey,
      action,
      orGroupId: '',
      expression: '',
    });
    this.eventKey = eventKey;
    this.expression = expression;
  }
}