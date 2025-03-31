import base64
import hmac
import hashlib
import json
import os

from pydantic import BaseModel


class Claims(BaseModel):
    sub: str
    server_url: str
    grpc_broadcast_address: str


def get_tenant_id_from_jwt(token: str) -> str:
    return extract_claims_from_jwt(token).sub


def get_addresses_from_jwt(token: str) -> tuple[str, str]:
    claims = extract_claims_from_jwt(token)

    return claims.server_url, claims.grpc_broadcast_address


def extract_claims_from_jwt(token: str) -> Claims:
    parts = token.split(".")
    if len(parts) != 3:
        raise ValueError("Invalid token format")

    # Verify the JWT signature
    verify_jwt_signature(parts)

    claims_part = parts[1]
    claims_part += "=" * ((4 - len(claims_part) % 4) % 4)  # Padding for base64 decoding
    claims_data = base64.urlsafe_b64decode(claims_part)

    return Claims.model_validate_json(claims_data)


def verify_jwt_signature(parts: list[str]) -> None:
    """
    Verify the JWT signature using the algorithm specified in the header.
    This implementation supports HS256 (HMAC-SHA256) by default.
    """
    header_part, payload_part, signature_part = parts
    
    # Decode the header to get the algorithm
    header_part_padded = header_part + "=" * ((4 - len(header_part) % 4) % 4)
    header_data = base64.urlsafe_b64decode(header_part_padded)
    header = json.loads(header_data.decode('utf-8'))
    
    # Get the algorithm
    alg = header.get('alg', 'HS256')
    
    # Currently only supporting HS256
    if alg != 'HS256':
        raise ValueError(f"Unsupported JWT algorithm: {alg}. Only HS256 is supported.")
    
    # Base64url decode the signature
    signature_part_padded = signature_part + "=" * ((4 - len(signature_part) % 4) % 4)
    signature = base64.urlsafe_b64decode(signature_part_padded)
    
    # Create the message to be verified (header.payload)
    message = f"{header_part}.{payload_part}".encode('utf-8')
    
    # Get the secret key from environment variable
    secret_key = os.environ.get("JWT_SECRET_KEY")
    if not secret_key:
        raise ValueError("JWT_SECRET_KEY environment variable is not set")
    secret_key = secret_key.encode('utf-8')
    
    # Create the expected signature using HMAC-SHA256
    expected_signature = hmac.new(
        secret_key,
        message,
        hashlib.sha256
    ).digest()
    
    # Compare the expected signature with the actual signature
    if not hmac.compare_digest(signature, expected_signature):
        raise ValueError("Invalid JWT signature")