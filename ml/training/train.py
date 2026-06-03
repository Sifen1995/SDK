"""
Train behavioral intent model on engineered features.
Saves artifact to models/intent_model.pkl
"""

from __future__ import annotations

import os
import sys

import joblib
import pandas as pd
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import classification_report
from sklearn.model_selection import train_test_split

# Allow running as script from ml/
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from training.dataset_builder import OUTPUT_FILE, build_synthetic_dataset, save_dataset
from training.feature_engineering import feature_column_order

RANDOM_STATE = 42
TEST_SIZE = 0.2
CONFIDENCE_THRESHOLD = 0.70
MODEL_DIR = os.path.join(os.path.dirname(__file__), "..", "models")
MODEL_PATH = os.path.join(MODEL_DIR, "intent_model.pkl")


def load_or_build_dataset() -> pd.DataFrame:
    if not os.path.exists(OUTPUT_FILE):
        print("Building synthetic behavioral dataset...")
        df = build_synthetic_dataset()
        save_dataset(df)
    else:
        df = pd.read_csv(OUTPUT_FILE)
    return df


def train() -> str:
    df = load_or_build_dataset()
    feature_cols = [c for c in feature_column_order() if c in df.columns]
    X = df[feature_cols].fillna(0.0)
    y = df["label"]

    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=TEST_SIZE, random_state=RANDOM_STATE, stratify=y
    )

    clf = RandomForestClassifier(
        n_estimators=300,
        max_depth=24,
        class_weight="balanced",
        random_state=RANDOM_STATE,
        n_jobs=-1,
    )
    clf.fit(X_train, y_train)

    y_pred = clf.predict(X_test)
    print(classification_report(y_test, y_pred))

    os.makedirs(MODEL_DIR, exist_ok=True)
    artifact = {
        "model": clf,
        "feature_columns": feature_cols,
        "threshold": CONFIDENCE_THRESHOLD,
        "intents": sorted(clf.classes_.tolist()),
        "model_version": "3.0-behavioral",
        "training_samples": len(X_train),
    }
    joblib.dump(artifact, MODEL_PATH)
    print(f"Model saved → {MODEL_PATH}")
    return MODEL_PATH


if __name__ == "__main__":
    train()
