# syntax=docker/dockerfile:1
FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Serve image does not need scikit-learn/pandas (training-only).
COPY requirements-serve.txt .

# Long timeout + retries: tensorflow wheel is ~500MB+ and often times out on slow links.
# BuildKit pip cache keeps partial downloads across rebuild attempts.
ENV PIP_DEFAULT_TIMEOUT=1000
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --upgrade pip && \
    pip install --retries 15 --timeout 1000 -r requirements-serve.txt

COPY . .

EXPOSE 8000

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
