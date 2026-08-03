"""Distill the trained sklearn scam detector into a TFLite float32 model.

RandomForest / LightGBM cannot be converted to TFLite directly, so we:
  1. Rebuild hybrid feature matrices from the processed datasets
  2. Train a compact Keras MLP that mimics the sklearn soft labels
  3. Export float32 (and optionally dynamic-range) ``.tflite`` files
  4. Write a manifest JSON describing the input vector for mobile clients

Run from the ``fraude_ml/`` directory:

    python -m src.export.convert_to_tflite
"""

from __future__ import annotations

import json
import os
from typing import Tuple

import joblib
import numpy as np
import pandas as pd
import tensorflow as tf
from scipy.sparse import hstack
from sklearn.metrics import accuracy_score, roc_auc_score

from src.features.build_features import FeatureExtractor
from src.models.predict import ScamDetector


def _repo_root() -> str:
    return os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))


def _load_processed(data_dir: str) -> Tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
    processed = os.path.join(data_dir, "processed")
    train = pd.read_csv(os.path.join(processed, "train_dataset.csv"))
    val = pd.read_csv(os.path.join(processed, "val_dataset.csv"))
    test = pd.read_csv(os.path.join(processed, "test_dataset.csv"))
    return train, val, test


def _build_matrix(
    detector: ScamDetector, df: pd.DataFrame
) -> Tuple[np.ndarray, np.ndarray]:
    X = detector.build_feature_matrix(df["text"].astype(str).tolist())
    y = df["label"].astype(int).values
    return X.toarray().astype(np.float32), y


def _build_keras_student(input_dim: int) -> tf.keras.Model:
    """Compact MLP designed for on-device TFLite conversion."""
    inputs = tf.keras.Input(shape=(input_dim,), name="features")
    x = tf.keras.layers.Dense(256, activation="relu", name="dense_1")(inputs)
    x = tf.keras.layers.Dropout(0.25, name="dropout_1")(x)
    x = tf.keras.layers.Dense(64, activation="relu", name="dense_2")(x)
    x = tf.keras.layers.Dropout(0.15, name="dropout_2")(x)
    outputs = tf.keras.layers.Dense(1, activation="sigmoid", name="spam_prob")(x)
    model = tf.keras.Model(inputs=inputs, outputs=outputs, name="scam_detector_tflite")
    model.compile(
        optimizer=tf.keras.optimizers.Adam(1e-3),
        loss="binary_crossentropy",
        metrics=["accuracy", tf.keras.metrics.AUC(name="auc")],
    )
    return model


def _write_tflite(model: tf.keras.Model, path: str, quantize: bool = False) -> int:
    converter = tf.lite.TFLiteConverter.from_keras_model(model)
    if quantize:
        converter.optimizations = [tf.lite.Optimize.DEFAULT]
    blob = converter.convert()
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "wb") as f:
        f.write(blob)
    return len(blob)


def _verify_tflite(path: str, X: np.ndarray, y: np.ndarray, n: int = 200) -> dict:
    interpreter = tf.lite.Interpreter(model_path=path)
    interpreter.allocate_tensors()
    inp = interpreter.get_input_details()[0]
    out = interpreter.get_output_details()[0]

    sample_x = X[:n]
    sample_y = y[:n]
    probs = []
    for row in sample_x:
        interpreter.set_tensor(inp["index"], row.reshape(1, -1).astype(np.float32))
        interpreter.invoke()
        probs.append(float(interpreter.get_tensor(out["index"])[0][0]))

    probs_arr = np.asarray(probs)
    preds = (probs_arr >= 0.5).astype(int)
    return {
        "samples": int(len(sample_y)),
        "accuracy": float(round(accuracy_score(sample_y, preds), 4)),
        "roc_auc": float(round(roc_auc_score(sample_y, probs_arr), 4)),
    }


