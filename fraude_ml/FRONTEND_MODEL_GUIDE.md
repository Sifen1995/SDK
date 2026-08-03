# Frontend Guide: SMS + URL Scam Detector

This guide explains how Flutter / mobile engineers use the `fraude_ml` scam
detector for:

1. **SMS / notification inspection** — is this message phishing?
2. **Browser / in-app link inspection** — is this URL risky?

The Python teacher model (RandomForest / LightGBM) and the on-device **TFLite**
student share the same hybrid feature idea:

```
hybrid_vector = TF-IDF(text)  +  structural/URL heuristics
output        = P(spam) ∈ [0, 1]
```

---

## What the model can (and cannot) do

| Use case | Supported? | How |
|---|---|---|
| Spam / phishing SMS body | Yes | `predict_message(text)` or TFLite with hybrid features |
| Suspicious URL while browsing | Yes | `predict_url(url)` — URL is scored as text so TF-IDF + URL heuristics both fire |
| Image / QR-only phishing with no text/URL | No | Needs OCR / other pipeline first |
| Guaranteed bank-grade blocking | No | Treat score as a risk signal; keep a human / allowlist policy |

URL-only browsing works because `FeatureExtractor.extract_url_features()`
already encodes high-risk TLDs, IP hosts, shorteners, entropy, hyphens, and
subdomain depth. Feeding the URL string through TF-IDF also catches brand /
urgency tokens inside the path.

---

## Artifacts you care about

After training + export (from `fraude_ml/`):

```bash
python -m src.models.train
python -m src.export.convert_to_tflite
```

| Path | Purpose |
|---|---|
| `models/artifacts/scam_detector.joblib` | Server-side teacher model |
| `models/artifacts/tfidf_vectorizer.joblib` | TF-IDF vocabulary (Python) |
| `models/artifacts/feature_config.joblib` | Tabular feature order |
| `models/tflite/scam_detector_float32.tflite` | **Preferred on-device model** |
| `models/tflite/scam_detector_dynamic.tflite` | Smaller dynamic-range quantized model |
| `models/tflite/model_manifest.json` | Input/output contract for mobile |
| `models/tflite/tfidf_vocabulary.json` | Vocab + IDF for on-device TF-IDF |

Copy into your Flutter asset bundle, for example:

```
assets/ml/scam_detector_float32.tflite
assets/ml/model_manifest.json
assets/ml/tfidf_vocabulary.json
```

---

## Option A — Call Python / backend (simplest)

Use this from a backend service or during SDK prototyping.

```python
from src.models.predict import ScamDetector

detector = ScamDetector(artifact_dir="models/artifacts")

# Incoming SMS
sms = detector.predict_message(
    "Urgent Telebirr Alert: verify now http://telebirr-kyc.xyz/login"
)
# → { "is_spam": True, "spam_probability": 0.97, "mode": "message", ... }

# Browser navigation / tapped link
url = detector.predict_url("http://192.168.1.105/cbe-sec-login")
# → { "is_spam": True, "spam_probability": 0.91, "mode": "url",
#     "url_features": { "ip_in_url": 1, ... }, ... }

if sms["spam_probability"] >= 0.5:
    # warn / quarantine / require confirm
    ...
```

Response fields:

- `is_spam` — boolean at the configured threshold (default `0.5`)
- `spam_probability` — model confidence in \([0, 1]\)
- `mode` — `"message"` or `"url"`
- `url_features` — transparent heuristics useful for UI badges

---

## Option B — On-device TFLite (Flutter)

### 1. Load the model

```dart
import 'package:tflite_flutter/tflite_flutter.dart';

final interpreter = await Interpreter.fromAsset(
  'ml/scam_detector_float32.tflite',
);
```

Read `model_manifest.json` for the expected input length:

```json
"input": { "shape": [1, 3016], "dtype": "float32" }
"output": { "shape": [1, 1], "dtype": "float32" }
"threshold": 0.5
```

Your length may differ slightly after re-training — **always trust the
manifest**, not a hard-coded constant.

### 2. Build the input vector

The TFLite model does **not** tokenize text for you. Build the same hybrid
vector the Python trainer used:

1. **Clean text** — lowercase, NFC normalize, keep URLs (see `TextCleaner`).
2. **TF-IDF slice** — length = `tfidf_vocab_size` from the manifest.
   - Load `tfidf_vocabulary.json`
   - Tokenize into word unigrams + bigrams (`ngram_range: [1, 2]`)
   - Apply `sublinear_tf` (`1 + log(tf)`) × IDF for each vocab term
   - Missing terms = `0`
3. **Tabular slice** — append features in exact `tabular_features` order:

| Feature | SMS message | Bare URL browse |
|---|---|---|
| `text_length`, `word_count`, `digit_count`, `uppercase_count` | from raw body | from URL string |
| `has_brand_mention`, `has_urgency_word`, `has_amharic` | keyword / script checks | usually 0 unless path has brands |
| `has_url`, `url_count`, `url_length` | from embedded links | `1`, `1`, `len(url)` |
| `domain_entropy` | Shannon entropy of registrable domain | same |
| `is_high_risk_tld` | `xyz/top/sbs/cfd/...` | same |
| `is_url_shortener` | bit.ly / tinyurl / t.co / cutt.ly | same |
| `ip_in_url` | host is IPv4 | same |
| `subdomain_count` | labels beyond apex | same |
| `has_hyphen_in_domain` | hyphen in apex label | same |

Reference implementations (Python — port the logic 1:1):

- Cleaning → `src/data/cleaner.py`
- Tabular / URL heuristics → `src/features/build_features.py`
- End-to-end scoring → `src/models/predict.py`

### 3. Run inference

```dart
final input = [featureVector]; // shape [1, inputSize], float32
final output = List.generate(1, (_) => List.filled(1, 0.0));

interpreter.run(input, output);

final spamProbability = output[0][0] as double;
final isSpam = spamProbability >= 0.5;
```

### 4. UX recommendations

- **SMS shield:** if `spamProbability >= 0.7` → hard warn; `0.5–0.7` → soft badge.
- **Browser shield:** block / interstitial when `ip_in_url == 1` **or**
  (`is_high_risk_tld == 1` and `spamProbability >= 0.5`).
- Always show the destination host to the user before they continue.
- Do not silently drop messages — false positives happen.

---

## Quick decision tree for the app

```
incoming event
 ├─ SMS / push body ──────────────► predict_message / TFLite(hybrid)
 └─ URL navigation / link tap ────► predict_url / TFLite(hybrid)
         │
         ▼
   spam_probability
         ├─ < 0.5  → allow
         ├─ 0.5–0.7 → warn
         └─ ≥ 0.7  → strong warn / quarantine
```

---

## Retraining / refreshing the on-device model

```bash
cd fraude_ml
source .venv/bin/activate   # or your env
python -m src.models.train
python -m src.export.convert_to_tflite
pytest tests/ -q
```

Ship the new `.tflite` + `model_manifest.json` + `tfidf_vocabulary.json`
together. Never mix a vocabulary file from one train run with a TFLite file
from another.

---

## Accuracy notes (honest)

- The teacher sklearn model is the source of truth for **server** scoring.
- The TFLite file is a **distilled student** (small MLP). Expect a small drop
  vs the teacher; check `model_manifest.json → keras_metrics / tflite_verification`
  after each export.
- Synthetic Ethiopian SMS templates improve local phishing recall but public
  English spam corpora still dominate the training mix — continue collecting
  real local labeled samples.
