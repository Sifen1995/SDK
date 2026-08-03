# fraude_ml — SMS & URL Scam Detector

Hybrid spam/phishing detector for Ethiopian SMS traffic and in-app / browser URLs.

## Layout

```
fraude_ml/
├── data/raw|processed     # datasets
├── models/artifacts       # sklearn teacher + TF-IDF
├── models/tflite          # on-device export
├── src/data               # loaders, cleaner, synthetic generator
├── src/features           # TF-IDF helpers + URL/structural features
├── src/models             # train / evaluate / predict
├── src/export             # TFLite distillation
├── tests/                 # unit tests
└── FRONTEND_MODEL_GUIDE.md
```

## Quick start

```bash
cd fraude_ml
python -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt

# Sanitize downloaded threat feeds before committing or training. Malicious
# URLs can contain real third-party API credentials that trigger secret scans.
python -m src.data.sanitize_threat_feed data/raw/PhishTank_2026.csv

# optional: regenerate local synthetic SMS
python -m src.data.local_data_generator

python -m src.models.train
python -m src.export.convert_to_tflite
pytest tests/ -q
```

## Inference

```python
from src.models.predict import ScamDetector

d = ScamDetector()
print(d.predict_message("Urgent Telebirr… http://telebirr-kyc.xyz/login"))
print(d.predict_url("http://192.168.1.105/cbe-sec"))
```

Both message and URL scoring are supported. See
[`FRONTEND_MODEL_GUIDE.md`](FRONTEND_MODEL_GUIDE.md) for Flutter / TFLite integration.
