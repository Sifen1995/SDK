# Flutter on-device intent prediction (TFLite)

Brief guide for embedding hosts: collect session signals → build a **71-float** vector → run `intent_model.tflite` → map the top class.

## Assets

Ship these from `skykin-sdk/lib/ml/assets/`:

| File | Purpose |
|------|---------|
| `intent_model.tflite` | Quantized classifier (input/output **float32**) |
| `label_map.json` | Index → intent name |
| `model_metadata.json` | `feature_size: 71`, `confidence_threshold: 0.7`, class list |

Add to `pubspec.yaml`:

```yaml
dependencies:
  tflite_flutter: ^0.11.0   # or current stable

flutter:
  assets:
    - packages/skykin_sdk/lib/ml/assets/intent_model.tflite
    - packages/skykin_sdk/lib/ml/assets/label_map.json
    - packages/skykin_sdk/lib/ml/assets/model_metadata.json
```

(Adjust asset paths to match how you package the SDK.)

---

## 1. What Dart must collect before prediction

Build one **session object** (same shape the Python pipeline expects). Do **not** pass raw event logs into TFLite — aggregate first.

### App usage (categories)

Track time and switches per category. Package names map via `ml/data/app_category_map.py` (unknown → `other`).

Fixed category order (12 slots):

`fashion`, `shopping`, `crypto`, `fintech`, `coffee`, `food`, `news`, `social`, `travel`, `fitness`, `banking`, `other`

```dart
/// Per category during the session
class CategoryUsage {
  double minutes;
  int switches;
}

Map<String, CategoryUsage> appUsage = {
  // only categories with activity; missing → 0
  'shopping': CategoryUsage(minutes: 12, switches: 5),
  'fashion': CategoryUsage(minutes: 6, switches: 3),
};
```

### UI text signals

Count keyword hits from accessibility / screen text using `ml/data/ui_keyword_map.py`.  
UI categories (8 slots): `fashion`, `crypto`, `coffee`, `fintech`, `travel`, `fitness`, `shopping`, `food`.

```dart
Map<String, int> uiSignals = {
  'shopping': 10,
  'fashion': 6,
};
```

### In-app behavioral events (funnel)

Host apps should emit **generic** actions (map ecommerce / Telebirr flows to these names):

| Action key | Ecommerce example | Super-app example |
|------------|-------------------|-------------------|
| `browseCategory` | Browse category | Open billers list |
| `viewItem` | View product | View bill detail |
| `stageTransaction` | Add to cart | Set amount |
| `initiateCheckout` | Checkout | PIN / confirm modal |
| `abandonTransaction` | Leave cart | Cancel PIN |

Category tags for behavioral counts (8):  
`coffee`, `fashion`, `crypto`, `fintech`, `travel`, `fitness`, `shopping`, `food`

```dart
class BehavioralEvents {
  /// 1.0 if the host emitted any in-app events this session; else 0.0
  final double hasData;
  final Map<String, int> actions;
  final Map<String, int> categories;

  BehavioralEvents({
    required this.hasData,
    required this.actions,
    required this.categories,
  });
}

// Example: accumulate counts as events arrive, then snapshot at predict time.
final behavioral = BehavioralEvents(
  hasData: 1.0,
  actions: {
    'browseCategory': 10,
    'viewItem': 12,
    'stageTransaction': 8,
    'initiateCheckout': 5,
    'abandonTransaction': 4,
  },
  categories: {
    'shopping': 20,
    'fashion': 10,
    'fintech': 4,
    // others default 0
  },
);
```

### Session + optional history

```dart
class SessionInput {
  final Map<String, CategoryUsage> appUsage;
  final Map<String, int> uiSignals;
  final BehavioralEvents? behavioralEvents;
  final DateTime sessionStart;
  final double sessionDurationMinutes;
  final int totalSwitches;
  final double isFirstSession; // 1.0 or 0.0

  /// Consented / logged-in only; omit or null for anonymous → zeros at [43–46]
  final HistoricalSignals? historical;
}
```

---

## 2. Build the 71-feature vector (Dart)

Mirror `ml/training/feature_engineering.py` exactly. Output: `Float32List(71)`.

