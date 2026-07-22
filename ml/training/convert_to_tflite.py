# skykin-ml/training/convert_to_tflite.py

import json
import os
import numpy as np
import pandas as pd
import tensorflow as tf

from feature_engineering import FEATURE_SIZE


def convert_to_tflite():
    print("Loading trained model...")
    model_path = "../models/best_model.keras"
    if not os.path.exists(model_path):
        raise FileNotFoundError(f"Model file not found at {model_path}")

    model = tf.keras.models.load_model(model_path)

    # ── Load Label Map ─────────────────────────────────────────
    label_map_src = "../models/label_map.json"
    if not os.path.exists(label_map_src):
        raise FileNotFoundError(f"Label map not found at {label_map_src}")

    with open(label_map_src, "r") as f:
        label_map = json.load(f)

    # ── Conversion 1: Float32 (baseline, higher accuracy) ──────
    converter = tf.lite.TFLiteConverter.from_keras_model(model)
    tflite_float = converter.convert()

    float_path = "../models/intent_model_float32.tflite"
    os.makedirs(os.path.dirname(float_path), exist_ok=True)
    with open(float_path, "wb") as f:
        f.write(tflite_float)
    float_size = os.path.getsize(float_path) / 1024
    print(f"Float32 model: {float_size:.1f} KB")

    # ── Conversion 2: INT8 Quantization (smaller, faster) ──────
    data_path = "../data/processed/training_data.csv"
    if not os.path.exists(data_path):
        raise FileNotFoundError(f"Training data for calibration not found at {data_path}")

    df = pd.read_csv(data_path)
    feature_cols = [f"f_{i}" for i in range(FEATURE_SIZE)]

    # Robust validation check for features
    missing_cols = [c for c in feature_cols if c not in df.columns]
    if missing_cols:
        raise KeyError(f"Missing feature columns in dataset: {missing_cols}")

    calibration_data = df[feature_cols].values[:500].astype(np.float32)

    def representative_dataset():
        for sample in calibration_data:
            yield [sample.reshape(1, -1)]

    converter_int8 = tf.lite.TFLiteConverter.from_keras_model(model)
    converter_int8.optimizations = [tf.lite.Optimize.DEFAULT]
    converter_int8.representative_dataset = representative_dataset
    converter_int8.target_spec.supported_ops = [
        tf.lite.OpsSet.TFLITE_BUILTINS_INT8
    ]
    converter_int8.inference_input_type = tf.float32
    converter_int8.inference_output_type = tf.float32

    tflite_int8 = converter_int8.convert()

    int8_path = "../models/intent_model_int8.tflite"
    with open(int8_path, "wb") as f:
        f.write(tflite_int8)
    int8_size = os.path.getsize(int8_path) / 1024
    print(f"INT8 model:    {int8_size:.1f} KB")

    # ── Verify accuracy after quantization ─────────────────────
    print("\nVerifying quantized model accuracy...")

    interpreter = tf.lite.Interpreter(model_path=int8_path)
    interpreter.allocate_tensors()

    input_details = interpreter.get_input_details()
    output_details = interpreter.get_output_details()

    test_samples = df[feature_cols].values[:1000].astype(np.float32)
    test_labels = df["label_idx"].values[:1000]

    correct = 0
    for i, sample in enumerate(test_samples):
        interpreter.set_tensor(
            input_details[0]["index"],
            sample.reshape(1, -1)
        )
        interpreter.invoke()
        output = interpreter.get_tensor(output_details[0]["index"])
        if np.argmax(output[0]) == test_labels[i]:
            correct += 1

    accuracy = correct / len(test_samples)
    print(f"Quantized accuracy: {accuracy:.4f}")

    is_quantized = accuracy >= 0.80
    if not is_quantized:
        print("WARNING: Quantized accuracy below 80% threshold")
        print("Using float32 model instead")
        final_model = tflite_float
    else:
        print("Quantized model passes accuracy threshold")
        final_model = tflite_int8

    # ── Copy to Flutter SDK assets ─────────────────────────────
    sdk_dir = "../../skykin-sdk/lib/ml/assets"
    os.makedirs(sdk_dir, exist_ok=True)

    sdk_asset_path = os.path.join(sdk_dir, "intent_model.tflite")
    with open(sdk_asset_path, "wb") as f:
        f.write(final_model)

    final_size = os.path.getsize(sdk_asset_path) / 1024
    print(f"\nFinal model copied to SDK: {final_size:.1f} KB")

    # ── Copy label map to SDK ──────────────────────────────────
    sdk_label_path = os.path.join(sdk_dir, "label_map.json")
    with open(sdk_label_path, "w") as f:
        json.dump(label_map, f, indent=2)

    # ── Save model metadata ────────────────────────────────────
    metadata = {
        "model_version": "2.0.0",
        "feature_size": FEATURE_SIZE,
        "num_classes": len(label_map),
        "intent_classes": list(label_map.values()),
        "quantized": is_quantized,
        "accuracy": round(accuracy, 4),
        "model_size_kb": round(final_size, 1),
        "confidence_threshold": 0.70,
    }

    metadata_path = os.path.join(sdk_dir, "model_metadata.json")
    with open(metadata_path, "w") as f:
        json.dump(metadata, f, indent=2)

    print(f"Metadata saved → {metadata_path}")
    print("\nConversion complete.")
    return metadata


if __name__ == "__main__":
    convert_to_tflite()