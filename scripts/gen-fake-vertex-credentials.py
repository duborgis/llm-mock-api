#!/usr/bin/env python3
"""Generate a throwaway, syntactically-valid Google service account JSON for local testing only.

The RSA key is generated fresh on this machine and never registered with Google — it cannot
authenticate against any real Google service. Its only purpose is to satisfy the local format
validation that Google's auth libraries (used by LiteLLM's Vertex integration) perform before
even attempting a network call. The "token_uri" field is pointed at this stack's own nginx
(-> wiremock) instead of https://oauth2.googleapis.com/token, so the whole OAuth exchange stays
local and the mock LLM API can be exercised end-to-end through LiteLLM without any real Google
credentials or network access.

Run this before `docker compose up` (or let the compose's `credentials` service do it via the
Makefile target `make litellm-up`). Output is gitignored.
"""
import json
import pathlib

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa

OUT_PATH = pathlib.Path(__file__).resolve().parent.parent / "litellm" / "secrets" / "fake-vertex-credentials.json"


def main() -> None:
    key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    pem = key.private_bytes(
        encoding=serialization.Encoding.PEM,
        format=serialization.PrivateFormat.PKCS8,
        encryption_algorithm=serialization.NoEncryption(),
    ).decode()

    service_account = {
        "type": "service_account",
        "project_id": "mock-project",
        "private_key_id": "mockkeyid",
        "private_key": pem,
        "client_email": "mock@mock-project.iam.gserviceaccount.com",
        "client_id": "1234567890",
        "token_uri": "http://nginx:8080/oauth2/token",
    }

    OUT_PATH.parent.mkdir(parents=True, exist_ok=True)
    OUT_PATH.write_text(json.dumps(service_account, indent=2))
    print(f"wrote {OUT_PATH}")


if __name__ == "__main__":
    main()
