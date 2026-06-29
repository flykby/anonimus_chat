# syntax=docker/dockerfile:1

FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt pyproject.toml ./
COPY shared/ ./shared/
COPY ai/ ./ai/
RUN pip install --no-cache-dir -r requirements.txt
EXPOSE 8001
CMD ["uvicorn", "ai.main:app", "--host", "0.0.0.0", "--port", "8001"]