| Index | Content |
|------:|---------|
| 0–11 | App time ratio per category (`minutes / totalMinutes`) |
| 12–23 | Switch ratio per category (`switches / totalSwitches`) |
| 24–31 | UI ratio per UI category (`count / totalUi`) |
| 32–37 | Temporal: `sin/cos(hour)`, `sin/cos(weekday)`, weekend flag, morning flag |
| 38–42 | Session length, switch density, dominance, diversity entropy, first-session |
| 43–46 | History (or zeros) |
| 47–70 | Behavioral funnel (only if `hasData > 0`; else leave 47–69 at 0, set `[70]`) |

**Important formulas (behavioral):**

```text
totalActions = browse + view + stage + checkout + abandon
denom = max(totalActions, 1)

[47]=browse/denom … [51]=abandon/denom
[52]=stage/max(browse,1)
[53]=checkout/max(stage,1)
[54]=abandon/max(stage,1)
[55]=min((stage + 2*checkout + 3*abandon) / denom, 1)
[56–63]=categoryEvent / sum(behavioralCategories)
[64]=shoppingTimeRatio * stageRatio      // features[1] * [49]
[65]=fintechTimeRatio * stageRatio       // features[3] * [49]
[66]=travelTimeRatio * stageRatio        // features[8] * [49]
[67]=shoppingUiRatio * abandonRatio      // features[30] * [51]
[68]=fintechUiRatio * abandonRatio       // features[27] * [51]
[69]=(stage + checkout) / denom
[70]=hasData                             // 1.0 or 0.0
```

Use safe denominators (`max(x, 1)` / `or 1.0`) the same way Python does so zeros don’t NaN.

---

## 3. Run TFLite

Input shape: `[1, 71]`, dtype **float32**.  
Output shape: `[1, 9]`, softmax scores (float32).

```dart
import 'dart:typed_data';
import 'package:tflite_flutter/tflite_flutter.dart';

Future<IntentResult> predict(Float32List features71) async {
  assert(features71.length == 71);

  final interpreter = await Interpreter.fromAsset(
    'packages/skykin_sdk/lib/ml/assets/intent_model.tflite',
  );

  final input = features71.reshape([1, 71]);
  final output = List.filled(1 * 9, 0.0).reshape([1, 9]);

  interpreter.run(input, output);
  interpreter.close();

  final scores = List<double>.from(output[0] as List);
  var bestIdx = 0;
  var bestScore = scores[0];
  for (var i = 1; i < scores.length; i++) {
    if (scores[i] > bestScore) {
      bestScore = scores[i];
      bestIdx = i;
    }
  }

  // label_map.json: "0" → "fashion_interest", …
  final intent = labelMap['$bestIdx']!;
  const threshold = 0.7; // model_metadata.json

  return IntentResult(
    intent: intent,
    confidence: bestScore,
    rewardTriggered: bestScore >= threshold && intent != 'no_clear_intent',
    topSignals: /* optional: top app/UI categories by weight */,
  );
}
```

### Output intents (9)

`fashion_interest`, `crypto_interest`, `coffee_interest`, `fintech_interest`, `travel_intent`, `fitness_interest`, `shopping_interest`, `food_interest`, `no_clear_intent`

---

## 4. Recommended predict timing

1. Accumulate usage / UI / behavioral counts during the session.  
2. On a cadence (e.g. every N minutes or after funnel milestones), snapshot → `buildFeatures` → `interpreter.run`.  
3. If `confidence >= 0.7` and intent ≠ `no_clear_intent`, treat as a reward / delivery signal.  
4. For anonymous users, leave historical features `[43–46]` at `0`.

---

## 5. Sanity checklist

- Feature length is **exactly 71** (`float32`).  
- Category / UI / behavioral key spellings match the tables above (camelCase for actions).  
- `has_data` is `1.0` only when the host actually sent in-app events.  
- Model input/output stay float32 (INT8 weights; float I/O).  
- After retraining (`cd ml && python main.py`), copy new assets from `skykin-sdk/lib/ml/assets/`.

Python reference implementation: `ml/training/feature_engineering.py` (`extract_features`). Keep Dart in lockstep with that file.
