from __future__ import annotations

import json
from pathlib import Path
from dataclasses import dataclass
from typing import Any

import numpy as np
import tensorflow as tf

ML_ROOT = Path(__file__).resolve().parent.parent
MODELS_DIR = ML_ROOT / "models"


@dataclass
class ModelArtifact:
    model_version: str
    intents: list[str]
    threshold: float
    feature_columns: list[str]
    model: Any


def load_model() -> ModelArtifact:
    label_map_path = MODELS_DIR / "label_map.json"
    model_path = MODELS_DIR / "best_model.keras"

    if not label_map_path.exists():
        raise FileNotFoundError(
            f"Missing label map at {label_map_path}. Run training first."
        )
    if not model_path.exists():
        raise FileNotFoundError(
            f"Missing model at {model_path}. Run training first."
        )

    with label_map_path.open() as f:
        label_map = json.load(f)

    model = tf.keras.models.load_model(model_path)
    feature_columns = [f"f_{i}" for i in range(47)]
    intents = [label_map[str(i)] for i in range(len(label_map))]

    return ModelArtifact(
        model_version="1.0.0",
        intents=intents,
        threshold=0.70,
        feature_columns=feature_columns,
        model=model,
    )