def convert(
    artifact_dir: str = "models/artifacts",
    export_dir: str = "models/tflite",
    data_dir: str = "data",
    epochs: int = 12,
    batch_size: int = 256,
) -> dict:
    root = _repo_root()
    artifact_dir = artifact_dir if os.path.isabs(artifact_dir) else os.path.join(root, artifact_dir)
    export_dir = export_dir if os.path.isabs(export_dir) else os.path.join(root, export_dir)
    data_dir = data_dir if os.path.isabs(data_dir) else os.path.join(root, data_dir)

    print("📦 Loading sklearn artifacts...")
    detector = ScamDetector(artifact_dir=artifact_dir)

    print("📥 Loading processed datasets...")
    train_df, val_df, test_df = _load_processed(data_dir)

    print("⚙️ Building feature matrices (same hybrid pipeline as training)...")
    X_train, y_train = _build_matrix(detector, train_df)
    X_val, y_val = _build_matrix(detector, val_df)
    X_test, y_test = _build_matrix(detector, test_df)
    input_dim = X_train.shape[1]
    print(f"   input_dim={input_dim}")

    # Soft labels from the teacher, blended with hard labels for stable distillation.
    teacher_train = detector.model.predict_proba(X_train)[:, 1].astype(np.float32)
    teacher_val = detector.model.predict_proba(X_val)[:, 1].astype(np.float32)
    y_train_f = y_train.astype(np.float32)
    y_val_f = y_val.astype(np.float32)
    # 70% teacher soft score + 30% hard label
    soft_train = 0.7 * teacher_train + 0.3 * y_train_f
    soft_val = 0.7 * teacher_val + 0.3 * y_val_f

    print("🧠 Distilling teacher → Keras student MLP...")
    student = _build_keras_student(input_dim)
    callbacks = [
        tf.keras.callbacks.EarlyStopping(
            monitor="val_loss",
            mode="min",
            patience=3,
            restore_best_weights=True,
            verbose=1,
        )
    ]
    student.fit(
        X_train,
        soft_train,
        validation_data=(X_val, soft_val),
        epochs=epochs,
        batch_size=batch_size,
        callbacks=callbacks,
        verbose=1,
    )

    # Hard-label sanity check vs real labels
    val_prob = student.predict(X_val, verbose=0).ravel()
    test_prob = student.predict(X_test, verbose=0).ravel()
    keras_metrics = {
        "val_accuracy": float(round(accuracy_score(y_val, (val_prob >= 0.5).astype(int)), 4)),
        "val_roc_auc": float(round(roc_auc_score(y_val, val_prob), 4)),
        "test_accuracy": float(round(accuracy_score(y_test, (test_prob >= 0.5).astype(int)), 4)),
        "test_roc_auc": float(round(roc_auc_score(y_test, test_prob), 4)),
    }
    print(f"📊 Keras student metrics: {keras_metrics}")

    keras_path = os.path.join(export_dir, "scam_detector.keras")
    os.makedirs(export_dir, exist_ok=True)
    student.save(keras_path)

    float_path = os.path.join(export_dir, "scam_detector_float32.tflite")
    dyn_path = os.path.join(export_dir, "scam_detector_dynamic.tflite")
    float_bytes = _write_tflite(student, float_path, quantize=False)
    dyn_bytes = _write_tflite(student, dyn_path, quantize=True)
    print(f"✅ float32 TFLite → {float_path} ({float_bytes / 1024:.1f} KB)")
    print(f"✅ dynamic TFLite → {dyn_path} ({dyn_bytes / 1024:.1f} KB)")

    print("🔎 Verifying TFLite interpreter parity on test sample...")
    tflite_metrics = _verify_tflite(float_path, X_test, y_test, n=min(300, len(y_test)))
    print(f"   TFLite check: {tflite_metrics}")

    # Persist tabular feature schema for mobile feature builders
    extractor = FeatureExtractor()
    sample_tabular = extractor.extract_message_features("example http://example.com")
    tabular_names = detector.tabular_feature_names or list(sample_tabular.keys())

    manifest = {
        "model_version": "1.0.0",
        "model_type": "binary_spam_detector",
        "labels": {"0": "ham", "1": "spam"},
        "input": {
            "name": "features",
            "dtype": "float32",
            "shape": [1, input_dim],
            "description": (
                "Hybrid vector = TF-IDF (from saved vocabulary) + tabular "
                "structural/URL features in the order listed under tabular_features."
            ),
        },
        "output": {
            "name": "spam_prob",
            "dtype": "float32",
            "shape": [1, 1],
            "description": "Sigmoid spam probability in [0, 1].",
        },
        "threshold": ScamDetector.DEFAULT_THRESHOLD,
        "tfidf_vocab_size": int(len(detector.vectorizer.get_feature_names_out())),
        "tabular_features": tabular_names,
        "recommended_files": {
            "tflite_float32": "scam_detector_float32.tflite",
            "tflite_dynamic": "scam_detector_dynamic.tflite",
            "keras": "scam_detector.keras",
            "sklearn_teacher": "../artifacts/scam_detector.joblib",
            "tfidf_vectorizer": "../artifacts/tfidf_vectorizer.joblib",
            "feature_config": "../artifacts/feature_config.joblib",
        },
        "modes": {
            "message": "Pass SMS body as text; build hybrid features; score.",
            "url": "Pass URL string (optionally prefixed with http://); build hybrid features; score.",
        },
        "keras_metrics": keras_metrics,
        "tflite_verification": tflite_metrics,
        "model_size_kb": {
            "float32": round(float_bytes / 1024, 1),
            "dynamic": round(dyn_bytes / 1024, 1),
        },
    }

    manifest_path = os.path.join(export_dir, "model_manifest.json")
    with open(manifest_path, "w", encoding="utf-8") as f:
        json.dump(manifest, f, indent=2)
    print(f"📄 Manifest → {manifest_path}")

    # Also stash a copy of the TF-IDF vocabulary as plain JSON for mobile ports
    vocab = {term: int(idx) for term, idx in detector.vectorizer.vocabulary_.items()}
    vocab_path = os.path.join(export_dir, "tfidf_vocabulary.json")
    with open(vocab_path, "w", encoding="utf-8") as f:
        json.dump(
            {
                "vocabulary": vocab,
                "idf": detector.vectorizer.idf_.astype(float).tolist(),
                "ngram_range": list(detector.vectorizer.ngram_range),
                "sublinear_tf": bool(detector.vectorizer.sublinear_tf),
                "max_features": int(detector.vectorizer.max_features or len(vocab)),
            },
            f,
        )
    print(f"📄 TF-IDF vocabulary → {vocab_path}")

    # Save the distilled keras weights beside sklearn artifacts for convenience
    joblib.dump(
        {"input_dim": input_dim, "tabular_features": tabular_names},
        os.path.join(artifact_dir, "tflite_export_meta.joblib"),
    )
    return manifest


def main():
    # Keep TF log noise down for CLI runs
    os.environ.setdefault("TF_CPP_MIN_LOG_LEVEL", "2")
    convert()


if __name__ == "__main__":
    main()
