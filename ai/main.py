import os

from fastapi import FastAPI

app = FastAPI(title="anonimus-ai", version="0.1.0")


@app.get("/health")
def health() -> dict[str, str | bool]:
    return {
        "status": "ok",
        "service": "ai",
        "runpod_llm_configured": bool(os.getenv("RUNPOD_LLM_URL")),
        "runpod_embedding_configured": bool(os.getenv("RUNPOD_EMBEDDING_URL")),
    }
