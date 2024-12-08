FROM python:3.12.7-slim AS build

RUN apt-get update && apt-get install -y --no-install-recommends gcc

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .

RUN useradd -u 1001 nonroot && chown -R 1001:1001 /app

USER nonroot

RUN pip wheel --no-cache-dir --no-deps --wheel-dir /app/wheels -r requirements.txt -f https://download.pytorch.org/whl/torch_stable.html

# ----

FROM python:3.12.7-slim

WORKDIR /app

RUN useradd -u 1001 nonroot && mkdir /home/nonroot && chown -R 1001:1001 /home/nonroot

COPY --from=build /app/wheels /wheels
COPY --from=build /app/requirements.txt .
COPY src src

RUN pip install --no-cache /wheels/*

USER nonroot

ENTRYPOINT ["python3", "/app/src/main.py"]