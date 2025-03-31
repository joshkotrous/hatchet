export function getTenantIdFromJWT(token: string): string {
  const claims = extractClaimsFromJWT(token);
  return claims.sub;
}

export function getAddressesFromJWT(token: string): {
  serverUrl: string;
  grpcBroadcastAddress: string;
} {
  const claims = extractClaimsFromJWT(token);
  return {
    serverUrl: claims.server_url,
    grpcBroadcastAddress: claims.grpc_broadcast_address,
  };
}

/**
 * Interface for JWT verification.
 * The application should implement this interface using a proper JWT library.
 */
export interface JWTVerifier {
  /**
   * Verify a JWT token's signature.
   * 
   * @param token The JWT token to verify
   * @returns true if the token's signature is valid, false otherwise
   */
  verify(token: string): boolean;
}

// The JWT verifier implementation (null by default)
let jwtVerifier: JWTVerifier | null = null;
// Whether verification is enforced (true by default for security)
let enforcedVerification: boolean = true;

/**
 * Set a custom JWT verifier.
 * The application must implement a proper verifier using a JWT library.
 * 
 * @param verifier The JWT verifier implementation
 */
export function setJWTVerifier(verifier: JWTVerifier): void {
  jwtVerifier = verifier;
}

/**
 * SECURITY WARNING: Disables JWT verification enforcement.
 * This should ONLY be used in testing environments or legacy compatibility.
 * Using this in production creates a serious security vulnerability.
 */
export function disableJWTVerification(): void {
  enforcedVerification = false;
  console.warn(
    'SECURITY WARNING: JWT verification has been disabled. ' +
    'This is a significant security risk in production environments.'
  );
}

/**
 * Re-enables JWT verification enforcement (on by default).
 */
export function enableJWTVerification(): void {
  enforcedVerification = true;
}

/**
 * Extract claims from a JWT token.
 * If JWT verification is enabled, the token's signature will be verified.
 * 
 * @param token The JWT token
 * @returns The decoded claims from the token
 * @throws Error if the token format is invalid or signature verification fails
 */
function extractClaimsFromJWT(token: string): any {
  const parts = token.split('.');
  if (parts.length !== 3) {
    throw new Error('Invalid token format');
  }

  // Verify the token if verification is enabled
  if (jwtVerifier !== null) {
    // Use the provided verifier
    if (!jwtVerifier.verify(token)) {
      throw new Error('JWT signature verification failed');
    }
  } else if (enforcedVerification) {
    // No verifier but verification is enforced
    throw new Error(
      'JWT verification is required but no verifier is configured. ' +
      'Call setJWTVerifier() to provide a verification implementation, ' +
      'or disableJWTVerification() to disable verification (NOT RECOMMENDED).'
    );
  } else {
    // No verifier and verification is not enforced (dangerous)
    console.warn(
      'SECURITY WARNING: Processing JWT without signature verification. ' +
      'This is a significant security risk.'
    );
  }

  const claimsPart = parts[1];
  const claimsData = atob(claimsPart.replace(/-/g, '+').replace(/_/g, '/'));
  const claims = JSON.parse(claimsData);

  return claims;
}