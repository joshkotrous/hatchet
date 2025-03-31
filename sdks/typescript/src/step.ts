/* eslint-disable max-classes-per-file */
[Previous code content unchanged until line 752]

// Helper function to validate CEL expressions with basic security checks
function validateCelExpression(expr: string): boolean {
  // Check if expression is empty
  if (!expr || expr.trim() === '') {
    return false;
  }

  // Maximum length check to prevent DoS
  if (expr.length > 1000) {
    return false;
  }

  // Check for balanced parentheses, brackets, and braces
  const stack: string[] = [];
  for (let i = 0; i < expr.length; i++) {
    const char = expr[i];
    
    if (char === '(') stack.push(')');
    else if (char === '{') stack.push('}');
    else if (char === '[') stack.push(']');
    else if (char === ')' || char === '}' || char === ']') {
      if (stack.length === 0 || stack.pop() !== char) return false;
    }
  }
  
  // If stack is not empty, there are unbalanced opening characters
  if (stack.length > 0) {
    return false;
  }

  // Check for suspicious patterns that might indicate injection attempts
  const suspiciousPatterns = [
    /\beval\s*\(/i,      // JavaScript eval
    /\bFunction\s*\(/i,  // JavaScript Function constructor
    /\bnew\s+Function/i, // JavaScript Function constructor alternative
    /\bimport\s*\(/i,    // Dynamic imports
    /\brequire\s*\(/i,   // Node.js require
    /\bprocess\b/i,      // Node.js process
    /\bglobal\b/i        // Node.js global
  ];

  for (const pattern of suspiciousPatterns) {
    if (pattern.test(expr)) {
      return false;
    }
  }

  return true;
}

[Remaining code content unchanged]