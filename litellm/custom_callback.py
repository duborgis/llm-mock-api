"""Custom LiteLLM callback that forwards each call's full original_response to the
openmeter-mock service, which persists it into MongoDB (collection "raw_responses").

This exists because LiteLLM's built-in "otel" callback only exposes original_response as
span attributes (see litellm/integrations/opentelemetry.py, set_raw_request_attributes) —
useful for a quick look in Jaeger, but not queryable/storable on its own. Registered in
litellm/config.yaml as litellm_settings.callbacks: ["custom_callback.raw_response_logger"].
"""

import json

import httpx
from litellm.integrations.custom_logger import CustomLogger

RAW_RESPONSE_ENDPOINT = "http://openmeter-mock:9010/api/v1/raw-responses"


def _to_jsonable(value):
    if value is None:
        return None
    if isinstance(value, str):
        try:
            return json.loads(value)
        except (TypeError, ValueError):
            return value
    if hasattr(value, "model_dump"):
        return value.model_dump(mode="json")
    if hasattr(value, "dict"):
        return value.dict()
    return value


class RawResponseLogger(CustomLogger):
    async def async_log_success_event(self, kwargs, response_obj, start_time, end_time):
        del response_obj, start_time, end_time
        try:
            metadata = kwargs.get("litellm_params", {}).get("metadata", {}) or {}
            payload = {
                "call_id": kwargs.get("litellm_call_id", ""),
                "model": kwargs.get("model", ""),
                "route": metadata.get("user_api_key_request_route", ""),
                "raw_response": _to_jsonable(kwargs.get("original_response")),
            }
            async with httpx.AsyncClient(timeout=5) as client:
                await client.post(RAW_RESPONSE_ENDPOINT, json=payload)
        except Exception:
            # Best-effort/non-blocking: never let this callback break the actual request.
            pass


raw_response_logger = RawResponseLogger()
