from fastapi.testclient import TestClient

from ai.main import app


def test_health_returns_ok() -> None:
    client = TestClient(app)
    response = client.get("/health")
    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    assert body["service"] == "ai"
    assert "runpod_llm_configured" in body
    assert "runpod_embedding_configured" in body
